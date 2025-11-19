package deepseek

import (
	"fmt"
	"math"

	"github.com/commoddity/devint/llm"
)

// CalculateCost calculates the cost in dollars for a given model and usage.
// Returns the cost rounded to 4 decimal places (1/100th of a cent precision).
// cacheHitTokens is the number of input tokens that were cache hits.
func CalculateCost(model DeepSeekModel, usage llm.Usage, cacheHitTokens int) (float64, error) {
	pricing, err := GetModelPricing(model)
	if err != nil {
		return 0, fmt.Errorf("failed to get pricing for model %s: %w", model, err)
	}

	// Calculate cache hit cost: (cacheHitTokens / 1_000_000) * inputPriceCacheHit
	cacheHitCost := (float64(cacheHitTokens) / 1_000_000.0) * pricing.InputPricePerMillionCacheHit

	// Calculate cache miss cost: ((promptTokens - cacheHitTokens) / 1_000_000) * inputPriceCacheMiss
	cacheMissTokens := usage.PromptTokens - cacheHitTokens
	if cacheMissTokens < 0 {
		cacheMissTokens = 0 // Safety check
	}
	cacheMissCost := (float64(cacheMissTokens) / 1_000_000.0) * pricing.InputPricePerMillionCacheMiss

	// Calculate output cost: (completionTokens / 1_000_000) * outputPrice
	outputCost := (float64(usage.CompletionTokens) / 1_000_000.0) * pricing.OutputPricePerMillion

	// Total cost
	totalCost := cacheHitCost + cacheMissCost + outputCost

	// Round to 4 decimal places (1/100th of a cent)
	roundedCost := math.Round(totalCost*10000) / 10000

	return roundedCost, nil
}

