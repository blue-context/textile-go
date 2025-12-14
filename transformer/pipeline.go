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

// Package transformer provides the transformation pipeline system for textile.
package transformer

import (
	"context"
	"fmt"
	"sync"
)

// Transformer transforms messages before LLM calls and responses after.
// This is re-exported from the parent package for convenience.
type Transformer func(ctx context.Context, tc interface{}) error

// Pipeline manages an ordered list of transformers.
// Transformers are executed sequentially in the order they were added.
//
// Pipeline is safe for concurrent use.
type Pipeline struct {
	transformers []Transformer
	mu           sync.RWMutex
}

// NewPipeline creates a new pipeline with the given transformers.
// Transformers are executed in the order they appear.
func NewPipeline(transformers ...Transformer) *Pipeline {
	return &Pipeline{
		transformers: transformers,
	}
}

// Transform applies all transformers to the context sequentially.
// If a transformer returns an error:
// - With ErrorStrategyFailClosed: returns immediately
// - With ErrorStrategyFailOpen: logs and continues (errors collected and returned as PipelineError)
//
// The errorStrategy and errorHandler should be provided via the execution context.
func (p *Pipeline) Transform(ctx context.Context, tc interface{}, errorStrategy int, errorHandler func(context.Context, error)) error {
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
			// Wrap error with transformer index for debugging
			transformErr := &TransformError{
				Transformer: fmt.Sprintf("transformer[%d]", i),
				Err:         err,
			}

			// Handle based on strategy
			if errorStrategy == 1 { // ErrorStrategyFailClosed
				return transformErr
			}

			// ErrorStrategyFailOpen: log and continue
			if errorHandler != nil {
				errorHandler(ctx, transformErr)
			}
			errs = append(errs, transformErr)
		}
	}

	// Return collected errors if any
	if len(errs) > 0 {
		return &PipelineError{Errors: errs}
	}

	return nil
}

// Append creates a new pipeline with additional transformers at the end.
// The original pipeline is not modified.
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
// The original pipeline is not modified.
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

// TransformError represents an error during transformation.
type TransformError struct {
	Transformer string
	Err         error
}

// Error returns a formatted error message.
func (e *TransformError) Error() string {
	return fmt.Sprintf("transform error in %s: %v", e.Transformer, e.Err)
}

// Unwrap returns the underlying error.
func (e *TransformError) Unwrap() error {
	return e.Err
}

// PipelineError represents multiple errors from pipeline execution.
type PipelineError struct {
	Errors []error
}

// Error returns a formatted error message.
func (e *PipelineError) Error() string {
	if len(e.Errors) == 0 {
		return "pipeline error: no errors"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("pipeline error: %v", e.Errors[0])
	}
	return fmt.Sprintf("pipeline encountered %d errors", len(e.Errors))
}

// Unwrap returns the list of errors.
func (e *PipelineError) Unwrap() []error {
	return e.Errors
}
