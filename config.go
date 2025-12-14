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
	"log"

	"github.com/blue-context/warp"
)

// Option configures the textile client.
// Options are applied during client creation via New().
type Option func(*clientOptions) error

// clientOptions holds configuration options during client construction.
type clientOptions struct {
	warpClient    warp.Client
	transformers  []Transformer
	pipeline      *Pipeline
	errorStrategy ErrorStrategy
	errorHandler  func(context.Context, error)
}

// defaultOptions returns the default client options.
func defaultOptions() *clientOptions {
	return &clientOptions{
		errorStrategy: ErrorStrategyFailOpen,
		errorHandler:  defaultErrorHandler,
	}
}

// defaultErrorHandler logs errors to the standard logger.
func defaultErrorHandler(ctx context.Context, err error) {
	log.Printf("[textile] transformer error: %v", err)
}

// WithWarpClient sets the underlying warp client to wrap.
// This option is required - New() will fail if no warp client is provided.
func WithWarpClient(client warp.Client) Option {
	return func(opts *clientOptions) error {
		if client == nil {
			return fmt.Errorf("warp client cannot be nil")
		}
		opts.warpClient = client
		return nil
	}
}

// WithTransformer adds a single transformer to the pipeline.
// Transformers are executed in the order they are added.
//
// Multiple WithTransformer options can be provided to build a pipeline.
func WithTransformer(t Transformer) Option {
	return func(opts *clientOptions) error {
		if t == nil {
			return fmt.Errorf("transformer cannot be nil")
		}
		opts.transformers = append(opts.transformers, t)
		return nil
	}
}

// WithTransformers adds multiple transformers to the pipeline.
// Transformers are executed in the order they appear in the slice.
func WithTransformers(ts ...Transformer) Option {
	return func(opts *clientOptions) error {
		if len(ts) == 0 {
			return fmt.Errorf("transformers slice cannot be empty")
		}
		for i, t := range ts {
			if t == nil {
				return fmt.Errorf("transformer at index %d cannot be nil", i)
			}
		}
		opts.transformers = append(opts.transformers, ts...)
		return nil
	}
}

// WithPipeline sets the transformation pipeline directly.
// This overrides any transformers added via WithTransformer/WithTransformers.
//
// If both WithPipeline and WithTransformer options are provided,
// the last option wins.
func WithPipeline(p *Pipeline) Option {
	return func(opts *clientOptions) error {
		if p == nil {
			return fmt.Errorf("pipeline cannot be nil")
		}
		opts.pipeline = p
		opts.transformers = nil // Clear individual transformers
		return nil
	}
}

// WithErrorStrategy sets the error handling strategy for transformers.
//
// ErrorStrategyFailOpen (default): Logs errors but continues pipeline execution.
// ErrorStrategyFailClosed: Returns errors immediately, stopping execution.
func WithErrorStrategy(strategy ErrorStrategy) Option {
	return func(opts *clientOptions) error {
		if strategy != ErrorStrategyFailOpen && strategy != ErrorStrategyFailClosed {
			return fmt.Errorf("invalid error strategy: %v", strategy)
		}
		opts.errorStrategy = strategy
		return nil
	}
}

// WithErrorHandler sets a custom error handler for transformer errors.
// The handler is called when transformers fail and ErrorStrategyFailOpen is used.
//
// The default handler logs errors using the standard log package.
func WithErrorHandler(handler func(context.Context, error)) Option {
	return func(opts *clientOptions) error {
		if handler == nil {
			return fmt.Errorf("error handler cannot be nil")
		}
		opts.errorHandler = handler
		return nil
	}
}

// New creates a new textile client with the provided options.
//
// The WithWarpClient option is required. Other options are optional.
//
// Example:
//
//	client, err := textile.New(
//	    textile.WithWarpClient(warpClient),
//	    textile.WithTransformer(myTransformer),
//	    textile.WithErrorStrategy(textile.ErrorStrategyFailOpen),
//	)
func New(opts ...Option) (*Client, error) {
	options := defaultOptions()

	// Apply all options
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate required options
	if options.warpClient == nil {
		return nil, fmt.Errorf("warp client is required (use WithWarpClient option)")
	}

	// Build pipeline from transformers if not explicitly provided
	var pipeline *Pipeline
	if options.pipeline != nil {
		pipeline = options.pipeline
	} else if len(options.transformers) > 0 {
		pipeline = NewPipeline(options.transformers...)
	} else {
		// Empty pipeline (no-op)
		pipeline = NewPipeline()
	}

	// Create client config
	config := &clientConfig{
		errorStrategy: options.errorStrategy,
		errorHandler:  options.errorHandler,
	}

	return &Client{
		inner:    options.warpClient,
		pipeline: pipeline,
		config:   config,
	}, nil
}
