package orchestration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/EchoCog/echollama/api"
)

// ─────────────────────────────────────────────────────────────────────────────
// EchoselfCoreAgent unit tests
// ─────────────────────────────────────────────────────────────────────────────

func TestNewEchoselfCoreAgent(t *testing.T) {
	agent := NewEchoselfCoreAgent(t.TempDir())

	if agent == nil {
		t.Fatal("NewEchoselfCoreAgent should return a non-nil agent")
	}
	if agent.ID() == "" {
		t.Error("agent ID should be non-empty")
	}
	arena := agent.Arena()
	if arena == nil {
		t.Fatal("arena should be non-nil")
	}
	if arena.ID == "" {
		t.Error("arena ID should be non-empty")
	}
	if arena.Name != "echoself-arena" {
		t.Errorf("expected arena name 'echoself-arena', got %q", arena.Name)
	}
}

func TestEchoselfCoreAgent_ArenaOperations(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())
	arena := agent.Arena()

	// GetArena by valid id
	got, err := agent.GetArena(ctx, arena.ID)
	if err != nil {
		t.Fatalf("GetArena: %v", err)
	}
	if got.ID != arena.ID {
		t.Errorf("expected arena id %s, got %s", arena.ID, got.ID)
	}

	// GetArena with unknown id
	_, err = agent.GetArena(ctx, "unknown-id")
	if err == nil {
		t.Error("GetArena with unknown id should return an error")
	}
}

func TestEchoselfCoreAgent_RegisterAgent(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())

	if err := agent.RegisterAgent(ctx, "agent-1"); err != nil {
		t.Fatalf("RegisterAgent: %v", err)
	}
	if err := agent.RegisterAgent(ctx, "agent-2"); err != nil {
		t.Fatalf("RegisterAgent: %v", err)
	}

	// Idempotent – registering the same agent twice should be a no-op.
	if err := agent.RegisterAgent(ctx, "agent-1"); err != nil {
		t.Fatalf("RegisterAgent (duplicate): %v", err)
	}

	arena := agent.Arena()
	found := 0
	for _, id := range arena.ActiveAgents {
		if id == "agent-1" || id == "agent-2" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("expected 2 distinct active agents, got %d (%v)", found, arena.ActiveAgents)
	}

	// Observe relations should have been created.
	relations, err := agent.GetRelations(ctx, arena.ID)
	if err != nil {
		t.Fatalf("GetRelations: %v", err)
	}
	if len(relations) < 2 {
		t.Errorf("expected at least 2 relations, got %d", len(relations))
	}
	for _, r := range relations {
		if r.Type != RelationTypeObserves {
			t.Errorf("expected relation type %s, got %s", RelationTypeObserves, r.Type)
		}
	}
}

func TestEchoselfCoreAgent_UnregisterAgent(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())

	_ = agent.RegisterAgent(ctx, "agent-a")
	_ = agent.RegisterAgent(ctx, "agent-b")

	agent.UnregisterAgent(ctx, "agent-a")

	arena := agent.Arena()
	for _, id := range arena.ActiveAgents {
		if id == "agent-a" {
			t.Error("agent-a should have been removed from active agents")
		}
	}

	relations, _ := agent.GetRelations(ctx, arena.ID)
	for _, r := range relations {
		if r.FromAgentID == "agent-a" || r.ToAgentID == "agent-a" {
			t.Errorf("relations involving agent-a should have been removed")
		}
	}
}

func TestEchoselfCoreAgent_AddRemoveRelation(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())
	arena := agent.Arena()

	rel := &Relation{
		FromAgentID: agent.ID(),
		ToAgentID:   "some-agent",
		Type:        RelationTypeSupervises,
		Strength:    0.9,
	}
	if err := agent.AddRelation(ctx, rel); err != nil {
		t.Fatalf("AddRelation: %v", err)
	}
	if rel.ID == "" {
		t.Error("relation ID should be set after AddRelation")
	}
	if rel.ArenaID != arena.ID {
		t.Errorf("relation ArenaID should be %s, got %s", arena.ID, rel.ArenaID)
	}

	// Remove it.
	if err := agent.RemoveRelation(ctx, rel.ID); err != nil {
		t.Fatalf("RemoveRelation: %v", err)
	}

	// Removing non-existent relation should error.
	err := agent.RemoveRelation(ctx, rel.ID)
	if err == nil {
		t.Error("RemoveRelation of non-existent relation should return error")
	}
}

