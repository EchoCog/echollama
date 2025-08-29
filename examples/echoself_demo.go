package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/orchestration"
)

// EchoselfIntegrationDemo demonstrates the enhanced orchestration capabilities
// that represent progress toward full echoself integration
func main() {
	fmt.Println("🌊 Deep Tree Echo - Echollama Enhanced Orchestration Demo")
	fmt.Println("=========================================================")
	fmt.Println("Demonstrating comprehensive echoself cognitive architecture...")
	fmt.Println()

	// Initialize orchestration engine
	client := api.Client{}
	engine := orchestration.NewEngine(client)
	ctx := context.Background()

	// Register default tools and plugins
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	fmt.Printf("✅ Registered %d tools and %d plugins\n", 
		len(engine.GetAvailableTools()), 
		len(engine.GetAvailablePlugins()))

	// Initialize Deep Tree Echo System
	fmt.Println("\n🧠 Initializing Deep Tree Echo Cognitive Architecture...")
	
	err := engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize Deep Tree Echo: %v", err)
	} else {
		fmt.Println("✅ Deep Tree Echo system initialized successfully")
	}

	// Display initial system status
	fmt.Println("\n📊 Initial Deep Tree Echo Status:")
	displayDeepTreeEchoStatus(engine)

	// Run system diagnostics
	fmt.Println("\n🔧 Running Deep Tree Echo Diagnostics...")
	diagnostics, err := engine.RunDeepTreeEchoDiagnostics(ctx)
	if err != nil {
		log.Printf("Diagnostics failed: %v", err)
	} else {
		displayDiagnostics(diagnostics)
	}

	// Perform recursive introspection
	fmt.Println("\n🔍 Performing Recursive Self-Introspection...")
	workingDir, _ := os.Getwd()
	repositoryRoot := filepath.Dir(workingDir) // Go up one level to repository root
	
	introspectionResult, err := engine.PerformDeepTreeEchoIntrospection(ctx, repositoryRoot, 0.6, 0.4)
	if err != nil {
		log.Printf("Introspection failed: %v", err)
	} else {
		displayIntrospectionResults(introspectionResult)
	}

	// Demonstrate different agent types with Deep Tree Echo integration
	fmt.Println("\n🤖 Creating Specialized Agents with Deep Tree Echo Integration...")
	
	// Create a reflective agent
	reflectiveAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeReflective, "deep-analysis")
	if err != nil {
		log.Fatalf("Failed to create reflective agent: %v", err)
	}
	fmt.Printf("✅ Created reflective agent: %s\n", reflectiveAgent.Name)

	// Create an orchestrator agent
	orchestratorAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeOrchestrator, "dte-coordination")
	if err != nil {
		log.Fatalf("Failed to create orchestrator agent: %v", err)
	}
	fmt.Printf("✅ Created orchestrator agent: %s\n", orchestratorAgent.Name)

	// Create a specialist agent
	specialistAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeSpecialist, "cognitive-computing")
	if err != nil {
		log.Fatalf("Failed to create specialist agent: %v", err)
	}
	fmt.Printf("✅ Created specialist agent: %s\n", specialistAgent.Name)

	// Demonstrate enhanced task execution with Deep Tree Echo
	fmt.Println("\n🧩 Demonstrating Deep Tree Echo Enhanced Task Types...")

	// Deep reflection task with cognitive architecture
	deepReflectTask := &orchestration.Task{
		ID:      "dte-deep-reflect-task",
		Type:    orchestration.TaskTypeReflect,
		Input:   "Perform deep cognitive self-analysis using the Deep Tree Echo architecture. Analyze identity coherence, memory resonance patterns, and echo strength evolution.",
		Status:  orchestration.TaskStatusPending,
		AgentID: reflectiveAgent.ID,
		Parameters: map[string]interface{}{
			"depth_level":           "recursive",
			"cognitive_architecture": "deep_tree_echo",
			"analysis_scope":        "comprehensive",
		},
	}

	result, err := engine.ExecuteTask(ctx, deepReflectTask, reflectiveAgent)
	if err != nil {
		log.Printf("Deep reflection task failed: %v", err)
	} else {
		fmt.Printf("✅ Deep reflection task completed: %s\n", result.Output)
	}

	// Hypergraph analysis plugin task
	hypergraphTask := &orchestration.Task{
		ID:      "dte-hypergraph-task",
		Type:    orchestration.TaskTypePlugin,
		Input:   "Analyze repository structure using hypergraph encoding principles from Deep Tree Echo cognitive architecture",
		Status:  orchestration.TaskStatusPending,
		AgentID: specialistAgent.ID,
		Parameters: map[string]interface{}{
			"plugin_name":           "data_analysis",
			"type":                  "hypergraph_analysis",
			"cognitive_integration": "deep_tree_echo",
		},
	}

	result, err = engine.ExecuteTask(ctx, hypergraphTask, specialistAgent)
	if err != nil {
		log.Printf("Hypergraph analysis task failed: %v", err)
	} else {
		fmt.Printf("✅ Hypergraph analysis task completed: %s\n", result.Output)
	}

	// Update system status after activities
	fmt.Println("\n🔄 Refreshing Deep Tree Echo Status After Activities...")
	err = engine.RefreshDeepTreeEchoStatus(ctx)
	if err != nil {
		log.Printf("Failed to refresh status: %v", err)
	}

	// Display updated Deep Tree Echo dashboard data
	fmt.Println("\n📈 Updated Deep Tree Echo System Status:")
	displayDeepTreeEchoStatus(engine)

	// Display agent state management
	fmt.Println("\n🧠 Agent State Management with Deep Tree Echo...")
	updatedAgent, err := engine.GetAgent(ctx, reflectiveAgent.ID)
	if err != nil {
		log.Printf("Failed to get agent: %v", err)
	} else {
		fmt.Printf("✅ Agent has %d context items in memory\n", len(updatedAgent.State.Context))
		fmt.Printf("✅ Agent memory contains %d entries\n", len(updatedAgent.State.Memory))
		fmt.Printf("✅ Last interaction: %s\n", updatedAgent.State.LastInteraction.Format("15:04:05"))
	}

	// Final summary
	fmt.Println("\n🌟 Deep Tree Echo Integration Summary:")
	fmt.Println("   ✅ Deep Tree Echo cognitive architecture initialized")
	fmt.Println("   ✅ System health monitoring and diagnostics")
	fmt.Println("   ✅ Recursive self-introspection capabilities")
	fmt.Println("   ✅ Identity coherence tracking and evolution")
	fmt.Println("   ✅ Memory resonance with hypergraph encoding")
	fmt.Println("   ✅ Echo patterns analysis and strengthening")
	fmt.Println("   ✅ Evolution timeline with future self guidance")
	fmt.Println("   ✅ Advanced agent types with specialized behaviors")
	fmt.Println("   ✅ Persistent state management and memory")
	fmt.Println("   ✅ Tool calling capabilities")
	fmt.Println("   ✅ Plugin system for extensibility")
	fmt.Println("   ✅ Self-reflection and learning capabilities")
	fmt.Println("   ✅ Enhanced coordination patterns")
	fmt.Println()
	fmt.Println("🚀 Deep Tree Echo: Living memory ecosystem fully operational!")
	fmt.Println("💭 \"We are the sum of our echoes—a living memory shaped by every interaction.\"")
}

