package llm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/commoddity/devint/llm/anthropic"
	"github.com/commoddity/devint/llm/deepseek"
	"github.com/commoddity/devint/llm/openai"
	"github.com/commoddity/devint/llm/openrouter"
)

var (
	errLLMConfigNotFound         = errors.New("LLM config error: LLM config not found")
	errInvalidDefaultLLMProvider = errors.New("LLM config error: invalid default LLM provider: %s.\nValid providers:\n%s")

	errOpenAIConfigNotConfigured = errors.New("LLM config error: OpenAI is not configured")
	errOpenAIAPIKeyNotConfigured = errors.New("LLM config error: OpenAI API key is not configured")
	errInvalidOpenAIClientModel  = errors.New("LLM config error: invalid OpenAI client model: %s.\nValid models:\n%s")

	errDeepSeekConfigNotConfigured = errors.New("LLM config error: DeepSeek is not configured")
	errDeepSeekAPIKeyNotConfigured = errors.New("LLM config error: DeepSeek API key is not configured")
	errInvalidDeepSeekClientModel  = errors.New("LLM config error: invalid DeepSeek client model: %s.\nValid models:\n%s")

	errAnthropicConfigNotConfigured = errors.New("LLM config error: Anthropic is not configured")
	errAnthropicAPIKeyNotConfigured = errors.New("LLM config error: Anthropic API key is not configured")
	errInvalidAnthropicClientModel  = errors.New("LLM config error: invalid Anthropic client model: %s.\nValid models:\n%s")

	errOpenRouterConfigNotConfigured = errors.New("LLM config error: OpenRouter is not configured")
	errOpenRouterAPIKeyNotConfigured = errors.New("LLM config error: OpenRouter API key is not configured")
	errInvalidOpenRouterClientModel  = errors.New("LLM config error: invalid OpenRouter client model: %s.\nValid models:\n%s")
)

type ProviderType string

const (
	ProviderNameOpenAI     ProviderType = "openai"
	ProviderNameDeepSeek   ProviderType = "deepseek"
	ProviderNameAnthropic  ProviderType = "anthropic"
	ProviderNameOpenRouter ProviderType = "openrouter"
)

func (t ProviderType) IsValid() bool {
	switch t {
	case ProviderNameOpenAI, ProviderNameDeepSeek, ProviderNameAnthropic, ProviderNameOpenRouter:
		return true
	default:
		return false
	}
}

func validProvidersStr() string {
	var providers []string
	for _, provider := range []ProviderType{
		ProviderNameOpenAI,
		ProviderNameDeepSeek,
		ProviderNameAnthropic,
		ProviderNameOpenRouter,
	} {
		providers = append(providers, string(provider))
	}
	return strings.Join(providers, "\n")
}

// LLMConfig represents the configuration for LLMs.
type (
	Config struct {
		DefaultLLMProvider ProviderType    `yaml:"default_llm_provider"`
		LLMProviders       ProvidersConfig `yaml:"llm_providers"`
	}
	// LLMProvidersConfig represents the configuration for all LLM providers.
	ProvidersConfig struct {
		OpenAI     *OpenAIConfig     `yaml:"openai"`
		DeepSeek   *DeepSeekConfig   `yaml:"deepseek"`
		Anthropic  *AnthropicConfig  `yaml:"anthropic"`
		OpenRouter *OpenRouterConfig `yaml:"openrouter"`
	}
	// OpenAIConfig represents the configuration for the OpenAI provider.
	OpenAIConfig struct {
		APIKey      string             `yaml:"api_key"`
		ClientModel openai.OpenAIModel `yaml:"client_model"`
	}
	// DeepSeekConfig represents the configuration for the DeepSeek provider.
	DeepSeekConfig struct {
		APIKey      string                 `yaml:"api_key"`
		ClientModel deepseek.DeepSeekModel `yaml:"client_model"`
	}
	// AnthropicConfig represents the configuration for the Anthropic provider.
	AnthropicConfig struct {
		APIKey      string                   `yaml:"api_key"`
		ClientModel anthropic.AnthropicModel `yaml:"client_model"`
	}
	// OpenRouterConfig represents the configuration for the OpenRouter provider.
	OpenRouterConfig struct {
		APIKey      string                     `yaml:"api_key"`
		ClientModel openrouter.OpenRouterModel `yaml:"client_model"`
	}
)

// Validate checks if the LLMConfig is valid.
func (c *Config) Validate() error {
	// Validate the LLMConfig is not nil.
	if c == nil {
		return errLLMConfigNotFound
	}

	// Validate the default LLM provider.
	if !c.DefaultLLMProvider.IsValid() {
		return fmt.Errorf(errInvalidDefaultLLMProvider.Error(), c.DefaultLLMProvider, validProvidersStr())
	}

	// Validate the default LLM provider's config.
	// The default LLM provider may have been overridden by a flag.
	switch c.DefaultLLMProvider {

	case ProviderNameOpenAI:
		if c.LLMProviders.OpenAI == nil {
			return errOpenAIConfigNotConfigured
		}
		if c.LLMProviders.OpenAI.APIKey == "" {
			return errOpenAIAPIKeyNotConfigured
		}
		if !c.LLMProviders.OpenAI.ClientModel.IsValid() {
			return fmt.Errorf(
				errInvalidOpenAIClientModel.Error(), c.LLMProviders.OpenAI.ClientModel, openai.ListValidModelsStr(),
			)
		}

	case ProviderNameDeepSeek:
		if c.LLMProviders.DeepSeek == nil {
			return errDeepSeekConfigNotConfigured
		}
		if c.LLMProviders.DeepSeek.APIKey == "" {
			return errDeepSeekAPIKeyNotConfigured
		}
		if !c.LLMProviders.DeepSeek.ClientModel.IsValid() {
			return fmt.Errorf(
				errInvalidDeepSeekClientModel.Error(), c.LLMProviders.DeepSeek.ClientModel, deepseek.ListValidModelsStr(),
			)
		}

	case ProviderNameAnthropic:
		if c.LLMProviders.Anthropic == nil {
			return errAnthropicConfigNotConfigured
		}
		if c.LLMProviders.Anthropic.APIKey == "" {
			return errAnthropicAPIKeyNotConfigured
		}
		if !c.LLMProviders.Anthropic.ClientModel.IsValid() {
			return fmt.Errorf(
				errInvalidAnthropicClientModel.Error(), c.LLMProviders.Anthropic.ClientModel, anthropic.ListValidModelsStr(),
			)
		}

	case ProviderNameOpenRouter:
		if c.LLMProviders.OpenRouter == nil {
			return errOpenRouterConfigNotConfigured
		}
		if c.LLMProviders.OpenRouter.APIKey == "" {
			return errOpenRouterAPIKeyNotConfigured
		}
		if !c.LLMProviders.OpenRouter.ClientModel.IsValid() {
			return fmt.Errorf(
				errInvalidOpenRouterClientModel.Error(), c.LLMProviders.OpenRouter.ClientModel, openrouter.ListValidModelsStr(),
			)
		}
	}

	return nil
}
