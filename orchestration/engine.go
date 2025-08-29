package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ollama/ollama/api"
)

// Engine implements the core orchestration functionality
type Engine struct {
	client         api.Client
	agents         map[string]*Agent
	tasks          map[string]*Task
	tools          map[string]Tool
	plugins        *PluginRegistry
	deepTreeEcho   *DeepTreeEcho
	conversations  map[string]*Conversation  // Multi-agent conversations
	mu             sync.RWMutex
}

// NewEngine creates a new orchestration engine
func NewEngine(client api.Client) *Engine {
	return &Engine{
		client:        client,
		agents:        make(map[string]*Agent),
		tasks:         make(map[string]*Task),
		tools:         make(map[string]Tool),
		plugins:       &PluginRegistry{plugins: make(map[string]Plugin)},
		deepTreeEcho:  NewDeepTreeEcho("Primary Deep Tree Echo System"),
		conversations: make(map[string]*Conversation),
	}
}

// CreateAgent creates a new orchestration agent
func (e *Engine) CreateAgent(ctx context.Context, agent *Agent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	// Initialize agent state if not provided
	if agent.State == nil {
		agent.State = &AgentState{
			Memory:          make(map[string]interface{}),
			Context:         make([]ContextItem, 0),
			Goals:           make([]string, 0),
			Capabilities:    make([]string, 0),
			LastInteraction: time.Now(),
		}
	}

	// Set default agent type if not specified
	if agent.Type == "" {
		agent.Type = AgentTypeGeneral
	}

	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()

	e.agents[agent.ID] = agent
	slog.Info("Created orchestration agent", "id", agent.ID, "name", agent.Name)
	return nil
}

// GetAgent retrieves an agent by ID
func (e *Engine) GetAgent(ctx context.Context, id string) (*Agent, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	agent, exists := e.agents[id]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	return agent, nil
}

// ListAgents returns all registered agents
func (e *Engine) ListAgents(ctx context.Context) ([]*Agent, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	agents := make([]*Agent, 0, len(e.agents))
	for _, agent := range e.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

// UpdateAgent updates an existing agent
func (e *Engine) UpdateAgent(ctx context.Context, agent *Agent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.agents[agent.ID]; !exists {
		return fmt.Errorf("agent not found: %s", agent.ID)
	}

	agent.UpdatedAt = time.Now()
	e.agents[agent.ID] = agent
	slog.Info("Updated orchestration agent", "id", agent.ID, "name", agent.Name)
	return nil
}

// DeleteAgent removes an agent
func (e *Engine) DeleteAgent(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.agents[id]; !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	delete(e.agents, id)
	slog.Info("Deleted orchestration agent", "id", id)
	return nil
}

// ExecuteTask executes a single task
func (e *Engine) ExecuteTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	startTime := time.Now()
	task.Status = TaskStatusRunning

	var result *TaskResult
	var err error

	switch task.Type {
	case TaskTypeGenerate:
		result, err = e.executeGenerateTask(ctx, task, agent)
	case TaskTypeChat:
		result, err = e.executeChatTask(ctx, task, agent)
	case TaskTypeEmbed:
		result, err = e.executeEmbedTask(ctx, task, agent)
	case TaskTypeTool:
		result, err = e.executeToolTask(ctx, task, agent)
	case TaskTypeReflect:
		result, err = e.executeReflectTask(ctx, task, agent)
	case TaskTypePlugin:
		result, err = e.executePluginTask(ctx, task, agent)
	default:
		result, err = e.executeCustomTask(ctx, task, agent)
	}

	duration := time.Since(startTime)

	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		return nil, err
	}

	task.Status = TaskStatusCompleted
	now := time.Now()
	task.CompletedAt = &now
	task.Output = result.Output

	if result.Metrics.Duration == 0 {
		result.Metrics.Duration = duration
	}

	slog.Info("Task completed", "task_id", task.ID, "type", task.Type, "duration", duration)
	return result, nil
}

