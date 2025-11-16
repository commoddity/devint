// Package chat provides a beautiful terminal UI for interacting with LLM providers.
//
// This file handles LLM response streaming and non-streaming response handling.
package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/commoddity/devint/llm"
)

// HandleResponse orchestrates sending a prompt to the LLM and displaying the response.
// It handles both streaming and non-streaming providers.
func (ui *ChatUI) HandleResponse(ctx context.Context, prompt string, promptFlags []llm.PromptFlag) error {
	// Get current provider and model for display
	providerName, modelName := ui.GetProviderInfo()

	// Add assistant message prefix to chat view
	assistantPrefix := FormatAssistantMessage(providerName, modelName)
	fmt.Fprint(ui.chatView, assistantPrefix)

	// Check if provider supports streaming
	streamingProvider, supportsStreaming := ui.llmProvider.(llm.StreamingLLMProvider)

	var responseText string
	var err error

	if supportsStreaming {
		// Use streaming mode
		responseText, err = ui.streamToView(ctx, streamingProvider, prompt, promptFlags)
	} else {
		// Use non-streaming mode with spinner
		responseText, err = ui.nonStreamingResponse(ctx, prompt, promptFlags)
	}

	if err != nil {
		errorMsg := fmt.Sprintf("\n[red]❌ Error: %v[-]\n\n", err)
		fmt.Fprint(ui.chatView, errorMsg)
		return err
	}

	// Add line breaks after response
	fmt.Fprint(ui.chatView, "\n\n")

	// Add response to history
	ui.history.AddAssistantMessage(responseText)

	// Scroll to bottom
	ui.chatView.ScrollToEnd()

	return nil
}

// streamToView handles streaming responses from the LLM provider.
// It updates the chat view in real-time as chunks arrive with a loading indicator.
func (ui *ChatUI) streamToView(ctx context.Context, provider llm.StreamingLLMProvider, prompt string, flags []llm.PromptFlag) (string, error) {
	var fullResponse string

	// Set processing state and disable input
	ui.isProcessing = true
	ui.app.QueueUpdateDraw(func() {
		ui.inputField.SetDisabled(true)
		// Keep title simple and consistent
		ui.inputField.SetTitle(" Input ")
		// Clear any existing text
		ui.inputField.SetText("", true)
	})

	// Start animated loading indicators (header and status view)
	stopHeaderLoading := make(chan bool, 1)
	stopStatusLoading := make(chan bool, 1)
	go ui.animateHeaderLoading(stopHeaderLoading)
	go ui.animateStatusLoading(stopStatusLoading)

	// Buffer to accumulate the response for proper markdown conversion
	var buffer string

	// Store the current chat view content before adding the response
	chatContentBeforeResponse := ui.chatView.GetText(false)

	err := provider.StreamPrompt(ctx, prompt, func(chunk string) error {
		// Append chunk to buffer and full response
		buffer += chunk
		fullResponse += chunk

		// Convert the entire accumulated buffer to tview formatting
		formattedResponse := ConvertMarkdownToTview(buffer)

		// Update UI by redrawing from the response start point
		ui.app.QueueUpdateDraw(func() {
			// Clear the chat view and restore content before response
			ui.chatView.Clear()
			fmt.Fprint(ui.chatView, chatContentBeforeResponse)

			// Add the formatted response
			fmt.Fprint(ui.chatView, formattedResponse)
			ui.chatView.ScrollToEnd()
		})

		return nil
	}, flags...)

	// Stop loading animations when streaming completes
	stopHeaderLoading <- true
	stopStatusLoading <- true

	// Small delay to ensure animations stop cleanly
	time.Sleep(100 * time.Millisecond)

	// Clear processing state and re-enable input
	ui.isProcessing = false
	ui.app.QueueUpdateDraw(func() {
		ui.statusView.SetText("")
		ui.inputField.SetDisabled(false)
		ui.inputField.SetPlaceholder("Type your message here...")
		// Keep title simple and consistent
		ui.inputField.SetTitle(" Input ")
	})

	if err != nil {
		return "", err
	}

	return fullResponse, nil
}

