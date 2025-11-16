package thaura

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/commoddity/devint/llm"
)

// For more details please see the following documentation:
//   -https://thaura.ai/api-platform
//
// Much love to the folks at Tech for Palestine. 🇵🇸❤️‍🔥🇵🇸
//   - Check them out at: https://techforpalestine.org/

var _ llm.LLMProvider = &ThauraProvider{}

type ThauraProvider struct {
	logger      *slog.Logger
	client      *http.Client
	apiKey      string
	clientModel ThauraModel
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

	return chatResp.Choices[0].Message.Content, nil
}
