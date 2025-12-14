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
)

// NewPipeline creates a new pipeline with the given transformers.
// Transformers are executed in the order they appear.
func NewPipeline(transformers ...Transformer) *Pipeline {
	return &Pipeline{
		transformers: transformers,
	}
}

// Transform applies all transformers to the context sequentially.
// Handles errors according to the configured error strategy.
func (p *Pipeline) Transform(ctx context.Context, tc *TransformContext, config *clientConfig) error {
	p.mu.RLock()
	transformers := p.transformers
	p.mu.RUnlock()

	if len(transformers) == 0 {
		return nil
	}

	var errs []error

	for i, transformer := range transformers {
		// Check context cancellation before each transformer
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := transformer(ctx, tc); err != nil {
			// Wrap error with context
			transformErr := &TransformError{
				Stage:       tc.Stage,
				Transformer: fmt.Sprintf("transformer[%d]", i),
				Err:         err,
			}

			// Handle based on strategy
			if config.errorStrategy == ErrorStrategyFailClosed {
				return transformErr
			}

			// ErrorStrategyFailOpen: log and continue
			if config.errorHandler != nil {
				config.errorHandler(ctx, transformErr)
			}
			errs = append(errs, transformErr)
		}
	}

	// Return collected errors if any (only in fail-open mode)
	if len(errs) > 0 {
		return &PipelineError{Errors: errs}
	}

	return nil
}

// Append creates a new pipeline with additional transformers at the end.
// The original pipeline is not modified (immutable operation).
func (p *Pipeline) Append(transformers ...Transformer) *Pipeline {
	p.mu.RLock()
	defer p.mu.RUnlock()

	newTransformers := make([]Transformer, len(p.transformers)+len(transformers))
	copy(newTransformers, p.transformers)
	copy(newTransformers[len(p.transformers):], transformers)

	return &Pipeline{
		transformers: newTransformers,
	}
}

// Prepend creates a new pipeline with transformers at the beginning.
// The original pipeline is not modified (immutable operation).
func (p *Pipeline) Prepend(transformers ...Transformer) *Pipeline {
	p.mu.RLock()
	defer p.mu.RUnlock()

	newTransformers := make([]Transformer, len(transformers)+len(p.transformers))
	copy(newTransformers, transformers)
	copy(newTransformers[len(transformers):], p.transformers)

	return &Pipeline{
		transformers: newTransformers,
	}
}

// Len returns the number of transformers in the pipeline.
func (p *Pipeline) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.transformers)
}

// PipelineBuilder provides a fluent API for building pipelines.
type PipelineBuilder struct {
	transformers []Transformer
}

// NewPipelineBuilder creates a new pipeline builder.
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		transformers: make([]Transformer, 0),
	}
}

// Add appends a transformer to the pipeline.
func (b *PipelineBuilder) Add(transformer Transformer) *PipelineBuilder {
	b.transformers = append(b.transformers, transformer)
	return b
}

// AddFunc adds a simple function as a transformer.
// The function receives only the TransformContext and doesn't need to handle context.Context.
func (b *PipelineBuilder) AddFunc(fn func(*TransformContext) error) *PipelineBuilder {
	return b.Add(func(ctx context.Context, tc *TransformContext) error {
		return fn(tc)
	})
}

// Build creates the final pipeline.
func (b *PipelineBuilder) Build() *Pipeline {
	return &Pipeline{
		transformers: b.transformers,
	}
}