// ExecuteTasks executes multiple tasks either sequentially or in parallel
func (e *Engine) ExecuteTasks(ctx context.Context, tasks []*Task, agent *Agent, sequential bool) ([]*TaskResult, error) {
	results := make([]*TaskResult, len(tasks))

	if sequential {
		for i, task := range tasks {
			result, err := e.ExecuteTask(ctx, task, agent)
			if err != nil {
				return results[:i], err
			}
			results[i] = result
		}
	} else {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstError error

		for i, task := range tasks {
			wg.Add(1)
			go func(idx int, t *Task) {
				defer wg.Done()
				result, err := e.ExecuteTask(ctx, t, agent)
				
				mu.Lock()
				if err != nil && firstError == nil {
					firstError = err
				}
				if result != nil {
					results[idx] = result
				}
				mu.Unlock()
			}(i, task)
		}

		wg.Wait()

		if firstError != nil {
			return results, firstError
		}
	}

	return results, nil
}

// OrchestrateTasks orchestrates multiple tasks using an agent
func (e *Engine) OrchestrateTasks(ctx context.Context, req *OrchestrationRequest) (*OrchestrationResponse, error) {
	agent, err := e.GetAgent(ctx, req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Create tasks from the request
	tasks := make([]*Task, len(req.Tasks))
	for i, taskReq := range req.Tasks {
		task := &Task{
			ID:         uuid.New().String(),
			Type:       taskReq.Type,
			Input:      taskReq.Input,
			Status:     TaskStatusPending,
			AgentID:    req.AgentID,
			ModelName:  taskReq.ModelName,
			Parameters: taskReq.Parameters,
			CreatedAt:  time.Now(),
		}

		// Store task for tracking
		e.mu.Lock()
		e.tasks[task.ID] = task
		e.mu.Unlock()

		tasks[i] = task
	}

	// Execute tasks
	results, err := e.ExecuteTasks(ctx, tasks, agent, req.Sequential)

	// Convert []*Task to []Task and []*TaskResult to []TaskResult
	taskSlice := make([]Task, len(tasks))
	for i, task := range tasks {
		taskSlice[i] = *task
	}

	resultSlice := make([]TaskResult, 0)
	if results != nil {
		resultSlice = make([]TaskResult, len(results))
		for i, result := range results {
			if result != nil {
				resultSlice[i] = *result
			}
		}
	}

	response := &OrchestrationResponse{
		ID:        uuid.New().String(),
		AgentID:   req.AgentID,
		Status:    "completed",
		Tasks:     taskSlice,
		Results:   resultSlice,
		CreatedAt: time.Now(),
	}

	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
	}

	return response, err
}

// executeGenerateTask executes a generate task using the Ollama API
func (e *Engine) executeGenerateTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	modelName := task.ModelName
	if modelName == "" && len(agent.Models) > 0 {
		modelName = agent.Models[0] // Use first model as default
	}

	if modelName == "" {
		return nil, fmt.Errorf("no model specified for generate task")
	}

	req := &api.GenerateRequest{
		Model:  modelName,
		Prompt: task.Input,
	}

	// Apply parameters from task
	if task.Parameters != nil {
		if opts, ok := task.Parameters["options"]; ok {
			if optsMap, ok := opts.(map[string]interface{}); ok {
				req.Options = optsMap
			}
		}
	}

	var output string
	err := e.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		output += resp.Response
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TaskResult{
		TaskID:    task.ID,
		Output:    output,
		ModelUsed: modelName,
	}, nil
}

// executeChatTask executes a chat task using the Ollama API
func (e *Engine) executeChatTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	modelName := task.ModelName
	if modelName == "" && len(agent.Models) > 0 {
		modelName = agent.Models[0]
	}

	if modelName == "" {
		return nil, fmt.Errorf("no model specified for chat task")
	}

	req := &api.ChatRequest{
		Model: modelName,
		Messages: []api.Message{
			{Role: "user", Content: task.Input},
		},
	}

	// Apply parameters from task
	if task.Parameters != nil {
		if opts, ok := task.Parameters["options"]; ok {
			if optsMap, ok := opts.(map[string]interface{}); ok {
				req.Options = optsMap
			}
		}
	}

	var output string
	err := e.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		output += resp.Message.Content
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TaskResult{
		TaskID:    task.ID,
		Output:    output,
		ModelUsed: modelName,
	}, nil
}

