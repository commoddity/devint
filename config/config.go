package config

import (
	"fmt"
	"io"
	"os"

	_ "embed"

	"gopkg.in/yaml.v3"

	"github.com/commoddity/devint/config/git"
	"github.com/commoddity/devint/config/llm"
)

const configFileName = ".devint.config.yaml"

// eg. /Users/greg/.devint.config.yaml
var ConfigFilePath = fmt.Sprintf("%s/%s", os.Getenv("HOME"), configFileName)

// Config represents the configuration for LLM providers and Git.
type Config struct {
	Git  *git.Config `yaml:"git_config"`
	LLMs *llm.Config `yaml:"llm_config"`
}

// LoadConfig loads the configuration from a YAML file.
func LoadConfig() (*Config, error) {
	file, err := os.Open(ConfigFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func InitEmptyConfig() error {
	config := &Config{
		Git: &git.Config{},
		LLMs: &llm.Config{
			LLMProviders: llm.ProvidersConfig{
				DeepSeek:   &llm.DeepSeekConfig{},
				OpenRouter: &llm.OpenRouterConfig{},
				Thaura:     &llm.ThauraConfig{},
			},
		},
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFilePath, data, 0644)
}

// Embed the schema file in the binary for use in the config command.
// This is to allow the schema file to be loaded in the config command,
// regardless of where the binary is run from.
//
//go:embed config.schema.yaml
var schemaYaml []byte

// LoadSchema loads the embedded schema from the config.schema.yaml file.
func LoadSchema() (*yaml.Node, error) {
	var schemaNode yaml.Node
	if err := yaml.Unmarshal(schemaYaml, &schemaNode); err != nil {
		return nil, err
	}

	// Ensure that the schemaNode is the mapping node. If the provided schemaNode
	// is a document node (which typically has Kind != yaml.MappingNode and its
	// actual content is in Content[0]), then extract it.
	if schemaNode.Kind != yaml.MappingNode && len(schemaNode.Content) > 0 {
		schemaNode = *schemaNode.Content[0]
	}

	return &schemaNode, nil
}
