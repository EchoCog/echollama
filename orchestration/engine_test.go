package orchestration

import (
	"context"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
)

func TestNewEngine(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)

	if engine == nil {
		t.Error("NewEngine should return a non-nil engine")
	}

	if engine.agents == nil {
		t.Error("Engine should have initialized agents map")
	}

	if engine.tasks == nil {
		t.Error("Engine should have initialized tasks map")
	}
}

func TestCreateAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent for unit testing",
		Models:      []string{"llama2"},
		Config:      map[string]interface{}{"key": "value"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	if agent.ID == "" {
		t.Error("Agent ID should be generated")
	}

	if agent.CreatedAt.IsZero() {
		t.Error("Agent CreatedAt should be set")
	}

	if agent.UpdatedAt.IsZero() {
		t.Error("Agent UpdatedAt should be set")
	}

	// Verify agent was stored
	stored, err := engine.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if stored.Name != agent.Name {
		t.Errorf("Expected agent name %s, got %s", agent.Name, stored.Name)
	}
}

func TestListAgents(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Initially should be empty
	agents, err := engine.ListAgents(ctx)
	if err != nil {
		t.Errorf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents))
	}

	// Create an agent
	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Models:      []string{"llama2"},
	}

	err = engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	// Now should have one agent
	agents, err = engine.ListAgents(ctx)
	if err != nil {
		t.Errorf("ListAgents failed: %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	if agents[0].Name != agent.Name {
		t.Errorf("Expected agent name %s, got %s", agent.Name, agents[0].Name)
	}
}

func TestUpdateAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create an agent first
	agent := &Agent{
		Name:        "test-agent",
		Description: "Original description",
		Models:      []string{"llama2"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	originalUpdatedAt := agent.UpdatedAt

	// Update the agent
	time.Sleep(time.Millisecond) // Ensure time difference
	agent.Description = "Updated description"
	err = engine.UpdateAgent(ctx, agent)
	if err != nil {
		t.Errorf("UpdateAgent failed: %v", err)
	}

	// Verify update
	updated, err := engine.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if updated.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", updated.Description)
	}

	if !updated.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestDeleteAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create an agent first
	agent := &Agent{
		Name:        "test-agent",
		Description: "To be deleted",
		Models:      []string{"llama2"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	// Delete the agent
	err = engine.DeleteAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("DeleteAgent failed: %v", err)
	}

	// Verify agent is gone
	_, err = engine.GetAgent(ctx, agent.ID)
	if err == nil {
		t.Error("GetAgent should have failed for deleted agent")
	}
}

func TestGetNonExistentAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	_, err := engine.GetAgent(ctx, "non-existent-id")
	if err == nil {
		t.Error("GetAgent should have failed for non-existent agent")
	}
}

func TestDeleteNonExistentAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	err := engine.DeleteAgent(ctx, "non-existent-id")
	if err == nil {
		t.Error("DeleteAgent should have failed for non-existent agent")
	}
}

func TestUpdateNonExistentAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	agent := &Agent{
		ID:          "non-existent-id",
		Name:        "test-agent",
		Description: "Test description",
		Models:      []string{"llama2"},
	}

	err := engine.UpdateAgent(ctx, agent)
	if err == nil {
		t.Error("UpdateAgent should have failed for non-existent agent")
	}
}

func TestNewAgentTypes(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Test creating different agent types
	testCases := []struct {
		agentType AgentType
		domain    string
	}{
		{AgentTypeReflective, "analysis"},
		{AgentTypeOrchestrator, "coordination"},
		{AgentTypeSpecialist, "coding"},
	}

	for _, tc := range testCases {
		agent, err := engine.CreateSpecializedAgent(ctx, tc.agentType, tc.domain)
		if err != nil {
			t.Errorf("CreateSpecializedAgent failed for type %s: %v", tc.agentType, err)
			continue
		}

		if agent.Type != tc.agentType {
			t.Errorf("Expected agent type %s, got %s", tc.agentType, agent.Type)
		}

		if agent.State == nil {
			t.Error("Agent state should be initialized")
		}

		if agent.State.Memory == nil {
			t.Error("Agent memory should be initialized")
		}
	}
}

func TestToolRegistration(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)

	// Register default tools
	RegisterDefaultTools(engine)

	tools := engine.GetAvailableTools()
	if len(tools) == 0 {
		t.Error("Expected tools to be registered")
	}

	// Check for specific tools
	expectedTools := []string{"web_search", "calculator"}
	for _, expectedTool := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool == expectedTool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool '%s' not found in registered tools", expectedTool)
		}
	}
}

