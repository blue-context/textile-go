# Textile-Go Examples

This directory contains example applications demonstrating the usage of textile-go.

## Prerequisites

Set your OpenAI API key:

```bash
export OPENAI_API_KEY=sk-...
```

## Examples

### Basic Usage

Demonstrates basic usage with a simple transformer that adds a prefix to user messages.

```bash
cd basic
go run main.go
```

### Custom Transformers

Shows how to build custom transformers and combine them in a pipeline:
- LoggingTransformer: logs requests and responses
- UppercaseTransformer: converts user messages to uppercase
- MetadataTransformer: adds custom metadata

```bash
cd custom_transformer
go run main.go
```

### Streaming

Demonstrates streaming responses with real-time chunk transformation.

```bash
cd streaming
go run main.go
```

## Common Patterns

### Creating a Simple Transformer

```go
func MyTransformer(ctx context.Context, tc *textile.TransformContext) error {
    if tc.Stage == textile.StageRequest {
        // Modify request messages
        for i := range tc.Messages {
            // Transform message
        }
    }
    return nil
}
```

### Building a Pipeline

```go
pipeline := textile.NewPipelineBuilder().
    Add(transformer1).
    Add(transformer2).
    AddFunc(func(tc *textile.TransformContext) error {
        // Inline transformer
        return nil
    }).
    Build()
```

### Error Handling Strategies

```go
// Fail open (default): continue on errors
client, _ := textile.New(
    textile.WithWarpClient(warpClient),
    textile.WithErrorStrategy(textile.ErrorStrategyFailOpen),
)

// Fail closed: return errors immediately
client, _ := textile.New(
    textile.WithWarpClient(warpClient),
    textile.WithErrorStrategy(textile.ErrorStrategyFailClosed),
)
```
