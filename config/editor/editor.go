package editor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/commoddity/devint/log"
)

type (
	// FieldName is the name of a field in the YAML config, used to perform custom field handling for specific config YAML formats
	FieldName string
	// CustomFieldHandler is a function that performs custom field handling for a specific field in the YAML config.
	// It is called when the user edits a specific field defined in the command for which a custom handler is registered.
	CustomFieldHandler func(node *yaml.Node, fieldPath, fieldValue string)
	// WithCustomFieldHandlerFunc is a function that registers a custom field handler for a specific field in the YAML config.
	// It is used in the command for which a custom handler is registered.
	WithCustomFieldHandlerFunc func(yamlEditor *YAMLEditor)
)

// YAMLEditor allows traversal and editing of ANY configuration YAML file in an
// interactive terminal editor.
//
// It is initialized with the following:
// - A YAML config file path. This is the absolute path to the YAML config file that will be edited.
// - A schema node. This is an unmarshaled yaml.Node of the config.schema.yaml file.
// - A list of custom field handler functions. These are functions that perform custom field handling
// for specific fields in the YAML config.
//
// The custom field handler functions are used to perform custom field handling
// for specific fields in the YAML config.
type YAMLEditor struct {
	// program is the name of the program that is running the editor.
	program string
	// configNode is the unmarshaled yaml.Node of the YAML config file.
	configNode *yaml.Node
	// configFilePath is the absolute path to the YAML config file that will be edited.
	configFilePath string
	// schemaNode is the unmarshaled yaml.Node of the config.schema.yaml file.
	// It is used to extract descriptions and enum options for fields in the YAML config.
	schemaNode *yaml.Node
	// customFieldHandlers is a map of custom field handler functions for specific fields in the YAML config.
	customFieldHandlers map[FieldName]CustomFieldHandler
}

// NewYAMLEditor loads a YAML document into a Node and sets up the editor.
func NewYAMLEditor(
	program string,
	configFilePath string,
	schemaNode *yaml.Node,
	customFieldHandlerFuncs ...WithCustomFieldHandlerFunc,
) (*YAMLEditor, error) {
	// Set up a channel to listen for interrupt signals.
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle the signal if the user interrupts the editor.
	go func() {
		<-signalChannel
		ClearTerminal()
		fmt.Println(log.Yellow + "👋 Exited the configuration YAML editor early. Changes not saved." + log.ResetColor)
		os.Exit(0)
	}()

	configNode, err := loadConfigNode(configFilePath)
	if err != nil {
		return nil, err
	}

	yamlEditor := &YAMLEditor{
		program:             program,
		configNode:          configNode,
		configFilePath:      configFilePath,
		schemaNode:          schemaNode,
		customFieldHandlers: make(map[FieldName]CustomFieldHandler),
	}

	// Add each custom field handler using WithCustomFieldHandler
	for _, customFieldHandlerFunc := range customFieldHandlerFuncs {
		customFieldHandlerFunc(yamlEditor)
	}

	return yamlEditor, nil
}

// loadConfigNode reads the YAML file and unmarshals it into a yaml.Node.
func loadConfigNode(configFilePath string) (*yaml.Node, error) {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		return nil, err
	}

	// The top-level YAML document is the first content node.
	if len(node.Content) > 0 {
		return node.Content[0], nil
	}
	return &node, nil
}

// SetCustomFieldHandler registers a custom field handler for a specific field in the YAML config.
func (e *YAMLEditor) SetCustomFieldHandler(fieldName FieldName, handler CustomFieldHandler) {
	e.customFieldHandlers[fieldName] = handler
}

func (e *YAMLEditor) GetConfigNode() *yaml.Node {
	return e.configNode
}

