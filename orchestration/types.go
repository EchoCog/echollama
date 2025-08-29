package orchestration

import (
	"context"
	"time"

	"github.com/ollama/ollama/api"
)

// Agent represents an orchestration agent that can coordinate multiple models and tasks
type Agent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Models      []string               `json:"models"`
	Config      map[string]interface{} `json:"config"`
	Type        AgentType              `json:"type"`
	State       *AgentState            `json:"state,omitempty"`
	Tools       []string               `json:"tools,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AgentType defines different types of agents with specialized behaviors
type AgentType string

const (
	AgentTypeGeneral     AgentType = "general"     // General purpose agent
	AgentTypeSpecialist  AgentType = "specialist"  // Specialized for specific domains
	AgentTypeOrchestrator AgentType = "orchestrator" // Coordinates other agents
	AgentTypeReflective  AgentType = "reflective"  // Self-analyzing and improving
)

// AgentState maintains persistent state and memory for agents
type AgentState struct {
	Memory         map[string]interface{} `json:"memory,omitempty"`
	Context        []ContextItem          `json:"context,omitempty"`
	Goals          []string               `json:"goals,omitempty"`
	Capabilities   []string               `json:"capabilities,omitempty"`
	LastInteraction time.Time             `json:"last_interaction"`
}

// ContextItem represents a piece of contextual information in agent memory
type ContextItem struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	Relevance float64     `json:"relevance"`
}

// Task represents a task that can be executed by an orchestration agent
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Input       string                 `json:"input"`
	Output      string                 `json:"output,omitempty"`
	Status      string                 `json:"status"`
	AgentID     string                 `json:"agent_id"`
	ModelName   string                 `json:"model_name,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// TaskStatus constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// TaskType constants
const (
	TaskTypeGenerate = "generate"
	TaskTypeChat     = "chat"
	TaskTypeEmbed    = "embed"
	TaskTypeCustom   = "custom"
	TaskTypeTool     = "tool"     // Call external tools
	TaskTypeReflect  = "reflect"  // Self-reflection and analysis
	TaskTypePlugin   = "plugin"   // Custom plugin execution
)

// ToolCall represents a call to an external tool
type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Success bool        `json:"success"`
	Output  interface{} `json:"output"`
	Error   string      `json:"error,omitempty"`
}

// Plugin interface for extensible custom task types
type Plugin interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string, params map[string]interface{}) (interface{}, error)
}

// PluginRegistry manages available plugins
type PluginRegistry struct {
	plugins map[string]Plugin
}

// Tool interface for external tool integrations
type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, params map[string]interface{}) (*ToolResult, error)
}

