// ---------------------------------------------------------------------------
// File: llm.go
// Package: llm
//
// Purpose:
//
//	This file defines the core interfaces and types for interacting with Language
//	Model (LLM) providers. It sets the contract for sending prompts to an LLM and
//	receiving responses. Additionally, it provides configuration types and utility
//	functions (via functional flags) to modify LLM prompt configurations, such as
//	overriding the default model.
//
// Features:
//   - LLMProvider interface: Defines a method to send prompts with a configurable setup.
//   - PromptConfig struct: Holds configuration data for LLM prompts, such as the model name.
//   - PromptFlag type: Implements the functional options pattern to modify PromptConfig.
//   - WithLLMModelOverride: A PromptFlag that allows overriding the default model.
//
// Usage:
//
//	Implement the LLMProvider interface for any custom LLM provider integration. Use
//	PromptFlag options to adjust or override prompt configuration values when sending
//	prompts to the LLM.
//
// ---------------------------------------------------------------------------
package llm

import (
	"context" // Standard library context package for managing request-scoped values and cancellation.
)

// LLMProvider is the interface that any Language Model provider must implement.
// It defines the method for sending a prompt and receiving a string response along with any errors.
type LLMProvider interface {
	// SendPrompt sends the given prompt to the LLM provider along with optional configuration
	// flags and returns the provider's response or an error.
	SendPrompt(ctx context.Context, prompt string, flags ...PromptFlag) (string, error)
}

// PromptConfig holds configuration details for an LLM prompt.
// The Model field specifies which LLM model should be used to process the prompt.
type PromptConfig struct {
	Model string // The LLM model identifier.
}

// PromptFlag is a functional option type that allows modification of the PromptConfig.
type PromptFlag func(cfg *PromptConfig)

// WithLLMModelOverride is a PromptFlag that overrides the default LLM model in the PromptConfig.
// By applying this flag, the user can specify a different model when sending a prompt.
func WithLLMModelOverride(model string) PromptFlag {
	return func(cfg *PromptConfig) {
		cfg.Model = model
	}
}