// InteractiveEditConfig runs the interactive loop for editing the YAML config.
func (e *YAMLEditor) InteractiveEditConfig() {
	for {
		ClearTerminal()
		e.editFieldRecursive(e.configNode, "")

		// Ask for confirmation to continue editing.
		cont := ReadInput(log.Green + "Do you want to continue editing? (y for yes, n to save and exit): " + log.ResetColor)
		cont = strings.ToLower(cont)
		if cont == "n" {
			e.SaveConfigAndExit()
		} else if cont != "y" {
			fmt.Println(log.Red + "Invalid input, returning to main prompt." + log.ResetColor)
		}
		// if "y", loop continues.
	}
}

// GetEnumOptionsForPath traverses the schemaNode using a dot-delimited field path and returns the allowed enum options if available.
func (e *YAMLEditor) GetEnumOptionsForPath(fieldPath string) []string {
	parts := strings.Split(fieldPath, ".")
	props := GetMappingValue(e.schemaNode, "properties")
	if props == nil {
		return nil
	}

	current := props
	for i, part := range parts {
		node := GetMappingValue(current, part)
		if node == nil {
			return nil
		}
		if i == len(parts)-1 {
			enumNode := GetMappingValue(node, "enum")
			if enumNode == nil || enumNode.Kind != yaml.SequenceNode {
				return nil
			}
			var options []string
			for _, item := range enumNode.Content {
				options = append(options, item.Value)
			}
			return options
		}
		next := GetMappingValue(node, "properties")
		if next == nil {
			return nil
		}
		current = next
	}
	return nil
}

// SaveConfigAndExit writes the updated YAML Node configuration to the config file,
// then clears the terminal, prints a success message, and exits.
func (e *YAMLEditor) SaveConfigAndExit() {
	file, err := os.Create(e.configFilePath)
	if err != nil {
		fmt.Printf(log.Red+"Failed to open config file for writing: %v"+log.ResetColor, err)
	}
	defer file.Close()
	encoder := yaml.NewEncoder(file)
	err = encoder.Encode(e.configNode)
	if err != nil {
		fmt.Printf(log.Red+"Failed to encode config to YAML: %v"+log.ResetColor, err)
	}

	ClearTerminal()

	fmt.Println(log.Green + "✅ All changes saved successfully to " + e.configFilePath + ".\n" + log.Blue + "💡 To view the updated raw config, run `" + e.program + " config --show`." + log.ResetColor)

	os.Exit(0)
}

// editFieldRecursive recursively traverses the YAML node tree for interactive editing.
// It allows users to navigate through the YAML structure, edit scalar values (or invoke custom handlers), and save changes.
func (e *YAMLEditor) editFieldRecursive(currentNode *yaml.Node, path string) {
	// Check if the current node is a mapping node (i.e., a dictionary-like structure).
	if currentNode.Kind == yaml.MappingNode {
		for {
			ClearTerminal() // Clear the terminal for a fresh display.
			fmt.Println(log.Green + "❓ Which field would you like to edit? (or type " + log.Yellow + "s" + log.Green + " to save and exit)" + log.ResetColor)

			// If we're at a nested level, provide an option to go up one level.
			if path != "" {
				fmt.Println("0. Go up one level")
			}

			// Extract keys from the current mapping node for display.
			keys := extractMappingKeys(currentNode)
			printMappingOptions(keys, path, e) // Display the keys with descriptions.

			// Prompt the user for a choice.
			choiceInput := ReadInput(log.Cyan + "📝 Enter choice (number, or " + log.Yellow + "s" + log.Cyan + " to save and exit): " + log.ResetColor)
			choiceInput = strings.ToLower(choiceInput)
			if choiceInput == "s" {
				e.SaveConfigAndExit()
			}

			choice, err := strconv.Atoi(choiceInput)
			if err != nil {
				fmt.Println(log.Red + "Invalid input, please try again." + log.ResetColor)
				continue
			}

			// If we're in a nested level and input is 0, then go up one level.
			if path != "" && choice == 0 {
				return
			}

			// Validate the user's choice.
			if choice < 1 || choice > len(keys) {
				fmt.Println(log.Red + "Invalid choice, please try again." + log.ResetColor)
				continue
			}

			// Calculate the index of the selected key in the node's content.
			index := (choice-1)*2 + 1
			selectedKey := keys[choice-1]
			selectedValueNode := currentNode.Content[index]
			newPath := path + selectedKey

			// Handle the selected YAML node based on its type.
			switch selectedValueNode.Kind {
			case yaml.MappingNode:
				// If the selected node is a mapping, recurse into it.
				e.editFieldRecursive(selectedValueNode, newPath+".")

			case yaml.ScalarNode:
				e.handleScalarEditing(selectedValueNode, newPath)

			default:
				fmt.Println(log.Red + "Unsupported node type." + log.ResetColor)
				return
			}
		}
	} else {
		// If the current node is a standalone primitive, handle its editing directly.
		e.handlePrimitiveNodeEditing(currentNode, path)
	}
}

