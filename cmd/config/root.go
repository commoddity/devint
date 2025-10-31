// ---------------------------------------------------------------------------
// File: root.go
// Package: config
//
// Purpose:
//
//	This command implements an interactive configuration editor for the
//	Developer Interface (DI). It allows the user to traverse and edit
//	the YAML configuration file (~/.devint.config.yaml) interactively, based on the
//	schema defined in ./config/config.schema.yaml. The command supports editing
//	of nested fields, enum selection with allowed values (displayed in purple),
//	and provider-specific validation (e.g., ensuring that a default LLM provider
//	is properly configured before it can be selected).
//
// Features:
//   - Interactive traversal of config fields with options to "go up" a level.
//   - Dynamic prompts that display the field's schema description.
//   - Enum-based selections with allowed values.
//   - Provider validation: if a default LLM provider is selected which
//     lacks configuration (api_key or client_model), the user is prompted to fill
//     in the necessary details. The client_model field uses enum options.
//   - Colorized output and emojis for improved readability and guidance.
//   - The ability to save and exit from any prompt by typing 's' (save option) in yellow.
//   - Clear text prompts for errors, field names, and schema descriptions.
//
// Usage:
//
//	Running the "devint config" command will launch the interactive configuration editor.
//	It supports flags:
//	   --show (-s): Show the current configuration.
//	   --editor (-e): Open the configuration in a text editor instead of interactive mode.
//
// ---------------------------------------------------------------------------
package config

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/commoddity/devint/config"
	cfgEditor "github.com/commoddity/devint/config/editor"
	"github.com/commoddity/devint/log"
	"gopkg.in/yaml.v3"
)

var (
	show   bool
	editor string
)

// init sets up flags for the config command.
func init() {
	ConfigCmd.Flags().BoolVarP(&show, "show", "s", false, "Show the configuration.")
	ConfigCmd.Flags().StringVarP(&editor, "editor", "e", "", "Edit the configuration in the given text editor.")
}

// ConfigCmd represents the interactive configuration command.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit the configuration for the application.",
	Long: `Edit the configuration for the application.

This command is used to modify the YAML configuration file for the Developer Interface.
It uses an interactive command-line interface to traverse and update configuration fields,
using the schema defined in ./config/config.schema.yaml. You can navigate through nested fields,
edit values (with enum validation where applicable), and ensure that required fields for providers
(such as LLM configurations) are appropriately set. You may also choose to save and exit at any
time by entering the save command.
	  
Flags:
  --show (-s)   : Show the current config file.
  --editor (-e) : Open the config file in a specified text editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the --show flag: print the configuration file.
		if show {
			showConfig()
			return
		}

		// Handle the --editor flag: open the file in the given text editor.
		if editor != "" {
			editConfig(editor)
			return
		}

		// Otherwise, start the interactive configuration editor.
		schema, err := config.LoadSchema()
		if err != nil {
			fmt.Printf("Failed to load schema: %v", err)
			os.Exit(1)
		}

		// Define custom field handlers.
		customHandlerFuncs := []cfgEditor.WithCustomFieldHandlerFunc{
			// This handler logs a warning if the user selects a default LLM provider that is not configured.
			WithDefaultLLMProviderHandler,
		}

		yamlEditor, err := cfgEditor.NewYAMLEditor(
			"devint",
			config.ConfigFilePath,
			schema,
			customHandlerFuncs...,
		)
		if err != nil {
			fmt.Printf("Failed to create editor: %v", err)
			os.Exit(1)
		}

		// Start the interactive editor.
		yamlEditor.InteractiveEditConfig()
	},
}

// showConfig prints the current configuration file to stdout.
func showConfig() {
	data, err := os.ReadFile(config.ConfigFilePath)
	if err != nil {
		fmt.Printf("Failed to read config file: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// editConfig opens the configuration file in the user's preferred text editor.
func editConfig(editor string) {
	cmd := exec.Command(editor, config.ConfigFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}

/* ---------- Custom Field Handlers ---------- */

const (
	// defaultLLMProviderFieldName is the field name for the default LLM provider in the DI config file.
	defaultLLMProviderFieldName = "llm_config.default_llm_provider"
)

// WithDefaultLLMProviderHandler sets a custom field handler for the default LLM provider field.
func WithDefaultLLMProviderHandler(yamlEditor *cfgEditor.YAMLEditor) {
	// customLLMProviderHandler logs a warning if the user
	// selects a default LLM provider that is not configured.
	defaultLLMProviderHandler := func(node *yaml.Node, fieldPath, provider string) {
		// Find the llm_providers section in the YAML config
		llmConfigNode := cfgEditor.GetMappingValue(node, "llm_config")
		if llmConfigNode == nil {
			fmt.Printf("%s⚠️  No LLM configuration found!%s\n", log.Red, log.ResetColor)
			return
		}

		llmProvidersNode := cfgEditor.GetMappingValue(llmConfigNode, "llm_providers")
		if llmProvidersNode == nil {
			fmt.Printf("%s⚠️  No LLM providers configured!%s\n", log.Red, log.ResetColor)
			return
		}

		// Check if the specified provider is configured
		providerNode := cfgEditor.GetMappingValue(llmProvidersNode, provider)
		if providerNode == nil || cfgEditor.GetMappingValue(providerNode, "api_key").Value == "" || cfgEditor.GetMappingValue(providerNode, "client_model").Value == "" {
			fmt.Printf("%s🚨 Provider '%s' is not fully configured!%s\n", log.Red, provider, log.ResetColor)
			fmt.Printf("%sPlease use the interactive configuration editor to configure provider '%s' by adding the missing fields: api_key and client_model.%s\n", log.Yellow, provider, log.ResetColor)
		}
	}

	// Set the custom field handler for the default LLM provider field.
	yamlEditor.SetCustomFieldHandler(
		defaultLLMProviderFieldName,
		defaultLLMProviderHandler,
	)
}