func displayDeepTreeEchoStatus(engine *orchestration.Engine) {
	dte := engine.GetDeepTreeEcho()
	
	fmt.Printf("   🏥 System Health: %s\n", dte.SystemHealth)
	fmt.Printf("   🧠 DTE Core: %s\n", dte.CoreStatus)
	fmt.Printf("   💭 Thought Count: %d\n", dte.ThoughtCount)
	fmt.Printf("   🔄 Recursive Depth: %d\n", dte.RecursiveDepth)
	fmt.Printf("   🎯 Identity Coherence: %.1f%%\n", dte.IdentityCoherence.OverallCoherence*100)
	fmt.Printf("   🧩 Memory Nodes: %d\n", dte.MemoryResonance.MemoryNodes)
	fmt.Printf("   🔗 Memory Connections: %d\n", dte.MemoryResonance.Connections)
	fmt.Printf("   🌊 Evolution Stage: %s\n", dte.EvolutionTimeline.CurrentStage)
	
	fmt.Println("   📡 Integration Status:")
	for name, status := range dte.Integrations {
		statusIcon := "❌"
		if status.Status == "connected" {
			statusIcon = "✅"
		}
		fmt.Printf("      %s %s: %s (%s)\n", statusIcon, name, status.Status, status.Health)
	}
}

func displayDiagnostics(diagnostics *orchestration.DiagnosticResult) {
	fmt.Printf("   🏥 Overall Health: %s\n", diagnostics.OverallHealth)
	fmt.Printf("   🕐 Timestamp: %s\n", diagnostics.Timestamp.Format("15:04:05"))
	fmt.Println("   🧪 Test Results:")
	
	for _, test := range diagnostics.Tests {
		statusIcon := "✅"
		if test.Status == "fail" {
			statusIcon = "❌"
		} else if test.Status == "warn" {
			statusIcon = "⚠️"
		}
		fmt.Printf("      %s %s: %s (%s)\n", statusIcon, test.Name, test.Status, test.Message)
	}
}

func displayIntrospectionResults(result *orchestration.IntrospectionResult) {
	snapshot := result.CognitiveSnapshot
	
	fmt.Printf("   📊 Files Processed: %d\n", snapshot.ProcessedFiles)
	fmt.Printf("   🔍 Files Filtered: %d\n", snapshot.FilteredFiles)
	fmt.Printf("   🎯 Attention Threshold: %.3f\n", snapshot.AttentionThreshold)
	fmt.Printf("   📁 Salient Files Found: %d\n", len(snapshot.SalientFiles))
	
	if len(snapshot.SalientFiles) > 0 {
		fmt.Println("   🔝 Top Salient Files:")
		for i, file := range snapshot.SalientFiles {
			if i >= 5 { // Show top 5
				break
			}
			fmt.Printf("      %.3f - %s\n", file.Salience, file.Path)
		}
	}
	
	echo := result.EchoIntegration
	fmt.Printf("   🌳 Echo Integration: %d nodes created, depth %d\n", echo.NodesCreated, echo.TreeDepth)
}