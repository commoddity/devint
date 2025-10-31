package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sashabaranov/go-openai"

	"github.com/commoddity/devint/llm"
)

var _ llm.LLMProvider = &OpenAIProvider{}

type OpenAIProvider struct {
	logger      *slog.Logger
	client      *openai.Client
	clientModel OpenAIModel
}

type Config struct {
	Logger      *slog.Logger
	APIKey      string
	ClientModel OpenAIModel
}

func NewOpenAIProvider(cfg Config) *OpenAIProvider {
	return &OpenAIProvider{
		logger:      cfg.Logger,
		client:      openai.NewClient(cfg.APIKey),
		clientModel: cfg.ClientModel,
	}
}

func (p *OpenAIProvider) SendPrompt(ctx context.Context, prompt string, flags ...llm.PromptFlag) (string, error) {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	openaiModel := OpenAIModel(cfg.Model)
	if !openaiModel.IsValid() {
		return "", fmt.Errorf("invalid OpenAI model: %s.\nValid models:\n%s", cfg.Model, ListValidModelsStr())
	}

	req := openai.ChatCompletionRequest{
		Model: string(openaiModel),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	p.logger.Info("🤖 Sending prompt to OpenAI...", "model", openaiModel)

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
