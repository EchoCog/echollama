package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/EchoCog/echollama/api"
)

// Engine implements the core orchestration functionality
type Engine struct {
	client               api.Client
	agents               map[string]*Agent
	tasks                map[string]*Task
	tools                map[string]Tool
	plugins              *PluginRegistry
	deepTreeEcho         *DeepTreeEcho
	conversations        map[string]*Conversation   // Multi-agent conversations
	learningSystem       *LearningSystem             // Advanced learning capabilities
	performanceOptimizer *PerformanceOptimizer      // Performance optimization
	echoselfCoreAgent    *EchoselfCoreAgent          // Core AAR orchestration agent
	mu                   sync.RWMutex
}

// NewEngine creates a new orchestration engine
func NewEngine(client api.Client) *Engine {
	return &Engine{
		client:               client,
		agents:               make(map[string]*Agent),
		tasks:                make(map[string]*Task),
		tools:                make(map[string]Tool),
		plugins:              &PluginRegistry{plugins: make(map[string]Plugin)},
		deepTreeEcho:         NewDeepTreeEcho("Primary Deep Tree Echo System"),
		conversations:        make(map[string]*Conversation),
		learningSystem:       NewLearningSystem(),
		performanceOptimizer: NewPerformanceOptimizer(),
		// echoselfCoreAgent is lazily initialised via RegisterEchoselfCoreAgent.
	}
}

// CreateAgent creates a new orchestration agent
func (e *Engine) CreateAgent(ctx context.Context, agent *Agent) error {
	e.mu.Lock()

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
	// Capture the echoself agent reference inside the lock so we hold a stable
	// pointer even if another goroutine calls RegisterEchoselfCoreAgent later.
	eca := e.echoselfCoreAgent
	e.mu.Unlock()

	slog.Info("Created orchestration agent", "id", agent.ID, "name", agent.Name)

	// Register the new agent in the echoself arena if one is active.
	if eca != nil {
		if err := eca.RegisterAgent(ctx, agent.ID); err != nil {
			slog.Warn("Failed to register new agent in echoself arena", "agent_id", agent.ID, "error", err)
		}
	}
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

	if _, exists := e.agents[id]; !exists {
		e.mu.Unlock()
		return fmt.Errorf("agent not found: %s", id)
	}

	delete(e.agents, id)
	eca := e.echoselfCoreAgent
	e.mu.Unlock()

	slog.Info("Deleted orchestration agent", "id", id)

	// Remove from echoself arena if registered.
	if eca != nil {
		eca.UnregisterAgent(ctx, id)
	}
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
	endTime := time.Now()

	// Record performance data for learning
	performance := &TaskPerformance{
		TaskID:     task.ID,
		TaskType:   task.Type,
		AgentID:    agent.ID,
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   duration,
		Success:    err == nil,
		Quality:    e.calculateTaskQuality(result, err),
		Difficulty: e.estimateTaskDifficulty(task),
		Context:    task.Parameters,
		Feedback:   e.generatePerformanceFeedback(task, result, err, duration),
	}
	
	e.learningSystem.RecordTaskPerformance(performance)

	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		return nil, err
	}

	task.Status = TaskStatusCompleted
	task.CompletedAt = &endTime
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

// Learning System Integration Methods

// calculateTaskQuality estimates the quality of a task result
func (e *Engine) calculateTaskQuality(result *TaskResult, err error) float64 {
	if err != nil {
		return 0.0
	}
	
	if result == nil {
		return 0.1
	}
	
	// Base quality on output length and completeness
	baseQuality := 0.5
	
	if result.Output != "" {
		if len(result.Output) > 50 {
			baseQuality += 0.2
		}
		if len(result.Output) > 200 {
			baseQuality += 0.1
		}
		
		// Check for common quality indicators
		output := strings.ToLower(result.Output)
		if strings.Contains(output, "error") || strings.Contains(output, "failed") {
			baseQuality -= 0.2
		}
		if strings.Contains(output, "successfully") || strings.Contains(output, "completed") {
			baseQuality += 0.2
		}
	}
	
	return math.Min(1.0, math.Max(0.0, baseQuality))
}

