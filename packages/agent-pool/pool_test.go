package agentpool

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	config := Config{
		MaxAgents:   10,
		MemoryLimit: 512 * MB,
	}

	pool := New(config)
	if pool == nil {
		t.Fatal("expected pool to be created")
	}

	if pool.config.MaxAgents != 10 {
		t.Errorf("expected MaxAgents to be 10, got %d", pool.config.MaxAgents)
	}

	// Clean up
	pool.Close()
}

func TestNewWithDefaults(t *testing.T) {
	pool := New(Config{})
	if pool == nil {
		t.Fatal("expected pool to be created with defaults")
	}

	if pool.config.MaxAgents <= 0 {
		t.Error("expected MaxAgents to be set to default")
	}

	pool.Close()
}

func TestSpawn(t *testing.T) {
	pool := New(Config{
		MaxAgents:   5,
		MemoryLimit: 256 * MB,
	})
	defer pool.Close()

	agent, err := pool.Spawn("agent-1", AgentConfig{
		Name:   "Test Agent",
		Memory: 64 * MB,
	})
	if err != nil {
		t.Fatalf("failed to spawn agent: %v", err)
	}

	if agent == nil {
		t.Fatal("expected agent to be returned")
	}

	if agent.ID != "agent-1" {
		t.Errorf("expected agent ID to be 'agent-1', got %s", agent.ID)
	}

	if agent.Name != "Test Agent" {
		t.Errorf("expected agent name to be 'Test Agent', got %s", agent.Name)
	}

	if agent.Status != StatusRunning {
		t.Errorf("expected agent status to be running, got %s", agent.Status)
	}
}

func TestSpawnDuplicate(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	_, err := pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))
	if err != nil {
		t.Fatalf("failed to spawn first agent: %v", err)
	}

	_, err = pool.Spawn("agent-1", DefaultAgentConfig("Agent 1 Duplicate"))
	if err != ErrAgentExists {
		t.Errorf("expected ErrAgentExists, got %v", err)
	}
}

func TestSpawnMemoryLimit(t *testing.T) {
	pool := New(Config{
		MaxAgents:   10,
		MemoryLimit: 128 * MB,
	})
	defer pool.Close()

	// First agent should succeed
	_, err := pool.Spawn("agent-1", AgentConfig{
		Name:   "Agent 1",
		Memory: 64 * MB,
	})
	if err != nil {
		t.Fatalf("failed to spawn first agent: %v", err)
	}

	// Second agent should succeed (64 + 64 = 128)
	_, err = pool.Spawn("agent-2", AgentConfig{
		Name:   "Agent 2",
		Memory: 64 * MB,
	})
	if err != nil {
		t.Fatalf("failed to spawn second agent: %v", err)
	}

	// Third agent should fail (would exceed 128 MB limit)
	_, err = pool.Spawn("agent-3", AgentConfig{
		Name:   "Agent 3",
		Memory: 64 * MB,
	})
	if err != ErrMemoryLimit {
		t.Errorf("expected ErrMemoryLimit, got %v", err)
	}
}

func TestSpawnMaxAgents(t *testing.T) {
	pool := New(Config{
		MaxAgents:   2,
		MemoryLimit: 1 * GB,
	})
	defer pool.Close()

	// Spawn first two agents
	_, err := pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))
	if err != nil {
		t.Fatalf("failed to spawn agent-1: %v", err)
	}

	_, err = pool.Spawn("agent-2", DefaultAgentConfig("Agent 2"))
	if err != nil {
		t.Fatalf("failed to spawn agent-2: %v", err)
	}

	// Third should fail
	_, err = pool.Spawn("agent-3", DefaultAgentConfig("Agent 3"))
	if err != ErrPoolFull {
		t.Errorf("expected ErrPoolFull, got %v", err)
	}
}

func TestKill(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	_, err := pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))
	if err != nil {
		t.Fatalf("failed to spawn agent: %v", err)
	}

	err = pool.Kill("agent-1")
	if err != nil {
		t.Fatalf("failed to kill agent: %v", err)
	}

	// Verify agent is gone
	_, err = pool.Get("agent-1")
	if err != ErrAgentNotFound {
		t.Errorf("expected ErrAgentNotFound after kill, got %v", err)
	}
}

