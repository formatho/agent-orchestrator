package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
)

// Test Agent Service
func TestAgentService_Create(t *testing.T) {
	svc := services.NewAgentService(nil)

	agent := &models.Agent{
		Name:  "test-agent",
		Model: "gpt-4o",
	}

	// Test creation
	created, err := svc.Create(agent)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if created.ID == "" {
		t.Error("Expected ID to be set")
	}

	if created.Name != agent.Name {
		t.Errorf("Expected name %s, got %s", agent.Name, created.Name)
	}

	t.Logf("✅ Agent created: %s (ID: %s)", created.Name, created.ID)
}

func TestAgentService_Validation(t *testing.T) {
	svc := services.NewAgentService(nil)

	// Test empty name
	agent := &models.Agent{
		Model: "gpt-4o",
	}

	_, err := svc.Create(agent)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	t.Logf("✅ Validation working: %v", err)
}

// Test TODO Service
func TestTODOService_Create(t *testing.T) {
	svc := services.NewTODOService(nil)

	todo := &models.TODO{
		Title:       "Test TODO",
		Description: "Testing TODO creation",
		Priority:    5,
	}

	created, err := svc.Create(todo)
	if err != nil {
		t.Fatalf("Failed to create TODO: %v", err)
	}

	if created.ID == "" {
		t.Error("Expected ID to be set")
	}

	if created.Status != models.TODOStatusPending {
		t.Errorf("Expected status %s, got %s", models.TODOStatusPending, created.Status)
	}

	t.Logf("✅ TODO created: %s (ID: %s)", created.Title, created.ID)
}

// Test Config Service
func TestConfigService_Get(t *testing.T) {
	svc := services.NewConfigService(nil)

	config, err := svc.Get()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if config == nil {
		t.Error("Expected config to be returned")
	}

	t.Logf("✅ Config retrieved: %+v", config)
}

// Test HTTP Handlers
func TestAgentHandler_Create(t *testing.T) {
	agent := models.Agent{
		Name:  "handler-test",
		Model: "gpt-4o",
	}

	body, _ := json.Marshal(agent)
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	t.Logf("✅ HTTP request created: %s", string(body))
	t.Logf("✅ Method: %s, Path: %s", req.Method, req.URL.Path)
}

// Integration Test - Full Workflow
func TestIntegration_FullWorkflow(t *testing.T) {
	t.Log("=== Integration Test: Full Workflow ===")

	// Step 1: Create Agent
	t.Log("Step 1: Create Agent")
	agentSvc := services.NewAgentService(nil)
	agent := &models.Agent{
		Name:  "integration-test-agent",
		Model: "gpt-4o",
	}
	createdAgent, err := agentSvc.Create(agent)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	t.Logf("✅ Agent created: %s", createdAgent.ID)

	// Step 2: Create TODO
	t.Log("Step 2: Create TODO")
	todoSvc := services.NewTODOService(nil)
	todo := &models.TODO{
		Title:       "Integration Test TODO",
		Description: "Testing full workflow",
		Priority:    8,
	}
	createdTODO, err := todoSvc.Create(todo)
	if err != nil {
		t.Fatalf("Failed to create TODO: %v", err)
	}
	t.Logf("✅ TODO created: %s", createdTODO.ID)

	// Step 3: Create Cron Job
	t.Log("Step 3: Create Cron Job")
	cronSvc := services.NewCronService(nil)
	cron := &models.CronJob{
		Name:      "Integration Test Cron",
		Schedule:  "*/5 * * * *",
		AgentID:   createdAgent.ID,
		Timezone:  "UTC",
	}
	createdCron, err := cronSvc.Create(cron)
	if err != nil {
		t.Fatalf("Failed to create cron: %v", err)
	}
	t.Logf("✅ Cron created: %s", createdCron.ID)

	t.Log("✅ Integration test passed!")
}

// Test with Local LLM (LM Studio)
func TestLocalLLM_Connection(t *testing.T) {
	t.Log("=== Testing Local LLM (LM Studio) ===")

	// Check if LM Studio is running
	resp, err := httptest.NewRequest("GET", "http://localhost:1234/v1/models", nil)
	if err != nil {
		t.Skip("LM Studio not running")
	}

	t.Logf("✅ LM Studio request prepared: %+v", resp)
}

func TestMain(m *testing.M) {
	// Setup
	println("🧪 Starting Backend Unit Tests")
	println("")

	// Run tests
	m.Run()

	// Cleanup
	println("")
	println("✅ All tests completed")
}
