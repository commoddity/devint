// ---------------------------------------------------------------------------
// File: chat.go
// Package: cmd
//
// Purpose:
//   This command allows users to send prompts to the configured LLM provider
//   and receive responses. It supports both one-shot mode (with arguments) and
//   interactive mode (without arguments) with conversation context preservation.
//
// Features:
//   - Interactive mode: Chat with the LLM while preserving conversation context
//   - One-shot mode: Send a single prompt and receive a response
//   - Supports overriding the default LLM provider and model using flags
//   - Colorized output for better readability
//   - Exit interactive mode with Ctrl+C
//   - Uses structured JSON logging using slog
//
// ---------------------------------------------------------------------------

package cmd

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"

	"github.com/commoddity/devint/config"
	llmCfg "github.com/commoddity/devint/config/llm"
	"github.com/commoddity/devint/llm"
	"github.com/commoddity/devint/log"
)

// LLM config flags
var chatProviderOverride string
var chatModelOverride string

func init() {
	// Initialize LLM-related flags.
	chatCmd.Flags().StringVarP(&chatProviderOverride, "provider-override", "p", "", "LLM provider override. If set the default provider in the config will be overridden. [OPTIONAL]")
	chatCmd.Flags().StringVarP(&chatModelOverride, "model-override", "m", "", "LLM model override. If set the default model in the config will be overridden. [OPTIONAL]")
}

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
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
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration from the config YAML file.
		cfg, err := config.LoadConfig()
		if err != nil {
			stdlog.Fatalf("failed to load config: %v", err)
		}

		// Determine mode based on whether arguments were provided
		var logger *slog.Logger
		if len(args) > 0 {
			// One-shot mode - use regular logger
			logger = log.NewJSONLogger()
		} else {
			// Interactive mode - use silent logger for cleaner UI
			logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		}

		// Get additional provider flags based on any overrides set.
		providerFlags := getChatProviderFlags()
		// Initialize the LLM provider with potential provider overrides.
		llmProvider, err := llmCfg.NewLLMProvider(logger, cfg.LLMs, providerFlags...)
		if err != nil {
			stdlog.Fatalf("failed to get LLM provider: %v", err)
		}

		if len(args) > 0 {
			// One-shot mode - show logs
			logger.Info("✅ Initialization successful.")
			prompt := strings.Join(args, " ")
			runOneShotMode(llmProvider, logger, prompt)
		} else {
			// Interactive mode
			runInteractiveMode(llmProvider, logger, cfg)
		}
	},
}

// runOneShotMode sends a single prompt and prints the response
func runOneShotMode(llmProvider llm.LLMProvider, logger *slog.Logger, prompt string) {
	logger = logger.With(
		"prompt_length", len(prompt),
		"mode", "one-shot",
	)

	logger.Info("Sending prompt to LLM...")

	// Get prompt flags based on any model override.
	promptFlags := getChatPromptFlags()
	// Send the prompt to the LLM provider.
	response, err := llmProvider.SendPrompt(context.Background(), prompt, promptFlags...)
	if err != nil {
		stdlog.Fatalf("failed to send prompt: %v", err)
	}

	// Print the response.
	fmt.Println(response)
	logger.Info("✅ Response received successfully.")
}

