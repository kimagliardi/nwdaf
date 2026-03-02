package agent

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/free5gc/nwdaf/internal/logger"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/tools"
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
	Config          AgentConfig
	LLM             *LLMClient
	AgentExecutor   *agents.Executor
	ctx             context.Context
	cancel          context.CancelFunc
	MonitorRunning  bool
	LastSteerTime   time.Time
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

	agent := &Agent{
		Config: config,
		LLM:    NewLLMClient(config.OllamaBase, config.ModelName),
	}

	// Initialize LangChainGo agent with tools
	if err := agent.initLangChainAgent(); err != nil {
		logger.AppLog.Warnf("Failed to initialize LangChain agent: %v. Falling back to simple LLM.", err)
	}

	return agent
}

// initLangChainAgent initializes the LangChainGo agent with tools
func (a *Agent) initLangChainAgent() error {
	// Create Ollama LLM for LangChain
	llm, err := ollama.New(
		ollama.WithServerURL(a.Config.OllamaBase),
		ollama.WithModel(a.Config.ModelName),
	)
	if err != nil {
		return fmt.Errorf("failed to create Ollama LLM: %w", err)
	}

	// Create tools that wrap our existing functions
	agentTools := []tools.Tool{
		NewMetricsTool(a),
		NewSteerTrafficTool(a),
	}

	// Create the agent with ReAct-style reasoning
	oneShotAgent := agents.NewOneShotAgent(
		llm,
		agentTools,
		agents.WithMaxIterations(5),
	)

	// Create the executor
	a.AgentExecutor = agents.NewExecutor(oneShotAgent)

	logger.AppLog.Infoln("LangChain agent initialized with tools: get_upf_network_metrics, steer_traffic")
	return nil
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

// Process processes a user chat request using the LangChain agent
func (a *Agent) Process(request string) (string, error) {
	// Use LangChain agent if available (automatic tool calling)
	if a.AgentExecutor != nil {
		return a.processWithLangChain(request)
	}

	// Fallback to simple LLM query with context injection
	return a.processWithSimpleLLM(request)
}

// processWithLangChain uses the LangChain agent executor for automatic tool calling
func (a *Agent) processWithLangChain(request string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Enhance the request with context about what the agent can do
	enhancedRequest := fmt.Sprintf(`You are a 5G Traffic Steering Agent. You have access to tools for monitoring UPF network metrics and steering traffic.

User Request: %s

Think step by step:
1. If the user asks about metrics, traffic, or network status, use get_upf_network_metrics first
2. If the user wants to steer or redirect traffic, use steer_traffic with "edge1" or "edge2"
3. Provide a helpful response based on the tool results

Do not write code. Use the available tools to accomplish the task.`, request)

	result, err := chains.Run(ctx, a.AgentExecutor, enhancedRequest)
	if err != nil {
		logger.AppLog.Warnf("LangChain agent error: %v. Falling back to simple LLM.", err)
		return a.processWithSimpleLLM(request)
	}

	return result, nil
}

// processWithSimpleLLM uses the basic LLM client with context injection
func (a *Agent) processWithSimpleLLM(request string) (string, error) {
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
