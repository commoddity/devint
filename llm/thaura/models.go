package thaura

import (
	"fmt"
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

// ModelPricing holds pricing information for a Thaura model.
// Prices are in dollars per million tokens.
type ModelPricing struct {
	InputPricePerMillion  float64 // Price per million input tokens
	OutputPricePerMillion float64 // Price per million output tokens
}

var (
	// pricingMap maps ThauraModel to its pricing information
	pricingMap = map[ThauraModel]ModelPricing{
		modelThaura: {
			InputPricePerMillion:  0.50, // $0.50 per million input tokens
			OutputPricePerMillion: 2.00, // $2.00 per million output tokens
		},
	}
)

// GetModelPricing returns the pricing information for a specific Thaura model.
// Returns an error if the model is invalid or pricing is not available.
func GetModelPricing(model ThauraModel) (ModelPricing, error) {
	if !model.IsValid() {
		return ModelPricing{}, fmt.Errorf("invalid Thaura model: %s", model)
	}

	pricing, exists := pricingMap[model]
	if !exists {
		return ModelPricing{}, fmt.Errorf("pricing not available for model: %s", model)
	}

	return pricing, nil
}
