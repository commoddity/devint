package thaura

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/commoddity/devint/llm"
)

// For more details please see the following documentation:
//   -https://thaura.ai/api-platform
//
// Much love to the folks at Tech for Palestine. 🇵🇸❤️‍🔥🇵🇸
//   - Check them out at: https://techforpalestine.org/

var _ llm.LLMProvider = &ThauraProvider{}
var _ llm.StreamingLLMProvider = &ThauraProvider{}
var _ llm.UsageTrackingProvider = &ThauraProvider{}

type ThauraProvider struct {
	logger      *slog.Logger
	client      *http.Client
	apiKey      string
	clientModel ThauraModel

	// Usage tracking
	lastUsage llm.Usage
	lastModel string
}

type Config struct {
	Logger      *slog.Logger
	APIKey      string
	ClientModel ThauraModel
}

func NewThauraProvider(cfg Config) *ThauraProvider {
	return &ThauraProvider{
		logger:      cfg.Logger,
		client:      &http.Client{},
		apiKey:      cfg.APIKey,
		clientModel: cfg.ClientModel,
	}
}

type chatCompletionRequest struct {
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type usage struct {
	PromptTokens    int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

type choice struct {
	Message message `json:"message"`
}

func (p *ThauraProvider) SendPrompt(ctx context.Context, prompt string, flags ...llm.PromptFlag) (string, error) {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	thauraModel := ThauraModel(cfg.Model)
	if !thauraModel.IsValid() {
		return "", fmt.Errorf("invalid Thaura model: %s.\nValid models:\n%s", cfg.Model, ListValidModelsStr())
	}

	reqBody := chatCompletionRequest{
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://backend.thaura.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	p.logger.Info("🤖 Sending prompt to Thaura...", "model", thauraModel)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no content returned from Thaura")
	}

	// Store usage and model for cost tracking
	p.lastUsage = llm.Usage{
		PromptTokens:    chatResp.Usage.PromptTokens,
		CompletionTokens: chatResp.Usage.CompletionTokens,
		TotalTokens:     chatResp.Usage.TotalTokens,
	}
	p.lastModel = string(thauraModel)

	return chatResp.Choices[0].Message.Content, nil
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage usage `json:"usage"`
}

func (p *ThauraProvider) StreamPrompt(ctx context.Context, prompt string, onChunk func(chunk string) error, flags ...llm.PromptFlag) error {
	cfg := llm.PromptConfig{
		Model: string(p.clientModel),
	}

	for _, flag := range flags {
		flag(&cfg)
	}

	thauraModel := ThauraModel(cfg.Model)
	if !thauraModel.IsValid() {
		return fmt.Errorf("invalid Thaura model: %s.\nValid models:\n%s", cfg.Model, ListValidModelsStr())
	}

	reqBody := chatCompletionRequest{
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://backend.thaura.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	p.logger.Info("🤖 Streaming prompt to Thaura...", "model", thauraModel)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse Server-Sent Events (SSE) format
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check for [DONE] marker
		if strings.Contains(line, "[DONE]") {
			break
		}

		// Parse SSE data lines (format: "data: {...}")
		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")
			jsonStr = strings.TrimSpace(jsonStr)

			// Skip empty data lines
			if jsonStr == "" {
				continue
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
				// Skip malformed JSON chunks
				continue
			}

			// Extract usage from final chunk (usage is typically in the last chunk before [DONE])
			if chunk.Usage.TotalTokens > 0 {
				p.lastUsage = llm.Usage{
					PromptTokens:    chunk.Usage.PromptTokens,
					CompletionTokens: chunk.Usage.CompletionTokens,
					TotalTokens:     chunk.Usage.TotalTokens,
				}
				p.lastModel = string(thauraModel)
			}

			// Extract content from chunk
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				if err := onChunk(chunk.Choices[0].Delta.Content); err != nil {
					return fmt.Errorf("error in chunk callback: %w", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// GetLastUsage returns the usage information from the last prompt sent.
func (p *ThauraProvider) GetLastUsage() (llm.Usage, error) {
	if p.lastUsage.TotalTokens == 0 {
		return llm.Usage{}, fmt.Errorf("no usage information available")
	}
	return p.lastUsage, nil
}

// GetLastModel returns the model identifier used for the last prompt.
func (p *ThauraProvider) GetLastModel() string {
	return p.lastModel
}
