# go-llm-client

**Unified Go client for LLM providers.** One interface for OpenAI, Anthropic, Ollama, and local models.

[![Go Reference](https://pkg.go.dev/badge/github.com/formatho/agent-orchestrator/packages/llm-client.svg)](https://pkg.go.dev/github.com/formatho/agent-orchestrator/packages/llm-client)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## Features

- ✅ **Multi-provider** — OpenAI, Anthropic, Ollama, local models
- ✅ **Unified interface** — Same API for all providers
- ✅ **Streaming support** — Real-time token streaming
- ✅ **Token counting** — Accurate token counting per provider
- ✅ **Retry logic** — Exponential backoff on failures
- ✅ **Per-request override** — Change provider/model per request

---

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/llm-client
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "os"

    llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
    client := llm.NewClient(llm.Config{
        Provider: llm.ProviderOpenAI,
        Model:    "gpt-4o",
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    })

    response, err := client.Simple(context.Background(), "Hello, world!")
    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Streaming

```go
ch, err := client.Stream(ctx, llm.Request{
    Messages: []llm.Message{
        {Role: "user", Content: "Tell me a story"},
    },
})
if err != nil {
    panic(err)
}

for chunk := range ch {
    fmt.Print(chunk.Delta.Content)
    if chunk.Finished {
        break
    }
}
```

### Per-Request Provider Override

```go
// Default is OpenAI, but use Anthropic for this request
response, err := client.Complete(ctx, llm.Request{
    Provider: llm.ProviderAnthropic,
    Model:    "claude-3-opus",
    Messages: []llm.Message{
        {Role: "user", Content: "Hello!"},
    },
})
```

### Ollama (Local)

```go
client := llm.NewClient(llm.Config{
    Provider: llm.ProviderOllama,
    Model:    "llama3",
    BaseURL:  "http://localhost:11434",
})
```

---

## Configuration

```go
type Config struct {
    Provider   Provider  // openai, anthropic, ollama, local
    Model      string    // Model to use
    APIKey     string    // API key (if required)
    BaseURL    string    // Custom endpoint
    MaxRetries int       // Retry count (default: 3)
    Timeout    int       // Request timeout in seconds
    Debug      bool      // Enable debug logging
}
```

---

## Providers

| Provider | Status | Notes |
|----------|--------|-------|
| OpenAI | ✅ | GPT-4, GPT-3.5, etc. |
| Anthropic | 🚧 | Claude models |
| Ollama | 🚧 | Local models |
| Local | 📋 | Custom endpoints |

---

## API Reference

### `NewClient(config Config) *Client`

Creates a new LLM client.

### `Complete(ctx, Request) (*Response, error)`

Sends a completion request.

### `Stream(ctx, Request) (<-chan StreamChunk, error)`

Streams completion tokens.

### `Simple(ctx, prompt) (string, error)`

Convenience method for simple prompts.

### `CountTokens(text) int`

Counts tokens in text.

---

## License

MIT

---

*Part of [Agent Orchestrator](https://github.com/formatho/agent-orchestrator)*