func TestEchoselfCoreAgent_RouteTask_NoAgents(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())
	task := &Task{
		ID:   "task-1",
		Type: TaskTypeGenerate,
	}

	decision, err := agent.RouteTask(ctx, task, agent.Arena())
	if err != nil {
		t.Fatalf("RouteTask: %v", err)
	}
	// With no active agents, echoself itself should be the target.
	if decision.TargetAgentID != agent.ID() {
		t.Errorf("expected self-routing when no agents in arena, got %s", decision.TargetAgentID)
	}
}

func TestEchoselfCoreAgent_RouteTask_WithAgents(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())
	_ = agent.RegisterAgent(ctx, "worker-1")
	_ = agent.RegisterAgent(ctx, "worker-2")

	task := &Task{
		ID:   "task-2",
		Type: TaskTypeReflect,
	}

	decision, err := agent.RouteTask(ctx, task, agent.Arena())
	if err != nil {
		t.Fatalf("RouteTask: %v", err)
	}
	if decision.TargetAgentID == "" {
		t.Error("routing decision should have a target agent")
	}
	if decision.Confidence < 0 || decision.Confidence > 1 {
		t.Errorf("confidence should be in [0,1], got %f", decision.Confidence)
	}
	if !decision.Timestamp.Before(time.Now().Add(time.Second)) {
		t.Error("routing decision timestamp should be set to now")
	}
}

func TestEchoselfCoreAgent_RouteTask_SupervisionRelation(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())
	_ = agent.RegisterAgent(ctx, "worker-supervised")
	_ = agent.RegisterAgent(ctx, "worker-other")

	// Assert explicit supervision.
	_ = agent.AddRelation(ctx, &Relation{
		FromAgentID: agent.ID(),
		ToAgentID:   "worker-supervised",
		Type:        RelationTypeSupervises,
		Strength:    0.95,
	})

	task := &Task{ID: "task-3", Type: TaskTypeChat}
	decision, err := agent.RouteTask(ctx, task, agent.Arena())
	if err != nil {
		t.Fatalf("RouteTask: %v", err)
	}
	if decision.TargetAgentID != "worker-supervised" {
		t.Errorf("expected supervised agent to be selected, got %s", decision.TargetAgentID)
	}
	if decision.RelationType != RelationTypeDelegates {
		t.Errorf("expected relation type Delegates, got %s", decision.RelationType)
	}
}

func TestEchoselfCoreAgent_RefreshArena(t *testing.T) {
	ctx := context.Background()
	// Use a real temp dir with a few files so introspection has something to find.
	dir := t.TempDir()

	agent := NewEchoselfCoreAgent(dir)
	if err := agent.RefreshArena(ctx); err != nil {
		t.Fatalf("RefreshArena: %v", err)
	}

	arena := agent.Arena()
	if arena.UpdatedAt.IsZero() {
		t.Error("arena.UpdatedAt should be set after refresh")
	}
	if _, ok := arena.CognitiveContext["snapshot_timestamp"]; !ok {
		t.Error("arena cognitive context should contain snapshot_timestamp")
	}
}

