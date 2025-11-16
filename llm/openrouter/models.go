package openrouter

type OpenRouterModel string

// IsValid returns true for any non-empty model string.
// OpenRouter supports any model string, so we don't restrict which models can be used.
// See https://openrouter.ai/models for available models.
func (m OpenRouterModel) IsValid() bool {
	return string(m) != ""
}

// ListValidModelsStr returns a message indicating that any model string is valid.
func ListValidModelsStr() string {
	return "Any model string is valid. See https://openrouter.ai/models for available models."
}
