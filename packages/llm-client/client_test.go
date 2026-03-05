package llm

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient(Config{
		Provider: ProviderOpenAI,
		Model:    "gpt-4o",
		APIKey:   "test-key",
	})

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.config.Provider != ProviderOpenAI {
		t.Errorf("Expected provider %s, got %s", ProviderOpenAI, client.config.Provider)
	}
}

func TestRegisterOpenAI(t *testing.T) {
	client := NewClient(Config{
		Provider: ProviderOpenAI,
		Model:    "gpt-4o",
		APIKey:   "test-key",
	})

	// Register OpenAI provider
	RegisterOpenAI(client, OpenAIConfig{
		APIKey: "test-key",
	})

	// Verify provider is registered
	if _, ok := client.providers[ProviderOpenAI]; !ok {
		t.Error("Expected OpenAI provider to be registered")
	}
}

func TestSimpleWithoutProvider(t *testing.T) {
	client := NewClient(Config{
		Provider: ProviderOpenAI,
		Model:    "gpt-4o",
		APIKey:   "test-key",
	})

	// Should fail because provider not registered
	_, err := client.Simple(context.Background(), "Hello")
	if err == nil {
		t.Error("Expected error when provider not registered")
	}
}
