# go-agent-pool

A thread-safe Go library for managing concurrent AI agents with resource limits, health monitoring, and lifecycle hooks.

## Features

- **Thread-safe operations** using `sync.RWMutex`
- **Resource limits**: MaxAgents, MemoryLimit, CPULimit
- **Health monitoring**: Periodic health checks with configurable intervals
- **Lifecycle hooks**: OnStart, OnComplete, OnError, OnPause, OnResume, OnKill callbacks
- **Agent states**: Idle, Running, Paused, Stopped, Error, Complete
- **Context support**: Each agent has its own context for cancellation
- **Simple API**: Clean and intuitive interface

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/agent-pool
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    agentpool "github.com/formatho/agent-orchestrator/packages/agent-pool"
)

func main() {
    // Create a pool with configuration
    pool := agentpool.New(agentpool.Config{
        MaxAgents:   10,
        MemoryLimit: 512 * agentpool.MB,
        CPULimit:    100, // 100% total CPU
        HealthCheckInterval: 30 * time.Second,
    })
    defer pool.Close()

    // Spawn an agent
    agent, err := pool.Spawn("agent-1", agentpool.AgentConfig{
        Name:    "Task Runner",
        Memory:  64 * agentpool.MB,
        CPU:     25,
        Timeout: 5 * time.Minute,
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Agent %s started with status: %s\n", agent.ID, agent.Status)

    // List all agents
    agents := pool.List()
    fmt.Printf("Pool has %d agents\n", len(agents))

    // Pause an agent
    pool.Pause("agent-1")
    fmt.Printf("Agent status: %s\n", agent.Status)

    // Resume an agent
    pool.Resume("agent-1")

    // Kill when done
    pool.Kill("agent-1")
}
```

## API Reference

### Pool

#### Creating a Pool

```go
// With custom configuration
pool := agentpool.New(agentpool.Config{
    MaxAgents:           10,
    MemoryLimit:         512 * agentpool.MB,
    CPULimit:            100,
    HealthCheckInterval: 30 * time.Second,
})

// With defaults
pool := agentpool.New(agentpool.DefaultConfig())
```

#### Pool Methods

```go
// Spawn a new agent
agent, err := pool.Spawn(id string, config AgentConfig) (*Agent, error)

// Kill an agent
err := pool.Kill(id string) error

// Pause a running agent
err := pool.Pause(id string) error

// Resume a paused agent
err := pool.Resume(id string) error

// List all agents
agents := pool.List() []*Agent

// Get a specific agent
agent, err := pool.Get(id string) (*Agent, error)

// Get pool size
count := pool.Size() int

// Get resource usage
memory, cpu := pool.ResourceUsage() (int64, int)

// Get pool statistics
stats := pool.GetStats() Stats

// Close the pool (stops all agents)
err := pool.Close() error

// Set lifecycle hooks
pool.SetHooks(hooks LifecycleHooks)
```

### Agent

#### Agent Properties

```go
type Agent struct {
    ID          string
    Name        string
    Status      AgentStatus
    CreatedAt   time.Time
    StartedAt   time.Time
    CompletedAt time.Time
    Config      AgentConfig
    Metadata    map[string]interface{}
    Error       error
}
```

#### Agent Methods

```go
// Get current status
status := agent.GetStatus() AgentStatus

// Check if active (running or paused)
active := agent.IsActive() bool

// Get runtime duration
runtime := agent.Runtime() time.Duration

// Health check
healthy := agent.IsHealthy() bool

// Get context for cancellation
ctx := agent.Context() context.Context

// Mark as complete
agent.Complete() error

// Set an error
agent.SetError(err error)
```

### Agent Status

```go
const (
    StatusIdle     AgentStatus = "idle"
    StatusRunning  AgentStatus = "running"
    StatusPaused   AgentStatus = "paused"
    StatusStopped  AgentStatus = "stopped"
    StatusError    AgentStatus = "error"
    StatusComplete AgentStatus = "complete"
)
```

## Examples

### With Lifecycle Hooks

```go
pool := agentpool.New(agentpool.DefaultConfig())

pool.SetHooks(agentpool.LifecycleHooks{
    OnStart: func(agent *Agent) {
        fmt.Printf("Agent %s started\n", agent.ID)
    },
    OnComplete: func(agent *Agent) {
        fmt.Printf("Agent %s completed\n", agent.ID)
    },
    OnError: func(agent *Agent, err error) {
        fmt.Printf("Agent %s error: %v\n", agent.ID, err)
    },
    OnPause: func(agent *Agent) {
        fmt.Printf("Agent %s paused\n", agent.ID)
    },
    OnResume: func(agent *Agent) {
        fmt.Printf("Agent %s resumed\n", agent.ID)
    },
    OnKill: func(agent *Agent) {
        fmt.Printf("Agent %s killed\n", agent.ID)
    },
})
```

### Resource Management

```go
// Create pool with resource limits
pool := agentpool.New(agentpool.Config{
    MaxAgents:   20,
    MemoryLimit: 1 * agentpool.GB,
    CPULimit:    200, // 200% = 2 CPU cores
})

// Spawn agents respecting limits
for i := 0; i < 20; i++ {
    agent, err := pool.Spawn(fmt.Sprintf("agent-%d", i), agentpool.AgentConfig{
        Name:   fmt.Sprintf("Worker %d", i),
        Memory: 50 * agentpool.MB,
        CPU:    10,
    })
    
    if err == agentpool.ErrPoolFull {
        fmt.Println("Pool is full!")
        break
    }
    if err == agentpool.ErrMemoryLimit {
        fmt.Println("Memory limit reached!")
        break
    }
}

// Monitor usage
mem, cpu := pool.ResourceUsage()
fmt.Printf("Using %d bytes memory, %d%% CPU\n", mem, cpu)
```

### Pool Statistics

```go
stats := pool.GetStats()

fmt.Printf("Total Agents: %d\n", stats.TotalAgents)
fmt.Printf("Running: %d\n", stats.Running)
fmt.Printf("Paused: %d\n", stats.Paused)
fmt.Printf("Idle: %d\n", stats.Idle)
fmt.Printf("Error: %d\n", stats.Error)
fmt.Printf("Complete: %d\n", stats.Complete)
fmt.Printf("Used Memory: %d bytes\n", stats.UsedMemory)
fmt.Printf("Used CPU: %d%%\n", stats.UsedCPU)
fmt.Printf("Max Agents: %d\n", stats.MaxAgents)
fmt.Printf("Memory Limit: %d bytes\n", stats.MemoryLimit)
fmt.Printf("CPU Limit: %d%%\n", stats.CPULimit)
```

### With Context Cancellation

```go
agent, _ := pool.Spawn("agent-1", agentpool.AgentConfig{
    Name:    "Task Worker",
    Timeout: 30 * time.Second,
})

// Use agent's context in your goroutine
go func() {
    select {
    case <-agent.Context().Done():
        fmt.Println("Agent was cancelled")
        return
    case <-time.After(20 * time.Second):
        agent.Complete()
    }
}()
```

### With Metadata

```go
agent, _ := pool.Spawn("agent-1", agentpool.AgentConfig{
    Name: "Data Processor",
    Metadata: map[string]interface{}{
        "task":     "process-queue",
        "priority": "high",
        "owner":    "user-123",
    },
})

// Access metadata
priority := agent.Metadata["priority"].(string)
fmt.Printf("Agent priority: %s\n", priority)
```

## Error Handling

```go
agent, err := pool.Spawn("agent-1", config)
if err != nil {
    switch err {
    case agentpool.ErrPoolFull:
        // Pool has reached max agents
    case agentpool.ErrAgentExists:
        // Agent with this ID already exists
    case agentpool.ErrMemoryLimit:
        // Memory limit exceeded
    case agentpool.ErrCPULimit:
        // CPU limit exceeded
    case agentpool.ErrPoolClosed:
        // Pool is closed
    default:
        // Other error
    }
}
```

## Testing

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

## License

MIT License
