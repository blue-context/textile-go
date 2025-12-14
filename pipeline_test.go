// Copyright 2025 Blue Context Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package textile

import (
	"context"
	"errors"
	"testing"

	"github.com/blue-context/warp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeline_Transform(t *testing.T) {
	t.Run("empty pipeline", func(t *testing.T) {
		pipeline := NewPipeline()
		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailOpen,
			errorHandler:  func(ctx context.Context, err error) {},
		}

		err := pipeline.Transform(context.Background(), tc, config)
		assert.NoError(t, err)
	})

	t.Run("single transformer success", func(t *testing.T) {
		var called bool
		transformer := func(ctx context.Context, tc *TransformContext) error {
			called = true
			return nil
		}

		pipeline := NewPipeline(transformer)
		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailOpen,
			errorHandler:  func(ctx context.Context, err error) {},
		}

		err := pipeline.Transform(context.Background(), tc, config)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("multiple transformers execute in order", func(t *testing.T) {
		var order []int

		t1 := func(ctx context.Context, tc *TransformContext) error {
			order = append(order, 1)
			return nil
		}
		t2 := func(ctx context.Context, tc *TransformContext) error {
			order = append(order, 2)
			return nil
		}
		t3 := func(ctx context.Context, tc *TransformContext) error {
			order = append(order, 3)
			return nil
		}

		pipeline := NewPipeline(t1, t2, t3)
		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailOpen,
			errorHandler:  func(ctx context.Context, err error) {},
		}

		err := pipeline.Transform(context.Background(), tc, config)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, order)
	})

	t.Run("transformer error with fail closed", func(t *testing.T) {
		expectedErr := errors.New("transformer failed")
		var t2Called bool

		t1 := func(ctx context.Context, tc *TransformContext) error {
			return expectedErr
		}
		t2 := func(ctx context.Context, tc *TransformContext) error {
			t2Called = true
			return nil
		}

		pipeline := NewPipeline(t1, t2)
		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailClosed,
			errorHandler:  func(ctx context.Context, err error) {},
		}

		err := pipeline.Transform(context.Background(), tc, config)
		assert.Error(t, err)
		assert.False(t, t2Called, "second transformer should not be called")

		var transformErr *TransformError
		assert.True(t, errors.As(err, &transformErr))
	})

	t.Run("transformer error with fail open", func(t *testing.T) {
		expectedErr := errors.New("transformer failed")
		var errorHandled bool
		var t2Called bool

		t1 := func(ctx context.Context, tc *TransformContext) error {
			return expectedErr
		}
		t2 := func(ctx context.Context, tc *TransformContext) error {
			t2Called = true
			return nil
		}

		pipeline := NewPipeline(t1, t2)
		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailOpen,
			errorHandler: func(ctx context.Context, err error) {
				errorHandled = true
			},
		}

		err := pipeline.Transform(context.Background(), tc, config)
		assert.Error(t, err) // PipelineError is returned in fail-open mode
		assert.True(t, errorHandled, "error handler should be called")
		assert.True(t, t2Called, "second transformer should be called")

		var pipelineErr *PipelineError
		assert.True(t, errors.As(err, &pipelineErr))
		assert.Len(t, pipelineErr.Errors, 1)
	})

	t.Run("context cancellation", func(t *testing.T) {
		var t2Called bool

		t1 := func(ctx context.Context, tc *TransformContext) error {
			return nil
		}
		t2 := func(ctx context.Context, tc *TransformContext) error {
			t2Called = true
			return nil
		}

		pipeline := NewPipeline(t1, t2)
		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailOpen,
			errorHandler:  func(ctx context.Context, err error) {},
		}

		// Cancel context before execution
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := pipeline.Transform(ctx, tc, config)
		assert.Equal(t, context.Canceled, err)
		assert.False(t, t2Called)
	})
}

func TestPipeline_Append(t *testing.T) {
	t1 := func(ctx context.Context, tc *TransformContext) error { return nil }
	t2 := func(ctx context.Context, tc *TransformContext) error { return nil }
	t3 := func(ctx context.Context, tc *TransformContext) error { return nil }

	pipeline := NewPipeline(t1)
	assert.Equal(t, 1, pipeline.Len())

	// Append creates new pipeline
	newPipeline := pipeline.Append(t2, t3)
	assert.Equal(t, 1, pipeline.Len(), "original pipeline should be unchanged")
	assert.Equal(t, 3, newPipeline.Len())
}

