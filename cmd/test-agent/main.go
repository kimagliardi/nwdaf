package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/free5gc/nwdaf/pkg/agent"
)

func main() {
	// Override environment for local testing
	os.Setenv("OLLAMA_API_BASE", "http://localhost:11434")
	os.Setenv("LLM_MODEL", "llama3.2:latest")

	fmt.Println("=== LangChain Agent Test ===")
	fmt.Println("Ollama URL: http://localhost:11434")
	fmt.Println("Model: llama3.2:latest")
	fmt.Println()

	// Create agent
	fmt.Println("Initializing agent...")
	a := agent.NewAgent()

	if a.AgentExecutor == nil {
		fmt.Println("WARNING: LangChain agent failed to initialize, using fallback mode")
	} else {
		fmt.Println("LangChain agent ready with tools:")
		fmt.Println("  - get_upf_network_metrics")
		fmt.Println("  - steer_traffic")
	}
	fmt.Println()

	// Interactive mode
	fmt.Println("Enter your requests (type 'quit' to exit):")
	fmt.Println("Example requests:")
	fmt.Println("  - 'What can you do?'")
	fmt.Println("  - 'Check the network metrics'")
	fmt.Println("  - 'Steer traffic to edge1'")
	fmt.Println("  - 'Hello, who are you?'")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		fmt.Println("\nProcessing...")
		response, err := a.Process(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("\n--- Response ---")
			fmt.Println(response)
			fmt.Println("----------------")
		}
		fmt.Println()
	}
}