// extractMappingKeys returns the keys (assumed at even indices) for a mapping node.
func extractMappingKeys(node *yaml.Node) []string {
	keys := []string{}
	for i := 0; i < len(node.Content); i += 2 {
		keys = append(keys, node.Content[i].Value)
	}
	return keys
}

// printMappingOptions prints the keys and schema descriptions for a mapping node.
func printMappingOptions(keys []string, path string, e *YAMLEditor) {
	for i, key := range keys {
		fullPath := key
		if path != "" {
			fullPath = path + key
		}
		desc := e.getDescriptionForPath(fullPath)
		if desc != "" {
			fmt.Printf("%d. "+log.Blue+"%s"+log.ResetColor+" - "+log.White+"%s"+log.ResetColor+"\n", i+1, key, desc)
		} else {
			fmt.Printf("%d. "+log.Blue+"%s"+log.ResetColor+"\n", i+1, key)
		}
	}
}

// handleScalarEditing manages editing of a scalar node (leaf) with optional enum options.
func (e *YAMLEditor) handleScalarEditing(node *yaml.Node, path string) {
	enumOptions := e.GetEnumOptionsForPath(path)

	// Display the current value at the beginning
	ClearTerminal()

	// Get the value directly from the node, or use our path lookup to find it
	currentValue := node.Value
	if currentValue == "" {
		// Try to get the value by path, which has hardcoded values for known problematic paths
		currentValue = e.getValueByPath(path)
		// Update the node's value too
		if currentValue != "" {
			node.Value = currentValue
		}
	}

	if !isSensitive(path) {
		fmt.Printf(log.Cyan+"Current value for %s:%s %s\n", path, log.ResetColor, currentValue)
	}

	var newValue string
	if len(enumOptions) > 0 {
		newValue = e.handleEnumSelection(path, enumOptions, currentValue)
		// If newValue is empty, user selected to go back
		if newValue == "" {
			return
		}
	} else {
		var goBack bool
		newValue, goBack = handleDirectScalarInput(path, currentValue, isSensitive(path))
		if goBack {
			return // Return directly without showing the continue prompt if user wants to go back
		}
	}

	if newValue != "" {
		node.Value = newValue
	}

	// After setting the new value, check if a custom handler is
	// registered for this field and call it if so.
	if handler, ok := e.customFieldHandlers[FieldName(path)]; ok {
		handler(e.configNode, path, newValue)
	}

	// After editing the scalar, ask if the user wants to continue editing.
	contLevel := ReadInput(log.Green + "Do you want to continue editing the config?" + log.Cyan + " (y for yes, n to save and exit): " + log.ResetColor)
	contLevel = strings.ToLower(contLevel)
	if contLevel == "n" {
		e.SaveConfigAndExit()
	} else if contLevel != "y" {
		return
	}
}

// getParentNodeForPath finds the parent node for a given path
func (e *YAMLEditor) getParentNodeForPath(path string) *yaml.Node {
	pathParts := strings.Split(path, ".")
	if len(pathParts) <= 1 {
		return e.configNode
	}

	// Navigate to the parent node
	currentNode := e.configNode
	for _, part := range pathParts[:len(pathParts)-1] {
		found := false
		if currentNode.Kind == yaml.MappingNode {
			for i := 0; i < len(currentNode.Content); i += 2 {
				if currentNode.Content[i].Value == part {
					currentNode = currentNode.Content[i+1]
					found = true
					break
				}
			}
		}
		if !found {
			return nil
		}
	}

	return currentNode
}

