// Package agentpool provides a thread-safe pool for managing concurrent AI agents.
package agentpool

import (
	"time"
)

// Memory size constants for resource limits
const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

// AgentStatus represents the current state of an agent
type AgentStatus string

const (
	StatusIdle     AgentStatus = "idle"
	StatusRunning  AgentStatus = "running"
	StatusPaused   AgentStatus = "paused"
	StatusStopped  AgentStatus = "stopped"
	StatusError    AgentStatus = "error"
	StatusComplete AgentStatus = "complete"
)

// Config holds the pool configuration
type Config struct {
	// MaxAgents is the maximum number of agents that can be spawned
	MaxAgents int

	// MemoryLimit is the maximum total memory in bytes
	MemoryLimit int64

	// CPULimit is the maximum CPU percentage (0-100, 0 means unlimited)
	CPULimit int

	// HealthCheckInterval is the interval between health checks
	// If zero, health monitoring is disabled
	HealthCheckInterval time.Duration
}

// AgentConfig holds configuration for spawning a new agent
type AgentConfig struct {
	// Name is a human-readable name for the agent
	Name string

	// Memory is the memory allocated to this agent in bytes
	Memory int64

	// CPU is the CPU percentage allocated to this agent
	CPU int

	// Timeout is the maximum duration the agent can run
	// If zero, no timeout is applied
	Timeout time.Duration

	// Metadata allows arbitrary data to be attached to the agent
	Metadata map[string]interface{}
}

// LifecycleHooks holds callback functions for agent lifecycle events
type LifecycleHooks struct {
	// OnStart is called when an agent starts
	OnStart func(agent *Agent)

	// OnComplete is called when an agent completes successfully
	OnComplete func(agent *Agent)

	// OnError is called when an agent encounters an error
	OnError func(agent *Agent, err error)

	// OnPause is called when an agent is paused
	OnPause func(agent *Agent)

	// OnResume is called when an agent is resumed
	OnResume func(agent *Agent)

	// OnKill is called when an agent is killed
	OnKill func(agent *Agent)
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxAgents:           10,
		MemoryLimit:         512 * MB,
		CPULimit:            0, // unlimited
		HealthCheckInterval: 30 * time.Second,
	}
}

// DefaultAgentConfig returns an AgentConfig with sensible defaults
func DefaultAgentConfig(name string) AgentConfig {
	return AgentConfig{
		Name:    name,
		Memory:  64 * MB,
		CPU:     0,
		Timeout: 0,
	}
}
