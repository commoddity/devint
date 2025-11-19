// Package chat implements the chat command for interacting with LLM providers.
//
// This file defines the Cobra command, flags, and handles both one-shot and
// interactive chat modes.
package chat

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/commoddity/devint/config"
	llmCfg "github.com/commoddity/devint/config/llm"
	"github.com/commoddity/devint/llm"
	"github.com/commoddity/devint/llm/deepseek"
	"github.com/commoddity/devint/llm/thaura"
)

// LLM config flags
var chatProviderOverride string
var chatModelOverride string

// ChatCmd represents the chat command.
var ChatCmd = &cobra.Command{
	Use:   "chat [prompt]",
	Short: "Chat with the configured LLM provider.",
	Long: `Chat with the configured LLM provider.

This command allows you to interact with your configured LLM provider in two modes:

1. Interactive Mode (no arguments):
   Start a chat session with the LLM that preserves conversation context.
   Type your messages and press Enter to send them. The conversation history
   is maintained throughout the session. Press Ctrl+C to exit.

   Example:
     devint chat

2. One-Shot Mode (with arguments):
   Send a single prompt and receive a response, then exit.

   Example:
     devint chat "tell me about winnie the pooh?"
     devint chat -p thaura "what is the capital of France?"

Flags:
  --provider-override (-p): Optional LLM provider override.
  --model-override (-m)   : Optional LLM model override.`,
	Run: runChatCommand,
}

func init() {
	// Initialize LLM-related flags
	ChatCmd.Flags().StringVarP(&chatProviderOverride, "provider-override", "p", "", "LLM provider override. If set the default provider in the config will be overridden. [OPTIONAL]")
	ChatCmd.Flags().StringVarP(&chatModelOverride, "model-override", "m", "", "LLM model override. If set the default model in the config will be overridden. [OPTIONAL]")
}

// runChatCommand is the main entry point for the chat command.
func runChatCommand(cmd *cobra.Command, args []string) {
	// Load configuration from the config YAML file
	cfg, err := config.LoadConfig()
	if err != nil {
		stdlog.Fatalf("failed to load config: %v", err)
	}

	// Use silent logger for both modes (no slog output)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Get additional provider flags based on any overrides set
	providerFlags := getChatProviderFlags()

	// Initialize the LLM provider with potential provider overrides
	llmProvider, err := llmCfg.NewLLMProvider(logger, cfg.LLMs, providerFlags...)
	if err != nil {
		stdlog.Fatalf("failed to get LLM provider: %v", err)
	}

	if len(args) > 0 {
		// One-shot mode
		prompt := strings.Join(args, " ")
		runOneShotMode(llmProvider, cfg, prompt)
	} else {
		// Interactive mode
		RunInteractiveMode(llmProvider, logger, cfg)
	}
}

