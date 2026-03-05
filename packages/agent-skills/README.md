# go-agent-skills

A Go library for building AI agent skill execution systems with permission-based access control.

## Overview

This library provides a clean, safe way to execute AI agent skills with configurable permissions. It includes:

- **Skill Interface**: A standard interface that all skills must implement
- **Runner**: An execution engine that manages skill registration and execution
- **Permission System**: Fine-grained access control with allow/deny lists
- **Built-in Skills**: File, Web, and Shell operations ready to use

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/agent-skills
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    agent "github.com/formatho/agent-orchestrator/packages/agent-skills"
    "github.com/formatho/agent-orchestrator/packages/agent-skills/skills"
)

func main() {
    // Create a runner with permission configuration
    runner := agent.NewRunner(agent.Config{
        Allowed: []string{"file:read", "file:write", "web:*"},
        Denied:  []string{"shell:*"}, // Block all shell commands
    })
    
    // Register built-in skills
    runner.Register(skills.NewFileSkill(""))  // Empty baseDir = no restriction
    runner.Register(skills.NewWebSkill())
    runner.Register(skills.NewShellSkill())
    
    // Execute a file read
    result, err := runner.Execute(context.Background(), agent.Action{
        Skill:  "file",
        Action: "read",
        Params: map[string]any{"path": "/tmp/example.txt"},
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("File contents: %v\n", result.Data)
}
```

## Permission System

The permission system uses pattern matching to control which skills and actions can be executed.

### Permission Patterns

| Pattern | Meaning |
|---------|---------|
| `*` | Allow/deny everything |
| `file:*` | Allow/deny all actions for `file` skill |
| `file:read` | Allow/deny specific action |
| `*:delete` | Allow/deny `delete` action on all skills |

### Permission Rules

1. **Deny by default**: If `Allowed` list is empty, nothing is permitted
2. **Explicit deny overrides allow**: Patterns in `Denied` always win
3. **Order doesn't matter**: Lists are checked for matches

### Examples

```go
// Allow all file operations, deny file deletion
config := agent.Config{
    Allowed: []string{"file:*"},
    Denied:  []string{"file:delete"},
}

// Allow read-only access to all skills
config := agent.Config{
    Allowed: []string{"*:read"},
}

// Allow specific actions only
config := agent.Config{
    Allowed: []string{
        "file:read",
        "file:list",
        "web:fetch",
    },
}

// Block dangerous commands but allow everything else
config := agent.Config{
    Allowed: []string{"*"},
    Denied:  []string{"shell:run", "file:delete"},
}
```

## Built-in Skills

### FileSkill

File system operations with optional directory sandboxing.

```go
fileSkill := skills.NewFileSkill("/safe/directory") // Restrict to this directory
runner.Register(fileSkill)
```

**Actions:**
- `read` - Read file contents
  - Params: `path` (string)
- `write` - Write content to file
  - Params: `path` (string), `content` (string), `mkdir` (bool, optional)
- `delete` - Delete file or directory
  - Params: `path` (string), `recursive` (bool, optional)
- `list` - List directory contents
  - Params: `path` (string), `recursive` (bool, optional)

**Example:**
```go
// Read a file
result, _ := runner.Execute(ctx, agent.Action{
    Skill:  "file",
    Action: "read",
    Params: map[string]any{"path": "/path/to/file.txt"},
})
fmt.Println(result.Data) // File contents

// Write a file (create directories if needed)
result, _ := runner.Execute(ctx, agent.Action{
    Skill:  "file",
    Action: "write",
    Params: map[string]any{
        "path":    "/path/to/file.txt",
        "content": "Hello, World!",
        "mkdir":   true,
    },
})

// List directory contents
result, _ := runner.Execute(ctx, agent.Action{
    Skill:  "file",
    Action: "list",
    Params: map[string]any{
        "path":      "/path/to/dir",
        "recursive": true,
    },
})
// result.Data contains []map[string]any with file info
```

### WebSkill

HTTP operations for fetching web content.

```go
runner.Register(skills.NewWebSkill())
```

**Actions:**
- `fetch` - Perform HTTP GET request
  - Params: `url` (string), `headers` (map[string]string, optional)

**Example:**
```go
// Fetch a URL
result, _ := runner.Execute(ctx, agent.Action{
    Skill:  "web",
    Action: "fetch",
    Params: map[string]any{
        "url": "https://api.example.com/data",
        "headers": map[string]any{
            "Authorization": "Bearer token123",
        },
    },
})
fmt.Printf("Status: %d\n", result.Metadata["statusCode"])
fmt.Printf("Body: %s\n", result.Data)
```

### ShellSkill

Execute shell commands with safety controls.

```go
shellSkill := skills.NewShellSkill()
shellSkill.AllowedCommands = []string{"ls", "cat", "echo"} // Whitelist
shellSkill.ForbiddenCommands = []string{"rm -rf /"} // Blacklist
runner.Register(shellSkill)
```

**Actions:**
- `run` - Execute a shell command
  - Params: `command` (string), `cwd` (string, optional), `timeout` (ms, optional), `env` (map, optional)

**Example:**
```go
// Run a command
result, _ := runner.Execute(ctx, agent.Action{
    Skill:  "shell",
    Action: "run",
    Params: map[string]any{
        "command": "ls -la",
        "cwd":     "/path/to/dir",
        "timeout": 5000, // 5 seconds
    },
})
fmt.Printf("Output: %s\n", result.Data)
fmt.Printf("Exit Code: %d\n", result.Metadata["exitCode"])
```

## Creating Custom Skills

Implement the `Skill` interface:

```go
type MySkill struct{}

func (s *MySkill) Name() string {
    return "my-skill"
}

func (s *MySkill) Actions() []string {
    return []string{"do-something", "do-another-thing"}
}

func (s *MySkill) Execute(ctx context.Context, action string, params map[string]any) (agent.Result, error) {
    switch action {
    case "do-something":
        // Handle action
        return agent.Result{
            Success: true,
            Data:    "result data",
            Message: "Action completed",
        }, nil
    default:
        return agent.Result{}, agent.NewExecutionError("my-skill", action, "unknown action")
    }
}
```

## Error Handling

The library uses specific error types:

```go
result, err := runner.Execute(ctx, action)
if err != nil {
    switch {
    case errors.Is(err, &agent.SkillNotFoundError{}):
        // Skill not registered
    case errors.Is(err, &agent.PermissionDeniedError{}):
        // Action not permitted
    case errors.Is(err, &agent.ExecutionError{}):
        // Skill execution failed
    }
}
```

## Context Support

All operations support context for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := runner.Execute(ctx, action)
```

## API Reference

### Runner

```go
// Create new runner
runner := agent.NewRunner(config agent.Config)

// Register skills
runner.Register(skill agent.Skill)
runner.RegisterAll(skills ...agent.Skill)
runner.Unregister(skillName string)

// Execute actions
result, err := runner.Execute(ctx context.Context, action agent.Action)

// Query
skills := runner.ListSkills()
skill := runner.GetSkill(name string)
err := runner.CheckPermission(skillName, action string)
```

### Config

```go
config := agent.Config{
    Allowed: []string{"pattern1", "pattern2"},
    Denied:  []string{"pattern3"},
}
```

### Action

```go
action := agent.Action{
    Skill:  "skill-name",
    Action: "action-name",
    Params: map[string]any{
        "key": "value",
    },
}
```

### Result

```go
type Result struct {
    Success  bool
    Data     any
    Message  string
    Metadata map[string]any
}
```

## License

MIT License
