package agent

import (
	"os"
	"testing"
)

func TestNewAgent(t *testing.T) {
	os.Setenv("AF_ID", "test-agent")
	defer os.Unsetenv("AF_ID")

	agent := NewAgent()

	if agent.Config.AfId != "test-agent" {
		t.Errorf("Expected AF_ID to be 'test-agent', got %s", agent.Config.AfId)
	}

	if agent.LLM == nil {
		t.Error("Expected LLM client to be initialized")
	}
}

func TestSteerTrafficValidation(t *testing.T) {
	agent := NewAgent()

	res, _ := agent.SteerTraffic("invalid")
	if res == "" {
		t.Error("Expected error message for invalid target")
	}
}
