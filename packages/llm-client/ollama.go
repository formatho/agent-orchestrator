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
	"regexp"
	"time"
)

// OllamaProvider implements the ProviderClient interface for Ollama
type OllamaProvider struct {
	baseURL    string
	client     *http.Client
	debug      bool
	maxRetries int
}

// OllamaConfig is configuration for Ollama provider
type OllamaConfig struct {
	BaseURL    string // Optional, defaults to http://localhost:11434
	Timeout    int    // Timeout in seconds
	MaxRetries int    // Max retry attempts (default: 3)
	Debug      bool
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config OllamaConfig) *OllamaProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &OllamaProvider{
		baseURL:    baseURL,
		debug:      config.Debug,
		maxRetries: maxRetries,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// ollamaRequest represents the request body for Ollama API
type ollamaRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Stream           bool      `json:"stream"`
	Options          struct {
		Temperature      float64 `json:"temperature,omitempty"`
		TopP             float64 `json:"top_p,omitempty"`
		NumPredict       int     `json:"num_predict,omitempty"`
		Stop             []string `json:"stop,omitempty"`
		FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
		PresencePenalty  float64 `json:"presence_penalty,omitempty"`
	} `json:"options,omitempty"`
}

// ollamaResponse represents the response from Ollama API
type ollamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
	EvalCount       int `json:"eval_count"`
	PromptEvalCount int `json:"prompt_eval_count"`
	Error          string `json:"error,omitempty"`
}