func TestPluginRegistration(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)

	// Register default plugins
	RegisterDefaultPlugins(engine)

	plugins := engine.GetAvailablePlugins()
	if len(plugins) == 0 {
		t.Error("Expected plugins to be registered")
	}

	// Check for specific plugin
	expectedPlugin := "data_analysis"
	found := false
	for _, plugin := range plugins {
		if plugin == expectedPlugin {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected plugin '%s' not found in registered plugins", expectedPlugin)
	}
}

func TestEnhancedTaskExecution(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Register tools and plugins for testing
	RegisterDefaultTools(engine)
	RegisterDefaultPlugins(engine)

	// Create a general agent
	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent for enhanced features",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama2"},
		Tools:       []string{"calculator"},
	}

	err := engine.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Test tool task execution
	toolTask := &Task{
		ID:      "tool-task-1",
		Type:    TaskTypeTool,
		Input:   "Calculate 2 + 3",
		Status:  TaskStatusPending,
		AgentID: agent.ID,
		Parameters: map[string]interface{}{
			"tool": map[string]interface{}{
				"name": "calculator",
				"parameters": map[string]interface{}{
					"operation": "add",
					"a":         2.0,
					"b":         3.0,
				},
			},
		},
	}

	result, err := engine.ExecuteTask(ctx, toolTask, agent)
	if err != nil {
		t.Errorf("Tool task execution failed: %v", err)
	}

	if result == nil {
		t.Error("Tool task result should not be nil")
	}

	// Test plugin task execution
	pluginTask := &Task{
		ID:      "plugin-task-1",
		Type:    TaskTypePlugin,
		Input:   "Analyze this sample data",
		Status:  TaskStatusPending,
		AgentID: agent.ID,
		Parameters: map[string]interface{}{
			"plugin_name": "data_analysis",
			"type":        "summary",
		},
	}

	result, err = engine.ExecuteTask(ctx, pluginTask, agent)
	if err != nil {
		t.Errorf("Plugin task execution failed: %v", err)
	}

	if result == nil {
		t.Error("Plugin task result should not be nil")
	}
}

func TestReflectiveAgent(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create a reflective agent
	agent, err := engine.CreateSpecializedAgent(ctx, AgentTypeReflective, "self-analysis")
	if err != nil {
		t.Fatalf("CreateSpecializedAgent failed: %v", err)
	}

	// Test reflection task
	reflectTask := &Task{
		ID:      "reflect-task-1",
		Type:    TaskTypeReflect,
		Input:   "Analyze recent performance and learning patterns",
		Status:  TaskStatusPending,
		AgentID: agent.ID,
	}

	result, err := engine.ExecuteTask(ctx, reflectTask, agent)
	if err != nil {
		t.Errorf("Reflection task execution failed: %v", err)
	}

	if result == nil {
		t.Error("Reflection task result should not be nil")
	}

	// Check that agent state was updated
	updatedAgent, err := engine.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if updatedAgent.State == nil || len(updatedAgent.State.Context) == 0 {
		t.Error("Agent state should be updated with reflection context")
	}
}

// Tests for Multi-Agent Conversation functionality (Enhanced Echoself Integration)

func TestStartConversation(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create test agents
	agent1 := &Agent{
		Name:        "agent-1",
		Description: "Test agent 1",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama3.2"},
	}
	agent2 := &Agent{
		Name:        "agent-2", 
		Description: "Test agent 2",
		Type:        AgentTypeSpecialist,
		Models:      []string{"llama3.2"},
	}

	err := engine.CreateAgent(ctx, agent1)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	err = engine.CreateAgent(ctx, agent2)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Start conversation
	conversation, err := engine.StartConversation(ctx, []string{agent1.ID, agent2.ID}, "Test conversation")
	if err != nil {
		t.Fatalf("StartConversation failed: %v", err)
	}

	if conversation.ID == "" {
		t.Error("Conversation should have an ID")
	}

	if len(conversation.Participants) != 2 {
		t.Errorf("Expected 2 participants, got %d", len(conversation.Participants))
	}

	if conversation.Status != ConversationStatusActive {
		t.Errorf("Expected conversation status to be active, got %s", conversation.Status)
	}

	if conversation.Topic != "Test conversation" {
		t.Errorf("Expected topic 'Test conversation', got %s", conversation.Topic)
	}
}

