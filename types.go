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

// Package textile provides transparent message transformation middleware for LLM clients.
//
// Textile wraps the warp client with a powerful transformation pipeline system while
// maintaining zero API changes for consumers. This allows you to apply configurable
// transformations to messages before and after LLM calls transparently.
//
// Basic usage:
//
//	// Create warp client
//	warpClient, _ := warp.NewClient(
//	    warp.WithAPIKey("openai", apiKey),
//	)
//
//	// Wrap with textile
//	client, _ := textile.New(
//	    textile.WithWarpClient(warpClient),
//	    textile.WithTransformer(myTransformer),
//	)
//
//	// Use exactly like warp - zero API changes
//	resp, err := client.Completion(ctx, &warp.CompletionRequest{
//	    Model: "openai/gpt-4",
//	    Messages: []warp.Message{{Role: "user", Content: "Hello"}},
//	})
package textile

import (
	"context"
	"sync"

	"github.com/blue-context/warp"
)

// Transformer transforms messages before LLM calls and responses after.
// Transformers receive a TransformContext that can be modified in-place.
//
// Context is provided for cancellation and timeout support.
// Transformers should respect context cancellation and return early if ctx.Done() is signaled.
type Transformer func(ctx context.Context, tc *TransformContext) error

// TransformContext contains the data being transformed.
// Transformers may modify fields in-place during their execution.
// All data in TransformContext is deep-cloned before transformation,
// making it safe to modify without affecting the original request/response.
type TransformContext struct {
	// Stage indicates when transformation is occurring
	Stage TransformStage

	// Messages being sent to LLM (for Request stage)
	// This is a deep clone, safe to modify
	Messages []warp.Message

	// Request being sent to LLM (for full request access)
	// This is a deep clone, safe to modify
	Request *warp.CompletionRequest

	// Response from LLM (for Response stage)
	// This is a deep clone, safe to modify
	Response *warp.CompletionResponse

	// StreamChunk for streaming responses (for StreamChunk stage)
	// This is a deep clone, safe to modify
	StreamChunk *warp.CompletionChunk

	// Metadata for transformer communication
	// Thread-safe for reads, modifications require external synchronization
	Metadata map[string]interface{}
}

// TransformStage indicates when in the pipeline we're transforming.
type TransformStage int

const (
	// StageRequest indicates transformation before sending to LLM
	StageRequest TransformStage = iota
	// StageResponse indicates transformation after receiving from LLM
	StageResponse
	// StageStreamChunk indicates transformation during streaming response
	StageStreamChunk
)

// String returns a string representation of the transform stage.
func (s TransformStage) String() string {
	switch s {
	case StageRequest:
		return "Request"
	case StageResponse:
		return "Response"
	case StageStreamChunk:
		return "StreamChunk"
	default:
		return "Unknown"
	}
}

// Client wraps warp.Client with transformation capabilities.
// It is safe for concurrent use.
//
// Client implements the warp.Client interface, allowing it to be used
// as a drop-in replacement for warp.Client with zero API changes.
type Client struct {
	inner    warp.Client
	pipeline *Pipeline
	config   *clientConfig
	mu       sync.RWMutex // Protects pipeline modifications
}

// Ensure Client implements warp.Client interface at compile time.
var _ warp.Client = (*Client)(nil)

// Pipeline manages an ordered list of transformers.
// Transformers are executed sequentially in the order they were added.
type Pipeline struct {
	transformers []Transformer
	mu           sync.RWMutex // Protects transformer list
}

// TransformableStream wraps warp.Stream with transformation.
// Each chunk received from the stream is transformed before being returned.
type TransformableStream struct {
	inner    warp.Stream
	pipeline *Pipeline
	ctx      context.Context
	config   *clientConfig
}

// Ensure TransformableStream implements warp.Stream interface at compile time.
var _ warp.Stream = (*TransformableStream)(nil)

// clientConfig holds internal configuration for the textile client.
type clientConfig struct {
	errorStrategy ErrorStrategy
	errorHandler  func(context.Context, error)
}

// ErrorStrategy defines how to handle transformer errors.
type ErrorStrategy int

const (
	// ErrorStrategyFailOpen logs errors but continues (default)
	// This ensures system availability even if transformers fail
	ErrorStrategyFailOpen ErrorStrategy = iota

	// ErrorStrategyFailClosed returns errors immediately
	// This ensures transformations are never silently skipped
	ErrorStrategyFailClosed
)

// String returns a string representation of the error strategy.
func (s ErrorStrategy) String() string {
	switch s {
	case ErrorStrategyFailOpen:
		return "FailOpen"
	case ErrorStrategyFailClosed:
		return "FailClosed"
	default:
		return "Unknown"
	}
}
