// Package chat provides a beautiful terminal UI for interacting with LLM providers.
//
// This file handles message formatting and markdown conversion for tview.
package chat

import (
	"strings"
)

// FormatUserMessage creates a styled user message for display in the chat view.
// The message is prefixed with a green bullet point and "You:" label on its own line.
func FormatUserMessage(message string) string {
	return "[green]● [green::b]You:[-:-]\n\n" + message + "\n\n"
}

// FormatAssistantMessage creates a styled assistant message prefix for display.
// The prefix includes the provider name (capitalized) and model in blue on its own line.
func FormatAssistantMessage(providerName, modelName string) string {
	// Capitalize first letter of provider name
	if len(providerName) == 0 {
		return "[blue]● [blue::b]Assistant:[-:-]\n"
	}
	displayProvider := strings.ToUpper(providerName[:1]) + providerName[1:]
	if modelName == "" {
		return "[blue]● [blue::b]" + displayProvider + ":[-:-]\n"
	}
	return "[blue]● [blue::b]" + displayProvider + " (" + modelName + "):[-:-]\n"
}

// ConvertMarkdownToTview converts markdown formatting to tview color tags.
// Currently supports:
// - **bold text** → [::b]bold text[-::-]
// - ### Heading → [cyan::b]Heading[-::-]
func ConvertMarkdownToTview(text string) string {
	// Handle bold text (**text**)
	text = convertBoldMarkdown(text)

	// Handle headings (###, ##, #)
	text = convertHeadings(text)

	return text
}

// convertBoldMarkdown converts **text** to tview bold formatting.
func convertBoldMarkdown(text string) string {
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
				result.WriteString("[::b]" + boldText + "[-::-]")
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

// convertHeadings converts markdown headings to tview cyan bold formatting.
func convertHeadings(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for heading markers
		if strings.HasPrefix(trimmed, "### ") {
			heading := strings.TrimPrefix(trimmed, "### ")
			lines[i] = "[cyan::b]" + heading + "[-::-]"
		} else if strings.HasPrefix(trimmed, "## ") {
			heading := strings.TrimPrefix(trimmed, "## ")
			lines[i] = "[cyan::b]" + heading + "[-::-]"
		} else if strings.HasPrefix(trimmed, "# ") {
			heading := strings.TrimPrefix(trimmed, "# ")
			lines[i] = "[cyan::b]" + heading + "[-::-]"
		}
	}
	return strings.Join(lines, "\n")
}
