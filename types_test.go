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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformStage_String(t *testing.T) {
	tests := []struct {
		name  string
		stage TransformStage
		want  string
	}{
		{
			name:  "request stage",
			stage: StageRequest,
			want:  "Request",
		},
		{
			name:  "response stage",
			stage: StageResponse,
			want:  "Response",
		},
		{
			name:  "stream chunk stage",
			stage: StageStreamChunk,
			want:  "StreamChunk",
		},
		{
			name:  "unknown stage",
			stage: TransformStage(999),
			want:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stage.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorStrategy_String(t *testing.T) {
	tests := []struct {
		name     string
		strategy ErrorStrategy
		want     string
	}{
		{
			name:     "fail closed strategy",
			strategy: ErrorStrategyFailClosed,
			want:     "FailClosed",
		},
		{
			name:     "fail open strategy",
			strategy: ErrorStrategyFailOpen,
			want:     "FailOpen",
		},
		{
			name:     "unknown strategy",
			strategy: ErrorStrategy(999),
			want:     "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.strategy.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
