//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/EchoCog/echollama/api"
	"github.com/EchoCog/echollama/orchestration"
)

// EnhancedOrchestrationDemo demonstrates the new advanced features
func main() {
	fmt.Println("🚀 Enhanced Echollama Orchestration Demo")
	fmt.Println("========================================")
	fmt.Println("Demonstrating advanced learning and performance optimization features...")
	fmt.Println()

	// Initialize orchestration engine
	client := api.Client{}
	engine := orchestration.NewEngine(client)
	ctx := context.Background()

	// Register default tools and plugins
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	fmt.Printf("✅ Initialized engine with %d tools and %d plugins\n", 
		len(engine.GetAvailableTools()), 
		len(engine.GetAvailablePlugins()))

	// Initialize Deep Tree Echo System
	fmt.Println("\n🧠 Initializing Deep Tree Echo System...")
	err := engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize Deep Tree Echo: %v", err)
	} else {
		fmt.Println("✅ Deep Tree Echo system initialized")
	}

	// Create specialized agents
	fmt.Println("\n🤖 Creating Specialized Agents...")
	
	// Create a performance-optimized agent
	performanceAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeSpecialist, "performance-optimization")
	if err != nil {
		log.Fatalf("Failed to create performance agent: %v", err)
	}
	fmt.Printf("✅ Created performance-optimized agent: %s\n", performanceAgent.Name)

	// Create a learning-focused agent
	learningAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeReflective, "machine-learning")
	if err != nil {
		log.Fatalf("Failed to create learning agent: %v", err)
	}
	fmt.Printf("✅ Created learning-focused agent: %s\n", learningAgent.Name)

	// Create an orchestrator agent
	orchestratorAgent, err := engine.CreateSpecializedAgent(ctx, orchestration.AgentTypeOrchestrator, "advanced-coordination")
	if err != nil {
		log.Fatalf("Failed to create orchestrator agent: %v", err)
	}
	fmt.Printf("✅ Created orchestrator agent: %s\n", orchestratorAgent.Name)

	// Demonstrate optimized task execution
	fmt.Println("\n⚡ Demonstrating Performance-Optimized Task Execution...")
	
	task1 := &orchestration.Task{
		Type:  orchestration.TaskTypeTool,
		Input: "Calculate the performance impact of recent optimizations",
		Parameters: map[string]interface{}{
			"tool": map[string]interface{}{
				"name": "calculator",
				"parameters": map[string]interface{}{
					"operation": "performance_analysis",
					"metrics":   []string{"throughput", "latency", "efficiency"},
				},
			},
		},
	}

	deadline := time.Now().Add(2 * time.Minute)
	result1, err := engine.ExecuteTaskOptimized(ctx, task1, orchestration.TaskPriorityHigh, deadline)
	if err != nil {
		log.Printf("Optimized task execution failed: %v", err)
	} else {
		fmt.Printf("✅ Optimized task completed: %s\n", truncateString(result1.Output, 80))
	}

	// Demonstrate learning system
	fmt.Println("\n🧠 Demonstrating Learning System...")
	
	// Execute several tasks to build learning history
	for i := 0; i < 3; i++ {
		task := &orchestration.Task{
			ID:    fmt.Sprintf("learning-task-%d", i+1),
			Type:  orchestration.TaskTypeReflect,
			Input: fmt.Sprintf("Perform learning iteration %d: analyze patterns and improve", i+1),
		}
		
		_, err := engine.ExecuteTask(ctx, task, learningAgent)
		if err != nil {
			log.Printf("Learning task %d failed: %v", i+1, err)
		} else {
			fmt.Printf("✅ Learning task %d completed\n", i+1)
		}
	}

	// Get learning model for the agent
	learningSystem := engine.GetLearningSystem()
	model := learningSystem.GetLearningModel(learningAgent.ID)
	
	fmt.Printf("📊 Learning Model Stats:\n")
	fmt.Printf("   - Current Performance: %.1f%%\n", model.LearningTrajectory.CurrentPerformance*100)
	fmt.Printf("   - Learning Rate: %.1f%%\n", model.LearningRate*100)
	fmt.Printf("   - Specialization Areas: %v\n", model.SpecializationAreas)
	fmt.Printf("   - Collaboration Style: %s\n", model.CollaborationStyle)

	// Demonstrate agent adaptation
	fmt.Println("\n🔄 Demonstrating Agent Adaptation...")
	adaptationResult, err := engine.AdaptAgent(ctx, learningAgent.ID)
	if err != nil {
		log.Printf("Agent adaptation failed: %v", err)
	} else {
		fmt.Printf("✅ Agent adaptation completed\n")
		fmt.Printf("   - Expected Improvement: %.1f%%\n", adaptationResult.ExpectedImprovement*100)
		fmt.Printf("   - Risk Level: %s\n", adaptationResult.RiskLevel)
		if len(adaptationResult.RecommendedActions) > 0 {
			fmt.Printf("   - Recommendations: %v\n", adaptationResult.RecommendedActions)
		}
	}

	// Demonstrate optimal agent prediction
	fmt.Println("\n🎯 Demonstrating Optimal Agent Prediction...")
	predictionTask := &orchestration.Task{
		Type:  orchestration.TaskTypePlugin,
		Input: "Advanced data analysis and pattern recognition",
		Parameters: map[string]interface{}{
			"plugin_name": "data_analysis",
			"complexity":  "high",
		},
	}

	optimalAgent, confidence, err := engine.PredictOptimalAgentForTask(ctx, predictionTask)
	if err != nil {
		log.Printf("Optimal agent prediction failed: %v", err)
	} else {
		fmt.Printf("✅ Predicted optimal agent: %s (confidence: %.1f%%)\n", optimalAgent.Name, confidence*100)
	}

	// Show performance metrics
	fmt.Println("\n📈 Current System Performance Metrics:")
	systemMetrics := engine.GetSystemMetrics()
	if systemMetrics != nil {
		fmt.Printf("   - Total Tasks: %d\n", systemMetrics.TotalTasks)
		fmt.Printf("   - Completed Tasks: %d\n", systemMetrics.CompletedTasks)
		fmt.Printf("   - Failed Tasks: %d\n", systemMetrics.FailedTasks)
		fmt.Printf("   - System Health: %.1f%%\n", systemMetrics.SystemHealth*100)
		fmt.Printf("   - Throughput: %.2f TPS\n", systemMetrics.ThroughputTPS)
	}

	// Show resource usage
	fmt.Println("\n💾 Resource Usage:")
	resourceUsage := engine.GetResourceUsage()
	if len(resourceUsage) > 0 {
		for agentID, usage := range resourceUsage {
			fmt.Printf("   - Agent %s: CPU=%.1f%%, Memory=%.1fGB\n", 
				truncateString(agentID, 8), usage.CPUUsage*100, usage.MemoryUsageGB)
		}
	} else {
		fmt.Printf("   - No active resource usage\n")
	}

	// Check for any alerts
	fmt.Println("\n🚨 System Alerts:")
	alerts := engine.GetActiveAlerts()
	if len(alerts) > 0 {
		for _, alert := range alerts {
			fmt.Printf("   - %s: %s\n", alert.Severity, alert.Message)
		}
	} else {
		fmt.Printf("   - No active alerts\n")
	}

	// Multi-agent workflow demonstration
	fmt.Println("\n🤝 Demonstrating Multi-Agent Workflow...")
	
	workflow := &orchestration.ConversationWorkflow{
		ID:           "enhanced-workflow",
		Name:         "Enhanced Orchestration Workflow",
		Description:  "Demonstrate advanced multi-agent collaboration with learning integration",
		Participants: []string{orchestratorAgent.ID, performanceAgent.ID, learningAgent.ID},
		Steps: []orchestration.ConversationStep{
			{
				ID:              "step1",
				Name:            "Performance Analysis",
				FromAgentID:     orchestratorAgent.ID,
				ToAgentID:       performanceAgent.ID,
				MessageTemplate: "Analyze current system performance and identify optimization opportunities",
			},
			{
				ID:              "step2",
				Name:            "Learning Insights",
				FromAgentID:     orchestratorAgent.ID,
				ToAgentID:       learningAgent.ID,
				MessageTemplate: "Share learning insights from recent adaptations and performance improvements",
			},
			{
				ID:              "step3",
				Name:            "Synthesis",
				FromAgentID:     performanceAgent.ID,
				ToAgentID:       learningAgent.ID,
				MessageTemplate: "Synthesize performance data with learning insights for comprehensive system improvement",
			},
		},
	}

	workflowResult, err := engine.ExecuteConversationWorkflow(ctx, workflow)
	if err != nil {
		log.Printf("Workflow execution failed: %v", err)
	} else {
		fmt.Printf("✅ Multi-agent workflow completed successfully\n")
		fmt.Printf("   - Duration: %v\n", workflowResult.Duration)
		fmt.Printf("   - Success: %t\n", workflowResult.Success)
		fmt.Printf("   - Steps Completed: %d\n", len(workflowResult.StepResults))
		fmt.Printf("   - Insights Generated: %d\n", len(workflowResult.Insights))
		if len(workflowResult.Insights) > 0 {
			fmt.Printf("   - Key Insight: %s\n", workflowResult.Insights[0])
		}
	}

	fmt.Println("\n🎉 Enhanced Orchestration Demo Complete!")
	fmt.Println("The system now includes advanced learning, performance optimization,")
	fmt.Println("intelligent agent selection, resource management, and comprehensive monitoring.")
	fmt.Println()
	fmt.Println("✨ Key Enhancements Demonstrated:")
	fmt.Println("   • Machine learning-based agent improvement")
	fmt.Println("   • Performance-optimized task execution")
	fmt.Println("   • Intelligent resource management")
	fmt.Println("   • Advanced load balancing")
	fmt.Println("   • Proactive health monitoring")
	fmt.Println("   • Adaptive agent behaviors")
	fmt.Println("   • Multi-agent collaboration workflows")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