// runInteractiveMode starts an interactive chat session with conversation context
func runInteractiveMode(llmProvider llm.LLMProvider, logger *slog.Logger, cfg *config.Config) {
	// Set up signal handler for graceful exit on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Conversation history to maintain context
	var conversationHistory []string

	// Determine provider and model info
	providerName := string(cfg.LLMs.DefaultLLMProvider)
	var providerEmoji string
	var modelName string

	// Apply any overrides
	if chatProviderOverride != "" {
		providerName = chatProviderOverride
	}

	switch providerName {
	case "thaura":
		providerEmoji = "🍉" // Watermelon emoji - symbol of Palestinian solidarity
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

	// Print welcome message
	fmt.Println()
	fmt.Printf("%s╔══════════════════════════════════════════════════════════════╗%s\n", log.Cyan, log.ResetColor)
	fmt.Printf("%s║          🤖 Interactive Chat Mode - LLM Assistant            ║%s\n", log.Cyan, log.ResetColor)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════╝%s\n", log.Cyan, log.ResetColor)
	fmt.Println()
	fmt.Printf("%s💡 Provider:%s %s %s%s%s\n", log.Yellow, log.ResetColor, providerEmoji, log.Cyan+log.Bold, providerName, log.ResetColor)
	fmt.Printf("%s🔧 Model:%s    %s%s%s\n", log.Yellow, log.ResetColor, log.Green+log.Bold, modelName, log.ResetColor)
	fmt.Println()
	fmt.Printf("%sType your messages and press Enter to send.%s\n", log.White, log.ResetColor)
	fmt.Printf("%sPress Ctrl+C to exit at any time.%s\n", log.White, log.ResetColor)
	fmt.Println()

	// Create readline instance with custom prompt
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              fmt.Sprintf("%s└─▶%s ", log.Green, log.ResetColor),
		HistoryFile:         "/tmp/devint-chat-history.tmp",
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: nil,
	})
	if err != nil {
		stdlog.Fatalf("failed to create readline: %v", err)
	}
	defer rl.Close()

	// Handle Ctrl+C gracefully
	go func() {
		<-sigChan
		fmt.Println()
		fmt.Printf("\n%s👋 Goodbye! Exiting chat session...%s\n", log.Yellow, log.ResetColor)
		rl.Close()
		os.Exit(0)
	}()

	for {
		// Show the [You] header before each prompt
		fmt.Printf("%s┌─[You]%s\n", log.Green, log.ResetColor)

		// Read line with full terminal support
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println()
				fmt.Printf("\n%s👋 Goodbye! Exiting chat session...%s\n", log.Yellow, log.ResetColor)
				return
			}
			stdlog.Fatalf("error reading input: %v", err)
			return
		}

		userInput := strings.TrimSpace(line)

		// Skip empty input
		if userInput == "" {
			continue
		}

		// Add user message to conversation history
		conversationHistory = append(conversationHistory, fmt.Sprintf("User: %s", userInput))

		// Build the full prompt with conversation context
		fullPrompt := buildConversationPrompt(conversationHistory)

		// Show animated thinking indicator with spinner
		fmt.Println()
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Prefix = fmt.Sprintf("%s", log.Blue)
		s.Suffix = fmt.Sprintf(" Thinking...%s", log.ResetColor)
		s.Color("cyan", "bold")
		s.Start()

		// Get prompt flags based on any model override.
		promptFlags := getChatPromptFlags()

		// Send the prompt to the LLM provider.
		response, err := llmProvider.SendPrompt(context.Background(), fullPrompt, promptFlags...)

		// Stop the spinner
		s.Stop()

		if err != nil {
			fmt.Printf("%s❌ Error: %v%s\n\n", log.Red, err, log.ResetColor)
			continue
		}

		// Add assistant response to conversation history
		conversationHistory = append(conversationHistory, fmt.Sprintf("Assistant: %s", response))

		// Print the response with formatting
		fmt.Println()
		fmt.Printf("%s┌─[Assistant]%s\n", log.Blue, log.ResetColor)
		fmt.Printf("%s└─▶%s\n", log.Blue, log.ResetColor)

		// Print response preserving formatting (paragraphs, lists, etc.)
		printFormattedResponse(response)
		fmt.Println()
		fmt.Println()
	}
}

// buildConversationPrompt builds a prompt that includes the full conversation history
func buildConversationPrompt(history []string) string {
	if len(history) == 0 {
		return ""
	}

	// For the first message, just send it as-is
	if len(history) == 1 {
		return strings.TrimPrefix(history[0], "User: ")
	}

	// For subsequent messages, include conversation context
	var prompt strings.Builder
	prompt.WriteString("Previous conversation:\n\n")

	// Include all history except the last user message
	for i := 0; i < len(history)-1; i++ {
		prompt.WriteString(history[i])
		prompt.WriteString("\n\n")
	}

	// Add the current user message
	prompt.WriteString("Current question:\n")
	prompt.WriteString(strings.TrimPrefix(history[len(history)-1], "User: "))

	return prompt.String()
}