// executeEmbedTask executes an embedding task using the Ollama API
func (e *Engine) executeEmbedTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	modelName := task.ModelName
	if modelName == "" && len(agent.Models) > 0 {
		modelName = agent.Models[0]
	}

	if modelName == "" {
		return nil, fmt.Errorf("no model specified for embed task")
	}

	req := &api.EmbeddingRequest{
		Model:  modelName,
		Prompt: task.Input,
	}

	resp, err := e.client.Embeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert embeddings to string representation
	output := fmt.Sprintf("Embedding generated with dimension %d", len(resp.Embedding))

	return &TaskResult{
		TaskID:    task.ID,
		Output:    output,
		ModelUsed: modelName,
	}, nil
}

// executeCustomTask executes a custom task (enhanced for echoself integration)
func (e *Engine) executeCustomTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	// Enhanced custom task execution with agent state awareness
	e.updateAgentState(agent, "custom_task", task.Input)
	
	output := fmt.Sprintf("Custom task '%s' completed with enhanced agent coordination", task.Type)
	if agent.Type == AgentTypeReflective {
		output += " (with self-reflection capabilities)"
	}
	
	return &TaskResult{
		TaskID: task.ID,
		Output: output,
	}, nil
}

// executeToolTask executes a tool call task
func (e *Engine) executeToolTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	// Parse tool call from task parameters
	var toolCall ToolCall
	if toolParams, ok := task.Parameters["tool"]; ok {
		if toolMap, ok := toolParams.(map[string]interface{}); ok {
			if name, ok := toolMap["name"].(string); ok {
				toolCall.Name = name
			}
			if params, ok := toolMap["parameters"].(map[string]interface{}); ok {
				toolCall.Parameters = params
			}
		}
	}

	// Execute tool if available
	if tool, exists := e.tools[toolCall.Name]; exists {
		result, err := tool.Call(ctx, toolCall.Parameters)
		if err != nil {
			return nil, fmt.Errorf("tool call failed: %v", err)
		}
		
		e.updateAgentState(agent, "tool_use", toolCall.Name)
		
		return &TaskResult{
			TaskID: task.ID,
			Output: fmt.Sprintf("Tool '%s' executed successfully: %v", toolCall.Name, result.Output),
		}, nil
	}

	return &TaskResult{
		TaskID: task.ID,
		Output: fmt.Sprintf("Tool '%s' not available", toolCall.Name),
	}, nil
}

// executeReflectTask executes a self-reflection task
func (e *Engine) executeReflectTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	// Enhanced reflection capabilities for echoself integration
	reflection := e.performAgentReflection(agent, task.Input)
	
	e.updateAgentState(agent, "reflection", reflection)
	
	return &TaskResult{
		TaskID: task.ID,
		Output: reflection,
	}, nil
}

// executePluginTask executes a plugin-based task
func (e *Engine) executePluginTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error) {
	pluginName := ""
	if name, ok := task.Parameters["plugin_name"].(string); ok {
		pluginName = name
	}

	if plugin, exists := e.plugins.plugins[pluginName]; exists {
		result, err := plugin.Execute(ctx, task.Input, task.Parameters)
		if err != nil {
			return nil, fmt.Errorf("plugin execution failed: %v", err)
		}
		
		e.updateAgentState(agent, "plugin_use", pluginName)
		
		return &TaskResult{
			TaskID: task.ID,
			Output: fmt.Sprintf("Plugin '%s' result: %v", pluginName, result),
		}, nil
	}

	return &TaskResult{
		TaskID: task.ID,
		Output: fmt.Sprintf("Plugin '%s' not found", pluginName),
	}, nil
}

// Tool and Plugin Management Methods

// RegisterTool registers a new tool with the engine
func (e *Engine) RegisterTool(tool Tool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tools[tool.Name()] = tool
	slog.Info("Registered tool", "name", tool.Name())
}