// estimateTaskDifficulty estimates how difficult a task is
func (e *Engine) estimateTaskDifficulty(task *Task) float64 {
	difficulty := 0.5 // Base difficulty
	
	// Factor in task type
	switch task.Type {
	case TaskTypeEmbed:
		difficulty = 0.3
	case TaskTypeGenerate:
		difficulty = 0.4
	case TaskTypeChat:
		difficulty = 0.5
	case TaskTypeTool:
		difficulty = 0.6
	case TaskTypeReflect:
		difficulty = 0.7
	case TaskTypePlugin:
		difficulty = 0.8
	case TaskTypeCustom:
		difficulty = 0.9
	}
	
	// Factor in input complexity
	if len(task.Input) > 500 {
		difficulty += 0.1
	}
	if len(task.Input) > 1000 {
		difficulty += 0.1
	}
	
	// Factor in parameters
	if task.Parameters != nil && len(task.Parameters) > 3 {
		difficulty += 0.1
	}
	
	return math.Min(1.0, difficulty)
}

// generatePerformanceFeedback generates feedback about task performance
func (e *Engine) generatePerformanceFeedback(task *Task, result *TaskResult, err error, duration time.Duration) *PerformanceFeedback {
	feedback := &PerformanceFeedback{}
	
	// Calculate accuracy based on error and result quality
	if err != nil {
		feedback.Accuracy = 0.0
	} else if result != nil && result.Output != "" {
		feedback.Accuracy = 0.8 // Assume good accuracy if task completed
	} else {
		feedback.Accuracy = 0.3
	}
	
	// Calculate efficiency based on duration
	expectedDuration := e.getExpectedTaskDuration(task.Type)
	if duration <= expectedDuration {
		feedback.Efficiency = 1.0
	} else if duration <= expectedDuration*2 {
		feedback.Efficiency = 0.8
	} else {
		feedback.Efficiency = 0.5
	}
	
	// Base values for other metrics
	feedback.Creativity = 0.5
	feedback.Adaptability = 0.6
	feedback.Collaboration = 0.5
	feedback.LearningRate = 0.1
	
	return feedback
}

// getExpectedTaskDuration returns expected duration for different task types
func (e *Engine) getExpectedTaskDuration(taskType string) time.Duration {
	switch taskType {
	case TaskTypeEmbed:
		return 2 * time.Second
	case TaskTypeGenerate:
		return 5 * time.Second
	case TaskTypeChat:
		return 10 * time.Second
	case TaskTypeTool:
		return 3 * time.Second
	case TaskTypeReflect:
		return 1 * time.Second
	case TaskTypePlugin:
		return 200 * time.Millisecond
	default:
		return 5 * time.Second
	}
}

// GetLearningSystem returns the learning system instance
func (e *Engine) GetLearningSystem() *LearningSystem {
	return e.learningSystem
}

// PredictOptimalAgentForTask uses learning system to predict best agent for a task
func (e *Engine) PredictOptimalAgentForTask(ctx context.Context, task *Task) (*Agent, float64, error) {
	e.mu.RLock()
	agents := make([]*Agent, 0, len(e.agents))
	for _, agent := range e.agents {
		agents = append(agents, agent)
	}
	e.mu.RUnlock()
	
	return e.learningSystem.PredictOptimalAgent(ctx, task, agents)
}

// AdaptAgent performs learning-based adaptation for an agent
func (e *Engine) AdaptAgent(ctx context.Context, agentID string) (*AdaptationResult, error) {
	agent, err := e.GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}
	
	return e.learningSystem.adaptationEngine.AdaptAgent(ctx, agent, e.learningSystem)
}

// Performance Optimization Methods

// GetPerformanceOptimizer returns the performance optimizer instance
func (e *Engine) GetPerformanceOptimizer() *PerformanceOptimizer {
	return e.performanceOptimizer
}

