//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EchoCog/echollama/api"
	"github.com/EchoCog/echollama/orchestration"
)

func main() {
	fmt.Println("🌊 Deep Tree Echo API Server")
	fmt.Println("============================")
	fmt.Println()

	// Initialize orchestration engine
	client := api.Client{}
	engine := orchestration.NewEngine(client)
	ctx := context.Background()

	// Register default tools and plugins
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	// Initialize Deep Tree Echo system
	fmt.Println("🧠 Initializing Deep Tree Echo System...")
	err := engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Deep Tree Echo: %v", err)
	}
	fmt.Println("✅ Deep Tree Echo system initialized")

	// Create some default agents
	fmt.Println("🤖 Creating default agents...")
	_, err = engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeReflective, "self-analysis")
	if err != nil {
		log.Printf("Warning: Failed to create reflective agent: %v", err)
	}

	_, err = engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeOrchestrator, "coordination")
	if err != nil {
		log.Printf("Warning: Failed to create orchestrator agent: %v", err)
	}

	_, err = engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeSpecialist, "cognitive-computing")
	if err != nil {
		log.Printf("Warning: Failed to create specialist agent: %v", err)
	}

	fmt.Println("✅ Default agents created")

	// Create and start API server
	server := orchestration.NewAPIServer(engine)
	
	fmt.Println("🌐 Starting API server on port 8080...")
	fmt.Println()
	fmt.Println("📍 Available endpoints:")
	fmt.Println("   GET  /api/deep-tree-echo/status         - Get DTE system status")
	fmt.Println("   GET  /api/deep-tree-echo/dashboard      - Get dashboard data")
	fmt.Println("   POST /api/deep-tree-echo/initialize     - Initialize DTE system")
	fmt.Println("   POST /api/deep-tree-echo/diagnostics    - Run system diagnostics")
	fmt.Println("   POST /api/deep-tree-echo/refresh        - Refresh system status")
	fmt.Println("   POST /api/deep-tree-echo/introspection  - Perform introspection")
	fmt.Println()
	fmt.Println("   GET  /api/agents                        - List all agents")
	fmt.Println("   POST /api/agents                        - Create new agent")
	fmt.Println("   GET  /api/agents/:id                    - Get agent by ID")
	fmt.Println("   PUT  /api/agents/:id                    - Update agent")
	fmt.Println("   POST /api/agents/:id/tasks              - Execute task")
	fmt.Println()
	fmt.Println("   POST /api/orchestration                 - Orchestrate tasks")
	fmt.Println("   GET  /api/orchestration/tools           - Get available tools")
	fmt.Println("   GET  /api/orchestration/plugins         - Get available plugins")
	fmt.Println()
	fmt.Println("🚀 Server ready at http://localhost:8080")
	fmt.Println()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n\n🛑 Shutting down server...")
		os.Exit(0)
	}()

	// Start server
	if err := server.Run(8080); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
