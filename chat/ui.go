// Package chat provides a beautiful terminal UI for interacting with LLM providers.
//
// This file handles tview UI component creation and management.
package chat

import (
	"fmt"
	"log/slog"

	"github.com/commoddity/devint/config"
	"github.com/commoddity/devint/llm"
	"github.com/rivo/tview"
)

// ChatUI holds all the tview components for the interactive chat interface.
type ChatUI struct {
	app        *tview.Application
	chatView   *tview.TextView
	inputField *tview.TextArea
	headerView *tview.TextView
	statusView *tview.TextView
	layout     *tview.Flex

	// Dependencies
	llmProvider llm.LLMProvider
	logger      *slog.Logger
	config      *config.Config
	history     *ConversationHistory

	// Provider info
	providerName  string
	providerEmoji string
	modelName     string
	modelOverride string

	// State
	isProcessing bool
}

// NewChatUI creates and initializes a new chat UI with all components.
func NewChatUI(llmProvider llm.LLMProvider, logger *slog.Logger, cfg *config.Config, providerOverride, modelOverride string) *ChatUI {
	ui := &ChatUI{
		app:           tview.NewApplication(),
		llmProvider:   llmProvider,
		logger:        logger,
		config:        cfg,
		history:       NewConversationHistory(),
		modelOverride: modelOverride,
	}

	// Determine provider and model info
	ui.providerName = string(cfg.LLMs.DefaultLLMProvider)
	if providerOverride != "" {
		ui.providerName = providerOverride
	}

	// Set provider-specific emoji and model
	switch ui.providerName {
	case "thaura":
		ui.providerEmoji = "🍉"
		if modelOverride != "" {
			ui.modelName = modelOverride
		} else if cfg.LLMs.LLMProviders.Thaura != nil {
			ui.modelName = string(cfg.LLMs.LLMProviders.Thaura.ClientModel)
		}
	case "deepseek":
		ui.providerEmoji = "🤖"
		if modelOverride != "" {
			ui.modelName = modelOverride
		} else if cfg.LLMs.LLMProviders.DeepSeek != nil {
			ui.modelName = string(cfg.LLMs.LLMProviders.DeepSeek.ClientModel)
		}
	case "openrouter":
		ui.providerEmoji = "🌐"
		if modelOverride != "" {
			ui.modelName = modelOverride
		} else if cfg.LLMs.LLMProviders.OpenRouter != nil {
			ui.modelName = string(cfg.LLMs.LLMProviders.OpenRouter.ClientModel)
		}
	default:
		ui.providerEmoji = "🤖"
		ui.modelName = "unknown"
	}

	ui.initializeComponents()
	return ui
}

// initializeComponents creates and configures all tview components.
func (ui *ChatUI) initializeComponents() {
	// Create header view
	ui.headerView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	headerText := fmt.Sprintf(
		"[cyan::b]🤖 Interactive Chat Mode - LLM Assistant[-::-]\n"+
			"[yellow]💡 Provider:[-] %s [cyan::b]%s[-::-]  [yellow]🔧 Model:[-] [green::b]%s[-::-]",
		ui.providerEmoji, ui.providerName, ui.modelName)
	ui.headerView.SetText(headerText)
	ui.headerView.SetBorder(true).SetBorderPadding(0, 0, 1, 1)

	// Create chat view (scrollable text area for conversation)
	ui.chatView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})
	ui.chatView.SetBorder(true).SetTitle(" Chat History ").SetBorderPadding(1, 1, 1, 1)

	// Enable mouse support for scrolling
	ui.app.EnableMouse(true)

	// Add welcome message
	welcomeMsg := "[white]Type your message below and press Enter to send.\n" +
		"Press Ctrl+C to exit at any time.[-:-]\n\n"
	fmt.Fprint(ui.chatView, welcomeMsg)

	// Create status view for loading indicator
	ui.statusView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	ui.statusView.SetBorder(false)

	// Create input field
	ui.inputField = tview.NewTextArea().
		SetPlaceholder("Type your message here...")
	ui.inputField.SetBorder(true).SetTitle(" Input ").SetBorderPadding(0, 0, 1, 1)

	// Make sure input field has focus by default
	ui.inputField.SetInputCapture(nil)
}

// GetApp returns the tview application instance.
func (ui *ChatUI) GetApp() *tview.Application {
	return ui.app
}

// GetChatView returns the chat history text view.
func (ui *ChatUI) GetChatView() *tview.TextView {
	return ui.chatView
}

// GetInputField returns the input text area.
func (ui *ChatUI) GetInputField() *tview.TextArea {
	return ui.inputField
}

// GetHistory returns the conversation history manager.
func (ui *ChatUI) GetHistory() *ConversationHistory {
	return ui.history
}

// GetLLMProvider returns the LLM provider instance.
func (ui *ChatUI) GetLLMProvider() llm.LLMProvider {
	return ui.llmProvider
}

// GetProviderInfo returns the provider name and model for display purposes.
func (ui *ChatUI) GetProviderInfo() (providerName, modelName string) {
	// Use override if provided
	if ui.modelOverride != "" {
		return ui.providerName, ui.modelOverride
	}
	return ui.providerName, ui.modelName
}