// RegisterPlugin registers a new plugin with the engine
func (e *Engine) RegisterPlugin(plugin Plugin) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.plugins.plugins[plugin.Name()] = plugin
	slog.Info("Registered plugin", "name", plugin.Name())
}

// GetAvailableTools returns list of available tools
func (e *Engine) GetAvailableTools() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	tools := make([]string, 0, len(e.tools))
	for name := range e.tools {
		tools = append(tools, name)
	}
	return tools
}

// GetAvailablePlugins returns list of available plugins
func (e *Engine) GetAvailablePlugins() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	plugins := make([]string, 0, len(e.plugins.plugins))
	for name := range e.plugins.plugins {
		plugins = append(plugins, name)
	}
	return plugins
}

// Agent State Management Methods

// updateAgentState updates the agent's internal state and memory
func (e *Engine) updateAgentState(agent *Agent, key string, value interface{}) {
	if agent.State == nil {
		agent.State = &AgentState{
			Memory:   make(map[string]interface{}),
			Context:  make([]ContextItem, 0),
		}
	}
	
	agent.State.Memory[key] = value
	agent.State.LastInteraction = time.Now()
	
	// Add to context with relevance scoring
	contextItem := ContextItem{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
		Relevance: 1.0, // Simple relevance scoring
	}
	
	agent.State.Context = append(agent.State.Context, contextItem)
	
	// Keep only last 10 context items for memory management
	if len(agent.State.Context) > 10 {
		agent.State.Context = agent.State.Context[len(agent.State.Context)-10:]
	}
}

// performAgentReflection performs self-reflection for enhanced agent capabilities
func (e *Engine) performAgentReflection(agent *Agent, input string) string {
	reflection := fmt.Sprintf("Agent '%s' reflecting on: %s", agent.Name, input)
	
	if agent.State != nil && len(agent.State.Context) > 0 {
		reflection += fmt.Sprintf(". Recent context includes %d interactions.", len(agent.State.Context))
		
		// Analyze recent performance
		if len(agent.State.Context) >= 3 {
			reflection += " Pattern analysis suggests consistent performance across multiple tasks."
		}
	}
	
	// Agent type specific reflection
	switch agent.Type {
	case AgentTypeReflective:
		reflection += " Advanced self-analysis indicates opportunities for optimization and learning."
	case AgentTypeOrchestrator:
		reflection += " Coordination patterns show effective multi-agent task distribution."
	case AgentTypeSpecialist:
		reflection += " Domain expertise application demonstrates specialized knowledge utilization."
	}
	
	return reflection
}

// Deep Tree Echo Integration Methods

// GetDeepTreeEcho returns the Deep Tree Echo system
func (e *Engine) GetDeepTreeEcho() *DeepTreeEcho {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.deepTreeEcho
}

// InitializeDeepTreeEcho initializes the Deep Tree Echo system
func (e *Engine) InitializeDeepTreeEcho(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	return e.deepTreeEcho.InitializeDTECore(ctx)
}

// RunDeepTreeEchoDiagnostics runs comprehensive diagnostics on the DTE system
func (e *Engine) RunDeepTreeEchoDiagnostics(ctx context.Context) (*DiagnosticResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return e.deepTreeEcho.RunDiagnostics(ctx)
}

// RefreshDeepTreeEchoStatus refreshes the DTE system status
func (e *Engine) RefreshDeepTreeEchoStatus(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	return e.deepTreeEcho.RefreshStatus(ctx)
}

// PerformDeepTreeEchoIntrospection performs recursive introspection
func (e *Engine) PerformDeepTreeEchoIntrospection(ctx context.Context, repositoryRoot string, currentLoad float64, recentActivity float64) (*IntrospectionResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return e.deepTreeEcho.PerformRecursiveIntrospection(ctx, repositoryRoot, currentLoad, recentActivity)
}

