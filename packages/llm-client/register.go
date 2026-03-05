package llm

// RegisterOpenAI registers the OpenAI provider with the client
func RegisterOpenAI(client *Client, config OpenAIConfig) {
	client.SetProvider(ProviderOpenAI, NewOpenAIProvider(config))
}
