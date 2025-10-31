package deepseek

import (
	"context"
	"fmt"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/commoddity/devint/llm"
)

var _ llm.LLMProvider = &DeepseekProvider{}

type DeepseekProvider struct {
	logger      polylog.Logger
	client      *deepseek.Client
	clientModel DeepSeekModel
}

type Config struct {
	Logger      polylog.Logger
	APIKey      string
	ClientModel DeepSeekModel
}

func NewDeepseekProvider(cfg Config) *DeepseekProvider {
	return &DeepseekProvider{
		logger:      cfg.Logger,
		client:      deepseek.NewClient(cfg.APIKey),
		clientModel: cfg.ClientModel,
	}
}

func (p *DeepseekProvider) SendPrompt(ctx context.Context, prompt string, flags ...llm.PromptFlag) (string, error) {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	deepSeekModel := DeepSeekModel(cfg.Model)
	if !deepSeekModel.IsValid() {
		return "", fmt.Errorf("invalid DeepSeek model: %s.\nValid models:\n%s", cfg.Model, ListValidModelsStr())
	}

	req := deepseek.ChatCompletionRequest{
		Model: string(deepSeekModel),
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    constants.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	p.logger.With("model", deepSeekModel).Info().Msg("Sending prompt to DeepSeek ...")

	resp, err := p.client.CreateChatCompletion(ctx, &req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no content returned from DeepSeek")
	}

	return resp.Choices[0].Message.Content, nil
}
