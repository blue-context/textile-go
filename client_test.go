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

// mockWarpClient is a mock implementation of warp.Client for testing.
type mockWarpClient struct {
	completionFunc       func(ctx context.Context, req *warp.CompletionRequest) (*warp.CompletionResponse, error)
	completionStreamFunc func(ctx context.Context, req *warp.CompletionRequest) (warp.Stream, error)
}

func (m *mockWarpClient) Completion(ctx context.Context, req *warp.CompletionRequest) (*warp.CompletionResponse, error) {
	if m.completionFunc != nil {
		return m.completionFunc(ctx, req)
	}
	return &warp.CompletionResponse{
		ID:    "test-id",
		Model: req.Model,
		Choices: []warp.Choice{
			{
				Index:   0,
				Message: warp.Message{Role: "assistant", Content: "mock response"},
			},
		},
	}, nil
}

func (m *mockWarpClient) CompletionStream(ctx context.Context, req *warp.CompletionRequest) (warp.Stream, error) {
	if m.completionStreamFunc != nil {
		return m.completionStreamFunc(ctx, req)
	}
	return &mockStream{}, nil
}

func (m *mockWarpClient) Embedding(ctx context.Context, req *warp.EmbeddingRequest) (*warp.EmbeddingResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) ImageGeneration(ctx context.Context, req *warp.ImageGenerationRequest) (*warp.ImageGenerationResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) ImageEdit(ctx context.Context, req *warp.ImageEditRequest) (*warp.ImageGenerationResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) ImageVariation(ctx context.Context, req *warp.ImageVariationRequest) (*warp.ImageGenerationResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) Transcription(ctx context.Context, req *warp.TranscriptionRequest) (*warp.TranscriptionResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) Speech(ctx context.Context, req *warp.SpeechRequest) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) Moderation(ctx context.Context, req *warp.ModerationRequest) (*warp.ModerationResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) Rerank(ctx context.Context, req *warp.RerankRequest) (*warp.RerankResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWarpClient) CompletionCost(resp *warp.CompletionResponse) (float64, error) {
	return 0, nil
}

func (m *mockWarpClient) Close() error {
	return nil
}

func (m *mockWarpClient) RegisterProvider(p warp.Provider) error {
	return nil
}

// mockStream is a mock implementation of warp.Stream for testing.
type mockStream struct {
	chunks []*warp.CompletionChunk
	idx    int
}

func (m *mockStream) Recv() (*warp.CompletionChunk, error) {
	if m.idx >= len(m.chunks) {
		return nil, io.EOF
	}
	chunk := m.chunks[m.idx]
	m.idx++
	return chunk, nil
}

func (m *mockStream) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	t.Run("successful creation with warp client", func(t *testing.T) {
		mock := &mockWarpClient{}
		client, err := New(WithWarpClient(mock))

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, mock, client.inner)
	})

	t.Run("error when warp client is missing", func(t *testing.T) {
		_, err := New()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "warp client is required")
	})

	t.Run("with transformer", func(t *testing.T) {
		mock := &mockWarpClient{}
		transformer := func(ctx context.Context, tc *TransformContext) error {
			return nil
		}

		client, err := New(
			WithWarpClient(mock),
			WithTransformer(transformer),
		)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, 1, client.pipeline.Len())
	})

	t.Run("with multiple transformers", func(t *testing.T) {
		mock := &mockWarpClient{}
		t1 := func(ctx context.Context, tc *TransformContext) error { return nil }
		t2 := func(ctx context.Context, tc *TransformContext) error { return nil }

		client, err := New(
			WithWarpClient(mock),
			WithTransformers(t1, t2),
		)

		require.NoError(t, err)
		assert.Equal(t, 2, client.pipeline.Len())
	})

	t.Run("with error strategy", func(t *testing.T) {
		mock := &mockWarpClient{}
		client, err := New(
			WithWarpClient(mock),
			WithErrorStrategy(ErrorStrategyFailClosed),
		)

		require.NoError(t, err)
		assert.Equal(t, ErrorStrategyFailClosed, client.config.errorStrategy)
	})
}

