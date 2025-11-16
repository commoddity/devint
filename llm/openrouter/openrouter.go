package openrouter

import (
	"context"
	"fmt"
	"log/slog"

	openrouter "github.com/revrost/go-openrouter"

	"github.com/commoddity/devint/llm"
)

var _ llm.LLMProvider = &OpenRouterProvider{}
var _ llm.StreamingLLMProvider = &OpenRouterProvider{}

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
				Content: openrouter.Content{Text: prompt},
			},
		},
	}

	p.logger.Info("🤖 Sending prompt to OpenRouter...", "model", openrouterModel)

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content.Text, nil
}

func (p *OpenRouterProvider) StreamPrompt(ctx context.Context, prompt string, onChunk func(chunk string) error, flags ...llm.PromptFlag) error {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	openrouterModel := OpenRouterModel(cfg.Model)
	if !openrouterModel.IsValid() {
		return fmt.Errorf("invalid OpenRouter model: model string cannot be empty")
	}

	req := openrouter.ChatCompletionRequest{
		Model: string(openrouterModel),
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{Text: prompt},
			},
		},
		Stream: true,
	}

	p.logger.Info("🤖 Streaming prompt to OpenRouter...", "model", openrouterModel)

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			// Check if it's EOF (end of stream) - OpenRouter may return different EOF representations
			errStr := err.Error()
			if errStr == "EOF" || errStr == "stream closed" || errStr == "io: read/write on closed pipe" {
				break
			}
			return fmt.Errorf("error receiving stream chunk: %w", err)
		}

		if len(response.Choices) == 0 {
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
