package thaura

import (
	"strings"
)

type ThauraModel string

// This is a simplified list of models that are supported by the Thaura API.
// This should cover any use case we have for now.
const (
	modelThaura ThauraModel = "thaura"
)

func (m ThauraModel) IsValid() bool {
	switch m {
	case modelThaura:
		return true
	default:
		return false
	}
}

func ListValidModelsStr() string {
	var models []string
	for _, model := range []ThauraModel{
		modelThaura,
	} {
		models = append(models, "- "+string(model))
	}
	return strings.Join(models, "\n")
}
