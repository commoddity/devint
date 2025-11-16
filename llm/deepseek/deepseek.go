package deepseek

import (
	"context"
	"fmt"
	"log/slog"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"

	"github.com/commoddity/devint/llm"
)

var _ llm.LLMProvider = &DeepseekProvider{}
var _ llm.StreamingLLMProvider = &DeepseekProvider{}

type DeepseekProvider struct {
	logger      *slog.Logger
	client      *deepseek.Client
	clientModel DeepSeekModel
}

type Config struct {
	Logger      *slog.Logger
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

	p.logger.Info("🤖 Sending prompt to DeepSeek...", "model", deepSeekModel)

	resp, err := p.client.CreateChatCompletion(ctx, &req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no content returned from DeepSeek")
	}

	return resp.Choices[0].Message.Content, nil
}

func (p *DeepseekProvider) StreamPrompt(ctx context.Context, prompt string, onChunk func(chunk string) error, flags ...llm.PromptFlag) error {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	deepSeekModel := DeepSeekModel(cfg.Model)
	if !deepSeekModel.IsValid() {
		return fmt.Errorf("invalid DeepSeek model: %s.\nValid models:\n%s", cfg.Model, ListValidModelsStr())
	}

	req := deepseek.StreamChatCompletionRequest{
		Model: string(deepSeekModel),
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    constants.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: true,
	}

	p.logger.Info("🤖 Streaming prompt to DeepSeek...", "model", deepSeekModel)

	stream, err := p.client.CreateChatCompletionStream(ctx, &req)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			// Check if it's EOF (end of stream)
			if err.Error() == "EOF" || err.Error() == "stream closed" {
				break
			}
			return fmt.Errorf("error receiving stream chunk: %w", err)
		}

		if response == nil || len(response.Choices) == 0 {
			continue
		}

		// Extract content from delta
		delta := response.Choices[0].Delta
		if delta.Content != "" {
			if err := onChunk(delta.Content); err != nil {
				return fmt.Errorf("error in chunk callback: %w", err)
			}
		}
	}

	return nil
}