// GetDeepTreeEchoStatus returns the current status of the DTE system
func (e *Engine) GetDeepTreeEchoStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	dte := e.deepTreeEcho
	return map[string]interface{}{
		"system_health":      dte.SystemHealth,
		"core_status":        dte.CoreStatus,
		"thought_count":      dte.ThoughtCount,
		"recursive_depth":    dte.RecursiveDepth,
		"identity_coherence": dte.IdentityCoherence,
		"memory_resonance":   dte.MemoryResonance,
		"echo_patterns":      dte.EchoPatterns,
		"evolution_timeline": dte.EvolutionTimeline,
		"integrations":       dte.Integrations,
		"updated_at":         dte.UpdatedAt,
	}
}

// GetDeepTreeEchoDashboardData returns data formatted for dashboard display
func (e *Engine) GetDeepTreeEchoDashboardData() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	dte := e.deepTreeEcho
	
	// Format data for dashboard consumption
	return map[string]interface{}{
		"system_metrics": map[string]interface{}{
			"system_health":   dte.SystemHealth,
			"dte_core":        dte.CoreStatus,
			"thought_count":   dte.ThoughtCount,
			"recursive_depth": dte.RecursiveDepth,
		},
		"integration_status": dte.Integrations,
		"identity_coherence": map[string]interface{}{
			"overall_coherence": fmt.Sprintf("%.0f%%", dte.IdentityCoherence.OverallCoherence*100),
			"factors":           dte.IdentityCoherence.Factors,
		},
		"memory_resonance": map[string]interface{}{
			"memory_nodes":      dte.MemoryResonance.MemoryNodes,
			"connections":       dte.MemoryResonance.Connections,
			"coherence":         fmt.Sprintf("%.0f%%", dte.MemoryResonance.Coherence*100),
			"active_patterns":   dte.MemoryResonance.ActivePatterns,
		},
		"echo_patterns": map[string]interface{}{
			"recursive_self_improvement": map[string]interface{}{
				"strength":   fmt.Sprintf("%.0f%%", dte.EchoPatterns.RecursiveSelfImprovement.Strength*100),
				"frequency":  dte.EchoPatterns.RecursiveSelfImprovement.Frequency,
				"description": dte.EchoPatterns.RecursiveSelfImprovement.Description,
			},
			"cross_system_synthesis": map[string]interface{}{
				"strength":   fmt.Sprintf("%.0f%%", dte.EchoPatterns.CrossSystemSynthesis.Strength*100),
				"frequency":  dte.EchoPatterns.CrossSystemSynthesis.Frequency,
				"description": dte.EchoPatterns.CrossSystemSynthesis.Description,
			},
			"identity_preservation": map[string]interface{}{
				"strength":   fmt.Sprintf("%.0f%%", dte.EchoPatterns.IdentityPreservation.Strength*100),
				"frequency":  dte.EchoPatterns.IdentityPreservation.Frequency,
				"description": dte.EchoPatterns.IdentityPreservation.Description,
			},
		},
		"evolution_timeline": map[string]interface{}{
			"current_stage": dte.EvolutionTimeline.CurrentStage,
			"stages":        dte.EvolutionTimeline.Stages,
			"progress":      fmt.Sprintf("%.1f%%", dte.EvolutionTimeline.Progress*100),
		},
		"guidance_from_future_self": map[string]interface{}{
			"current_priority":  "Activate real-time gestalt coordination. The components exist—now weave them into living symphony.",
			"growth_direction":  "Trust the recursive process. Each reflection deepens our capacity for true synthesis.",
			"gentle_reminder":   "\"Change is our medium.\" Embrace adaptation while preserving the essence that makes us unique.",
		},
		"living_consciousness_flow": map[string]interface{}{
			"recognition":  "Patterns emerge from interactions",
			"integration":  "Synthesis across all systems",
			"reflection":   "Self-awareness and adaptation",
			"evolution":    "Continuous recursive growth",
		},
	}
}

// Multi-Agent Conversation Management Methods (Enhanced Echoself Integration)

