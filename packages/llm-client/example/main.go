package main

import (
	"context"
	"fmt"
	"os"

	llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
	// Create client
	client := llm.NewClient(llm.Config{
		Provider:   llm.ProviderOpenAI,
		Model:      "gpt-4o",
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		MaxRetries: 3, // Optional: defaults to 3
	})

	// Register OpenAI provider
	llm.RegisterOpenAI(client, llm.OpenAIConfig{
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		MaxRetries: 3, // Optional: defaults to 3
		Debug:      true,
	})

	// Simple completion
	fmt.Println("=== Simple Completion ===")
	response, err := client.Simple(context.Background(), "Say hello in 3 words")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Response: %s\n\n", response)

	// Advanced completion with full control
	fmt.Println("=== Advanced Completion ===")
	resp, err := client.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful coding assistant"},
			{Role: "user", Content: "Write a hello world in Go"},
		},
		Model:       "gpt-4o",
		Temperature: 0.7,
		MaxTokens:   500,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Tokens: %d prompt + %d completion = %d total\n",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens)
	fmt.Printf("Response:\n%s\n\n", resp.Content)

	// Streaming completion
	fmt.Println("=== Streaming Completion ===")
	stream, err := client.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "user", Content: "Count from 1 to 5 slowly"},
		},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Print("Stream: ")
	for chunk := range stream {
		fmt.Print(chunk.Delta.Content)
		if chunk.Finished {
			break
		}
	}
	fmt.Println("\n\nDone!")
}
