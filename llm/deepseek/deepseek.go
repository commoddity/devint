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
var _ llm.UsageTrackingProvider = &DeepseekProvider{}

type DeepseekProvider struct {
	logger      *slog.Logger
	client      *deepseek.Client
	clientModel DeepSeekModel

	// Usage tracking
	lastUsage          llm.Usage
	lastModel          string
	lastCacheHitTokens int
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

	// Store usage and model for cost tracking
	if resp.Usage.TotalTokens > 0 {
		p.lastUsage = llm.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
		// Extract cache hit tokens from usage if available
		if resp.Usage.PromptCacheHitTokens > 0 {
			p.lastCacheHitTokens = resp.Usage.PromptCacheHitTokens
		} else {
			p.lastCacheHitTokens = 0
		}
	}
	p.lastModel = string(deepSeekModel)

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

		if response == nil {
			continue
		}

		// Extract usage from final chunk (usage is typically in the last chunk before EOF)
		if response.Usage != nil && response.Usage.TotalTokens > 0 {
			p.lastUsage = llm.Usage{
				PromptTokens:     response.Usage.PromptTokens,
				CompletionTokens: response.Usage.CompletionTokens,
				TotalTokens:      response.Usage.TotalTokens,
			}
			// Extract cache hit tokens from usage if available
			if response.Usage.PromptCacheHitTokens > 0 {
				p.lastCacheHitTokens = response.Usage.PromptCacheHitTokens
			} else {
				p.lastCacheHitTokens = 0
			}
			p.lastModel = string(deepSeekModel)
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

// GetLastUsage returns the usage information from the last prompt sent.
func (p *DeepseekProvider) GetLastUsage() (llm.Usage, error) {
	if p.lastUsage.TotalTokens == 0 {
		return llm.Usage{}, fmt.Errorf("no usage information available")
	}
	return p.lastUsage, nil
}

// GetLastModel returns the model identifier used for the last prompt.
func (p *DeepseekProvider) GetLastModel() string {
	return p.lastModel
}

// GetLastCacheHitTokens returns the number of cache hit tokens from the last prompt.
func (p *DeepseekProvider) GetLastCacheHitTokens() int {
	return p.lastCacheHitTokens
}
