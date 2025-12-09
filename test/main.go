package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cwww3/bifrost"
	"github.com/cwww3/bifrost/schemas"
)

type MyAccount struct{}

// Account interface needs to implement these 3 methods
func (a *MyAccount) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
	return []schemas.ModelProvider{schemas.OpenAI}, nil
}

func (a *MyAccount) GetKeysForProvider(ctx *context.Context, provider schemas.ModelProvider) ([]schemas.Key, error) {
	if provider == schemas.OpenAI {
		return []schemas.Key{{
			Value:  os.Getenv("OPENAI_API_KEY"),
			Models: []string{}, // Keep Models empty to use any model
			Weight: 1.0,
		}}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", provider)
}

func (a *MyAccount) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
	c := schemas.NetworkConfig{
		DefaultRequestTimeoutInSeconds: schemas.DefaultRequestTimeoutInSeconds,
		MaxRetries:                     schemas.DefaultMaxRetries,
		RetryBackoffInitial:            schemas.DefaultRetryBackoffInitial,
		RetryBackoffMax:                schemas.DefaultRetryBackoffMax,
	}
	c.BaseURL = "https://www.baidu.com"

	return &schemas.ProviderConfig{
		NetworkConfig:            c,
		ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
	}, nil
}

func main() {
	client, initErr := bifrost.Init(context.Background(), schemas.BifrostConfig{
		Account: &MyAccount{},
	})
	if initErr != nil {
		panic(initErr)
	}
	defer client.Shutdown()

	messages := []schemas.ChatMessage{
		{
			Role: schemas.ChatMessageRoleUser,
			Content: &schemas.ChatMessageContent{
				ContentStr: schemas.Ptr("Hello, Bifrost!"),
			},
		},
	}

	response, err := client.ChatCompletionRequest(context.Background(), &schemas.BifrostChatRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o-mini",
		Input:    messages,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Response:", *response.Choices[0].Message.Content.ContentStr)
}