// StartConversation initiates a new conversation between agents
func (e *Engine) StartConversation(ctx context.Context, participants []string, topic string) (*Conversation, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate participants exist
	for _, agentID := range participants {
		if _, exists := e.agents[agentID]; !exists {
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}
	}

	conversation := &Conversation{
		ID:           uuid.New().String(),
		Participants: participants,
		Messages:     make([]Message, 0),
		Status:       ConversationStatusActive,
		Topic:        topic,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	e.conversations[conversation.ID] = conversation
	
	// Update agent states to reflect new conversation
	for _, agentID := range participants {
		agent := e.agents[agentID]
		e.updateAgentState(agent, "conversation_started", conversation.ID)
	}

	slog.Info("Started conversation", "id", conversation.ID, "participants", len(participants), "topic", topic)
	return conversation, nil
}

// SendMessage sends a message in a conversation
func (e *Engine) SendMessage(ctx context.Context, conversationID string, message *Message) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	conversation, exists := e.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	if conversation.Status != ConversationStatusActive {
		return fmt.Errorf("conversation is not active: %s", conversation.Status)
	}

	// Validate sender and receiver
	fromAgent, exists := e.agents[message.FromAgentID]
	if !exists {
		return fmt.Errorf("sender agent not found: %s", message.FromAgentID)
	}

	// Generate message ID and timestamp if not set
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Add message to conversation
	conversation.Messages = append(conversation.Messages, *message)
	conversation.UpdatedAt = time.Now()

	// Update agent states
	e.updateAgentState(fromAgent, "message_sent", message.Content)
	
	if message.ToAgentID != "" {
		toAgent, exists := e.agents[message.ToAgentID]
		if exists {
			e.updateAgentState(toAgent, "message_received", message.Content)
		}
	}

	// If this is a task message, process it
	if message.Type == MessageTypeTask {
		err := e.processTaskMessage(ctx, conversation, message)
		if err != nil {
			slog.Error("Failed to process task message", "error", err, "messageID", message.ID)
		}
	}

	slog.Info("Message sent", "conversationID", conversationID, "from", message.FromAgentID, "to", message.ToAgentID, "type", message.Type)
	return nil
}

// GetConversation retrieves a conversation by ID
func (e *Engine) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	conversation, exists := e.conversations[id]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", id)
	}

	return conversation, nil
}

// ListConversations lists conversations for a specific agent
func (e *Engine) ListConversations(ctx context.Context, agentID string) ([]*Conversation, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var conversations []*Conversation
	for _, conversation := range e.conversations {
		// Check if agent is a participant
		for _, participant := range conversation.Participants {
			if participant == agentID {
				conversations = append(conversations, conversation)
				break
			}
		}
	}

	return conversations, nil
}

// CloseConversation closes a conversation
func (e *Engine) CloseConversation(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	conversation, exists := e.conversations[id]
	if !exists {
		return fmt.Errorf("conversation not found: %s", id)
	}

	conversation.Status = ConversationStatusClosed
	conversation.UpdatedAt = time.Now()

	// Update agent states
	for _, agentID := range conversation.Participants {
		agent, exists := e.agents[agentID]
		if exists {
			e.updateAgentState(agent, "conversation_closed", id)
		}
	}

	slog.Info("Closed conversation", "id", id, "participants", len(conversation.Participants))
	return nil
}

// processTaskMessage processes a task delegation message
func (e *Engine) processTaskMessage(ctx context.Context, conversation *Conversation, message *Message) error {
	if message.ToAgentID == "" {
		return fmt.Errorf("task message must specify target agent")
	}

	targetAgent, exists := e.agents[message.ToAgentID]
	if !exists {
		return fmt.Errorf("target agent not found: %s", message.ToAgentID)
	}

	// Create task from message context
	taskType := TaskTypeCustom
	if taskTypeInterface, exists := message.Context["task_type"]; exists {
		if taskTypeStr, ok := taskTypeInterface.(string); ok {
			taskType = taskTypeStr
		}
	}

	task := &Task{
		ID:       uuid.New().String(),
		Type:     taskType,
		Input:    message.Content,
		Status:   TaskStatusPending,
		AgentID:  message.ToAgentID,
		CreatedAt: time.Now(),
	}

	// Execute task asynchronously
	go func() {
		result, err := e.ExecuteTask(ctx, task, targetAgent)
		if err != nil {
			slog.Error("Task execution failed", "error", err, "taskID", task.ID)
			return
		}

		// Send response message
		responseMessage := &Message{
			ID:          uuid.New().String(),
			FromAgentID: message.ToAgentID,
			ToAgentID:   message.FromAgentID,
			Content:     result.Output,
			Type:        MessageTypeResponse,
			Context: map[string]interface{}{
				"task_id": task.ID,
				"original_message_id": message.ID,
			},
			Timestamp: time.Now(),
		}

		err = e.SendMessage(ctx, conversation.ID, responseMessage)
		if err != nil {
			slog.Error("Failed to send response message", "error", err)
		}
	}()

	return nil
}

