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
	"fmt"
	"io"

	"github.com/blue-context/textile-go/internal/clone"
	"github.com/blue-context/warp"
)

// Completion creates a chat completion with transformation pipeline applied.
// This method wraps the underlying warp.Client.Completion() with transparent transformation.
//
// Transformations are applied in two phases:
// 1. Request phase: transformers modify the request before sending to LLM
// 2. Response phase: transformers modify the response after receiving from LLM
func (c *Client) Completion(ctx context.Context, req *warp.CompletionRequest) (*warp.CompletionResponse, error) {
	// Deep clone request to ensure immutability
	clonedReq, err := clone.Deep(req)
	if err != nil {
		return nil, fmt.Errorf("failed to clone request: %w", err)
	}

	// Get current pipeline (thread-safe read)
	c.mu.RLock()
	pipeline := c.pipeline
	config := c.config
	c.mu.RUnlock()

	// Apply request transformations
	tc := &TransformContext{
		Stage:    StageRequest,
		Messages: clonedReq.Messages,
		Request:  clonedReq,
		Metadata: make(map[string]interface{}),
	}

	if err := pipeline.Transform(ctx, tc, config); err != nil {
		if config.errorStrategy == ErrorStrategyFailClosed {
			return nil, err
		}
		// Error already logged by pipeline, continue with original request
	}

	// Update request with transformed messages
	clonedReq.Messages = tc.Messages

	// Call underlying warp client
	resp, err := c.inner.Completion(ctx, clonedReq)
	if err != nil {
		return nil, err
	}

	// Deep clone response for transformation
	clonedResp, err := clone.Deep(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to clone response: %w", err)
	}

	// Apply response transformations
	tc = &TransformContext{
		Stage:    StageResponse,
		Response: clonedResp,
		Metadata: make(map[string]interface{}),
	}

	if err := pipeline.Transform(ctx, tc, config); err != nil {
		if config.errorStrategy == ErrorStrategyFailClosed {
			return nil, err
		}
		// Error already logged by pipeline, continue with original response
	}

	return tc.Response, nil
}

// CompletionStream creates a streaming chat completion with transformation pipeline applied.
// Returns a wrapped stream that applies transformations to each chunk.
func (c *Client) CompletionStream(ctx context.Context, req *warp.CompletionRequest) (warp.Stream, error) {
	// Deep clone request to ensure immutability
	clonedReq, err := clone.Deep(req)
	if err != nil {
		return nil, fmt.Errorf("failed to clone request: %w", err)
	}

	// Get current pipeline (thread-safe read)
	c.mu.RLock()
	pipeline := c.pipeline
	config := c.config
	c.mu.RUnlock()

	// Apply request transformations
	tc := &TransformContext{
		Stage:    StageRequest,
		Messages: clonedReq.Messages,
		Request:  clonedReq,
		Metadata: make(map[string]interface{}),
	}

	if err := pipeline.Transform(ctx, tc, config); err != nil {
		if config.errorStrategy == ErrorStrategyFailClosed {
			return nil, err
		}
		// Error already logged by pipeline, continue with original request
	}

	// Update request with transformed messages
	clonedReq.Messages = tc.Messages

	// Call underlying warp client
	stream, err := c.inner.CompletionStream(ctx, clonedReq)
	if err != nil {
		return nil, err
	}

	// Wrap stream with transformation
	return &TransformableStream{
		inner:    stream,
		pipeline: pipeline,
		ctx:      ctx,
		config:   config,
	}, nil
}

// Embedding creates embeddings.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) Embedding(ctx context.Context, req *warp.EmbeddingRequest) (*warp.EmbeddingResponse, error) {
	return c.inner.Embedding(ctx, req)
}

// ImageGeneration generates images from text prompts.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) ImageGeneration(ctx context.Context, req *warp.ImageGenerationRequest) (*warp.ImageGenerationResponse, error) {
	return c.inner.ImageGeneration(ctx, req)
}

// ImageEdit edits an image using AI based on a text prompt.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) ImageEdit(ctx context.Context, req *warp.ImageEditRequest) (*warp.ImageGenerationResponse, error) {
	return c.inner.ImageEdit(ctx, req)
}

// ImageVariation creates variations of an existing image.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) ImageVariation(ctx context.Context, req *warp.ImageVariationRequest) (*warp.ImageGenerationResponse, error) {
	return c.inner.ImageVariation(ctx, req)
}

// Transcription transcribes audio to text.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) Transcription(ctx context.Context, req *warp.TranscriptionRequest) (*warp.TranscriptionResponse, error) {
	return c.inner.Transcription(ctx, req)
}

// Speech converts text to speech.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) Speech(ctx context.Context, req *warp.SpeechRequest) (io.ReadCloser, error) {
	return c.inner.Speech(ctx, req)
}

// Moderation checks content for policy violations.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) Moderation(ctx context.Context, req *warp.ModerationRequest) (*warp.ModerationResponse, error) {
	return c.inner.Moderation(ctx, req)
}

// Rerank ranks documents by relevance to a query.
// This method forwards directly to the underlying warp client without transformation.
func (c *Client) Rerank(ctx context.Context, req *warp.RerankRequest) (*warp.RerankResponse, error) {
	return c.inner.Rerank(ctx, req)
}

// CompletionCost calculates the cost of a completion.
// This method forwards directly to the underlying warp client.
func (c *Client) CompletionCost(resp *warp.CompletionResponse) (float64, error) {
	return c.inner.CompletionCost(resp)
}

// Close closes the client and releases resources.
// This closes the underlying warp client.
func (c *Client) Close() error {
	return c.inner.Close()
}

// RegisterProvider registers a provider with the underlying warp client.
// This method forwards directly to the underlying warp client.
func (c *Client) RegisterProvider(p warp.Provider) error {
	return c.inner.RegisterProvider(p)
}

// WithTransformers creates a new client with additional transformers.
// This is useful for adding transformers at runtime without modifying the original client.
//
// The new client shares the underlying warp client but has an independent pipeline.
func (c *Client) WithTransformers(transformers ...Transformer) *Client {
	c.mu.RLock()
	newPipeline := c.pipeline.Append(transformers...)
	config := c.config
	inner := c.inner
	c.mu.RUnlock()

	return &Client{
		inner:    inner,
		pipeline: newPipeline,
		config:   config,
	}
}