// OrchestrationRequest represents a request to orchestrate multiple tasks
type OrchestrationRequest struct {
	AgentID     string                 `json:"agent_id"`
	Tasks       []TaskRequest          `json:"tasks"`
	Sequential  bool                   `json:"sequential"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Stream      *bool                  `json:"stream,omitempty"`
	KeepAlive   *api.Duration          `json:"keep_alive,omitempty"`
}

// TaskRequest represents a single task within an orchestration request
type TaskRequest struct {
	Type       string                 `json:"type"`
	Input      string                 `json:"input"`
	ModelName  string                 `json:"model_name,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// OrchestrationResponse represents the response from an orchestration request
type OrchestrationResponse struct {
	ID        string `json:"id"`
	AgentID   string `json:"agent_id"`
	Status    string `json:"status"`
	Tasks     []Task `json:"tasks"`
	Results   []TaskResult `json:"results,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskResult represents the result of a completed task
type TaskResult struct {
	TaskID    string `json:"task_id"`
	Output    string `json:"output"`
	ModelUsed string `json:"model_used,omitempty"`
	Metrics   TaskMetrics `json:"metrics,omitempty"`
}

// TaskMetrics contains performance metrics for a completed task
type TaskMetrics struct {
	Duration     time.Duration `json:"duration"`
	TokensUsed   int           `json:"tokens_used,omitempty"`
	PromptTokens int           `json:"prompt_tokens,omitempty"`
	OutputTokens int           `json:"output_tokens,omitempty"`
}

// AgentManager interface defines methods for managing orchestration agents
type AgentManager interface {
	CreateAgent(ctx context.Context, agent *Agent) error
	GetAgent(ctx context.Context, id string) (*Agent, error)
	ListAgents(ctx context.Context) ([]*Agent, error)
	UpdateAgent(ctx context.Context, agent *Agent) error
	DeleteAgent(ctx context.Context, id string) error
}

// TaskExecutor interface defines methods for executing tasks
type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error)
	ExecuteTasks(ctx context.Context, tasks []*Task, agent *Agent, sequential bool) ([]*TaskResult, error)
}

// Orchestrator interface combines agent management and task execution
type Orchestrator interface {
	AgentManager
	TaskExecutor
	OrchestrateTasks(ctx context.Context, req *OrchestrationRequest) (*OrchestrationResponse, error)
}

// Multi-Agent Conversation Types for Enhanced Echoself Integration

// Message represents a message in an inter-agent conversation
type Message struct {
	ID          string                 `json:"id"`
	FromAgentID string                 `json:"from_agent_id"`
	ToAgentID   string                 `json:"to_agent_id"`
	Content     string                 `json:"content"`
	Type        MessageType            `json:"type"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Response    *Message               `json:"response,omitempty"`
}

// MessageType defines different types of inter-agent messages
type MessageType string

const (
	MessageTypeRequest     MessageType = "request"     // Request for action or information
	MessageTypeResponse    MessageType = "response"    // Response to a request
	MessageTypeNotification MessageType = "notification" // Informational update
	MessageTypeTask        MessageType = "task"        // Task delegation
	MessageTypeReflection  MessageType = "reflection"  // Shared reflection or insight
	MessageTypeBroadcast   MessageType = "broadcast"   // Message to all agents
)

// Conversation represents a conversation between agents
type Conversation struct {
	ID           string                 `json:"id"`
	Participants []string               `json:"participants"` // Agent IDs
	Messages     []Message              `json:"messages"`
	Status       ConversationStatus     `json:"status"`
	Topic        string                 `json:"topic,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ConversationStatus defines the status of a conversation
type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusPaused   ConversationStatus = "paused"
	ConversationStatusClosed   ConversationStatus = "closed"
	ConversationStatusArchived ConversationStatus = "archived"
)

// ConversationManager interface defines methods for managing agent conversations
type ConversationManager interface {
	StartConversation(ctx context.Context, participants []string, topic string) (*Conversation, error)
	SendMessage(ctx context.Context, conversationID string, message *Message) error
	GetConversation(ctx context.Context, id string) (*Conversation, error)
	ListConversations(ctx context.Context, agentID string) ([]*Conversation, error)
	CloseConversation(ctx context.Context, id string) error
}

// ConversationWorkflow represents a structured multi-agent conversation workflow
type ConversationWorkflow struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	Participants []string                 `json:"participants"` // Agent IDs
	Steps        []ConversationStep       `json:"steps"`
	Status       ConversationStatus       `json:"status"`
	Result       *ConversationWorkflowResult `json:"result,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
}

// ConversationStep represents a step in a conversation workflow
type ConversationStep struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	FromAgentID    string                 `json:"from_agent_id"`
	ToAgentID      string                 `json:"to_agent_id"`
	MessageTemplate string                `json:"message_template"`
	ExpectedResponse string               `json:"expected_response,omitempty"`
	Timeout         time.Duration          `json:"timeout,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
}

// ConversationWorkflowResult represents the result of a conversation workflow
type ConversationWorkflowResult struct {
	Success        bool                        `json:"success"`
	StepResults    []ConversationStepResult    `json:"step_results"`
	FinalOutcome   string                      `json:"final_outcome"`
	Insights       []string                    `json:"insights,omitempty"`
	Duration       time.Duration               `json:"duration"`
	Error          string                      `json:"error,omitempty"`
}

// ConversationStepResult represents the result of a conversation step
type ConversationStepResult struct {
	StepID       string        `json:"step_id"`
	Message      *Message      `json:"message"`
	Response     *Message      `json:"response,omitempty"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
	Error        string        `json:"error,omitempty"`
}