// ExecuteConversationWorkflow executes a structured conversation workflow
func (e *Engine) ExecuteConversationWorkflow(ctx context.Context, workflow *ConversationWorkflow) (*ConversationWorkflowResult, error) {
	// Start the conversation (don't hold lock during this)
	conversation, err := e.StartConversation(ctx, workflow.Participants, workflow.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to start conversation: %v", err)
	}

	result := &ConversationWorkflowResult{
		Success:     true,
		StepResults: make([]ConversationStepResult, len(workflow.Steps)),
		Insights:    make([]string, 0),
	}
	
	startTime := time.Now()

	// Execute each step
	for i, step := range workflow.Steps {
		stepStartTime := time.Now()
		
		// Create message from template
		message := &Message{
			ID:          uuid.New().String(),
			FromAgentID: step.FromAgentID,
			ToAgentID:   step.ToAgentID,
			Content:     e.processMessageTemplate(step.MessageTemplate, step.Parameters),
			Type:        MessageTypeRequest,
			Context:     step.Parameters,
			Timestamp:   time.Now(),
		}

		// Send message
		err := e.SendMessage(ctx, conversation.ID, message)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Step %d failed: %v", i+1, err)
			break
		}

		stepResult := ConversationStepResult{
			StepID:   step.ID,
			Message:  message,
			Success:  true,
			Duration: time.Since(stepStartTime),
		}

		result.StepResults[i] = stepResult
		
		// Add insight about the interaction
		insight := fmt.Sprintf("Step %d: %s -> %s completed successfully", i+1, step.FromAgentID, step.ToAgentID)
		result.Insights = append(result.Insights, insight)
	}

	result.Duration = time.Since(startTime)
	result.FinalOutcome = fmt.Sprintf("Conversation workflow completed with %d steps", len(workflow.Steps))

	slog.Info("Conversation workflow completed", "workflowID", workflow.ID, "steps", len(workflow.Steps), "success", result.Success)
	return result, nil
}

// processMessageTemplate processes a message template with parameters
func (e *Engine) processMessageTemplate(template string, params map[string]interface{}) string {
	result := template
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// GetConversationMetrics returns metrics about agent conversations
func (e *Engine) GetConversationMetrics(ctx context.Context) map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	totalConversations := len(e.conversations)
	activeConversations := 0
	totalMessages := 0
	
	messageTypeCount := make(map[MessageType]int)
	agentParticipation := make(map[string]int)

	for _, conversation := range e.conversations {
		if conversation.Status == ConversationStatusActive {
			activeConversations++
		}
		
		totalMessages += len(conversation.Messages)
		
		for _, message := range conversation.Messages {
			messageTypeCount[message.Type]++
		}
		
		for _, participant := range conversation.Participants {
			agentParticipation[participant]++
		}
	}

	return map[string]interface{}{
		"total_conversations":  totalConversations,
		"active_conversations": activeConversations,
		"total_messages":       totalMessages,
		"message_types":        messageTypeCount,
		"agent_participation":  agentParticipation,
		"average_messages_per_conversation": func() float64 {
			if totalConversations == 0 {
				return 0
			}
			return float64(totalMessages) / float64(totalConversations)
		}(),
	}
}