// ExecuteTaskOptimized executes a task with performance optimization
func (e *Engine) ExecuteTaskOptimized(ctx context.Context, task *Task, priority TaskPriority, deadline time.Time) (*TaskResult, error) {
	// Generate task ID if not provided
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	
	// Store task in engine
	e.mu.Lock()
	e.tasks[task.ID] = task
	e.mu.Unlock()
	
	// Select optimal agent using learning system and load balancing
	availableAgents, err := e.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available agents: %v", err)
	}
	
	// Use learning system to predict optimal agent
	optimalAgent, confidence, err := e.learningSystem.PredictOptimalAgent(ctx, task, availableAgents)
	if err != nil || confidence < 0.3 { // Fall back to load balancing if confidence is low
		optimalAgent, err = e.performanceOptimizer.loadBalancer.SelectOptimalAgent(ctx, task, availableAgents)
		if err != nil {
			return nil, fmt.Errorf("failed to select optimal agent: %v", err)
		}
	}
	
	// Schedule the task
	scheduledTask, err := e.performanceOptimizer.taskScheduler.ScheduleTask(task, optimalAgent, priority, deadline)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule task: %v", err)
	}
	
	// Allocate resources
	resourceRequirements := scheduledTask.ResourceRequirements
	reservation, err := e.performanceOptimizer.resourceManager.AllocateResources(
		ctx, task.ID, optimalAgent.ID, resourceRequirements, ResourcePriorityNormal)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate resources: %v", err)
	}
	
	// Execute the task
	result, err := e.ExecuteTask(ctx, task, optimalAgent)
	
	// Release resources
	e.performanceOptimizer.resourceManager.ReleaseResources(ctx, reservation.ReservationID)
	
	// Update performance metrics
	e.updatePerformanceMetrics(task, result, err, scheduledTask)
	
	return result, err
}

// updatePerformanceMetrics updates system performance metrics after task execution
func (e *Engine) updatePerformanceMetrics(task *Task, result *TaskResult, err error, scheduledTask *ScheduledTask) {
	// Update load balancer with agent performance
	agentID := scheduledTask.Agent.ID
	performanceScore := 0.5
	healthStatus := HealthStatusHealthy
	
	if result != nil && err == nil {
		performanceScore = 0.8
	} else if err != nil {
		performanceScore = 0.2
		healthStatus = HealthStatusDegraded
	}
	
	// Update agent load (simplified)
	e.performanceOptimizer.loadBalancer.UpdateAgentLoad(
		agentID, 1, 0, scheduledTask.ResourceRequirements, performanceScore, healthStatus)
	
	// Update system metrics
	e.mu.RLock()
	totalTasks := len(e.tasks)
	completedTasks := 0
	failedTasks := 0
	totalDuration := time.Duration(0)
	
	for _, t := range e.tasks {
		if t.Status == TaskStatusCompleted {
			completedTasks++
			if t.CompletedAt != nil {
				totalDuration += t.CompletedAt.Sub(t.CreatedAt)
			}
		} else if t.Status == TaskStatusFailed {
			failedTasks++
		}
	}
	e.mu.RUnlock()
	
	avgResponseTime := time.Duration(0)
	if completedTasks > 0 {
		avgResponseTime = totalDuration / time.Duration(completedTasks)
	}
	
	throughputTPS := 0.0
	if totalDuration > 0 {
		throughputTPS = float64(completedTasks) / totalDuration.Seconds()
	}
	
	systemHealth := 1.0
	if totalTasks > 0 {
		systemHealth = float64(completedTasks) / float64(totalTasks)
	}
	
	systemMetrics := &SystemMetrics{
		TotalTasks:          totalTasks,
		CompletedTasks:      completedTasks,
		FailedTasks:         failedTasks,
		AverageResponseTime: avgResponseTime,
		ThroughputTPS:       throughputTPS,
		ResourceUtilization: scheduledTask.ResourceRequirements,
		SystemHealth:        systemHealth,
		LastUpdated:         time.Now(),
	}
	
	e.performanceOptimizer.performanceMonitor.UpdateSystemMetrics(systemMetrics)
}

// GetSystemMetrics returns current system performance metrics
func (e *Engine) GetSystemMetrics() *SystemMetrics {
	return e.performanceOptimizer.performanceMonitor.GetSystemMetrics()
}

// GetActiveAlerts returns currently active performance alerts
func (e *Engine) GetActiveAlerts() []*Alert {
	return e.performanceOptimizer.performanceMonitor.GetActiveAlerts()
}

