package anthropic

import (
	"context"
	"fmt"
	"log/slog"

	anthropicapi "github.com/liushuangls/go-anthropic/v2"

	"github.com/commoddity/devint/llm"
)

var _ llm.LLMProvider = &AnthropicProvider{}

type AnthropicProvider struct {
	logger      *slog.Logger
	client      *anthropicapi.Client
	clientModel AnthropicModel
}

type Config struct {
	Logger      *slog.Logger
	APIKey      string
	ClientModel AnthropicModel
}

func NewAnthropicProvider(cfg Config) *AnthropicProvider {
	return &AnthropicProvider{
		logger:      cfg.Logger,
		client:      anthropicapi.NewClient(cfg.APIKey),
		clientModel: cfg.ClientModel,
	}
}

func (p *AnthropicProvider) SendPrompt(ctx context.Context, prompt string, flags ...llm.PromptFlag) (string, error) {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	anthropicModel := AnthropicModel(cfg.Model)
	if !anthropicModel.IsValid() {
		return "", fmt.Errorf("invalid Anthropic model: %s.\nValid models:\n%s", cfg.Model, ListValidModelsStr())
	}

	req := anthropicapi.MessagesRequest{
		Model:     anthropicapi.Model(anthropicModel),
		MaxTokens: 1000,
		Messages: []anthropicapi.Message{
			anthropicapi.NewUserTextMessage(prompt),
		},
	}

	p.logger.Info("🤖 Sending prompt to Anthropic...", "model", anthropicModel)

	resp, err := p.client.CreateMessages(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no content returned from Anthropic")
	}

	return resp.Content[0].GetText(), nil
}