func TestSendMessage(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create test agents
	agent1 := &Agent{
		Name:        "sender",
		Description: "Sender agent",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama3.2"},
	}
	agent2 := &Agent{
		Name:        "receiver",
		Description: "Receiver agent", 
		Type:        AgentTypeSpecialist,
		Models:      []string{"llama3.2"},
	}

	err := engine.CreateAgent(ctx, agent1)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	err = engine.CreateAgent(ctx, agent2)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Start conversation
	conversation, err := engine.StartConversation(ctx, []string{agent1.ID, agent2.ID}, "Message test")
	if err != nil {
		t.Fatalf("StartConversation failed: %v", err)
	}

	// Send message
	message := &Message{
		FromAgentID: agent1.ID,
		ToAgentID:   agent2.ID,
		Content:     "Hello, how are you?",
		Type:        MessageTypeRequest,
		Context:     map[string]interface{}{"test": true},
	}

	err = engine.SendMessage(ctx, conversation.ID, message)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Verify message was added to conversation
	updatedConversation, err := engine.GetConversation(ctx, conversation.ID)
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}

	if len(updatedConversation.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(updatedConversation.Messages))
	}

	sentMessage := updatedConversation.Messages[0]
	if sentMessage.Content != "Hello, how are you?" {
		t.Errorf("Expected message content 'Hello, how are you?', got %s", sentMessage.Content)
	}

	if sentMessage.Type != MessageTypeRequest {
		t.Errorf("Expected message type request, got %s", sentMessage.Type)
	}
}

func TestListConversations(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create test agents
	agent1 := &Agent{
		Name:        "agent-1",
		Description: "Test agent 1",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama3.2"},
	}
	agent2 := &Agent{
		Name:        "agent-2",
		Description: "Test agent 2",
		Type:        AgentTypeSpecialist,
		Models:      []string{"llama3.2"},
	}
	agent3 := &Agent{
		Name:        "agent-3",
		Description: "Test agent 3",
		Type:        AgentTypeOrchestrator,
		Models:      []string{"llama3.2"},
	}

	err := engine.CreateAgent(ctx, agent1)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	err = engine.CreateAgent(ctx, agent2)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	err = engine.CreateAgent(ctx, agent3)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Start multiple conversations
	conv1, err := engine.StartConversation(ctx, []string{agent1.ID, agent2.ID}, "Conversation 1")
	if err != nil {
		t.Fatalf("StartConversation failed: %v", err)
	}

	conv2, err := engine.StartConversation(ctx, []string{agent1.ID, agent3.ID}, "Conversation 2") 
	if err != nil {
		t.Fatalf("StartConversation failed: %v", err)
	}

	_, err = engine.StartConversation(ctx, []string{agent2.ID, agent3.ID}, "Conversation 3")
	if err != nil {
		t.Fatalf("StartConversation failed: %v", err)
	}

	// List conversations for agent1 (should have 2)
	conversations, err := engine.ListConversations(ctx, agent1.ID)
	if err != nil {
		t.Fatalf("ListConversations failed: %v", err)
	}

	if len(conversations) != 2 {
		t.Errorf("Expected 2 conversations for agent1, got %d", len(conversations))
	}

	// Verify conversation IDs
	conversationIDs := make(map[string]bool)
	for _, conv := range conversations {
		conversationIDs[conv.ID] = true
	}

	if !conversationIDs[conv1.ID] || !conversationIDs[conv2.ID] {
		t.Error("Expected conversations not found in list")
	}
}

func TestCloseConversation(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create test agents
	agent1 := &Agent{
		Name:        "agent-1",
		Description: "Test agent 1",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama3.2"},
	}
	agent2 := &Agent{
		Name:        "agent-2",
		Description: "Test agent 2",
		Type:        AgentTypeSpecialist,
		Models:      []string{"llama3.2"},
	}

	err := engine.CreateAgent(ctx, agent1)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	err = engine.CreateAgent(ctx, agent2)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Start conversation
	conversation, err := engine.StartConversation(ctx, []string{agent1.ID, agent2.ID}, "Test conversation")
	if err != nil {
		t.Fatalf("StartConversation failed: %v", err)
	}

	// Close conversation
	err = engine.CloseConversation(ctx, conversation.ID)
	if err != nil {
		t.Fatalf("CloseConversation failed: %v", err)
	}

	// Verify conversation is closed
	updatedConversation, err := engine.GetConversation(ctx, conversation.ID)
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}

	if updatedConversation.Status != ConversationStatusClosed {
		t.Errorf("Expected conversation status to be closed, got %s", updatedConversation.Status)
	}

	// Verify that sending a message to a closed conversation fails
	message := &Message{
		FromAgentID: agent1.ID,
		ToAgentID:   agent2.ID,
		Content:     "This should fail",
		Type:        MessageTypeRequest,
	}

	err = engine.SendMessage(ctx, conversation.ID, message)
	if err == nil {
		t.Error("Expected sending message to closed conversation to fail")
	}
}