// animateHeaderLoading shows an animated loading indicator in the header.
func (ui *ChatUI) animateHeaderLoading(stop chan bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	// Store original header text
	providerName, modelName := ui.GetProviderInfo()
	originalHeader := fmt.Sprintf(
		"[cyan::b]🤖 Interactive Chat Mode - LLM Assistant[-::-]\n"+
			"[yellow]💡 Provider:[-] %s [cyan::b]%s[-::-]  [yellow]🔧 Model:[-] [green::b]%s[-::-]",
		ui.providerEmoji, providerName, modelName)

	for {
		select {
		case <-stop:
			// Restore original header
			ui.app.QueueUpdateDraw(func() {
				ui.headerView.SetText(originalHeader)
			})
			return
		case <-ticker.C:
			frame := frames[frameIdx]
			ui.app.QueueUpdateDraw(func() {
				// Update header with loading indicator
				headerWithLoading := fmt.Sprintf(
					"[cyan::b]🤖 Interactive Chat Mode - LLM Assistant[-::-] [cyan]%s[-]\n"+
						"[yellow]💡 Provider:[-] %s [cyan::b]%s[-::-]  [yellow]🔧 Model:[-] [green::b]%s[-::-]",
					frame, ui.providerEmoji, providerName, modelName)
				ui.headerView.SetText(headerWithLoading)
			})
			frameIdx = (frameIdx + 1) % len(frames)
		}
	}
}

// animateStatusLoading shows an animated loading indicator in the status view with blue color.
func (ui *ChatUI) animateStatusLoading(stop chan bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			frame := frames[frameIdx]
			ui.app.QueueUpdateDraw(func() {
				// Set text with blue color tags for both spinner and text
				statusText := fmt.Sprintf("[blue]%s Waiting for response...[-]", frame)
				ui.statusView.SetText(statusText)
			})
			frameIdx = (frameIdx + 1) % len(frames)
		}
	}
}

// nonStreamingResponse handles non-streaming responses with a loading spinner.
func (ui *ChatUI) nonStreamingResponse(ctx context.Context, prompt string, flags []llm.PromptFlag) (string, error) {
	// Set processing state and disable input
	ui.isProcessing = true
	ui.app.QueueUpdateDraw(func() {
		ui.inputField.SetDisabled(true)
		// Keep title simple and consistent
		ui.inputField.SetTitle(" Input ")
		// Clear any existing text
		ui.inputField.SetText("", true)
	})

	// Start animated loading indicator in status view
	stopLoading := make(chan bool, 1)
	go ui.animateStatusLoading(stopLoading)

	// Show spinner in chat view
	spinnerMsg := "[cyan]⏳ Thinking...[-]"
	fmt.Fprint(ui.chatView, spinnerMsg)
	ui.app.Draw()

	// Create a channel to handle the response
	type result struct {
		text string
		err  error
	}
	resultChan := make(chan result, 1)

	// Start spinner in a goroutine
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()

	// Make the API call in a goroutine
	go func() {
		response, err := ui.llmProvider.SendPrompt(ctx, prompt, flags...)
		s.Stop()
		resultChan <- result{text: response, err: err}
	}()

	// Wait for response
	res := <-resultChan

	// Stop loading animation
	stopLoading <- true

	// Small delay to ensure animation stops cleanly
	time.Sleep(100 * time.Millisecond)

	// Clear spinner message by finding and removing it
	currentText := ui.chatView.GetText(true)
	currentText = currentText[:len(currentText)-len(spinnerMsg)]
	ui.chatView.Clear()
	fmt.Fprint(ui.chatView, currentText)

	// Clear processing state and re-enable input
	ui.isProcessing = false
	ui.app.QueueUpdateDraw(func() {
		ui.statusView.SetText("")
		ui.inputField.SetDisabled(false)
		ui.inputField.SetPlaceholder("Type your message here...")
		// Keep title simple and consistent
		ui.inputField.SetTitle(" Input ")
	})

	if res.err != nil {
		return "", res.err
	}

	// Format and display the response
	formattedResponse := ConvertMarkdownToTview(res.text)
	fmt.Fprint(ui.chatView, formattedResponse)

	return res.text, nil
}