func TestKillNotFound(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	err := pool.Kill("nonexistent")
	if err != ErrAgentNotFound {
		t.Errorf("expected ErrAgentNotFound, got %v", err)
	}
}

func TestList(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	// Start with empty list
	list := pool.List()
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d agents", len(list))
	}

	// Add agents
	_, _ = pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))
	_, _ = pool.Spawn("agent-2", DefaultAgentConfig("Agent 2"))
	_, _ = pool.Spawn("agent-3", DefaultAgentConfig("Agent 3"))

	list = pool.List()
	if len(list) != 3 {
		t.Errorf("expected 3 agents, got %d", len(list))
	}
}

func TestGet(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	_, _ = pool.Spawn("agent-1", AgentConfig{
		Name:     "Test Agent",
		Memory:   64 * MB,
		Metadata: map[string]interface{}{"key": "value"},
	})

	agent, err := pool.Get("agent-1")
	if err != nil {
		t.Fatalf("failed to get agent: %v", err)
	}

	if agent.ID != "agent-1" {
		t.Errorf("expected agent ID 'agent-1', got %s", agent.ID)
	}

	if agent.Name != "Test Agent" {
		t.Errorf("expected agent name 'Test Agent', got %s", agent.Name)
	}
}

func TestGetNotFound(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	_, err := pool.Get("nonexistent")
	if err != ErrAgentNotFound {
		t.Errorf("expected ErrAgentNotFound, got %v", err)
	}
}

func TestPauseResume(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	_, _ = pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))

	// Pause
	err := pool.Pause("agent-1")
	if err != nil {
		t.Fatalf("failed to pause agent: %v", err)
	}

	agent, _ := pool.Get("agent-1")
	if agent.Status != StatusPaused {
		t.Errorf("expected status paused, got %s", agent.Status)
	}

	// Resume
	err = pool.Resume("agent-1")
	if err != nil {
		t.Fatalf("failed to resume agent: %v", err)
	}

	agent, _ = pool.Get("agent-1")
	if agent.Status != StatusRunning {
		t.Errorf("expected status running, got %s", agent.Status)
	}
}

func TestPauseNotFound(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	err := pool.Pause("nonexistent")
	if err != ErrAgentNotFound {
		t.Errorf("expected ErrAgentNotFound, got %v", err)
	}
}

func TestResourceUsage(t *testing.T) {
	pool := New(Config{
		MaxAgents:   10,
		MemoryLimit: 512 * MB,
		CPULimit:    100,
	})
	defer pool.Close()

	// Initially zero
	mem, cpu := pool.ResourceUsage()
	if mem != 0 || cpu != 0 {
		t.Errorf("expected initial usage to be 0, got mem=%d, cpu=%d", mem, cpu)
	}

	// Spawn agents
	_, _ = pool.Spawn("agent-1", AgentConfig{
		Name:   "Agent 1",
		Memory: 64 * MB,
		CPU:    25,
	})

	_, _ = pool.Spawn("agent-2", AgentConfig{
		Name:   "Agent 2",
		Memory: 128 * MB,
		CPU:    30,
	})

	mem, cpu = pool.ResourceUsage()
	if mem != 192*MB {
		t.Errorf("expected memory usage 192MB, got %d", mem)
	}

	if cpu != 55 {
		t.Errorf("expected CPU usage 55, got %d", cpu)
	}
}

func TestSize(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	if pool.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", pool.Size())
	}

	_, _ = pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))
	_, _ = pool.Spawn("agent-2", DefaultAgentConfig("Agent 2"))

	if pool.Size() != 2 {
		t.Errorf("expected size 2, got %d", pool.Size())
	}

	_ = pool.Kill("agent-1")

	if pool.Size() != 1 {
		t.Errorf("expected size 1 after kill, got %d", pool.Size())
	}
}

