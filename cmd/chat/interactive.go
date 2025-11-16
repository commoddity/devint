// Package chat implements the chat command for interacting with LLM providers.
//
// This file handles the interactive chat mode entry point.
package chat

import (
	"log/slog"

	"github.com/commoddity/devint/chat"
	"github.com/commoddity/devint/config"
	"github.com/commoddity/devint/llm"
)

// RunInteractiveMode starts an interactive chat session with the LLM provider.
// It creates a beautiful TUI using tview for the chat interface.
func RunInteractiveMode(llmProvider llm.LLMProvider, logger *slog.Logger, cfg *config.Config) {
	// Create the chat UI
	ui := chat.NewChatUI(llmProvider, logger, cfg, chatProviderOverride, chatModelOverride)

	// Run the UI (this blocks until the user exits)
	if err := ui.Run(); err != nil {
		logger.Error("failed to run chat UI", "error", err)
	}
}
