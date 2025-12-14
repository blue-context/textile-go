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
	"fmt"
	"strings"
)

// TransformError represents an error during transformation.
// It wraps the underlying error with context about which stage and transformer failed.
type TransformError struct {
	Stage       TransformStage
	Transformer string // Name or description of transformer
	Err         error
}

// Error returns a formatted error message.
func (e *TransformError) Error() string {
	transformerDesc := "unknown transformer"
	if e.Transformer != "" {
		transformerDesc = e.Transformer
	}
	return fmt.Sprintf("transform error at stage %v in %s: %v", e.Stage, transformerDesc, e.Err)
}

// Unwrap returns the underlying error.
func (e *TransformError) Unwrap() error {
	return e.Err
}

// PipelineError represents multiple errors from pipeline execution.
// Used when ErrorStrategyFailOpen is configured and multiple transformers fail.
type PipelineError struct {
	Errors []error
}

// Error returns a formatted error message listing all errors.
func (e *PipelineError) Error() string {
	if len(e.Errors) == 0 {
		return "pipeline error: no errors"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("pipeline error: %v", e.Errors[0])
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("pipeline encountered %d errors:\n", len(e.Errors)))
	for i, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %v\n", i+1, err))
	}
	return sb.String()
}

// Unwrap returns the list of errors for errors.Is/As support.
func (e *PipelineError) Unwrap() []error {
	return e.Errors
}
