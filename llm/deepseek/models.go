package deepseek

import (
	"fmt"
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

// ModelPricing holds pricing information for a DeepSeek model.
// Prices are in dollars per million tokens.
type ModelPricing struct {
	InputPricePerMillionCacheHit  float64 // Price per million input tokens (cache hit)
	InputPricePerMillionCacheMiss float64 // Price per million input tokens (cache miss)
	OutputPricePerMillion         float64 // Price per million output tokens
}

var (
	// pricingMap maps DeepSeekModel to its pricing information
	pricingMap = map[DeepSeekModel]ModelPricing{
		modelDeepSeekChat: {
			InputPricePerMillionCacheHit:  0.028, // $0.028 per million input tokens (cache hit)
			InputPricePerMillionCacheMiss: 0.28,  // $0.28 per million input tokens (cache miss)
			OutputPricePerMillion:         0.42,  // $0.42 per million output tokens
		},
		modelDeepSeekReasoner: {
			InputPricePerMillionCacheHit:  0.028, // $0.028 per million input tokens (cache hit)
			InputPricePerMillionCacheMiss: 0.28,  // $0.28 per million input tokens (cache miss)
			OutputPricePerMillion:         0.42,  // $0.42 per million output tokens
		},
		// Note: deepseek-coder pricing not specified, using same as chat/reasoner for now
		modelDeepSeekCoder: {
			InputPricePerMillionCacheHit:  0.028, // $0.028 per million input tokens (cache hit)
			InputPricePerMillionCacheMiss: 0.28,  // $0.28 per million input tokens (cache miss)
			OutputPricePerMillion:         0.42,  // $0.42 per million output tokens
		},
	}
)

// GetModelPricing returns the pricing information for a specific DeepSeek model.
// Returns an error if the model is invalid or pricing is not available.
func GetModelPricing(model DeepSeekModel) (ModelPricing, error) {
	if !model.IsValid() {
		return ModelPricing{}, fmt.Errorf("invalid DeepSeek model: %s", model)
	}

	pricing, exists := pricingMap[model]
	if !exists {
		return ModelPricing{}, fmt.Errorf("pricing not available for model: %s", model)
	}

	return pricing, nil
}
