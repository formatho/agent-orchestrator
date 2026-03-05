package agentpool

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Agent represents a single agent instance in the pool
type Agent struct {
	// ID is the unique identifier for this agent
	ID string

	// Name is a human-readable name
	Name string

	// Status is the current state of the agent
	Status AgentStatus

	// CreatedAt is when the agent was created
	CreatedAt time.Time

	// StartedAt is when the agent started running
	StartedAt time.Time

	// CompletedAt is when the agent finished
	CompletedAt time.Time

	// Config holds the agent's configuration
	Config AgentConfig

	// Metadata holds arbitrary data attached to the agent
	Metadata map[string]interface{}

	// Error holds any error the agent encountered
	Error error

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Internal state
	mu        sync.RWMutex
	healthMu  sync.RWMutex
	lastCheck time.Time
	healthy   bool
}

// NewAgent creates a new agent with the given ID and configuration
func NewAgent(id string, config AgentConfig) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		ID:        id,
		Name:      config.Name,
		Status:    StatusIdle,
		CreatedAt: time.Now(),
		Config:    config,
		Metadata:  config.Metadata,
		ctx:       ctx,
		cancel:    cancel,
		healthy:   true,
	}
}

// Start transitions the agent to running state
func (a *Agent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Status == StatusRunning {
		return errors.New("agent is already running")
	}

	if a.Status == StatusStopped {
		return errors.New("cannot start a stopped agent")
	}

	a.Status = StatusRunning
	a.StartedAt = time.Now()

	return nil
}

// Pause transitions the agent to paused state
func (a *Agent) Pause() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Status != StatusRunning {
		return errors.New("can only pause a running agent")
	}

	a.Status = StatusPaused
	return nil
}

// Resume transitions a paused agent back to running state
func (a *Agent) Resume() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Status != StatusPaused {
		return errors.New("can only resume a paused agent")
	}

	a.Status = StatusRunning
	return nil
}

// Stop stops the agent completely
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Status == StatusStopped {
		return errors.New("agent is already stopped")
	}

	if a.cancel != nil {
		a.cancel()
	}

	a.Status = StatusStopped
	a.CompletedAt = time.Now()

	return nil
}

// Complete marks the agent as successfully completed
func (a *Agent) Complete() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}

	a.Status = StatusComplete
	a.CompletedAt = time.Now()

	return nil
}

// SetError marks the agent as errored with the given error
func (a *Agent) SetError(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}

	a.Status = StatusError
	a.Error = err
	a.CompletedAt = time.Now()
}

// Context returns the agent's context for cancellation
func (a *Agent) Context() context.Context {
	return a.ctx
}

// IsHealthy returns true if the agent is healthy
func (a *Agent) IsHealthy() bool {
	a.healthMu.RLock()
	defer a.healthMu.RUnlock()
	return a.healthy
}

// SetHealthy sets the agent's health status
func (a *Agent) SetHealthy(healthy bool) {
	a.healthMu.Lock()
	defer a.healthMu.Unlock()
	a.healthy = healthy
	a.lastCheck = time.Now()
}

// LastHealthCheck returns when the agent was last health checked
func (a *Agent) LastHealthCheck() time.Time {
	a.healthMu.RLock()
	defer a.healthMu.RUnlock()
	return a.lastCheck
}

// GetStatus returns the current status (thread-safe)
func (a *Agent) GetStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status
}

// IsActive returns true if the agent is running or paused
func (a *Agent) IsActive() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status == StatusRunning || a.Status == StatusPaused
}

// Runtime returns how long the agent has been running
func (a *Agent) Runtime() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.StartedAt.IsZero() {
		return 0
	}

	endTime := a.CompletedAt
	if endTime.IsZero() {
		endTime = time.Now()
	}

	return endTime.Sub(a.StartedAt)
}
