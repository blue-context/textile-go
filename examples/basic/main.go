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

// Package main demonstrates basic usage of textile-go.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/blue-context/textile-go"
	"github.com/blue-context/warp"
	"github.com/blue-context/warp/provider/openai"
)

func main() {
	// Create warp client with OpenAI
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Create warp client
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

	// Create a simple transformer that adds a prefix to user messages
	prefixTransformer := func(ctx context.Context, tc *textile.TransformContext) error {
		if tc.Stage == textile.StageRequest {
			for i := range tc.Messages {
				if tc.Messages[i].Role == "user" {
					if content, ok := tc.Messages[i].Content.(string); ok {
						tc.Messages[i].Content = "[IMPORTANT] " + content
					}
				}
			}
		}
		return nil
	}

	// Create textile client wrapping warp
	client, err := textile.New(
		textile.WithWarpClient(warpClient),
		textile.WithTransformer(prefixTransformer),
		textile.WithErrorStrategy(textile.ErrorStrategyFailOpen),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Use exactly like warp - zero API changes
	resp, err := client.Completion(context.Background(), &warp.CompletionRequest{
		Model: "openai/gpt-4o-mini",
		Messages: []warp.Message{
			{Role: "user", Content: "Say hello in exactly 5 words"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Print response
	fmt.Println("Response:", resp.Choices[0].Message.Content)
}