// handlePrimitiveNodeEditing handles editing when the current node is a non-mapping primitive.
func (e *YAMLEditor) handlePrimitiveNodeEditing(node *yaml.Node, path string) {
	ClearTerminal()

	// Get the value directly from the node, or use our path lookup to find it
	currentValue := node.Value
	if currentValue == "" {
		// Try to get the value by path, which has hardcoded values for known problematic paths
		currentValue = e.getValueByPath(path)
		// Update the node's value too
		if currentValue != "" {
			node.Value = currentValue
		}
	}

	if !isSensitive(path) {
		fmt.Printf(log.Cyan+"Current value for %s:%s %s\n", path, log.ResetColor, currentValue)
	}
	fmt.Println(log.White + "0. Go up one level" + log.ResetColor)

	newValue, goBack := handleDirectScalarInput(path, currentValue, isSensitive(path))
	if goBack {
		return
	}
	node.Value = newValue
}

// handleEnumSelection manages enum selection from given options.
// Returns the selected value.
func (e *YAMLEditor) handleEnumSelection(path string, options []string, currentVal string) string {
	ClearTerminal()
	fmt.Printf(log.Green+"Select one of the following options for "+log.Blue+"%s"+log.Green+" (or type "+log.Yellow+"s"+log.Green+" to save and exit):"+log.ResetColor+"\n", path)
	fmt.Println(log.White + "0. Go back" + log.ResetColor)
	// List options
	for i, option := range options {
		fmt.Printf("%d. "+log.Purple+"%s"+log.ResetColor+"\n", i+1, option)
	}
	enumInput := ReadInput(log.Cyan + "📝 Enter choice: " + log.ResetColor)
	enumInput = strings.ToLower(enumInput)
	if enumInput == "s" {
		e.SaveConfigAndExit()
	}
	choiceNum, err := strconv.Atoi(enumInput)
	if err != nil || choiceNum < 0 || choiceNum > len(options) {
		fmt.Println(log.Red + "Invalid choice, input ignored." + log.ResetColor)
		return currentVal
	} else if choiceNum == 0 {
		return ""
	} else {
		return options[choiceNum-1]
	}
}

// handleDirectScalarInput prompts the user directly for a new value for a scalar.
// It clears the terminal and prints a "0. Go up one level" option.
// Returns the new value and a bool indicating if the user wants to go back.
func handleDirectScalarInput(path string, currentValue string, sensitive bool) (string, bool) {
	// ClearTerminal() - Remove this as we're now clearing in handleScalarEditing
	// Don't display the value here as it's now shown in handleScalarEditing

	fmt.Println(log.White + "0. Go up one level" + log.ResetColor)

	newValue := promptForValue(path, sensitive)

	for newValue == "" {
		fmt.Println(log.Red + "Input cannot be empty. Please enter a valid newValue." + log.ResetColor)
		newValue = promptForValue(path, sensitive)
	}

	if newValue == "0" {
		return "", true
	}

	return newValue, false
}

// getDescriptionForPath retrieves the description for a given field path from the schemaNode.
func (e *YAMLEditor) getDescriptionForPath(fieldPath string) string {
	parts := strings.Split(fieldPath, ".")
	props := GetMappingValue(e.schemaNode, "properties")
	if props == nil {
		return ""
	}

	current := props
	for i, part := range parts {
		node := GetMappingValue(current, part)
		if node == nil {
			return ""
		}
		if i == len(parts)-1 {
			descNode := GetMappingValue(node, "description")
			if descNode != nil {
				return descNode.Value
			}
			return ""
		}
		next := GetMappingValue(node, "properties")
		if next == nil {
			return ""
		}
		current = next
	}
	return ""
}

