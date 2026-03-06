package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicProvider implements the ProviderClient interface for Anthropic
type AnthropicProvider struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	debug      bool
	maxRetries int
}

// AnthropicConfig is configuration for Anthropic provider
type AnthropicConfig struct {
	APIKey     string
	BaseURL    string // Optional, defaults to https://api.anthropic.com/v1
	Model      string // Default model (e.g., claude-3-opus-20240229)
	Timeout    int    // Timeout in seconds
	MaxRetries int    // Max retry attempts (default: 3)
	Debug      bool
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config AnthropicConfig) *AnthropicProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 120 // Anthropic can be slower, use 120s default
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	model := config.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Default to latest stable
	}

	return &AnthropicProvider{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		model:      model,
		debug:      config.Debug,
		maxRetries: maxRetries,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// anthropicRequest represents the request body for Anthropic API
type anthropicRequest struct {
	Model       string         `json:"model"`
	Messages    []anthropicMsg `json:"messages"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	Stop        []string       `json:"stop_sequences,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	System      string         `json:"system,omitempty"`
}

// anthropicMsg represents a message in Anthropic's format
type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents the response from Anthropic API
type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Complete sends a completion request to Anthropic with retry logic
func (p *AnthropicProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if p.debug {
				fmt.Printf("[Anthropic] Retry attempt %d after %v\n", attempt, backoff)
			}

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := p.doRequest(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if !p.shouldRetry(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries (%d) exceeded, last error: %w", p.maxRetries, lastErr)
}

// doRequest performs a single HTTP request without retry
func (p *AnthropicProvider) doRequest(ctx context.Context, req Request) (*Response, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	// Convert messages to Anthropic format
	var systemPrompt string
	anthropicMsgs := make([]anthropicMsg, 0, len(req.Messages))

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}
		anthropicMsgs = append(anthropicMsgs, anthropicMsg{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	anthropicReq := anthropicRequest{
		Model:       model,
		Messages:    anthropicMsgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      false,
		System:      systemPrompt,
	}

	if anthropicReq.MaxTokens == 0 {
		anthropicReq.MaxTokens = 4096 // Default max tokens for Anthropic
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	if p.debug {
		fmt.Printf("[Anthropic] Request: %s\n", string(body))
	}

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, &RetryableError{Err: err}
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if p.debug {
		fmt.Printf("[Anthropic] Response (status %d): %s\n", httpResp.StatusCode, string(respBody))
	}

	if p.shouldRetryHTTP(httpResp.StatusCode) {
		return nil, &RetryableError{
			Err:    fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respBody)),
			Status: httpResp.StatusCode,
		}
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if anthropicResp.Error != nil {
		return nil, p.convertAPIError(httpResp.StatusCode, anthropicResp.Error)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := anthropicResp.Content[0].Text
	return &Response{
		ID:      anthropicResp.ID,
		Model:   anthropicResp.Model,
		Content: content,
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: anthropicResp.StopReason,
			},
		},
	}, nil
}

// convertAPIError converts Anthropic API errors to our error types
func (p *AnthropicProvider) convertAPIError(statusCode int, apiErr *struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}) error {
	switch apiErr.Type {
	case "authentication_error":
		return &AuthenticationError{
			Provider: "anthropic",
			Message:  apiErr.Message,
		}
	case "rate_limit_error":
		return &RateLimitError{
			Provider: "anthropic",
			Message:  apiErr.Message,
		}
	case "not_found_error":
		return &ModelNotFoundError{
			Provider: "anthropic",
			Model:    "",
		}
	case "invalid_request_error":
		if strings.Contains(apiErr.Message, "context length") {
			return &ContextLengthExceededError{
				Provider: "anthropic",
				Message:  apiErr.Message,
			}
		}
		return &InvalidRequestError{
			Provider: "anthropic",
			Message:  apiErr.Message,
		}
	case "overloaded_error":
		return &ProviderUnavailableError{
			Provider: "anthropic",
			Status:   statusCode,
			Message:  apiErr.Message,
		}
	}

	return fmt.Errorf("Anthropic API error: %s (type: %s)", apiErr.Message, apiErr.Type)
}

// shouldRetry checks if an error is retryable
func (p *AnthropicProvider) shouldRetry(err error) bool {
	var retryErr *RetryableError
	return errors.As(err, &retryErr)
}

// shouldRetryHTTP checks if an HTTP status code is retryable
func (p *AnthropicProvider) shouldRetryHTTP(statusCode int) bool {
	retryableCodes := map[int]bool{
		429: true,
		500: true,
		502: true,
		503: true,
		504: true,
		529: true, // Overloaded
	}
	return retryableCodes[statusCode]
}

// Stream sends a streaming completion request to Anthropic
func (p *AnthropicProvider) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	model := req.Model
	if model == "" {
		model = p.model
	}

	var systemPrompt string
	anthropicMsgs := make([]anthropicMsg, 0, len(req.Messages))

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}
		anthropicMsgs = append(anthropicMsgs, anthropicMsg{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	anthropicReq := anthropicRequest{
		Model:       model,
		Messages:    anthropicMsgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      true,
		System:      systemPrompt,
	}

	if anthropicReq.MaxTokens == 0 {
		anthropicReq.MaxTokens = 4096
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check HTTP status before streaming
	if resp.StatusCode >= 400 {
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		close(ch)
		if err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Message *anthropicResponse `json:"message"`
			}

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				if p.debug {
					fmt.Printf("[Anthropic] Failed to parse chunk: %v\n", err)
				}
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta.Type == "text_delta" {
					ch <- StreamChunk{
						Delta: Message{
							Role:    "assistant",
							Content: event.Delta.Text,
						},
						Finished: false,
					}
				}
			case "message_stop":
				ch <- StreamChunk{Finished: true}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if p.debug {
				fmt.Printf("[Anthropic] Scanner error: %v\n", err)
			}
		}
	}()

	return ch, nil
}

// CountTokens counts tokens using Anthropic's tokenizer
func (p *AnthropicProvider) CountTokens(text string) int {
	// Approximation: 1 token ≈ 4 characters
	return len(text) / 4
}
