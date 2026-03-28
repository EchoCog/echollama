//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/EchoCog/echollama/api"
	"github.com/EchoCog/echollama/orchestration"
)

// EchoChatDemo demonstrates the shell integration capabilities
func main() {
	fmt.Println("🌊 EchoChat - Deep Tree Echo Shell Integration Demo")
	fmt.Println("==================================================")
	fmt.Println("Demonstrating natural language to shell command translation...")
	fmt.Println()

	// Initialize orchestration engine
	client := api.Client{}
	engine := orchestration.NewEngine(client)
	ctx := context.Background()

	// Register default tools and plugins
	orchestration.RegisterDefaultTools(engine)
	orchestration.RegisterDefaultPlugins(engine)

	fmt.Printf("✅ Initialized orchestration engine\n")

	// Initialize Deep Tree Echo System
	fmt.Println("🧠 Initializing Deep Tree Echo Cognitive Architecture...")
	
	err := engine.InitializeDeepTreeEcho(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize Deep Tree Echo: %v", err)
	} else {
		fmt.Println("✅ Deep Tree Echo system initialized successfully")
	}

	// Create EchoChat instance
	echoChat := orchestration.NewEchoChat(engine)
	
	fmt.Println("\n🔧 Running EchoChat Demo Commands...")
	
	// Demo some commands
	demoCommands := []string{
		"list files in current directory",
		"show current directory",
		"check disk space",
		"show system information",
		"find all text files",
	}
	
	for _, cmd := range demoCommands {
		fmt.Printf("\n🗣️  Input: '%s'\n", cmd)
		fmt.Print("⚙️  Processing with Deep Tree Echo...")
		
		err := echoChat.ProcessInput(ctx, cmd)
		if err != nil {
			fmt.Printf(" ❌ Error: %v\n", err)
		} else {
			fmt.Println(" ✅ Completed")
		}
		fmt.Println(strings.Repeat("-", 50))
	}
	
	// Show command history
	fmt.Println("\n📜 Command History:")
	history := echoChat.GetHistory()
	for i, cmd := range history {
		status := "✅"
		if cmd.ExitCode != 0 {
			status = "❌"
		}
		fmt.Printf("%d. %s '%s' -> '%s' (%.2fs)\n", 
			i+1, status, cmd.Input, cmd.Command, cmd.Duration.Seconds())
	}

	fmt.Println("\n🌊 EchoChat Demo Complete!")
	fmt.Println("To start interactive mode, run: go run examples/echochat_interactive.go")
}