func TestPipeline_Prepend(t *testing.T) {
	t1 := func(ctx context.Context, tc *TransformContext) error { return nil }
	t2 := func(ctx context.Context, tc *TransformContext) error { return nil }
	t3 := func(ctx context.Context, tc *TransformContext) error { return nil }

	pipeline := NewPipeline(t1)
	assert.Equal(t, 1, pipeline.Len())

	// Prepend creates new pipeline
	newPipeline := pipeline.Prepend(t2, t3)
	assert.Equal(t, 1, pipeline.Len(), "original pipeline should be unchanged")
	assert.Equal(t, 3, newPipeline.Len())

	// Verify order: prepended transformers should execute first
	var order []int
	t2Ordered := func(ctx context.Context, tc *TransformContext) error {
		order = append(order, 2)
		return nil
	}
	t3Ordered := func(ctx context.Context, tc *TransformContext) error {
		order = append(order, 3)
		return nil
	}
	t1Ordered := func(ctx context.Context, tc *TransformContext) error {
		order = append(order, 1)
		return nil
	}

	pipeline = NewPipeline(t1Ordered)
	newPipeline = pipeline.Prepend(t2Ordered, t3Ordered)

	tc := &TransformContext{
		Stage:    StageRequest,
		Messages: []warp.Message{{Role: "user", Content: "test"}},
		Metadata: make(map[string]interface{}),
	}
	config := &clientConfig{
		errorStrategy: ErrorStrategyFailOpen,
		errorHandler:  func(ctx context.Context, err error) {},
	}

	err := newPipeline.Transform(context.Background(), tc, config)
	require.NoError(t, err)
	assert.Equal(t, []int{2, 3, 1}, order)
}

func TestPipelineBuilder(t *testing.T) {
	t.Run("empty builder", func(t *testing.T) {
		builder := NewPipelineBuilder()
		pipeline := builder.Build()
		assert.Equal(t, 0, pipeline.Len())
	})

	t.Run("add transformers", func(t *testing.T) {
		t1 := func(ctx context.Context, tc *TransformContext) error { return nil }
		t2 := func(ctx context.Context, tc *TransformContext) error { return nil }

		builder := NewPipelineBuilder()
		builder.Add(t1).Add(t2)
		pipeline := builder.Build()

		assert.Equal(t, 2, pipeline.Len())
	})

	t.Run("add func", func(t *testing.T) {
		var called bool
		fn := func(tc *TransformContext) error {
			called = true
			return nil
		}

		builder := NewPipelineBuilder()
		builder.AddFunc(fn)
		pipeline := builder.Build()

		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailOpen,
			errorHandler:  func(ctx context.Context, err error) {},
		}

		err := pipeline.Transform(context.Background(), tc, config)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("fluent API", func(t *testing.T) {
		var order []int

		pipeline := NewPipelineBuilder().
			AddFunc(func(tc *TransformContext) error {
				order = append(order, 1)
				return nil
			}).
			AddFunc(func(tc *TransformContext) error {
				order = append(order, 2)
				return nil
			}).
			AddFunc(func(tc *TransformContext) error {
				order = append(order, 3)
				return nil
			}).
			Build()

		tc := &TransformContext{
			Stage:    StageRequest,
			Messages: []warp.Message{{Role: "user", Content: "test"}},
			Metadata: make(map[string]interface{}),
		}
		config := &clientConfig{
			errorStrategy: ErrorStrategyFailOpen,
			errorHandler:  func(ctx context.Context, err error) {},
		}

		err := pipeline.Transform(context.Background(), tc, config)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, order)
	})
}

func TestPipeline_ConcurrentAccess(t *testing.T) {
	transformer := func(ctx context.Context, tc *TransformContext) error {
		return nil
	}

	pipeline := NewPipeline(transformer)

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			tc := &TransformContext{
				Stage:    StageRequest,
				Messages: []warp.Message{{Role: "user", Content: "test"}},
				Metadata: make(map[string]interface{}),
			}
			config := &clientConfig{
				errorStrategy: ErrorStrategyFailOpen,
				errorHandler:  func(ctx context.Context, err error) {},
			}

			_ = pipeline.Transform(context.Background(), tc, config)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent Append/Prepend (creates new pipelines, doesn't modify original)
	for i := 0; i < 10; i++ {
		go func() {
			_ = pipeline.Append(transformer)
			_ = pipeline.Prepend(transformer)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
