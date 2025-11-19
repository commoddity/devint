package thaura

import (
	"fmt"
	"math"

	"github.com/commoddity/devint/llm"
)

// CalculateCost calculates the cost in dollars for a given model and usage.
// Returns the cost rounded to 4 decimal places (1/100th of a cent precision).
func CalculateCost(model ThauraModel, usage llm.Usage) (float64, error) {
	pricing, err := GetModelPricing(model)
	if err != nil {
		return 0, fmt.Errorf("failed to get pricing for model %s: %w", model, err)
	}

	// Calculate input cost: (promptTokens / 1_000_000) * inputPrice
	inputCost := (float64(usage.PromptTokens) / 1_000_000.0) * pricing.InputPricePerMillion

	// Calculate output cost: (completionTokens / 1_000_000) * outputPrice
	outputCost := (float64(usage.CompletionTokens) / 1_000_000.0) * pricing.OutputPricePerMillion

	// Total cost
	totalCost := inputCost + outputCost

	// Round to 4 decimal places (1/100th of a cent)
	roundedCost := math.Round(totalCost*10000) / 10000

	return roundedCost, nil
}