// ollamaGenerateResponse represents streaming response chunks
type ollamaGenerateResponse struct {
	Model           string  `json:"model"`
	CreatedAt       string  `json:"created_at"`
	Response        string  `json:"response"`
	Done            bool    `json:"done"`
	TotalDuration   float64 `json:"total_duration"`
	EvalCount       int     `json:"eval_count"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	Error           string  `json:"error,omitempty"`
}

// Complete sends a completion request to Ollama with retry logic
func (p *OllamaProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if p.debug {
				fmt.Printf("[Ollama] Retry attempt %d after %v\n", attempt, backoff)
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
func (p *OllamaProvider) doRequest(ctx context.Context, req Request) (*Response, error) {
	model := req.Model
	if model == "" {
		model = "llama3.2"
	}

	ollamaReq := ollamaRequest{
		Model:    model,
		Messages: req.Messages,
		Stream:   false,
	}
	ollamaReq.Options.Temperature = req.Temperature
	ollamaReq.Options.TopP = req.TopP
	ollamaReq.Options.Stop = req.Stop
	ollamaReq.Options.FrequencyPenalty = req.FrequencyPenalty
	ollamaReq.Options.PresencePenalty = req.PresencePenalty
	if req.MaxTokens > 0 {
		ollamaReq.Options.NumPredict = req.MaxTokens
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	if p.debug {
		fmt.Printf("[Ollama] Request: %s\n", string(body))
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
		fmt.Printf("[Ollama] Response (status %d): %s\n", httpResp.StatusCode, string(respBody))
	}

	if p.shouldRetryHTTP(httpResp.StatusCode) {
		return nil, &RetryableError{
			Err:    fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respBody)),
			Status: httpResp.StatusCode,
		}
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ollamaResp.Error != "" {
		return nil, p.convertAPIError(httpResp.StatusCode, ollamaResp.Error)
	}

	if !ollamaResp.Done {
		return nil, fmt.Errorf("incomplete response from Ollama")
	}

	return &Response{
		ID:      fmt.Sprintf("ollama-%d", time.Now().Unix()),
		Model:   ollamaResp.Model,
		Content: ollamaResp.Message.Content,
		Usage: Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    ollamaResp.Message.Role,
					Content: ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
	}, nil
}

// ollamaErrorResponse represents potential structured error response from Ollama
// While Ollama currently uses simple string errors, this allows for future expansion
type ollamaErrorResponse struct {
	Error string `json:"error"`
	Type  string `json:"type,omitempty"`  // For future structured errors
	Code  string `json:"code,omitempty"`  // For future structured errors
}

// ollamaErrorPattern represents a pattern for matching Ollama errors
type ollamaErrorPattern struct {
	pattern     *regexp.Regexp
	errorType   string
	extractModel bool
}

// ollamaErrorPatterns defines known error patterns from Ollama
var ollamaErrorPatterns = []ollamaErrorPattern{
	{
		pattern:     regexp.MustCompile(`(?i)model\s+['"]?([^'"\s]+)['"]?\s+not\s+found`),
		errorType:   "model_not_found",
		extractModel: true,
	},
	{
		pattern:     regexp.MustCompile(`(?i)model\s+not\s+found`),
		errorType:   "model_not_found",
		extractModel: false,
	},
	{
		pattern:     regexp.MustCompile(`(?i)context\s+length\s+exceeded`),
		errorType:   "context_length_exceeded",
		extractModel: false,
	},
	{
		pattern:     regexp.MustCompile(`(?i)context\s+window\s+exceeded`),
		errorType:   "context_length_exceeded",
		extractModel: false,
	},
}

// convertAPIError converts Ollama API errors to our error types
func (p *OllamaProvider) convertAPIError(statusCode int, errMsg string) error {
	// Try to parse as structured error first (for future compatibility)
	var structuredErr ollamaErrorResponse
	if err := json.Unmarshal([]byte(errMsg), &structuredErr); err == nil {
		if structuredErr.Type != "" || structuredErr.Code != "" {
			// Ollama may add structured errors in the future
			return p.convertStructuredError(statusCode, &structuredErr)
		}
		// If parsing succeeded but no type/code, use the error message
		errMsg = structuredErr.Error
	}

	// Use pattern matching for robust error detection
	for _, ep := range ollamaErrorPatterns {
		if matches := ep.pattern.FindStringSubmatch(errMsg); matches != nil {
			switch ep.errorType {
			case "model_not_found":
				modelName := ""
				if ep.extractModel && len(matches) > 1 {
					modelName = matches[1]
				}
				return &ModelNotFoundError{
					Provider: "ollama",
					Model:    modelName,
				}
			case "context_length_exceeded":
				return &ContextLengthExceededError{
					Provider: "ollama",
					Message:  errMsg,
				}
			}
		}
	}

	// Use HTTP status code as fallback for error classification
	if statusCode == http.StatusNotFound {
		return &ModelNotFoundError{
			Provider: "ollama",
			Model:    "",
		}
	}

	if statusCode >= 500 {
		return &ProviderUnavailableError{
			Provider: "ollama",
			Status:   statusCode,
			Message:  errMsg,
		}
	}

	// Default to invalid request error
	return &InvalidRequestError{
		Provider: "ollama",
		Message:  errMsg,
	}
}

// convertStructuredError handles potential future structured errors from Ollama
func (p *OllamaProvider) convertStructuredError(statusCode int, apiErr *ollamaErrorResponse) error {
	errType := apiErr.Type
	if errType == "" {
		errType = apiErr.Code
	}

	switch errType {
	case "model_not_found", "not_found_error":
		return &ModelNotFoundError{
			Provider: "ollama",
			Model:    "",
		}
	case "context_length_exceeded":
		return &ContextLengthExceededError{
			Provider: "ollama",
			Message:  apiErr.Error,
		}
	case "invalid_request_error":
		return &InvalidRequestError{
			Provider: "ollama",
			Message:  apiErr.Error,
		}
	default:
		// Fallback to pattern matching on the error message
		return p.convertAPIError(statusCode, apiErr.Error)
	}
}

// shouldRetry checks if an error is retryable
func (p *OllamaProvider) shouldRetry(err error) bool {
	var retryErr *RetryableError
	if errors.As(err, &retryErr) {
		return true
	}
	return false
}

// shouldRetryHTTP checks if an HTTP status code is retryable
func (p *OllamaProvider) shouldRetryHTTP(statusCode int) bool {
	retryableCodes := map[int]bool{
		429: true,
		500: true,
		502: true,
		503: true,
		504: true,
	}
	return retryableCodes[statusCode]
}

// Stream sends a streaming completion request to Ollama
func (p *OllamaProvider) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	model := req.Model
	if model == "" {
		model = "llama3.2"
	}

	ollamaReq := ollamaRequest{
		Model:    model,
		Messages: req.Messages,
		Stream:   true,
	}
	ollamaReq.Options.Temperature = req.Temperature
	ollamaReq.Options.TopP = req.TopP
	ollamaReq.Options.Stop = req.Stop
	if req.MaxTokens > 0 {
		ollamaReq.Options.NumPredict = req.MaxTokens
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check HTTP status before streaming
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		close(ch)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				select {
				case ch <- StreamChunk{Error: ctx.Err(), Finished: true}:
				default:
				}
				return
			default:
			}

			line := scanner.Text()

			if line == "" {
				continue
			}

			var ollamaResp ollamaGenerateResponse
			if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
				if p.debug {
					fmt.Printf("[Ollama] Failed to parse chunk: %v\n", err)
				}
				continue
			}

			chunk := StreamChunk{
				Delta: Message{
					Role:    "assistant",
					Content: ollamaResp.Response,
				},
				Finished: ollamaResp.Done,
			}

			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}

			if ollamaResp.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if p.debug {
				fmt.Printf("[Ollama] Scanner error: %v\n", err)
			}
			select {
			case ch <- StreamChunk{Error: err, Finished: true}:
			case <-ctx.Done():
			}
		}
	}()

	return ch, nil
}

// CountTokens counts tokens using Ollama's tokenizer
func (p *OllamaProvider) CountTokens(text string) int {
	// Approximation: 1 token ≈ 4 characters
	return len(text) / 4
}