func TestClient_Completion(t *testing.T) {
	t.Run("basic completion without transformers", func(t *testing.T) {
		mock := &mockWarpClient{}
		client, err := New(WithWarpClient(mock))
		require.NoError(t, err)

		req := &warp.CompletionRequest{
			Model:    "test/model",
			Messages: []warp.Message{{Role: "user", Content: "hello"}},
		}

		resp, err := client.Completion(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test/model", resp.Model)
	})

	t.Run("completion with request transformer", func(t *testing.T) {
		mock := &mockWarpClient{}
		var capturedMessages []warp.Message

		mock.completionFunc = func(ctx context.Context, req *warp.CompletionRequest) (*warp.CompletionResponse, error) {
			capturedMessages = req.Messages
			return &warp.CompletionResponse{
				ID:    "test-id",
				Model: req.Model,
				Choices: []warp.Choice{
					{Message: warp.Message{Role: "assistant", Content: "response"}},
				},
			}, nil
		}

		transformer := func(ctx context.Context, tc *TransformContext) error {
			if tc.Stage == StageRequest {
				for i := range tc.Messages {
					if content, ok := tc.Messages[i].Content.(string); ok {
						tc.Messages[i].Content = "TRANSFORMED: " + content
					}
				}
			}
			return nil
		}

		client, err := New(
			WithWarpClient(mock),
			WithTransformer(transformer),
		)
		require.NoError(t, err)

		req := &warp.CompletionRequest{
			Model:    "test/model",
			Messages: []warp.Message{{Role: "user", Content: "hello"}},
		}

		resp, err := client.Completion(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify transformation was applied
		require.Len(t, capturedMessages, 1)
		assert.Equal(t, "TRANSFORMED: hello", capturedMessages[0].Content)

		// Verify original request was not modified
		assert.Equal(t, "hello", req.Messages[0].Content)
	})

	t.Run("completion with response transformer", func(t *testing.T) {
		mock := &mockWarpClient{}
		transformer := func(ctx context.Context, tc *TransformContext) error {
			if tc.Stage == StageResponse {
				for i := range tc.Response.Choices {
					if content, ok := tc.Response.Choices[i].Message.Content.(string); ok {
						tc.Response.Choices[i].Message.Content = "TRANSFORMED: " + content
					}
				}
			}
			return nil
		}

		client, err := New(
			WithWarpClient(mock),
			WithTransformer(transformer),
		)
		require.NoError(t, err)

		req := &warp.CompletionRequest{
			Model:    "test/model",
			Messages: []warp.Message{{Role: "user", Content: "hello"}},
		}

		resp, err := client.Completion(context.Background(), req)
		require.NoError(t, err)

		// Verify transformation was applied to response
		content := resp.Choices[0].Message.Content.(string)
		assert.Contains(t, content, "TRANSFORMED:")
	})

	t.Run("transformer error with fail closed", func(t *testing.T) {
		mock := &mockWarpClient{}
		expectedErr := errors.New("transformer error")

		transformer := func(ctx context.Context, tc *TransformContext) error {
			return expectedErr
		}

		client, err := New(
			WithWarpClient(mock),
			WithTransformer(transformer),
			WithErrorStrategy(ErrorStrategyFailClosed),
		)
		require.NoError(t, err)

		req := &warp.CompletionRequest{
			Model:    "test/model",
			Messages: []warp.Message{{Role: "user", Content: "hello"}},
		}

		_, err = client.Completion(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transformer error")
	})

	t.Run("transformer error with fail open", func(t *testing.T) {
		mock := &mockWarpClient{}
		expectedErr := errors.New("transformer error")

		transformer := func(ctx context.Context, tc *TransformContext) error {
			return expectedErr
		}

		var errorLogged bool
		errorHandler := func(ctx context.Context, err error) {
			errorLogged = true
		}

		client, err := New(
			WithWarpClient(mock),
			WithTransformer(transformer),
			WithErrorStrategy(ErrorStrategyFailOpen),
			WithErrorHandler(errorHandler),
		)
		require.NoError(t, err)

		req := &warp.CompletionRequest{
			Model:    "test/model",
			Messages: []warp.Message{{Role: "user", Content: "hello"}},
		}

		// Should succeed despite transformer error
		resp, err := client.Completion(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, errorLogged)
	})
}

func TestClient_CompletionStream(t *testing.T) {
	t.Run("streaming without transformers", func(t *testing.T) {
		mock := &mockWarpClient{
			completionStreamFunc: func(ctx context.Context, req *warp.CompletionRequest) (warp.Stream, error) {
				return &mockStream{
					chunks: []*warp.CompletionChunk{
						{ID: "chunk-1", Model: req.Model},
						{ID: "chunk-2", Model: req.Model},
					},
				}, nil
			},
		}

		client, err := New(WithWarpClient(mock))
		require.NoError(t, err)

		req := &warp.CompletionRequest{
			Model:    "test/model",
			Messages: []warp.Message{{Role: "user", Content: "hello"}},
		}

		stream, err := client.CompletionStream(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, stream)

		// Read chunks
		chunk1, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, "chunk-1", chunk1.ID)

		chunk2, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, "chunk-2", chunk2.ID)

		_, err = stream.Recv()
		assert.Equal(t, io.EOF, err)
	})

	t.Run("streaming with chunk transformer", func(t *testing.T) {
		mock := &mockWarpClient{
			completionStreamFunc: func(ctx context.Context, req *warp.CompletionRequest) (warp.Stream, error) {
				return &mockStream{
					chunks: []*warp.CompletionChunk{
						{ID: "original", Model: req.Model},
					},
				}, nil
			},
		}

		transformer := func(ctx context.Context, tc *TransformContext) error {
			if tc.Stage == StageStreamChunk {
				tc.StreamChunk.ID = "transformed-" + tc.StreamChunk.ID
			}
			return nil
		}

		client, err := New(
			WithWarpClient(mock),
			WithTransformer(transformer),
		)
		require.NoError(t, err)

		req := &warp.CompletionRequest{
			Model:    "test/model",
			Messages: []warp.Message{{Role: "user", Content: "hello"}},
		}

		stream, err := client.CompletionStream(context.Background(), req)
		require.NoError(t, err)

		chunk, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, "transformed-original", chunk.ID)
	})
}

