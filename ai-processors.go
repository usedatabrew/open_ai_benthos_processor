package open_api_benthos_processor

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/sashabaranov/go-openai"
)

type AiProcessor interface {
	Ask(prompt string) (string, error)
}

type OpenAiProcessor struct {
	client *openai.Client
	model  string
}

func NewOpenAIProcessor(apiKey, model string) AiProcessor {
	return &OpenAiProcessor{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (oai *OpenAiProcessor) Ask(prompt string) (string, error) {
	resp, err := oai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: oai.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

type AzureProcessor struct {
	client *azopenai.Client
}

func NewAzureProcessor(apiKey, apiEndpoint string) AiProcessor {
	keyCredential, err := azopenai.NewKeyCredential(apiKey)

	if err != nil {
		// TODO: handle error
	}

	client, err := azopenai.NewClientWithKeyCredential(apiEndpoint, keyCredential, nil)

	if err != nil {
		panic(err)
	}
	return &AzureProcessor{
		client: client,
	}
}

func (az *AzureProcessor) Ask(prompt string) (string, error) {
	resp, err := az.client.GetCompletions(context.TODO(), azopenai.CompletionsOptions{
		Prompt:      []string{prompt},
		MaxTokens:   to.Ptr(int32(2048)),
		Temperature: to.Ptr(float32(0.0)),
	}, nil)

	if err != nil {
		return "", err
	}

	return *resp.Choices[0].Text, nil
}
