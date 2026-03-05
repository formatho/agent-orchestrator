# go-llm-client

**Unified Go client for LLM providers.** One interface for OpenAI, Anthropic, Ollama, and local models.

[![Go Reference](https://pkg.go.dev/badge/github.com/formatho/agent-orchestrator/packages/llm-client.svg)](https://pkg.go.dev/github.com/formatho/agent-orchestrator/packages/llm-client)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/formatho/agent-orchestrator/packages/llm-client)](https://goreportcard.com/report/github.com/formatho/agent-orchestrator/packages/llm-client)

---

## Table of Contents

- [Why go-llm-client?](#why-go-llm-client)
- [Installation](#installation)
- [Quick Start](#quick-start)
  - [OpenAI](#openai)
  - [Anthropic](#anthropic)
  - [Ollama](#ollama)
- [API Reference](#api-reference)
- [Streaming](#streaming)
- [Error Handling](#error-handling)
- [Configuration](#configuration)
- [Provider Comparison](#provider-comparison)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

---

## Why go-llm-client?

- **🔄 Multi-provider** — Same API for OpenAI, Anthropic, Ollama, and local models
- **⚡ Streaming support** — Real-time token streaming for responsive UX
- **🔁 Auto-retry** — Exponential backoff on transient failures (429, 500, etc.)
- **🎯 Type-safe errors** — Catch specific error types (rate limit, auth, context length)
- **📊 Token counting** — Built-in token estimation per provider
- **🔧 Per-request override** — Change provider/model on the fly
- **Production-ready** — Comprehensive error handling, retries, and logging

---

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/llm-client
```

Import in your code:

```go
import llm "github.com/formatho/agent-orchestrator/packages/llm-client"
```

---

## Quick Start

### OpenAI

```go
package main

import (
    "context"
    "fmt"
    "os"

    llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
    // Create client with OpenAI as default provider
    client := llm.NewClient(llm.Config{
        Provider: llm.ProviderOpenAI,
        Model:    "gpt-4o",
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    })

    // Register OpenAI provider
    llm.RegisterOpenAI(client, llm.OpenAIConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    // Simple completion
    response, err := client.Simple(context.Background(), "What is the capital of France?")
    if err != nil {
        panic(err)
    }

    fmt.Println(response)
    // Output: The capital of France is Paris.
}
```

### Anthropic

```go
package main

import (
    "context"
    "fmt"
    "os"

    llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
    // Create client with Anthropic as default provider
    client := llm.NewClient(llm.Config{
        Provider: llm.ProviderAnthropic,
        Model:    "claude-3-opus-20240229",
        APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
    })

    // Register Anthropic provider (when available)
    // llm.RegisterAnthropic(client, llm.AnthropicConfig{
    //     APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    // })

    // Complete with conversation history
    response, err := client.Complete(context.Background(), llm.Request{
        Messages: []llm.Message{
            {Role: "system", Content: "You are a helpful coding assistant"},
            {Role: "user", Content: "Write a Go function that reverses a string"},
        },
        MaxTokens: 500,
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(response.Content)
}
```

### Ollama

```go
package main

import (
    "context"
    "fmt"

    llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
    // Create client with Ollama for local models
    client := llm.NewClient(llm.Config{
        Provider: llm.ProviderOllama,
        Model:    "llama2",
        BaseURL:  "http://localhost:11434", // Ollama default
    })

    // Register Ollama provider (when available)
    // llm.RegisterOllama(client, llm.OllamaConfig{
    //     BaseURL: "http://localhost:11434",
    // })

    // Simple local completion
    response, err := client.Simple(context.Background(), "Explain quantum computing in one paragraph")
    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

---

## API Reference

### Client

#### `NewClient(config Config) *Client`

Creates a new LLM client with the specified configuration.

```go
client := llm.NewClient(llm.Config{
    Provider:   llm.ProviderOpenAI,
    Model:      "gpt-4o",
    APIKey:     "your-api-key",
    MaxRetries: 3,
    Timeout:    60,
    Debug:      false,
})
```

### Configuration Options

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Provider` | `Provider` | LLM provider to use | Required |
| `Model` | `string` | Default model for completions | Provider-specific |
| `APIKey` | `string` | API key for authentication | Required |
| `BaseURL` | `string` | Custom API endpoint | Provider default |
| `MaxRetries` | `int` | Number of retry attempts | 3 |
| `Timeout` | `int` | Request timeout in seconds | 60 |
| `Debug` | `bool` | Enable debug logging | false |

### Methods

#### `Complete(ctx context.Context, req Request) (*Response, error)`

Sends a completion request with full control over parameters.

```go
response, err := client.Complete(ctx, llm.Request{
    Messages: []llm.Message{
        {Role: "system", Content: "You are a helpful assistant"},
        {Role: "user", Content: "Explain recursion"},
    },
    Model:       "gpt-4o",              // Optional: override default
    Provider:    llm.ProviderOpenAI,     // Optional: override default
    MaxTokens:   1000,
    Temperature: 0.7,
    TopP:        0.9,
    Stop:        []string{"\n", "END"},
})
```

#### `Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)`

Streams completion tokens in real-time.

```go
stream, err := client.Stream(ctx, llm.Request{
    Messages: []llm.Message{
        {Role: "user", Content: "Tell me a story"},
    },
})
if err != nil {
    panic(err)
}

for chunk := range stream {
    fmt.Print(chunk.Delta.Content)
    if chunk.Finished {
        break
    }
}
```

#### `Simple(ctx context.Context, prompt string) (string, error)`

Convenience method for quick completions without message arrays.

```go
response, err := client.Simple(ctx, "What is machine learning?")
```

#### `SimpleStream(ctx context.Context, prompt string, writer io.Writer) error`

Convenience method for streaming directly to a writer.

```go
err := client.SimpleStream(ctx, "Explain Go concurrency", os.Stdout)
```

#### `CountTokens(text string) int`

Estimates token count for text (provider-specific).

```go
tokens := client.CountTokens("Hello, world!")
fmt.Printf("Token count: %d\n", tokens)
```

### Types

#### `Message`

```go
type Message struct {
    Role    string `json:"role"`              // "system", "user", "assistant"
    Content string `json:"content"`           // Message content
    Name    string `json:"name,omitempty"`    // Optional name for function calling
}
```

#### `Request`

```go
type Request struct {
    Messages          []Message  `json:"messages"`
    Model             string     `json:"model,omitempty"`
    Provider          Provider   `json:"provider,omitempty"`
    MaxTokens         int        `json:"max_tokens,omitempty"`
    Temperature       float64    `json:"temperature,omitempty"`
    Stream            bool       `json:"stream,omitempty"`
    TopP              float64    `json:"top_p,omitempty"`
    Stop              []string   `json:"stop,omitempty"`
    FrequencyPenalty  float64    `json:"frequency_penalty,omitempty"`
    PresencePenalty   float64    `json:"presence_penalty,omitempty"`
}
```

#### `Response`

```go
type Response struct {
    ID      string   `json:"id"`
    Model   string   `json:"model"`
    Content string   `json:"content"`
    Usage   Usage    `json:"usage"`
    Choices []Choice `json:"choices,omitempty"`
}
```

#### `Usage`

```go
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

#### `StreamChunk`

```go
type StreamChunk struct {
    Delta    Message `json:"delta"`
    Finished bool    `json:"finished"`
}
```

---

## Streaming

Streaming is ideal for long-form content, chat interfaces, or when you want to show progress to users.

### Basic Streaming

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

    llm.RegisterOpenAI(client, llm.OpenAIConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    // Stream a long response
    stream, err := client.Stream(context.Background(), llm.Request{
        Messages: []llm.Message{
            {Role: "system", Content: "You are a helpful assistant"},
            {Role: "user", Content: "Tell me a short story about a robot learning to paint"},
        },
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Streaming response:")
    fmt.Println("---")

    for chunk := range stream {
        fmt.Print(chunk.Delta.Content)
        if chunk.Finished {
            break
        }
    }

    fmt.Println("\n---")
}
```

### Streaming to File

```go
file, err := os.Create("output.txt")
if err != nil {
    panic(err)
}
defer file.Close()

err = client.SimpleStream(ctx, "Write a poem about coding", file)
if err != nil {
    panic(err)
}
```

### Streaming with Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := client.Stream(ctx, llm.Request{
    Messages: []llm.Message{
        {Role: "user", Content: "Generate a long response"},
    },
})
if err != nil {
    panic(err)
}

for chunk := range stream {
    select {
    case <-ctx.Done():
        fmt.Println("\nCancelled!")
        return
    default:
        fmt.Print(chunk.Delta.Content)
    }
}
```

---

## Error Handling

The library provides type-safe errors for better error handling and user feedback.

### Error Types

| Error Type | Description | When It Occurs |
|------------|-------------|----------------|
| `AuthenticationError` | Invalid API key or auth failure | Wrong API key, expired token |
| `RateLimitError` | Rate limit exceeded | Too many requests, quota exceeded |
| `ModelNotFoundError` | Model doesn't exist | Invalid model name |
| `ContextLengthExceededError` | Context too long | Prompt + completion > model limit |
| `InvalidRequestError` | Malformed request | Missing required fields, invalid parameters |
| `ProviderUnavailableError` | Provider temporarily down | HTTP 500, 502, 503, 504 |

### Error Checking Functions

```go
_, err := client.Simple(ctx, "Hello")
if err != nil {
    if llm.IsAuthenticationError(err) {
        // Handle invalid API key
        fmt.Println("❌ Authentication failed - check your API key")
    } else if llm.IsRateLimitError(err) {
        // Handle rate limiting
        fmt.Println("⏳ Rate limit exceeded - please wait and retry")
    } else if llm.IsModelNotFoundError(err) {
        // Handle invalid model
        fmt.Println("🔍 Model not found - check model name")
    } else if llm.IsContextLengthError(err) {
        // Handle context length exceeded
        fmt.Println("📏 Context too long - reduce message size")
    } else if llm.IsRetryable(err) {
        // Handle transient errors
        fmt.Println("🔄 Transient error - will retry automatically")
    } else {
        // Handle unknown errors
        fmt.Printf("❓ Unknown error: %v\n", err)
    }
}
```

### Comprehensive Error Handling Example

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
        Provider:   llm.ProviderOpenAI,
        Model:      "gpt-4o",
        APIKey:     "invalid-key", // Intentionally invalid for demo
        MaxRetries: 2,
    })

    llm.RegisterOpenAI(client, llm.OpenAIConfig{
        APIKey: "invalid-key",
    })

    // Try a completion
    _, err := client.Simple(context.Background(), "Hello")
    if err != nil {
        // Check error type
        if llm.IsAuthenticationError(err) {
            fmt.Println("❌ Authentication failed - check your API key")
            os.Exit(1)
        } else if llm.IsRateLimitError(err) {
            fmt.Println("⏳ Rate limit exceeded - please wait and retry")
            os.Exit(1)
        } else if llm.IsModelNotFoundError(err) {
            fmt.Println("🔍 Model not found - check model name")
            os.Exit(1)
        } else if llm.IsContextLengthError(err) {
            fmt.Println("📏 Context too long - reduce message size")
            os.Exit(1)
        } else if llm.IsRetryable(err) {
            fmt.Println("🔄 Transient error - will retry automatically")
        } else {
            fmt.Printf("❓ Unknown error: %v\n", err)
            os.Exit(1)
        }
    }

    fmt.Println("Success!")
}
```

### Accessing Error Details

```go
import "errors"

_, err := client.Simple(ctx, "Hello")
if err != nil {
    var rateLimitErr *llm.RateLimitError
    if errors.As(err, &rateLimitErr) {
        fmt.Printf("Rate limited. Retry after %d seconds\n", rateLimitErr.RetryAfter)
    }
    
    var modelErr *llm.ModelNotFoundError
    if errors.As(err, &modelErr) {
        fmt.Printf("Model '%s' not found\n", modelErr.Model)
    }
}
```

---

## Configuration

### Retry Configuration

The client automatically retries on transient failures with exponential backoff.

**Retryable errors:**
- HTTP 429 (rate limit)
- HTTP 500 (internal server error)
- HTTP 502 (bad gateway)
- HTTP 503 (service unavailable)
- HTTP 504 (gateway timeout)

**Backoff strategy:**
- Exponential: 1s, 2s, 4s, 8s, ...
- Default: 3 retries (4 total attempts)

```go
// Default retry behavior (3 retries)
client := llm.NewClient(llm.Config{
    Provider: llm.ProviderOpenAI,
    Model:    "gpt-4o",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
})

// Custom retry configuration (5 retries)
client := llm.NewClient(llm.Config{
    Provider:   llm.ProviderOpenAI,
    Model:      "gpt-4o",
    APIKey:     os.Getenv("OPENAI_API_KEY"),
    MaxRetries: 5, // 6 total attempts (1 initial + 5 retries)
})

// Register with matching config
llm.RegisterOpenAI(client, llm.OpenAIConfig{
    APIKey:     os.Getenv("OPENAI_API_KEY"),
    MaxRetries: 5,
    Debug:      true, // Enable debug to see retry attempts
})
```

### Timeout Configuration

```go
client := llm.NewClient(llm.Config{
    Provider: llm.ProviderOpenAI,
    Model:    "gpt-4o",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Timeout:  120, // 2 minutes
})

// Per-request timeout with context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := client.Simple(ctx, "Hello")
```

### Custom Endpoints

```go
// Use a custom OpenAI-compatible API
client := llm.NewClient(llm.Config{
    Provider: llm.ProviderOpenAI,
    Model:    "custom-model",
    APIKey:   "your-key",
    BaseURL:  "https://your-custom-api.com/v1",
})

llm.RegisterOpenAI(client, llm.OpenAIConfig{
    APIKey:  "your-key",
    BaseURL: "https://your-custom-api.com/v1",
})
```

### Per-Request Override

Override provider or model on a per-request basis:

```go
// Default is OpenAI, but use Anthropic for this request
response, err := client.Complete(ctx, llm.Request{
    Provider: llm.ProviderAnthropic,
    Model:    "claude-3-opus",
    Messages: []llm.Message{
        {Role: "user", Content: "Hello!"},
    },
})

// Use different model for specific task
response, err := client.Complete(ctx, llm.Request{
    Model: "gpt-4-turbo", // Override default gpt-4o
    Messages: []llm.Message{
        {Role: "user", Content: "Analyze this code"},
    },
})
```

### Debug Mode

Enable debug logging to see request/response details:

```go
llm.RegisterOpenAI(client, llm.OpenAIConfig{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Debug:  true,
})

// Output:
// [OpenAI] Request: {"model":"gpt-4o","messages":[...]}
// [OpenAI] Response (status 200): {"id":"...","choices":[...]}
```

---

## Provider Comparison

| Feature | OpenAI | Anthropic | Ollama | Local |
|---------|--------|-----------|--------|-------|
| **Status** | ✅ Stable | 🚧 In Progress | 🚧 In Progress | 📋 Planned |
| **Streaming** | ✅ Yes | 🚧 Planned | 🚧 Planned | 📋 Planned |
| **Auto-retry** | ✅ Yes | 🚧 Planned | 🚧 Planned | 📋 Planned |
| **Token counting** | ✅ Approximation | 🚧 Planned | 🚧 Planned | 📋 Planned |
| **Models** | GPT-4, GPT-3.5 | Claude 3 | Llama 2, Mistral | Custom |
| **API Key Required** | ✅ Yes | ✅ Yes | ❌ No | ❌ No |
| **Rate Limits** | ✅ Yes | ✅ Yes | ❌ No | ❌ No |
| **Cost** | $$ | $$$ | Free | Free |
| **Latency** | Low | Low | Varies | Varies |
| **Privacy** | Cloud | Cloud | Local | Local |

### When to Use Each Provider

**OpenAI**
- Production applications requiring high reliability
- Complex reasoning and coding tasks
- When you need the best overall performance
- Applications with budget for API costs

**Anthropic**
- When you need Constitutional AI principles
- Longer context windows (up to 200K tokens)
- Tasks requiring nuanced ethical considerations
- Alternative to OpenAI for diversity

**Ollama**
- Development and testing without API costs
- Privacy-sensitive applications
- Offline scenarios
- Custom fine-tuned models

**Local/Custom**
- Self-hosted deployments
- Proprietary models
- Air-gapped environments
- Maximum control over data

---

## Examples

The [`examples/`](./examples/) directory contains complete working examples:

| Example | Description | Run |
|---------|-------------|-----|
| [basic-usage](./examples/basic-usage) | Simple completion | `cd examples/basic-usage && go run main.go` |
| [streaming](./examples/streaming) | Real-time token streaming | `cd examples/streaming && go run main.go` |
| [error-handling](./examples/error-handling) | Type-safe error handling | `cd examples/error-handling && go run main.go` |
| [retry-configuration](./examples/retry-configuration) | Custom retry behavior | `cd examples/retry-configuration && go run main.go` |
| [concurrent](./examples/concurrent) | Multiple concurrent requests | `cd examples/concurrent && go run main.go` |

### Concurrent Requests

```go
package main

import (
    "context"
    "fmt"
    "os"
    "sync"
    "time"

    llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
    client := llm.NewClient(llm.Config{
        Provider:   llm.ProviderOpenAI,
        Model:      "gpt-4o",
        APIKey:     os.Getenv("OPENAI_API_KEY"),
        MaxRetries: 3,
    })

    llm.RegisterOpenAI(client, llm.OpenAIConfig{
        APIKey:     os.Getenv("OPENAI_API_KEY"),
        MaxRetries: 3,
    })

    prompts := []string{
        "What is 2+2?",
        "What is the capital of Japan?",
        "Name a primary color",
        "What planet do we live on?",
        "Name a popular programming language",
    }

    var wg sync.WaitGroup
    results := make([]string, len(prompts))
    errors := make([]error, len(prompts))

    start := time.Now()

    for i, prompt := range prompts {
        wg.Add(1)
        go func(idx int, p string) {
            defer wg.Done()

            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            resp, err := client.Simple(ctx, p)
            if err != nil {
                errors[idx] = err
                return
            }
            results[idx] = resp
        }(i, prompt)
    }

    wg.Wait()
    elapsed := time.Since(start)

    fmt.Printf("Completed %d requests in %v\n\n", len(prompts), elapsed)

    for i, prompt := range prompts {
        fmt.Printf("Q%d: %s\n", i+1, prompt)
        if errors[i] != nil {
            fmt.Printf("   Error: %v\n", errors[i])
        } else {
            fmt.Printf("   A: %s\n", results[i])
        }
        fmt.Println()
    }
}
```

---

## Contributing

We welcome contributions! Here's how to get started:

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/formatho/agent-orchestrator.git
   cd agent-orchestrator/packages/llm-client
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run tests**
   ```bash
   go test ./...
   ```

4. **Run examples**
   ```bash
   cd examples/basic-usage
   go run main.go
   ```

### Adding a New Provider

To add a new provider (e.g., Cohere, Gemini):

1. **Create provider file** (e.g., `cohere.go`)
   ```go
   type CohereProvider struct {
       apiKey  string
       baseURL string
       client  *http.Client
   }

   type CohereConfig struct {
       APIKey  string
       BaseURL string
       Timeout int
   }

   func NewCohereProvider(config CohereConfig) *CohereProvider {
       // Implementation
   }

   func (p *CohereProvider) Complete(ctx context.Context, req Request) (*Response, error) {
       // Implementation
   }

   func (p *CohereProvider) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
       // Implementation
   }

   func (p *CohereProvider) CountTokens(text string) int {
       // Implementation
   }
   ```

2. **Create registration function** in `register.go`
   ```go
   func RegisterCohere(client *Client, config CohereConfig) {
       client.SetProvider(ProviderCohere, NewCohereProvider(config))
   }
   ```

3. **Add provider constant** in `client.go`
   ```go
   const (
       ProviderCohere Provider = "cohere"
   )
   ```

4. **Write tests** (e.g., `cohere_test.go`)

5. **Create examples** in `examples/cohere/`

6. **Update README.md** with usage examples

### Code Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `go vet` and `golint`
- Write comprehensive tests
- Add documentation comments

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestOpenAIComplete

# Run benchmarks
go test -bench=. ./...
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Update documentation (README.md, godoc comments)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Reporting Issues

- Use [GitHub Issues](https://github.com/formatho/agent-orchestrator/issues)
- Include Go version, OS, and library version
- Provide minimal reproduction code
- Describe expected vs actual behavior

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2024 Formatho

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## Acknowledgments

- Built with ❤️ by the Formatho team
- Inspired by the need for a unified LLM interface in Go
- Thanks to all contributors and users

---

*Part of [Agent Orchestrator](https://github.com/formatho/agent-orchestrator)*

**Need help?** [Open an issue](https://github.com/formatho/agent-orchestrator/issues) or check the [documentation](https://pkg.go.dev/github.com/formatho/agent-orchestrator/packages/llm-client)