func TestClient_WithTransformers(t *testing.T) {
	mock := &mockWarpClient{}
	t1 := func(ctx context.Context, tc *TransformContext) error { return nil }

	client, err := New(
		WithWarpClient(mock),
		WithTransformer(t1),
	)
	require.NoError(t, err)
	assert.Equal(t, 1, client.pipeline.Len())

	// Add more transformers at runtime
	t2 := func(ctx context.Context, tc *TransformContext) error { return nil }
	newClient := client.WithTransformers(t2)

	// Original client unchanged
	assert.Equal(t, 1, client.pipeline.Len())

	// New client has additional transformer
	assert.Equal(t, 2, newClient.pipeline.Len())

	// Both clients share the same underlying warp client
	assert.Equal(t, client.inner, newClient.inner)
}

func TestClient_ConcurrentCompletion(t *testing.T) {
	mock := &mockWarpClient{}
	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	// Run concurrent completions
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := &warp.CompletionRequest{
				Model:    "test/model",
				Messages: []warp.Message{{Role: "user", Content: "hello"}},
			}

			_, err := client.Completion(context.Background(), req)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func BenchmarkClient_Completion(b *testing.B) {
	mock := &mockWarpClient{}
	client, _ := New(WithWarpClient(mock))

	req := &warp.CompletionRequest{
		Model:    "test/model",
		Messages: []warp.Message{{Role: "user", Content: "hello"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Completion(context.Background(), req)
	}
}

func BenchmarkClient_CompletionWithTransformer(b *testing.B) {
	mock := &mockWarpClient{}
	transformer := func(ctx context.Context, tc *TransformContext) error {
		if tc.Stage == StageRequest {
			for i := range tc.Messages {
				if content, ok := tc.Messages[i].Content.(string); ok {
					tc.Messages[i].Content = "PREFIX: " + content
				}
			}
		}
		return nil
	}

	client, _ := New(
		WithWarpClient(mock),
		WithTransformer(transformer),
	)

	req := &warp.CompletionRequest{
		Model:    "test/model",
		Messages: []warp.Message{{Role: "user", Content: "hello"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Completion(context.Background(), req)
	}
}

// Test all passthrough methods
func TestClient_Embedding(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.EmbeddingRequest{
		Model: "test/embedding",
		Input: "test input",
	}

	_, err = client.Embedding(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_ImageGeneration(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.ImageGenerationRequest{
		Prompt: "test prompt",
	}

	_, err = client.ImageGeneration(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_ImageEdit(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.ImageEditRequest{
		Prompt: "test prompt",
	}

	_, err = client.ImageEdit(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_ImageVariation(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.ImageVariationRequest{}

	_, err = client.ImageVariation(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_Transcription(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.TranscriptionRequest{
		Model: "test/whisper",
	}

	_, err = client.Transcription(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_Speech(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.SpeechRequest{
		Model: "test/tts",
		Input: "test input",
	}

	_, err = client.Speech(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_Moderation(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.ModerationRequest{
		Input: "test input",
	}

	_, err = client.Moderation(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_Rerank(t *testing.T) {
	mock := &mockWarpClient{}

	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	req := &warp.RerankRequest{
		Query:     "test query",
		Documents: []string{"doc1", "doc2"},
	}

	_, err = client.Rerank(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestClient_CompletionCost(t *testing.T) {
	mock := &mockWarpClient{}
	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	resp := &warp.CompletionResponse{
		ID:    "test-id",
		Model: "test/model",
	}

	cost, err := client.CompletionCost(resp)
	require.NoError(t, err)
	assert.Equal(t, 0.0, cost)
}

func TestClient_Close(t *testing.T) {
	mock := &mockWarpClient{}
	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestClient_RegisterProvider(t *testing.T) {
	mock := &mockWarpClient{}
	client, err := New(WithWarpClient(mock))
	require.NoError(t, err)

	// Pass nil as we're just testing the passthrough
	err = client.RegisterProvider(nil)
	assert.NoError(t, err)
}
