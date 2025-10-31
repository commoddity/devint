package deepseek

import (
	"strings"

	deepseek "github.com/cohesion-org/deepseek-go"
)

type DeepSeekModel string

// This is a simplified list of models that are supported by the DeepSeek API.
// This should cover any use case we have for now.
const (
	modelDeepSeekChat     DeepSeekModel = DeepSeekModel(deepseek.DeepSeekChat)
	modelDeepSeekCoder    DeepSeekModel = DeepSeekModel(deepseek.DeepSeekCoder)
	modelDeepSeekReasoner DeepSeekModel = DeepSeekModel(deepseek.DeepSeekReasoner)
)

func (m DeepSeekModel) IsValid() bool {
	switch m {
	case modelDeepSeekChat,
		modelDeepSeekCoder,
		modelDeepSeekReasoner:
		return true
	default:
		return false
	}
}

func ListValidModelsStr() string {
	var models []string
	for _, model := range []DeepSeekModel{
		modelDeepSeekChat,
		modelDeepSeekCoder,
		modelDeepSeekReasoner,
	} {
		models = append(models, "- "+string(model))
	}
	return strings.Join(models, "\n")
}
