// Package chat provides a beautiful terminal UI for interacting with LLM providers.
//
// This file handles layout assembly and event handling for the chat UI.
package chat

import (
	"context"
	"strings"

	"github.com/commoddity/devint/llm"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// BuildLayout creates the three-panel flex layout and sets up event handlers.
// Returns the root layout component ready to be displayed.
func (ui *ChatUI) BuildLayout() *tview.Flex {
	// Create main vertical flex layout
	ui.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.headerView, 4, 0, false).
		AddItem(ui.chatView, 0, 1, false).
		AddItem(ui.statusView, 1, 0, false).
		AddItem(ui.inputField, 6, 0, true)

	// Set up event handlers
	ui.setupEventHandlers()

	return ui.layout
}

// setupEventHandlers configures input handling for the chat interface.
func (ui *ChatUI) setupEventHandlers() {
	// Handle input field events (Enter to send, Ctrl+C to exit)
	ui.inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			// Don't process if already processing a response
			if ui.isProcessing {
				return nil
			}

			// Get user input
			userInput := strings.TrimSpace(ui.inputField.GetText())

			// Skip empty input
			if userInput == "" {
				return nil
			}

			// Clear input field
			ui.inputField.SetText("", true)

			// Add user message to chat view
			userMsg := FormatUserMessage(userInput)
			ui.chatView.Write([]byte(userMsg))
			ui.chatView.ScrollToEnd()

			// Add to history
			ui.history.AddUserMessage(userInput)

			// Build prompt with conversation context
			fullPrompt := ui.history.BuildPrompt()

			// Get prompt flags for model override
			var promptFlags []llm.PromptFlag
			if ui.modelOverride != "" {
				promptFlags = append(promptFlags, llm.WithLLMModelOverride(ui.modelOverride))
			}

			// Handle response in a goroutine to avoid blocking UI
			go func() {
				err := ui.HandleResponse(context.Background(), fullPrompt, promptFlags)
				if err != nil {
					// Error is already displayed in HandleResponse
					ui.app.Draw()
				}
			}()

			return nil

		case tcell.KeyCtrlC:
			// Exit the application
			ui.app.Stop()
			return nil
		}

		return event
	})

	// Handle global Ctrl+C
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			ui.app.Stop()
			return nil
		}
		return event
	})
}

// Run starts the tview application and blocks until it exits.
func (ui *ChatUI) Run() error {
	// Build the layout
	layout := ui.BuildLayout()

	// Set the root and run
	return ui.app.SetRoot(layout, true).Run()
}
