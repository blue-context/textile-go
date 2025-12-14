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
	"io"
	"testing"

	"github.com/blue-context/warp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errorStream is a mock stream that returns an error on Recv
type errorStream struct {
	err error
}

func (e *errorStream) Recv() (*warp.CompletionChunk, error) {
	return nil, e.err
}

func (e *errorStream) Close() error {
	return nil
}

func TestTransformableStream_Recv(t *testing.T) {
	t.Run("successful recv without transformers", func(t *testing.T) {
		mockInner := &mockStream{
			chunks: []*warp.CompletionChunk{
				{ID: "chunk-1", Model: "test/model"},
			},
		}

		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(),
			ctx:      context.Background(),
			config:   &clientConfig{errorStrategy: ErrorStrategyFailClosed},
		}

		chunk, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, "chunk-1", chunk.ID)
	})

	t.Run("EOF passthrough", func(t *testing.T) {
		mockInner := &mockStream{
			chunks: []*warp.CompletionChunk{},
		}

		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(),
			ctx:      context.Background(),
			config:   &clientConfig{errorStrategy: ErrorStrategyFailClosed},
		}

		_, err := stream.Recv()
		assert.Equal(t, io.EOF, err)
	})

	t.Run("error from inner stream", func(t *testing.T) {
		expectedErr := errors.New("stream error")
		mockInner := &errorStream{err: expectedErr}

		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(),
			ctx:      context.Background(),
			config:   &clientConfig{errorStrategy: ErrorStrategyFailClosed},
		}

		_, err := stream.Recv()
		assert.Equal(t, expectedErr, err)
	})

	t.Run("transformer error with fail closed", func(t *testing.T) {
		mockInner := &mockStream{
			chunks: []*warp.CompletionChunk{
				{ID: "chunk-1", Model: "test/model"},
			},
		}

		transformErr := errors.New("transform failed")
		transformer := func(ctx context.Context, tc *TransformContext) error {
			return transformErr
		}

		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(transformer),
			ctx:      context.Background(),
			config:   &clientConfig{errorStrategy: ErrorStrategyFailClosed},
		}

		_, err := stream.Recv()
		assert.Error(t, err)
	})

	t.Run("transformer error with fail open", func(t *testing.T) {
		mockInner := &mockStream{
			chunks: []*warp.CompletionChunk{
				{ID: "chunk-1", Model: "test/model"},
			},
		}

		transformErr := errors.New("transform failed")
		transformer := func(ctx context.Context, tc *TransformContext) error {
			return transformErr
		}

		var errorHandled bool
		errorHandler := func(ctx context.Context, err error) {
			errorHandled = true
		}

		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(transformer),
			ctx:      context.Background(),
			config: &clientConfig{
				errorStrategy: ErrorStrategyFailOpen,
				errorHandler:  errorHandler,
			},
		}

		chunk, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, "chunk-1", chunk.ID)
		assert.True(t, errorHandled)
	})

	t.Run("transformer modifies chunk", func(t *testing.T) {
		mockInner := &mockStream{
			chunks: []*warp.CompletionChunk{
				{ID: "original", Model: "test/model"},
			},
		}

		transformer := func(ctx context.Context, tc *TransformContext) error {
			if tc.Stage == StageStreamChunk {
				tc.StreamChunk.ID = "modified-" + tc.StreamChunk.ID
			}
			return nil
		}

		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(transformer),
			ctx:      context.Background(),
			config:   &clientConfig{errorStrategy: ErrorStrategyFailClosed},
		}

		chunk, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, "modified-original", chunk.ID)
	})
}

func TestTransformableStream_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		mockInner := &mockStream{}
		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(),
			ctx:      context.Background(),
			config:   &clientConfig{errorStrategy: ErrorStrategyFailClosed},
		}

		err := stream.Close()
		assert.NoError(t, err)
	})

	t.Run("close error propagation", func(t *testing.T) {
		expectedErr := errors.New("close error")
		mockInner := &errorClosingStream{err: expectedErr}

		stream := &TransformableStream{
			inner:    mockInner,
			pipeline: NewPipeline(),
			ctx:      context.Background(),
			config:   &clientConfig{errorStrategy: ErrorStrategyFailClosed},
		}

		err := stream.Close()
		assert.Equal(t, expectedErr, err)
	})
}

// errorClosingStream is a mock stream that returns an error on Close
type errorClosingStream struct {
	err error
}

func (e *errorClosingStream) Recv() (*warp.CompletionChunk, error) {
	return nil, io.EOF
}

func (e *errorClosingStream) Close() error {
	return e.err
}
