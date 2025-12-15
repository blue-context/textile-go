# Textile-Go

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/blue-context/textile-go)

Transparent message transformation middleware for LLM clients in Go.

Textile wraps the [warp](https://github.com/blue-context/warp) client with a powerful transformation pipeline system while maintaining zero API changes for consumers. This allows you to apply configurable transformations to messages before and after LLM calls transparently.

## Features

- **Zero API Changes**: Drop-in replacement for warp.Client
- **Transformation Pipeline**: Apply multiple transformers in sequence
- **Request & Response Transformation**: Transform messages before sending and responses after receiving
- **Streaming Support**: Transform streaming chunks in real-time
- **Thread-Safe**: Safe for concurrent use from multiple goroutines
- **Immutable Operations**: Deep cloning ensures transformations don't affect original data
- **Flexible Error Handling**: Choose fail-open (continue on errors) or fail-closed (stop on errors)
- **Composable**: Build complex pipelines with simple transformers

## Installation

```bash
go get github.com/blue-context/textile-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/blue-context/textile-go"
    "github.com/blue-context/warp"
    "github.com/blue-context/warp/provider/openai"
)

func main() {
    // Create warp client
    warpClient, _ := warp.NewClient()
    defer warpClient.Close()

    // Register provider
    provider, _ := openai.NewProvider(openai.WithAPIKey("sk-..."))
    warpClient.RegisterProvider(provider)

    // Create a simple transformer
    prefixTransformer := func(ctx context.Context, tc *textile.TransformContext) error {
        if tc.Stage == textile.StageRequest {
            for i := range tc.Messages {
                if content, ok := tc.Messages[i].Content.(string); ok {
                    tc.Messages[i].Content = "[IMPORTANT] " + content
                }
            }
        }
        return nil
    }

    // Wrap with textile
    client, _ := textile.New(
        textile.WithWarpClient(warpClient),
        textile.WithTransformer(prefixTransformer),
    )
    defer client.Close()

    // Use exactly like warp - zero API changes
    resp, _ := client.Completion(context.Background(), &warp.CompletionRequest{
        Model: "openai/gpt-4o-mini",
        Messages: []warp.Message{
            {Role: "user", Content: "Hello!"},
        },
    })

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Core Concepts

### Transformers

Transformers are functions that modify messages or responses:

```go
type Transformer func(ctx context.Context, tc *TransformContext) error
```

Transformers receive a `TransformContext` indicating which stage of processing is occurring:
- `StageRequest`: Before sending to LLM (modify `tc.Messages` or `tc.Request`)
- `StageResponse`: After receiving from LLM (modify `tc.Response`)
- `StageStreamChunk`: During streaming (modify `tc.StreamChunk`)

### Pipeline

Transformers are organized in a pipeline and executed sequentially:

```go
pipeline := textile.NewPipelineBuilder().
    Add(transformer1).
    Add(transformer2).
    AddFunc(func(tc *textile.TransformContext) error {
        // Inline transformer
        return nil
    }).
    Build()

client, _ := textile.New(
    textile.WithWarpClient(warpClient),
    textile.WithPipeline(pipeline),
)
```

### Error Strategies

Control how errors are handled:

```go
// Fail open (default): log errors and continue
client, _ := textile.New(
    textile.WithWarpClient(warpClient),
    textile.WithErrorStrategy(textile.ErrorStrategyFailOpen),
    textile.WithErrorHandler(func(ctx context.Context, err error) {
        log.Printf("Transformer error: %v", err)
    }),
)

// Fail closed: return errors immediately
client, _ := textile.New(
    textile.WithWarpClient(warpClient),
    textile.WithErrorStrategy(textile.ErrorStrategyFailClosed),
)
```

## Examples

### Request Transformation

```go
// Convert all user messages to uppercase
uppercaseTransformer := func(ctx context.Context, tc *textile.TransformContext) error {
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
```

### Response Transformation

```go
// Add metadata to responses
metadataTransformer := func(ctx context.Context, tc *textile.TransformContext) error {
    if tc.Stage == textile.StageResponse {
        // Store metadata for later use
        tc.Metadata["model_used"] = tc.Response.Model
        tc.Metadata["tokens_used"] = tc.Response.Usage.TotalTokens
    }
    return nil
}
```

### Streaming Transformation

```go
// Log each streaming chunk
loggingTransformer := func(ctx context.Context, tc *textile.TransformContext) error {
    if tc.Stage == textile.StageStreamChunk {
        log.Printf("Received chunk: %s", tc.StreamChunk.ID)
    }
    return nil
}
```

### Runtime Transformer Addition

```go
// Add transformers at runtime (creates new client)
baseClient, _ := textile.New(
    textile.WithWarpClient(warpClient),
    textile.WithTransformer(loggingTransformer),
)

// Add additional transformer for specific use case
enhancedClient := baseClient.WithTransformers(validationTransformer)
```

## Architecture

Textile uses deep cloning to ensure immutability:

1. **Request Phase**:
   - Deep clone request
   - Apply transformers to clone
   - Send transformed request to LLM

2. **Response Phase**:
   - Receive response from LLM
   - Deep clone response
   - Apply transformers to clone
   - Return transformed response

3. **Streaming Phase**:
   - Receive each chunk
   - Deep clone chunk
   - Apply transformers to clone
   - Return transformed chunk

This ensures transformers cannot accidentally modify the original request/response data.

## Thread Safety

All components are safe for concurrent use:
- `Client`: Can be called from multiple goroutines
- `Pipeline`: Immutable operations (Append/Prepend create new pipelines)
- `TransformContext`: Each request gets its own context

## Testing

Run tests with race detector:

```bash
go test -race ./...
```

Run with coverage:

```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance

Textile uses reflection-based deep cloning which has some overhead:
- Simple requests: ~1-2ms overhead
- Complex nested structures: ~5-10ms overhead

The overhead is typically negligible compared to LLM API latency (100-1000ms+).

Benchmarks:

```bash
go test -bench=. -benchmem ./...
```

## Contributing

Contributions are welcome! Please ensure:
- All tests pass with race detector
- Code follows Go idioms and style guide
- Proper error handling (no ignored errors)
- Documentation for exported types/functions

## License

Copyright 2025 Blue Context Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Related Projects

- [warp](https://github.com/blue-context/warp) - Unified Go SDK for 100+ LLM providers
- [textile (Python)](https://github.com/blue-context/textile) - Original Python implementation