func TestEchoselfCoreAgent_GeneratePromptContext(t *testing.T) {
	ctx := context.Background()
	agent := NewEchoselfCoreAgent(t.TempDir())

	prompt, err := agent.GeneratePromptContext(ctx)
	if err != nil {
		t.Fatalf("GeneratePromptContext: %v", err)
	}
	if !strings.Contains(prompt, "Hypergraph-Encoded Repository Analysis") {
		t.Errorf("prompt context should contain the expected header, got: %s", prompt)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Engine integration tests for the echoself core agent
// ─────────────────────────────────────────────────────────────────────────────

func TestEngine_RegisterEchoselfCoreAgent(t *testing.T) {
	ctx := context.Background()
	client := api.Client{}
	engine := NewEngine(client)

	// No echoself agent yet.
	if engine.GetEchoselfCoreAgent() != nil {
		t.Error("engine should have no echoself agent before registration")
	}

	eca := NewEchoselfCoreAgent(t.TempDir())
	if err := engine.RegisterEchoselfCoreAgent(ctx, eca); err != nil {
		t.Fatalf("RegisterEchoselfCoreAgent: %v", err)
	}

	if engine.GetEchoselfCoreAgent() == nil {
		t.Error("engine should have an echoself agent after registration")
	}

	// The echoself agent should be findable via GetAgent.
	stored, err := engine.GetAgent(ctx, eca.ID())
	if err != nil {
		t.Fatalf("GetAgent for echoself: %v", err)
	}
	if stored.Type != AgentTypeEchoself {
		t.Errorf("expected agent type %s, got %s", AgentTypeEchoself, stored.Type)
	}
}

func TestEngine_CreateAgent_RegistersInArena(t *testing.T) {
	ctx := context.Background()
	client := api.Client{}
	engine := NewEngine(client)

	eca := NewEchoselfCoreAgent(t.TempDir())
	_ = engine.RegisterEchoselfCoreAgent(ctx, eca)

	agent := &Agent{Name: "arena-worker", Type: AgentTypeGeneral}
	if err := engine.CreateAgent(ctx, agent); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	arena := eca.Arena()
	found := false
	for _, id := range arena.ActiveAgents {
		if id == agent.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("newly created agent %s should be in echoself arena", agent.ID)
	}
}

func TestEngine_DeleteAgent_UnregistersFromArena(t *testing.T) {
	ctx := context.Background()
	client := api.Client{}
	engine := NewEngine(client)

	eca := NewEchoselfCoreAgent(t.TempDir())
	_ = engine.RegisterEchoselfCoreAgent(ctx, eca)

	agent := &Agent{Name: "temp-worker", Type: AgentTypeGeneral}
	_ = engine.CreateAgent(ctx, agent)

	if err := engine.DeleteAgent(ctx, agent.ID); err != nil {
		t.Fatalf("DeleteAgent: %v", err)
	}

	arena := eca.Arena()
	for _, id := range arena.ActiveAgents {
		if id == agent.ID {
			t.Errorf("deleted agent %s should be removed from echoself arena", agent.ID)
		}
	}
}

func TestEngine_RouteTaskViaEchoself_NoAgent(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(api.Client{})

	task := &Task{ID: "t1", Type: TaskTypeGenerate}
	decision, err := engine.RouteTaskViaEchoself(ctx, task)
	if err != nil {
		t.Fatalf("RouteTaskViaEchoself: %v", err)
	}
	// No echoself agent → should return nil decision.
	if decision != nil {
		t.Error("expected nil decision when no echoself agent is registered")
	}
}

func TestEngine_RouteTaskViaEchoself_WithAgent(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(api.Client{})

	eca := NewEchoselfCoreAgent(t.TempDir())
	_ = engine.RegisterEchoselfCoreAgent(ctx, eca)

	worker := &Agent{Name: "route-worker"}
	_ = engine.CreateAgent(ctx, worker)

	task := &Task{ID: "t2", Type: TaskTypeReflect}
	decision, err := engine.RouteTaskViaEchoself(ctx, task)
	if err != nil {
		t.Fatalf("RouteTaskViaEchoself: %v", err)
	}
	if decision == nil {
		t.Fatal("expected a routing decision, got nil")
	}
	if decision.TaskID != task.ID {
		t.Errorf("expected task ID %s, got %s", task.ID, decision.TaskID)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// AAR types tests
// ─────────────────────────────────────────────────────────────────────────────

func TestAARTypes(t *testing.T) {
	arena := &Arena{
		ID:               "arena-1",
		Name:             "test-arena",
		CognitiveContext: map[string]interface{}{"key": "value"},
		SharedFacts:      map[string]interface{}{},
		ActiveAgents:     []string{"agent-a"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if arena.Name != "test-arena" {
		t.Errorf("Arena name mismatch: %s", arena.Name)
	}

	rel := &Relation{
		FromAgentID: "from",
		ToAgentID:   "to",
		Type:        RelationTypeCollaborates,
		Strength:    0.8,
	}
	if rel.Type != RelationTypeCollaborates {
		t.Errorf("Relation type mismatch: %s", rel.Type)
	}

	decision := &AARRoutingDecision{
		TaskID:        "task-x",
		TargetAgentID: "agent-a",
		RelationType:  RelationTypeDelegates,
		Confidence:    0.75,
		Timestamp:     time.Now(),
	}
	if decision.Confidence != 0.75 {
		t.Errorf("Confidence mismatch: %f", decision.Confidence)
	}
}

func TestAgentTypeEchoself(t *testing.T) {
	if AgentTypeEchoself != "echoself" {
		t.Errorf("AgentTypeEchoself constant value should be 'echoself', got %s", AgentTypeEchoself)
	}
}