/* ------------ Helper functions ------------ */

// GetMappingValue returns the value node corresponding to a key in a mapping node.
func GetMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		v := node.Content[i+1]
		if k.Value == key {
			return v
		}
	}
	return nil
}

// ClearTerminal clears the console for a fresh prompt display.
func ClearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// ReadInput displays a prompt and returns the user's input.
func ReadInput(prompt string) string {
	fmt.Print(prompt)
	// Use bufio.NewReader to read the entire line including empty responses
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(log.Red + "Error reading input." + log.ResetColor)
		return ReadInput(prompt)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println(log.Red + "Input cannot be empty. Please enter a valid value." + log.ResetColor)
		return ReadInput(prompt)
	}
	return input
}

// ReadHiddenInput reads input with hidden echo (used for sensitive fields).
func ReadHiddenInput(prompt string) string {
	fmt.Print(prompt)
	byteInput, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf(log.Red+"Failed to read hidden input: %v"+log.ResetColor, err)
	}
	fmt.Println("")
	return strings.TrimSpace(string(byteInput))
}

// isSensitive returns true if the field path indicates sensitive information.
func isSensitive(fieldPath string) bool {
	fieldPath = strings.ToLower(fieldPath)
	return strings.Contains(fieldPath, "api_key") ||
		strings.Contains(fieldPath, "personal_access_token") ||
		strings.Contains(fieldPath, "private_key") ||
		strings.Contains(fieldPath, "signature")
}

// promptForValue prompts the user for a new value for a field.
func promptForValue(path string, sensitive bool) string {
	var newValue string
	if sensitive {
		newValue = ReadHiddenInput(log.Cyan + "📝 Enter new value for " + log.Blue + path + log.Cyan + " (input hidden): " + log.ResetColor)
	} else {
		newValue = ReadInput(log.Cyan + "📝 Enter new value for " + log.Blue + path + log.Cyan + ": " + log.ResetColor)
	}
	return newValue
}

// GetOrCreateMappingNode retrieves a mapping node with the given key from the parent node,
// or creates a new one if it doesn't exist.
func GetOrCreateMappingNode(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1]
		}
	}
	node := &yaml.Node{Kind: yaml.MappingNode}
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, node)
	return node
}

// SetScalarValue sets the scalar value for a given key in a mapping node's content.
func SetScalarValue(content []*yaml.Node, key, value string) []*yaml.Node {
	for i := 0; i < len(content); i += 2 {
		if content[i].Value == key {
			content[i+1].Value = value
			return content
		}
	}
	return append(content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, &yaml.Node{Kind: yaml.ScalarNode, Value: value})
}

// SetMappingValue sets the mapping value for a given key in a mapping node's content.
func SetMappingValue(content []*yaml.Node, key string, value *yaml.Node) []*yaml.Node {
	for i := 0; i < len(content); i += 2 {
		if content[i].Value == key {
			content[i+1] = value
			return content
		}
	}
	return append(content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
}

// getValueByPath finds a scalar value by its dot-separated path in the YAML tree.
// This provides a more reliable way to get values than trying to extract them from current nodes.
func (e *YAMLEditor) getValueByPath(path string) string {
	pathParts := strings.Split(path, ".")
	if len(pathParts) == 0 {
		return ""
	}

	// Navigate through the config node
	currentNode := e.configNode
	for i, part := range pathParts {
		if currentNode.Kind != yaml.MappingNode {
			return ""
		}

		// Find the matching key in the mapping
		found := false
		for j := 0; j < len(currentNode.Content); j += 2 {
			if j+1 < len(currentNode.Content) && currentNode.Content[j].Value == part {
				// If this is the last part, return the value
				if i == len(pathParts)-1 {
					// We found the value node
					return currentNode.Content[j+1].Value
				}

				// Otherwise keep traversing
				currentNode = currentNode.Content[j+1]
				found = true
				break
			}
		}

		if !found {
			return ""
		}
	}

	return ""
}
