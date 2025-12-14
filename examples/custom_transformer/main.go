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

// Package main demonstrates custom transformers with textile-go.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/blue-context/textile-go"
	"github.com/blue-context/warp"
	"github.com/blue-context/warp/provider/openai"
)

// LoggingTransformer logs all requests and responses
func LoggingTransformer(ctx context.Context, tc *textile.TransformContext) error {
	switch tc.Stage {
	case textile.StageRequest:
		fmt.Printf("[LOG] Request to model: %s\n", tc.Request.Model)
		fmt.Printf("[LOG] Number of messages: %d\n", len(tc.Messages))

	case textile.StageResponse:
		fmt.Printf("[LOG] Response ID: %s\n", tc.Response.ID)
		if tc.Response.Usage != nil {
			fmt.Printf("[LOG] Tokens used: %d\n", tc.Response.Usage.TotalTokens)
		}

	case textile.StageStreamChunk:
		fmt.Printf("[LOG] Received chunk: %s\n", tc.StreamChunk.ID)
	}
	return nil
}

// UppercaseTransformer converts all user messages to uppercase
func UppercaseTransformer(ctx context.Context, tc *textile.TransformContext) error {
	if tc.Stage == textile.StageRequest {
		for i := range tc.Messages {
			if tc.Messages[i].Role == "user" {
				if content, ok := tc.Messages[i].Content.(string); ok {
					tc.Messages[i].Content = strings.ToUpper(content)
				}
			}
		}
	}
	return nil
}

// MetadataTransformer adds custom metadata to requests
func MetadataTransformer(ctx context.Context, tc *textile.TransformContext) error {
	if tc.Stage == textile.StageRequest {
		// Store information in metadata for other transformers
		tc.Metadata["transformed_by"] = "MetadataTransformer"
		tc.Metadata["message_count"] = len(tc.Messages)
	}
	return nil
}

func main() {
	// Create warp client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	warpClient, err := warp.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer warpClient.Close()

	// Register OpenAI provider
	provider, err := openai.NewProvider(openai.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	if err := warpClient.RegisterProvider(provider); err != nil {
		log.Fatal(err)
	}

	// Build pipeline with multiple transformers
	pipeline := textile.NewPipelineBuilder().
		Add(MetadataTransformer).
		Add(LoggingTransformer).
		Add(UppercaseTransformer).
		Build()

	// Create textile client with pipeline
	client, err := textile.New(
		textile.WithWarpClient(warpClient),
		textile.WithPipeline(pipeline),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("Sending request with transformers...")

	// Send completion request
	resp, err := client.Completion(context.Background(), &warp.CompletionRequest{
		Model: "openai/gpt-4o-mini",
		Messages: []warp.Message{
			{Role: "user", Content: "explain transformers in one sentence"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nResponse: %s\n", resp.Choices[0].Message.Content)
}