// printFormattedResponse prints the response preserving original formatting
// including paragraphs, lists, and other structure
func printFormattedResponse(text string) {
	// Get terminal width, default to 120 if can't determine
	terminalWidth := getTerminalWidth()
	if terminalWidth < 40 {
		terminalWidth = 120
	}

	// Leave some margin for readability (3 chars for indent)
	maxWidth := terminalWidth - 6

	// Split by double newlines first to preserve paragraphs
	paragraphs := strings.Split(text, "\n\n")

	for i, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// Check if this is a heading (starts with ###, ##, or #)
		if strings.HasPrefix(para, "### ") {
			// H3 heading - cyan, bold
			heading := strings.TrimPrefix(para, "### ")
			fmt.Printf("   %s%s%s\n", log.Cyan+log.Bold, heading, log.ResetColor)
			continue
		} else if strings.HasPrefix(para, "## ") {
			// H2 heading - cyan, bold, slightly larger feel
			heading := strings.TrimPrefix(para, "## ")
			fmt.Printf("   %s%s%s\n", log.Cyan+log.Bold, heading, log.ResetColor)
			continue
		} else if strings.HasPrefix(para, "# ") {
			// H1 heading - cyan, bold
			heading := strings.TrimPrefix(para, "# ")
			fmt.Printf("   %s%s%s\n", log.Cyan+log.Bold, heading, log.ResetColor)
			continue
		}

		// Check if this is a list item (starts with -, *, •, or number.)
		lines := strings.Split(para, "\n")
		isListBlock := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") ||
				strings.HasPrefix(trimmed, "•") || isNumberedListItem(trimmed) {
				isListBlock = true
				break
			}
		}

		if isListBlock {
			// Handle list formatting
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Format the line with bold text support
				formattedLine := formatMarkdown(line)
				printWrappedLine(formattedLine, maxWidth, "   ")
			}
		} else {
			// Regular paragraph - format markdown and wrap text
			formattedPara := formatMarkdown(para)
			printWrappedLine(formattedPara, maxWidth, "   ")
		}

		// Add spacing between paragraphs (but not after the last one)
		if i < len(paragraphs)-1 {
			fmt.Println()
		}
	}
}

// formatMarkdown converts markdown formatting to terminal escape codes
func formatMarkdown(text string) string {
	// Handle bold text (**text** or __text__)
	// Use a simple state machine to handle bold formatting
	var result strings.Builder
	i := 0
	for i < len(text) {
		// Check for **bold**
		if i+1 < len(text) && text[i:i+2] == "**" {
			// Find the closing **
			closeIdx := strings.Index(text[i+2:], "**")
			if closeIdx != -1 {
				// Found closing **, apply bold
				boldText := text[i+2 : i+2+closeIdx]
				result.WriteString(log.Bold + boldText + log.ResetColor)
				i += 2 + closeIdx + 2
				continue
			}
		}
		// No markdown formatting found, just append the character
		result.WriteByte(text[i])
		i++
	}

	return result.String()
}

// isNumberedListItem checks if a line starts with a number followed by . or )
func isNumberedListItem(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}

	// Check for patterns like "1. " or "1) "
	for i, ch := range trimmed {
		if ch >= '0' && ch <= '9' {
			continue
		}
		if (ch == '.' || ch == ')') && i > 0 {
			return true
		}
		return false
	}
	return false
}

// printWrappedLine prints a single line/paragraph with word wrapping
func printWrappedLine(text string, maxWidth int, indent string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	var currentLine strings.Builder
	currentLine.WriteString(indent)

	for i, word := range words {
		// Check if adding this word would exceed the width
		testLen := currentLine.Len() + len(word)
		if i > 0 {
			testLen++ // for the space
		}

		if testLen > maxWidth && currentLine.Len() > len(indent) {
			// Print current line and start a new one
			fmt.Println(currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(indent)
			currentLine.WriteString(word)
		} else {
			if currentLine.Len() > len(indent) {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	// Print any remaining text
	if currentLine.Len() > len(indent) {
		fmt.Println(currentLine.String())
	}
}

// getTerminalWidth returns the current terminal width
func getTerminalWidth() int {
	// Try to get from readline's terminal
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if width, err := strconv.Atoi(cols); err == nil && width > 0 {
			return width
		}
	}

	// Default to a reasonable width
	return 120
}

/*--------- LLM Provider and Prompt Flags ---------*/

// getChatProviderFlags returns the LLM provider flags, modifying the LLM configuration.
func getChatProviderFlags() []llmCfg.ProviderFlag {
	var flags []llmCfg.ProviderFlag

	// Append provider override flag if specified.
	if chatProviderOverride != "" {
		flags = append(flags, llmCfg.WithProviderOverride(llmCfg.ProviderType(chatProviderOverride)))
	}

	return flags
}

// getChatPromptFlags returns the LLM prompt flags, modifying the LLM prompt configuration.
func getChatPromptFlags() []llm.PromptFlag {
	var flags []llm.PromptFlag

	// Append model override flag if provided.
	if chatModelOverride != "" {
		flags = append(flags, llm.WithLLMModelOverride(chatModelOverride))
	}

	return flags
}
