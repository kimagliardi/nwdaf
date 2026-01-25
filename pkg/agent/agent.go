package agent

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/free5gc/nwdaf/internal/logger"
)

type AgentConfig struct {
	NefUrl                string
	PrometheusUrl         string
	AfId                  string
	Dnn                   string
	Sst                   int
	Sd                    string
	OllamaBase            string
	ModelName             string
	AutoSteerEnabled      bool
	AutoSteerInterval     int
	AutoSteerThresholdBps float64
	AutoSteerCooldown     int
}

type Agent struct {
	Config         AgentConfig
	LLM            *LLMClient
	ctx            context.Context
	cancel         context.CancelFunc
	MonitorRunning bool
	LastSteerTime  time.Time
}

func NewAgent() *Agent {
	config := AgentConfig{
		NefUrl:                getEnv("NEF_URL", "http://10.152.183.162:80"),
		PrometheusUrl:         getEnv("PROMETHEUS_URL", "http://prometheus-kube-prometheus-prometheus.monitoring:9090"),
		AfId:                  getEnv("AF_ID", "traffic-steering-agent"),
		Dnn:                   getEnv("DNN", "internet"),
		Sst:                   getEnvInt("SST", 1),
		Sd:                    getEnv("SD", "010203"),
		OllamaBase:            getEnv("OLLAMA_API_BASE", "http://192.168.0.149:11434"),
		ModelName:             getEnv("LLM_MODEL", "qwen2.5-coder"),
		AutoSteerEnabled:      getEnv("AUTO_STEER_ENABLED", "true") == "true",
		AutoSteerInterval:     getEnvInt("AUTO_STEER_INTERVAL", 30),
		AutoSteerThresholdBps: getEnvFloat("AUTO_STEER_THRESHOLD_BPS", 100000),
		AutoSteerCooldown:     getEnvInt("AUTO_STEER_COOLDOWN", 60),
	}

	return &Agent{
		Config: config,
		LLM:    NewLLMClient(config.OllamaBase, config.ModelName),
	}
}

func (a *Agent) Start(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	logger.AppLog.Infoln("🤖 Initializing Traffic Steering Agent...")

	// Start monitor
	a.StartAutoSteerMonitor()

	logger.AppLog.Infoln("✅ Agent ready with tools: get_upf_network_metrics, steer_traffic")
}

func (a *Agent) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}

// Process processes a user chat request
func (a *Agent) Process(request string) (string, error) {
	// Simple tool selection logic (simplified ReAct)
	// For now, we will just use a simple routing based on keywords,
	// or ask the LLM to generate tool calls if we wanted to be fancy.
	// But let's stick to a simpler "Chain of Thought" or direct LLM response for chat.

	// Note: The python version uses CodeAgent from smolagents which handles tool calling.
	// To keep it simple in Go without a full agent framework, we'll give the LLM context
	// about tools and execute if it asks, OR just answer if it is a general question.

	// BUT, the request might be "check metrics" or "steer to edge1".
	// Let's implement a basic command parser + LLM fallback.

	// 1. Direct tool invocation check (heuristic)
	// Because implementing a full ReAct loop in Go from scratch is out of scope for this migration,
	// we will use the LLM to ANSWER, and we can inject context into it.

	// However, if we want the agent to use tools, we should tell it about them in the prompt
	// and ask it to output a special command if it wants to use a tool.
	// For this simplified version, let's just expose the endpoints directly for tool usage
	// (like in the Python Flask app) and use the Chat endpoint mainly for Q&A or analysis.

	// Wait, the Python agent wraps the tools.
	// Let's allow the LLM to "Reason" and give a final answer.
	// If the user wants to take action, they might use the direct endpoints or we can add tool usage later.
	// For now, let's update the prompt to include current metrics context so the LLM can answer intelligently.

	metrics, _ := a.GetUPFNetworkMetrics()

	prompt := fmt.Sprintf(`You are a Traffic Steering Agent. Refuse to write code.
	
Context:
%s

User Request: %s`, metrics, request)

	return a.LLM.Query(prompt)
}

// Helpers
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if value, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return fallback
}
