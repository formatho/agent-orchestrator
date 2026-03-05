# go-agent-config

A flexible, thread-safe configuration management library for agent orchestration in Go. Supports multiple formats (YAML, TOML, JSON) with validation and hot reload capabilities.

## Features

- **Multi-format support**: YAML, TOML, and JSON with auto-detection
- **Per-agent overrides**: Global defaults with agent-specific customizations
- **Comprehensive validation**: Range checks, required fields, pattern validation
- **Thread-safe operations**: Safe for concurrent access
- **Hot reload**: Optional file watching with automatic configuration reload
- **Import/Export**: Easy serialization between formats

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/agent-config
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    agentconfig "github.com/formatho/agent-orchestrator/packages/agent-config"
)

func main() {
    // Create a manager with default config
    manager, err := agentconfig.New(agentconfig.NewConfig())
    if err != nil {
        log.Fatal(err)
    }

    // Load from file (auto-detects format by extension)
    if err := manager.Load("agents.yaml"); err != nil {
        log.Fatal(err)
    }

    // Get global configuration
    global := manager.GetGlobal()
    fmt.Printf("Default timeout: %d seconds\n", global.Timeout)

    // Get agent configuration (merged with global defaults)
    agent := manager.GetAgent("daily-report")
    if agent != nil {
        fmt.Printf("Agent model: %s\n", agent.LLM.Model)
    }

    // List all agents
    for _, name := range manager.ListAgents() {
        fmt.Printf("- %s\n", name)
    }
}
```

### Creating Configuration

```go
// Create config with defaults
config := agentconfig.NewConfig()

// Customize global settings
config.Global.Timeout = 600
config.Global.Debug = true
config.Global.LLM = &agentconfig.LLMConfig{
    Provider:    "anthropic",
    Model:       "claude-3-opus",
    Temperature: ptrFloat64(0.7),
    MaxTokens:   ptrInt(4000),
}

// Add an agent
config.Agents["code-reviewer"] = &agentconfig.AgentConfig{
    LLM: &agentconfig.LLMConfig{
        Model: "claude-3-sonnet", // Override global model
    },
    Skills: &agentconfig.SkillsConfig{
        Allowed: []string{"file.read", "file.write", "git.*"},
    },
    Timeout: ptrInt(1200), // Override global timeout
}

// Create manager
manager, err := agentconfig.New(config)
```

### Saving Configuration

```go
// Save to YAML
if err := manager.Save("config.yaml"); err != nil {
    log.Fatal(err)
}

// Save to TOML
if err := manager.Save("config.toml"); err != nil {
    log.Fatal(err)
}

// Save to JSON
if err := manager.Save("config.json"); err != nil {
    log.Fatal(err)
}
```

### Agent Operations

```go
// Set a new agent
agent := &agentconfig.AgentConfig{
    LLM: &agentconfig.LLMConfig{
        Model:       "gpt-4-turbo",
        Temperature: ptrFloat64(0.5),
    },
    Skills: &agentconfig.SkillsConfig{
        Allowed: []string{"http.*", "file.read"},
        Denied:  []string{"shell.run"},
    },
}
if err := manager.SetAgent("web-scraper", agent); err != nil {
    log.Fatal(err)
}

// Get agent config (merged with global defaults)
fullConfig := manager.GetAgent("web-scraper")
// fullConfig contains merged values from agent + global

// Get raw agent config (no global defaults)
rawConfig := manager.GetAgentRaw("web-scraper")

// Delete an agent
deleted := manager.DeleteAgent("web-scraper")
```

### Hot Reload

```go
// Create manager with reload callback
manager, err := agentconfig.New(agentconfig.NewConfig(),
    agentconfig.WithOnReload(func(config *agentconfig.Config) {
        fmt.Println("Configuration reloaded!")
        // Handle the new configuration
    }),
)

// Load initial config
if err := manager.Load("agents.yaml"); err != nil {
    log.Fatal(err)
}

// Start watching for changes
if err := manager.StartWatcher(); err != nil {
    log.Fatal(err)
}
defer manager.StopWatcher()

// Now any changes to agents.yaml will trigger the reload callback
```

### Validation

```go
// Validate entire configuration
if err := manager.Validate(); err != nil {
    if validationErrors, ok := err.(agentconfig.ValidationErrors); ok {
        for _, e := range validationErrors {
            fmt.Printf("Error: %s\n", e.Error())
        }
    }
}

// Add custom validation rules
manager, err := agentconfig.New(agentconfig.NewConfig(),
    agentconfig.WithValidationRules(func(config *agentconfig.Config) agentconfig.ValidationErrors {
        var errs agentconfig.ValidationErrors
        
        // Custom validation logic
        for name, agent := range config.Agents {
            if agent.LLM != nil && agent.LLM.Provider == "custom" {
                errs = append(errs, &agentconfig.ValidationError{
                    Field:   fmt.Sprintf("agents.%s.llm.provider", name),
                    Message: "custom provider requires additional setup",
                })
            }
        }
        
        return errs
    }),
)
```

### Import/Export

```go
// Export as YAML
yamlData, err := manager.Export()
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(yamlData))

