package openrouter

import (
	"context"
	"fmt"
	"log/slog"

	openrouter "github.com/revrost/go-openrouter"

	"github.com/commoddity/devint/llm"
)

var _ llm.LLMProvider = &OpenRouterProvider{}

type OpenRouterProvider struct {
	logger      *slog.Logger
	client      *openrouter.Client
	clientModel OpenRouterModel
}

type Config struct {
	Logger      *slog.Logger
	APIKey      string
	ClientModel OpenRouterModel
}

func NewOpenRouterProvider(cfg Config) *OpenRouterProvider {
	return &OpenRouterProvider{
		logger: cfg.Logger,
		client: openrouter.NewClient(
			cfg.APIKey,
		),
		clientModel: cfg.ClientModel,
	}
}

func (p *OpenRouterProvider) SendPrompt(ctx context.Context, prompt string, flags ...llm.PromptFlag) (string, error) {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	openrouterModel := OpenRouterModel(cfg.Model)
	if !openrouterModel.IsValid() {
		return "", fmt.Errorf("invalid OpenRouter model: model string cannot be empty")
	}

	req := openrouter.ChatCompletionRequest{
		Model: string(openrouterModel),
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	p.logger.Info("🤖 Sending prompt to OpenRouter...", "model", openrouterModel)

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