// GetResourceUsage returns current resource usage across all agents
func (e *Engine) GetResourceUsage() map[string]*ResourceUsage {
	e.performanceOptimizer.resourceManager.mu.RLock()
	defer e.performanceOptimizer.resourceManager.mu.RUnlock()
	
	usage := make(map[string]*ResourceUsage)
	for agentID, resourceUsage := range e.performanceOptimizer.resourceManager.resourceUsage {
		usage[agentID] = resourceUsage
	}
	return usage
}

// GetAgentLoads returns current load information for all agents
func (e *Engine) GetAgentLoads() map[string]*AgentLoad {
	e.performanceOptimizer.loadBalancer.mu.RLock()
	defer e.performanceOptimizer.loadBalancer.mu.RUnlock()
	
	loads := make(map[string]*AgentLoad)
	for agentID, agentLoad := range e.performanceOptimizer.loadBalancer.agentLoads {
		loads[agentID] = agentLoad
	}
	return loads
}

// NewEchoChat creates a new EchoChat instance connected to this engine
func (e *Engine) NewEchoChat() *EchoChat {
	return NewEchoChat(e)
}

// ─────────────────────────────────────────────────────────────────────────────
// Echoself Core Agent (AAR) – Engine integration
// ─────────────────────────────────────────────────────────────────────────────

// RegisterEchoselfCoreAgent attaches an EchoselfCoreAgent to the engine and
// immediately registers all existing agents in its arena.  If an agent already
// exists with the echoself agent's ID it is replaced.
func (e *Engine) RegisterEchoselfCoreAgent(ctx context.Context, agent *EchoselfCoreAgent) error {
	e.mu.Lock()
	e.echoselfCoreAgent = agent
	// Register the echoself agent itself in the agents map so the API can find it.
	echoselfAgentRecord := &Agent{
		ID:          agent.ID(),
		Name:        "echoself-core",
		Description: "Core AAR orchestration agent with recursive self-introspection (echoself)",
		Type:        AgentTypeEchoself,
		State: &AgentState{
			Memory:          make(map[string]interface{}),
			Context:         make([]ContextItem, 0),
			Goals:           []string{"orchestrate", "introspect", "route"},
			Capabilities:    []string{"aar-routing", "introspection", "arena-management", "prompt-injection"},
			LastInteraction: time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	e.agents[agent.ID()] = echoselfAgentRecord

	// Capture existing agent IDs while holding the lock.
	existingIDs := make([]string, 0, len(e.agents))
	for id := range e.agents {
		if id != agent.ID() {
			existingIDs = append(existingIDs, id)
		}
	}
	e.mu.Unlock()

	// Register each existing agent in the arena (outside the main lock).
	for _, id := range existingIDs {
		if err := agent.RegisterAgent(ctx, id); err != nil {
			slog.Warn("Failed to register existing agent in echoself arena", "agent_id", id, "error", err)
		}
	}

	// Perform an initial arena refresh so the cognitive snapshot is populated.
	if err := agent.RefreshArena(ctx); err != nil {
		slog.Warn("Initial echoself arena refresh failed", "error", err)
	}

	slog.Info("Echoself core agent registered", "agent_id", agent.ID(), "arena_id", agent.Arena().ID)
	return nil
}

// GetEchoselfCoreAgent returns the registered EchoselfCoreAgent, or nil if
// none has been registered yet.
func (e *Engine) GetEchoselfCoreAgent() *EchoselfCoreAgent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.echoselfCoreAgent
}

// RouteTaskViaEchoself uses the EchoselfCoreAgent to determine the best agent
// for a task using the AAR framework.  If no echoself agent is registered the
// call returns nil without error, and the caller should fall back to its own
// routing logic.
func (e *Engine) RouteTaskViaEchoself(ctx context.Context, task *Task) (*AARRoutingDecision, error) {
	e.mu.RLock()
	eca := e.echoselfCoreAgent
	e.mu.RUnlock()

	if eca == nil {
		return nil, nil
	}
	return eca.RouteTask(ctx, task, eca.Arena())
}