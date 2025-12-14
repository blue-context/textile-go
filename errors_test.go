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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformError(t *testing.T) {
	t.Run("error message includes transformer and stage", func(t *testing.T) {
		baseErr := errors.New("base error")
		transformErr := &TransformError{
			Transformer: "test-transformer",
			Stage:       StageRequest,
			Err:         baseErr,
		}

		errMsg := transformErr.Error()
		assert.Contains(t, errMsg, "test-transformer")
		assert.Contains(t, errMsg, "Request")
		assert.Contains(t, errMsg, "base error")
	})

	t.Run("unwrap returns original error", func(t *testing.T) {
		baseErr := errors.New("base error")
		transformErr := &TransformError{
			Transformer: "test-transformer",
			Stage:       StageResponse,
			Err:         baseErr,
		}

		unwrapped := transformErr.Unwrap()
		assert.Equal(t, baseErr, unwrapped)
	})

	t.Run("errors.Is works correctly", func(t *testing.T) {
		baseErr := errors.New("base error")
		transformErr := &TransformError{
			Transformer: "test-transformer",
			Stage:       StageRequest,
			Err:         baseErr,
		}

		assert.True(t, errors.Is(transformErr, baseErr))
	})

	t.Run("error with empty transformer name", func(t *testing.T) {
		baseErr := errors.New("base error")
		transformErr := &TransformError{
			Transformer: "",
			Stage:       StageRequest,
			Err:         baseErr,
		}

		errMsg := transformErr.Error()
		assert.Contains(t, errMsg, "unknown transformer")
	})
}

func TestPipelineError(t *testing.T) {
	t.Run("error message with single error", func(t *testing.T) {
		err1 := errors.New("error 1")
		pipelineErr := &PipelineError{
			Errors: []error{err1},
		}

		errMsg := pipelineErr.Error()
		assert.Contains(t, errMsg, "error 1")
	})

	t.Run("error message with multiple errors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		pipelineErr := &PipelineError{
			Errors: []error{err1, err2},
		}

		errMsg := pipelineErr.Error()
		assert.Contains(t, errMsg, "2 errors")
		assert.Contains(t, errMsg, "error 1")
		assert.Contains(t, errMsg, "error 2")
	})

	t.Run("error message with no errors", func(t *testing.T) {
		pipelineErr := &PipelineError{
			Errors: []error{},
		}

		errMsg := pipelineErr.Error()
		assert.Contains(t, errMsg, "no errors")
	})

	t.Run("unwrap returns error slice", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		pipelineErr := &PipelineError{
			Errors: []error{err1, err2},
		}

		unwrapped := pipelineErr.Unwrap()
		assert.Len(t, unwrapped, 2)
		assert.Equal(t, err1, unwrapped[0])
		assert.Equal(t, err2, unwrapped[1])
	})
}
