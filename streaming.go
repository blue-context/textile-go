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

	"github.com/blue-context/textile-go/internal/clone"
	"github.com/blue-context/warp"
)

// Recv receives the next chunk from the stream with transformation applied.
// Each chunk is deep-cloned and transformed before being returned.
//
// Returns io.EOF when the stream is complete.
// Returns other errors for failure conditions.
func (s *TransformableStream) Recv() (*warp.CompletionChunk, error) {
	// Receive chunk from underlying stream
	chunk, err := s.inner.Recv()
	if err != nil {
		return nil, err
	}

	// Deep clone chunk for safe transformation
	clonedChunk, err := clone.Deep(chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to clone chunk: %w", err)
	}

	// Apply transformations
	tc := &TransformContext{
		Stage:       StageStreamChunk,
		StreamChunk: clonedChunk,
		Metadata:    make(map[string]interface{}),
	}

	if err := s.pipeline.Transform(s.ctx, tc, s.config); err != nil {
		if s.config.errorStrategy == ErrorStrategyFailClosed {
			return nil, err
		}
		// Error already logged by pipeline, continue with original chunk
	}

	return tc.StreamChunk, nil
}

// Close closes the stream and releases resources.
// This closes the underlying warp stream.
func (s *TransformableStream) Close() error {
	return s.inner.Close()
}