func TestClose(t *testing.T) {
	pool := New(DefaultConfig())

	_, _ = pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))
	_, _ = pool.Spawn("agent-2", DefaultAgentConfig("Agent 2"))

	err := pool.Close()
	if err != nil {
		t.Fatalf("failed to close pool: %v", err)
	}

	if pool.Size() != 0 {
		t.Errorf("expected pool to be empty after close, got %d agents", pool.Size())
	}

	// Should be able to close multiple times without error
	err = pool.Close()
	if err != nil {
		t.Errorf("expected no error on second close, got %v", err)
	}
}

func TestSpawnAfterClose(t *testing.T) {
	pool := New(DefaultConfig())
	pool.Close()

	_, err := pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))
	if err != ErrPoolClosed {
		t.Errorf("expected ErrPoolClosed, got %v", err)
	}
}

func TestGetStats(t *testing.T) {
	pool := New(Config{
		MaxAgents:   10,
		MemoryLimit: 512 * MB,
		CPULimit:    100,
	})
	defer pool.Close()

	_, _ = pool.Spawn("agent-1", AgentConfig{Name: "Agent 1", Memory: 64 * MB, CPU: 20})
	_, _ = pool.Spawn("agent-2", AgentConfig{Name: "Agent 2", Memory: 64 * MB, CPU: 20})
	_ = pool.Pause("agent-2")

	stats := pool.GetStats()

	if stats.TotalAgents != 2 {
		t.Errorf("expected 2 total agents, got %d", stats.TotalAgents)
	}

	if stats.Running != 1 {
		t.Errorf("expected 1 running agent, got %d", stats.Running)
	}

	if stats.Paused != 1 {
		t.Errorf("expected 1 paused agent, got %d", stats.Paused)
	}

	if stats.UsedMemory != 128*MB {
		t.Errorf("expected 128MB used, got %d", stats.UsedMemory)
	}

	if stats.MaxAgents != 10 {
		t.Errorf("expected MaxAgents 10, got %d", stats.MaxAgents)
	}
}

func TestLifecycleHooks(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	var onStartCalled, onKillCalled bool
	var pausedAgent *Agent

	pool.SetHooks(LifecycleHooks{
		OnStart: func(agent *Agent) {
			onStartCalled = true
		},
		OnPause: func(agent *Agent) {
			pausedAgent = agent
		},
		OnKill: func(agent *Agent) {
			onKillCalled = true
		},
	})

	_, _ = pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))

	// Wait for goroutine to execute
	time.Sleep(10 * time.Millisecond)

	if !onStartCalled {
		t.Error("expected OnStart hook to be called")
	}

	_ = pool.Pause("agent-1")
	if pausedAgent == nil || pausedAgent.ID != "agent-1" {
		t.Error("expected OnPause hook to be called with correct agent")
	}

	_ = pool.Kill("agent-1")
	if !onKillCalled {
		t.Error("expected OnKill hook to be called")
	}
}

func TestHealthMonitoring(t *testing.T) {
	pool := New(Config{
		MaxAgents:           5,
		HealthCheckInterval: 100 * time.Millisecond,
	})
	defer pool.Close()

	agent, _ := pool.Spawn("agent-1", AgentConfig{
		Name:    "Agent 1",
		Timeout: 200 * time.Millisecond,
	})

	// Agent should be healthy initially
	if !agent.IsHealthy() {
		t.Error("expected agent to be healthy initially")
	}

	// Wait for health check to run
	time.Sleep(150 * time.Millisecond)

	// Health check should have updated lastCheck
	if agent.LastHealthCheck().IsZero() {
		t.Error("expected health check to have run")
	}
}

func TestAgentRuntime(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	agent, _ := pool.Spawn("agent-1", DefaultAgentConfig("Agent 1"))

	// Runtime should be positive
	time.Sleep(10 * time.Millisecond)
	runtime := agent.Runtime()
	if runtime <= 0 {
		t.Errorf("expected positive runtime, got %v", runtime)
	}

	// After stopping, runtime should be fixed
	_ = pool.Kill("agent-1")
	runtimeAfterKill := agent.Runtime()
	time.Sleep(20 * time.Millisecond)

	// Runtime should not increase after kill
	if agent.Runtime() > runtimeAfterKill+5*time.Millisecond {
		t.Error("runtime should not increase after agent is stopped")
	}
}