func TestExecuteConversationWorkflow(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create test agents
	orchestrator := &Agent{
		Name:        "orchestrator",
		Description: "Orchestrator agent",
		Type:        AgentTypeOrchestrator,
		Models:      []string{"llama3.2"},
	}
	specialist := &Agent{
		Name:        "specialist",
		Description: "Specialist agent",
		Type:        AgentTypeSpecialist,
		Models:      []string{"llama3.2"},
	}

	err := engine.CreateAgent(ctx, orchestrator)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	err = engine.CreateAgent(ctx, specialist)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Create workflow
	workflow := &ConversationWorkflow{
		ID:          "test-workflow",
		Name:        "Test Workflow",
		Description: "A test conversation workflow",
		Participants: []string{orchestrator.ID, specialist.ID},
		Steps: []ConversationStep{
			{
				ID:              "step1",
				Name:            "Initial Request",
				FromAgentID:     orchestrator.ID,
				ToAgentID:       specialist.ID,
				MessageTemplate: "Please analyze: {{task}}",
				Parameters: map[string]interface{}{
					"task": "test data analysis",
				},
			},
			{
				ID:              "step2", 
				Name:            "Follow-up",
				FromAgentID:     orchestrator.ID,
				ToAgentID:       specialist.ID,
				MessageTemplate: "Can you provide more details about {{aspect}}?",
				Parameters: map[string]interface{}{
					"aspect": "the results",
				},
			},
		},
		Status: ConversationStatusActive,
	}

	// Execute workflow
	result, err := engine.ExecuteConversationWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("ExecuteConversationWorkflow failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected workflow to succeed, got: %s", result.Error)
	}

	if len(result.StepResults) != 2 {
		t.Errorf("Expected 2 step results, got %d", len(result.StepResults))
	}

	if len(result.Insights) != 2 {
		t.Errorf("Expected 2 insights, got %d", len(result.Insights))
	}

	// Verify message template processing
	step1Result := result.StepResults[0]
	if step1Result.Message.Content != "Please analyze: test data analysis" {
		t.Errorf("Expected processed template, got: %s", step1Result.Message.Content)
	}
}

func TestGetConversationMetrics(t *testing.T) {
	client := api.Client{}
	engine := NewEngine(client)
	ctx := context.Background()

	// Create test agents
	agent1 := &Agent{
		Name:        "agent-1",
		Description: "Test agent 1",
		Type:        AgentTypeGeneral,
		Models:      []string{"llama3.2"},
	}
	agent2 := &Agent{
		Name:        "agent-2",
		Description: "Test agent 2",
		Type:        AgentTypeSpecialist,
		Models:      []string{"llama3.2"},
	}

	err := engine.CreateAgent(ctx, agent1)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	err = engine.CreateAgent(ctx, agent2)
	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	// Start conversation and send messages
	conversation, err := engine.StartConversation(ctx, []string{agent1.ID, agent2.ID}, "Metrics test")
	if err != nil {
		t.Fatalf("StartConversation failed: %v", err)
	}

	// Send a few messages
	messages := []*Message{
		{
			FromAgentID: agent1.ID,
			ToAgentID:   agent2.ID,
			Content:     "Hello",
			Type:        MessageTypeRequest,
		},
		{
			FromAgentID: agent2.ID,
			ToAgentID:   agent1.ID,
			Content:     "Hi there",
			Type:        MessageTypeResponse,
		},
		{
			FromAgentID: agent1.ID,
			ToAgentID:   agent2.ID,
			Content:     "How are you?",
			Type:        MessageTypeNotification,
		},
	}

	for _, message := range messages {
		err = engine.SendMessage(ctx, conversation.ID, message)
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}
	}

	// Get metrics
	metrics := engine.GetConversationMetrics(ctx)

	if metrics["total_conversations"] != 1 {
		t.Errorf("Expected 1 total conversation, got %v", metrics["total_conversations"])
	}

	if metrics["active_conversations"] != 1 {
		t.Errorf("Expected 1 active conversation, got %v", metrics["active_conversations"])
	}

	if metrics["total_messages"] != 3 {
		t.Errorf("Expected 3 total messages, got %v", metrics["total_messages"])
	}

	messageTypes := metrics["message_types"].(map[MessageType]int)
	if messageTypes[MessageTypeRequest] != 1 {
		t.Errorf("Expected 1 request message, got %d", messageTypes[MessageTypeRequest])
	}
	if messageTypes[MessageTypeResponse] != 1 {
		t.Errorf("Expected 1 response message, got %d", messageTypes[MessageTypeResponse])
	}
	if messageTypes[MessageTypeNotification] != 1 {
		t.Errorf("Expected 1 notification message, got %d", messageTypes[MessageTypeNotification])
	}

	agentParticipation := metrics["agent_participation"].(map[string]int)
	if agentParticipation[agent1.ID] != 1 {
		t.Errorf("Expected agent1 to participate in 1 conversation, got %d", agentParticipation[agent1.ID])
	}
	if agentParticipation[agent2.ID] != 1 {
		t.Errorf("Expected agent2 to participate in 1 conversation, got %d", agentParticipation[agent2.ID])
	}
}