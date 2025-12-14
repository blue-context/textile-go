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

// Package main demonstrates streaming with textile-go.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/blue-context/textile-go"
	"github.com/blue-context/warp"
	"github.com/blue-context/warp/provider/openai"
)

// StreamChunkTransformer transforms streaming chunks in real-time
func StreamChunkTransformer(ctx context.Context, tc *textile.TransformContext) error {
	if tc.Stage == textile.StageStreamChunk {
		// Add metadata or modify chunks as they arrive
		fmt.Print(".") // Show progress
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

	// Create textile client with streaming transformer
	client, err := textile.New(
		textile.WithWarpClient(warpClient),
		textile.WithTransformer(StreamChunkTransformer),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("Streaming response:")

	// Create streaming request
	stream, err := client.CompletionStream(context.Background(), &warp.CompletionRequest{
		Model: "openai/gpt-4o-mini",
		Messages: []warp.Message{
			{Role: "user", Content: "Write a haiku about Go programming"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	fmt.Print("\n\nContent: ")

	// Read chunks from stream
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Print chunk content
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				fmt.Print(delta.Content)
			}
		}
	}

	fmt.Println("\n\nStreaming complete!")
}
