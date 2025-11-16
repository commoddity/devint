// Package chat provides a beautiful terminal UI for interacting with LLM providers.
//
// This file handles conversation history management and context building.
package chat

import (
	"fmt"
	"strings"
)

// ConversationHistory manages the conversation context between user and assistant.
type ConversationHistory struct {
	messages []string
}

// NewConversationHistory creates a new conversation history manager.
func NewConversationHistory() *ConversationHistory {
	return &ConversationHistory{
		messages: make([]string, 0),
	}
}

// AddUserMessage appends a user message to the conversation history.
func (h *ConversationHistory) AddUserMessage(message string) {
	h.messages = append(h.messages, fmt.Sprintf("User: %s", message))
}

// AddAssistantMessage appends an assistant message to the conversation history.
func (h *ConversationHistory) AddAssistantMessage(message string) {
	h.messages = append(h.messages, fmt.Sprintf("Assistant: %s", message))
}

// BuildPrompt constructs a prompt that includes the full conversation history.
// For the first message, it returns just the user's message.
// For subsequent messages, it includes all previous context.
func (h *ConversationHistory) BuildPrompt() string {
	if len(h.messages) == 0 {
		return ""
	}

	// For the first message, just send it as-is
	if len(h.messages) == 1 {
		return strings.TrimPrefix(h.messages[0], "User: ")
	}

	// For subsequent messages, include conversation context
	var prompt strings.Builder
	prompt.WriteString("Previous conversation:\n\n")

	// Include all history except the last user message
	for i := 0; i < len(h.messages)-1; i++ {
		prompt.WriteString(h.messages[i])
		prompt.WriteString("\n\n")
	}

	// Add the current user message
	prompt.WriteString("Current question:\n")
	prompt.WriteString(strings.TrimPrefix(h.messages[len(h.messages)-1], "User: "))

	return prompt.String()
}