// Export as specific format
jsonData, err := manager.ExportAs(agentconfig.FormatJSON)
tomlData, err := manager.ExportAs(agentconfig.FormatTOML)

// Import from data
if err := manager.ImportYAML(yamlData); err != nil {
    log.Fatal(err)
}

if err := manager.ImportJSON(jsonData); err != nil {
    log.Fatal(err)
}

if err := manager.ImportTOML(tomlData); err != nil {
    log.Fatal(err)
}
```

## Configuration Structure

### Global Configuration

```yaml
global:
  llm:
    provider: openai           # Required: LLM provider
    model: gpt-4o              # Required: Model identifier
    temperature: 0.7           # Optional: 0-2
    max_tokens: 2000           # Optional: Max response tokens
    top_p: 0.9                 # Optional: 0-1
    frequency_penalty: 0.0     # Optional: 0-2
    presence_penalty: 0.0      # Optional: 0-2
    stop_sequences:            # Optional: Stop generation at these
      - "END"
    base_url: ""               # Optional: Custom API endpoint
    api_key: ""                # Optional: API key (use env vars)
  
  timeout: 300                 # Default timeout in seconds
  max_retries: 3               # Default retry attempts
  debug: false                 # Global debug mode
  max_concurrent_tasks: 5      # Default concurrency limit
```

### Agent Configuration

```yaml
agents:
  agent-name:
    llm:
      provider: anthropic      # Override global provider
      model: claude-3-opus     # Override global model
      temperature: 0.5         # Override global temperature
    
    skills:
      allowed:                 # Whitelist of permitted skills
        - file.read
        - file.write
        - git.*
      denied:                  # Blacklist of forbidden skills
        - shell.run
    
    timeout: 600               # Override global timeout
    max_retries: 5             # Override global retries
    max_concurrent_tasks: 10   # Override global concurrency
    debug: true                # Per-agent debug mode
    enabled: true              # Agent enabled flag
    
    metadata:                  # Custom metadata
      description: "Agent description"
      version: "1.0"
```

### Skill Patterns

Skill patterns support wildcards for flexible permission management:

- `file.read` - Specific skill
- `file.*` - All skills in file namespace
- `git.commit` - Specific git skill
- `admin.*` - All admin skills

## Validation Rules

The validator checks:

1. **Temperature**: Must be between 0 and 2
2. **Top-P**: Must be between 0 and 1
3. **Frequency/Presence Penalty**: Must be between 0 and 2
4. **Max Tokens**: Must be non-negative
5. **Timeout/Retries**: Must be non-negative
6. **Agent Names**: Alphanumeric, dashes, underscores only
7. **Skill Patterns**: Must match `namespace.action` or `namespace.*` format
8. **Skill Conflicts**: A skill cannot be in both allowed and denied lists

## Thread Safety

All manager operations are thread-safe:

```go
// Safe for concurrent use
go func() {
    for {
        manager.GetAgent("worker")
        time.Sleep(100 * time.Millisecond)
    }
}()

go func() {
    for {
        manager.SetAgent("worker", newConfig)
        time.Sleep(time.Second)
    }
}()
```

## Examples

See the `example.yaml` and `example.toml` files for complete configuration examples.

## API Reference

### Manager Methods

| Method | Description |
|--------|-------------|
| `New(config, opts...)` | Create a new manager |
| `Load(path)` | Load configuration from file |
| `Save(path)` | Save configuration to file |
| `GetGlobal()` | Get global configuration |
| `SetGlobal(global)` | Update global configuration |
| `GetAgent(name)` | Get agent config (merged with global) |
| `GetAgentRaw(name)` | Get agent config (no merging) |
| `SetAgent(name, agent)` | Set or update agent |
| `ListAgents()` | List all agent names |
| `DeleteAgent(name)` | Remove an agent |
| `Validate()` | Validate entire configuration |
| `Export()` | Export as YAML |
| `ExportAs(format)` | Export in specific format |
| `Import(data, format)` | Import from data |
| `ImportYAML(data)` | Import from YAML |
| `ImportTOML(data)` | Import from TOML |
| `ImportJSON(data)` | Import from JSON |
| `StartWatcher()` | Enable hot reload |
| `StopWatcher()` | Disable hot reload |

### Options

| Option | Description |
|--------|-------------|
| `WithOnReload(callback)` | Set reload callback |
| `WithHotReload(path)` | Enable hot reload for path |
| `WithValidationRules(rules...)` | Add custom validation rules |

## License

MIT License
