package anthropic

import (
	"strings"

	anthropic "github.com/liushuangls/go-anthropic/v2"
)

type AnthropicModel string

// This is a simplified list of models that are supported by the Anthropic API.
// This should cover any use case we have for now.
const (
	modelClaude3Dot5HaikuLatest  AnthropicModel = AnthropicModel(anthropic.ModelClaude3Dot5HaikuLatest)
	modelClaude3Dot5SonnetLatest AnthropicModel = AnthropicModel(anthropic.ModelClaude3Dot5SonnetLatest)
	modelClaude3Opus20240229     AnthropicModel = AnthropicModel(anthropic.ModelClaude3Opus20240229)
)

func (m AnthropicModel) IsValid() bool {
	switch m {
	case modelClaude3Dot5HaikuLatest,
		modelClaude3Dot5SonnetLatest,
		modelClaude3Opus20240229:
		return true
	default:
		return false
	}
}

func ListValidModelsStr() string {
	var models []string
	for _, model := range []AnthropicModel{
		modelClaude3Dot5HaikuLatest,
		modelClaude3Dot5SonnetLatest,
		modelClaude3Opus20240229,
	} {
		models = append(models, "- "+string(model))
	}
	return strings.Join(models, "\n")
}
