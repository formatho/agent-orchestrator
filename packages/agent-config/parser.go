package agentconfig

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// Format represents a configuration file format.
type Format string

const (
	FormatYAML Format = "yaml"
	FormatTOML Format = "toml"
	FormatJSON Format = "json"
)

// Parser handles parsing and serialization of configuration files.
type Parser struct{}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// DetectFormat determines the format based on file extension.
// Returns an error if the extension is not recognized.
func (p *Parser) DetectFormat(path string) (Format, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return FormatYAML, nil
	case ".toml":
		return FormatTOML, nil
	case ".json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown format for extension: %s", ext)
	}
}

// Parse parses configuration data based on the specified format.
func (p *Parser) Parse(data []byte, format Format) (*Config, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty configuration data")
	}

	config := NewConfig()

	switch format {
	case FormatYAML:
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case FormatTOML:
		if err := toml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse TOML: %w", err)
		}
	case FormatJSON:
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// Ensure agents map is initialized
	if config.Agents == nil {
		config.Agents = make(map[string]*AgentConfig)
	}

	// Set agent names from map keys
	for name, agent := range config.Agents {
		agent.Name = name
	}

	return config, nil
}

// ParseFile parses a configuration file, auto-detecting the format.
func (p *Parser) ParseFile(path string) (*Config, error) {
	format, err := p.DetectFormat(path)
	if err != nil {
		return nil, err
	}

	data, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(data, format)
}

// Serialize converts a Config to bytes in the specified format.
func (p *Parser) Serialize(config *Config, format Format) ([]byte, error) {
	if config == nil {
		return nil, fmt.Errorf("nil configuration")
	}

	switch format {
	case FormatYAML:
		data, err := yaml.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize YAML: %w", err)
		}
		return data, nil

	case FormatTOML:
		var buf strings.Builder
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(config); err != nil {
			return nil, fmt.Errorf("failed to serialize TOML: %w", err)
		}
		return []byte(buf.String()), nil

	case FormatJSON:
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize JSON: %w", err)
		}
		return data, nil

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ParseYAML is a convenience function to parse YAML data.
func ParseYAML(data []byte) (*Config, error) {
	return NewParser().Parse(data, FormatYAML)
}

// ParseTOML is a convenience function to parse TOML data.
func ParseTOML(data []byte) (*Config, error) {
	return NewParser().Parse(data, FormatTOML)
}

// ParseJSON is a convenience function to parse JSON data.
func ParseJSON(data []byte) (*Config, error) {
	return NewParser().Parse(data, FormatJSON)
}

// ParseString parses configuration from a string with the specified format.
func ParseString(data string, format Format) (*Config, error) {
	return NewParser().Parse([]byte(data), format)
}

// readFile is a helper that will be replaced by os.ReadFile when used in manager.
// This allows for easier testing.
var readFile = func(path string) ([]byte, error) {
	return readFileImpl(path)
}
