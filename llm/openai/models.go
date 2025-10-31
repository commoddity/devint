package openai

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

type OpenAIModel string

// This is a simplified list of models that are supported by the OpenAI API.
// This should cover any use case we have for now.
const (
	modelO1Mini OpenAIModel = OpenAIModel(openai.O1Mini)
	modelO1     OpenAIModel = OpenAIModel(openai.O1)
	modelO3Mini OpenAIModel = OpenAIModel(openai.O3Mini)
	modelGPT4o  OpenAIModel = OpenAIModel(openai.GPT4o)
)

func (m OpenAIModel) IsValid() bool {
	switch m {
	case modelO1Mini,
		modelO1,
		modelO3Mini,
		modelGPT4o:
		return true
	default:
		return false
	}
}

func ListValidModelsStr() string {
	var models []string
	for _, model := range []OpenAIModel{
		modelO1Mini,
		modelO1,
		modelO3Mini,
		modelGPT4o,
	} {
		models = append(models, "- "+string(model))
	}
	return strings.Join(models, "\n")
}
