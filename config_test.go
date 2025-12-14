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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithWarpClient(t *testing.T) {
	t.Run("sets warp client", func(t *testing.T) {
		mock := &mockWarpClient{}
		opts := &clientOptions{}

		err := WithWarpClient(mock)(opts)
		require.NoError(t, err)
		assert.Equal(t, mock, opts.warpClient)
	})

	t.Run("error when client is nil", func(t *testing.T) {
		opts := &clientOptions{}
		err := WithWarpClient(nil)(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "warp client cannot be nil")
	})
}

func TestWithTransformer(t *testing.T) {
	t.Run("adds transformer to list", func(t *testing.T) {
		transformer := func(ctx context.Context, tc *TransformContext) error {
			return nil
		}
		opts := defaultOptions()

		err := WithTransformer(transformer)(opts)
		require.NoError(t, err)
		assert.Len(t, opts.transformers, 1)
	})

	t.Run("error when transformer is nil", func(t *testing.T) {
		opts := defaultOptions()
		err := WithTransformer(nil)(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transformer cannot be nil")
	})
}

func TestWithTransformers(t *testing.T) {
	t.Run("adds multiple transformers", func(t *testing.T) {
		t1 := func(ctx context.Context, tc *TransformContext) error { return nil }
		t2 := func(ctx context.Context, tc *TransformContext) error { return nil }
		t3 := func(ctx context.Context, tc *TransformContext) error { return nil }

		opts := defaultOptions()
		err := WithTransformers(t1, t2, t3)(opts)
		require.NoError(t, err)
		assert.Len(t, opts.transformers, 3)
	})

	t.Run("error when any transformer is nil", func(t *testing.T) {
		t1 := func(ctx context.Context, tc *TransformContext) error { return nil }
		opts := defaultOptions()

		err := WithTransformers(t1, nil)(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transformer")
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("error with empty transformers", func(t *testing.T) {
		opts := defaultOptions()
		err := WithTransformers()(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestWithPipeline(t *testing.T) {
	t.Run("sets custom pipeline", func(t *testing.T) {
		transformer := func(ctx context.Context, tc *TransformContext) error {
			return nil
		}
		pipeline := NewPipeline(transformer)

		opts := defaultOptions()
		err := WithPipeline(pipeline)(opts)
		require.NoError(t, err)
		assert.Equal(t, pipeline, opts.pipeline)
		assert.Nil(t, opts.transformers)
	})

	t.Run("error when pipeline is nil", func(t *testing.T) {
		opts := defaultOptions()
		err := WithPipeline(nil)(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pipeline cannot be nil")
	})
}

func TestWithErrorStrategy(t *testing.T) {
	t.Run("sets fail closed strategy", func(t *testing.T) {
		opts := defaultOptions()
		err := WithErrorStrategy(ErrorStrategyFailClosed)(opts)
		require.NoError(t, err)
		assert.Equal(t, ErrorStrategyFailClosed, opts.errorStrategy)
	})

	t.Run("sets fail open strategy", func(t *testing.T) {
		opts := defaultOptions()
		err := WithErrorStrategy(ErrorStrategyFailOpen)(opts)
		require.NoError(t, err)
		assert.Equal(t, ErrorStrategyFailOpen, opts.errorStrategy)
	})

	t.Run("error with invalid strategy", func(t *testing.T) {
		opts := defaultOptions()
		// Use an invalid integer value
		err := WithErrorStrategy(ErrorStrategy(999))(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid error strategy")
	})
}

func TestWithErrorHandler(t *testing.T) {
	t.Run("sets error handler", func(t *testing.T) {
		handler := func(ctx context.Context, err error) {}
		opts := defaultOptions()

		err := WithErrorHandler(handler)(opts)
		require.NoError(t, err)
		assert.NotNil(t, opts.errorHandler)
	})

	t.Run("error when handler is nil", func(t *testing.T) {
		opts := defaultOptions()
		err := WithErrorHandler(nil)(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error handler cannot be nil")
	})
}

func TestDefaultErrorHandler(t *testing.T) {
	// Default error handler should not panic
	testErr := errors.New("test error")

	// Should not panic
	assert.NotPanics(t, func() {
		defaultErrorHandler(context.Background(), testErr)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	assert.Equal(t, ErrorStrategyFailOpen, opts.errorStrategy)
	assert.NotNil(t, opts.errorHandler)
	assert.Nil(t, opts.pipeline)
	assert.Nil(t, opts.transformers)
}
