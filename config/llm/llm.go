package llm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/commoddity/devint/llm/deepseek"
	"github.com/commoddity/devint/llm/openrouter"
	"github.com/commoddity/devint/llm/thaura"
)

var (
	errLLMConfigNotFound         = errors.New("LLM config error: LLM config not found")
	errInvalidDefaultLLMProvider = errors.New("LLM config error: invalid default LLM provider: %s.\nValid providers:\n%s")

	errDeepSeekConfigNotConfigured = errors.New("LLM config error: DeepSeek is not configured")
	errDeepSeekAPIKeyNotConfigured = errors.New("LLM config error: DeepSeek API key is not configured")
	errInvalidDeepSeekClientModel  = errors.New("LLM config error: invalid DeepSeek client model: %s.\nValid models:\n%s")

	errOpenRouterConfigNotConfigured = errors.New("LLM config error: OpenRouter is not configured")
	errOpenRouterAPIKeyNotConfigured = errors.New("LLM config error: OpenRouter API key is not configured")
	errInvalidOpenRouterClientModel  = errors.New("LLM config error: invalid OpenRouter client model: %s.\nValid models:\n%s")

	errThauraConfigNotConfigured = errors.New("LLM config error: Thaura is not configured")
	errThauraAPIKeyNotConfigured = errors.New("LLM config error: Thaura API key is not configured")
	errInvalidThauraClientModel  = errors.New("LLM config error: invalid Thaura client model: %s.\nValid models:\n%s")
)

type ProviderType string

const (
	ProviderNameDeepSeek   ProviderType = "deepseek"
	ProviderNameOpenRouter ProviderType = "openrouter"
	ProviderNameThaura     ProviderType = "thaura"
)

func (t ProviderType) IsValid() bool {
	switch t {
	case ProviderNameDeepSeek, ProviderNameOpenRouter, ProviderNameThaura:
		return true
	default:
		return false
	}
}

func validProvidersStr() string {
	var providers []string
	for _, provider := range []ProviderType{
		ProviderNameDeepSeek,
		ProviderNameOpenRouter,
		ProviderNameThaura,
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
		DeepSeek   *DeepSeekConfig   `yaml:"deepseek"`
		OpenRouter *OpenRouterConfig `yaml:"openrouter"`
		Thaura     *ThauraConfig     `yaml:"thaura"`
	}
	// DeepSeekConfig represents the configuration for the DeepSeek provider.
	DeepSeekConfig struct {
		APIKey      string                 `yaml:"api_key"`
		ClientModel deepseek.DeepSeekModel `yaml:"client_model"`
	}
	// OpenRouterConfig represents the configuration for the OpenRouter provider.
	OpenRouterConfig struct {
		APIKey      string                     `yaml:"api_key"`
		ClientModel openrouter.OpenRouterModel `yaml:"client_model"`
	}
	// ThauraConfig represents the configuration for the Thaura provider.
	ThauraConfig struct {
		APIKey      string             `yaml:"api_key"`
		ClientModel thaura.ThauraModel `yaml:"client_model"`
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

	case ProviderNameThaura:
		if c.LLMProviders.Thaura == nil {
			return errThauraConfigNotConfigured
		}
		if c.LLMProviders.Thaura.APIKey == "" {
			return errThauraAPIKeyNotConfigured
		}
		if !c.LLMProviders.Thaura.ClientModel.IsValid() {
			return fmt.Errorf(
				errInvalidThauraClientModel.Error(), c.LLMProviders.Thaura.ClientModel, thaura.ListValidModelsStr(),
			)
		}
	}

	return nil
}