// runOneShotMode sends a single prompt and prints the response with a spinner animation.
func runOneShotMode(llmProvider llm.LLMProvider, cfg *config.Config, prompt string) {
	// Get prompt flags based on any model override
	promptFlags := getChatPromptFlags()

	// Determine provider and model info for display
	providerName := string(cfg.LLMs.DefaultLLMProvider)
	if chatProviderOverride != "" {
		providerName = chatProviderOverride
	}

	var providerEmoji string
	var modelName string

	switch providerName {
	case "thaura":
		providerEmoji = "🍉"
		if chatModelOverride != "" {
			modelName = chatModelOverride
		} else if cfg.LLMs.LLMProviders.Thaura != nil {
			modelName = string(cfg.LLMs.LLMProviders.Thaura.ClientModel)
		}
	case "deepseek":
		providerEmoji = "🤖"
		if chatModelOverride != "" {
			modelName = chatModelOverride
		} else if cfg.LLMs.LLMProviders.DeepSeek != nil {
			modelName = string(cfg.LLMs.LLMProviders.DeepSeek.ClientModel)
		}
	case "openrouter":
		providerEmoji = "🌐"
		if chatModelOverride != "" {
			modelName = chatModelOverride
		} else if cfg.LLMs.LLMProviders.OpenRouter != nil {
			modelName = string(cfg.LLMs.LLMProviders.OpenRouter.ClientModel)
		}
	default:
		providerEmoji = "🤖"
		modelName = "unknown"
	}

	// Capitalize first letter of provider name
	displayProvider := strings.ToUpper(providerName[:1]) + providerName[1:]

	// Display provider and model info
	fmt.Printf("💡 Provider: %s %s  🔧 Model: %s\n", providerEmoji, displayProvider, modelName)

	// Check if provider supports streaming for real-time output
	streamingProvider, supportsStreaming := llmProvider.(llm.StreamingLLMProvider)

	if supportsStreaming {
		// Show loading animation while starting
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Prefix = " "
		s.Suffix = " Processing..."
		s.Start()

		// Start streaming in a goroutine
		firstChunk := make(chan bool, 1)
		streamDone := make(chan error, 1)
		go func() {
			err := streamingProvider.StreamPrompt(context.Background(), prompt, func(chunk string) error {
				// Stop spinner on first chunk
				select {
				case firstChunk <- true:
				default:
				}
				fmt.Print(chunk)
				return nil
			}, promptFlags...)
			streamDone <- err
		}()

		// Wait for first chunk
		<-firstChunk
		s.Stop()
		// Clear the spinner line
		fmt.Print("\r\033[K")

		// Wait for streaming to complete
		if err := <-streamDone; err != nil {
			stdlog.Fatalf("❌ Error: %v", err)
		}
		fmt.Println() // Add newline at the end

		// Calculate and display cost if provider supports usage tracking
		calculateAndDisplayCost(llmProvider, providerName, cfg)
	} else {
		// Non-streaming: show spinner while waiting
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Prefix = " "
		s.Suffix = " Processing..."
		s.Start()

		response, err := llmProvider.SendPrompt(context.Background(), prompt, promptFlags...)

		s.Stop()
		// Clear the spinner line
		fmt.Print("\r\033[K")

		if err != nil {
			stdlog.Fatalf("❌ Error: %v", err)
		}

		// Print the response
		fmt.Println(response)

		// Calculate and display cost if provider supports usage tracking
		calculateAndDisplayCost(llmProvider, providerName, cfg)
	}
}

// calculateAndDisplayCost calculates and displays the cost for a prompt if the provider supports usage tracking.
func calculateAndDisplayCost(llmProvider llm.LLMProvider, providerName string, cfg *config.Config) {
	if usageProvider, ok := llmProvider.(llm.UsageTrackingProvider); ok {
		usage, err := usageProvider.GetLastUsage()
		if err == nil {
			model := usageProvider.GetLastModel()
			var cost float64
			var costErr error

			// Calculate cost based on provider type
			switch providerName {
			case "thaura":
				thauraModel := thaura.ThauraModel(model)
				cost, costErr = thaura.CalculateCost(thauraModel, usage)
			case "deepseek":
				deepseekModel := deepseek.DeepSeekModel(model)
				// Try to get cache hit tokens if available
				cacheHitTokens := 0
				if deepseekProvider, ok := llmProvider.(*deepseek.DeepseekProvider); ok {
					cacheHitTokens = deepseekProvider.GetLastCacheHitTokens()
				}
				cost, costErr = deepseek.CalculateCost(deepseekModel, usage, cacheHitTokens)
			}

			if costErr == nil && cost > 0 {
				fmt.Printf("💰 Cost: $%.4f\n", cost)
			}
		}
	}
}

// getChatProviderFlags returns the LLM provider flags for configuration.
func getChatProviderFlags() []llmCfg.ProviderFlag {
	var flags []llmCfg.ProviderFlag

	// Append provider override flag if specified
	if chatProviderOverride != "" {
		flags = append(flags, llmCfg.WithProviderOverride(llmCfg.ProviderType(chatProviderOverride)))
	}

	return flags
}

// getChatPromptFlags returns the LLM prompt flags for model override.
func getChatPromptFlags() []llm.PromptFlag {
	var flags []llm.PromptFlag

	// Append model override flag if provided
	if chatModelOverride != "" {
		flags = append(flags, llm.WithLLMModelOverride(chatModelOverride))
	}

	return flags
}
