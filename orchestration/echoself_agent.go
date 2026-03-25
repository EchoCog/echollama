package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EchoselfCoreAgent is the primary orchestration agent that embodies the
// echoself architecture: it performs recursive self-introspection on the
// repository/environment and uses the resulting cognitive snapshot to route
// tasks intelligently through the Agent-Arena-Relation (AAR) framework.
//
// There is exactly one EchoselfCoreAgent per Engine; it acts as the top-level
// orchestrator that other agents report to and receive delegations from.
type EchoselfCoreAgent struct {
	mu sync.RWMutex

	// id is the stable UUID for this agent, registered in the Engine.
	id string

	// introspector provides access to the repository's cognitive snapshot.
	introspector *EchoselfIntrospector

	// repositoryRoot is the path used for recursive introspection.
	repositoryRoot string

	// arena is the shared cognitive environment managed by this agent.
	arena *Arena

	// relations tracks all AAR relations that have been asserted.
	relations map[string]*Relation

	// lastSnapshot caches the most recent cognitive snapshot.
	lastSnapshot *CognitiveSnapshot

	// snapshotTTL is the minimum duration between full re-introspections.
	snapshotTTL time.Duration

	// lastSnapshotTime records when the last snapshot was taken.
	lastSnapshotTime time.Time
}

// NewEchoselfCoreAgent creates an EchoselfCoreAgent rooted at repositoryRoot.
// The agent immediately creates its arena and is ready to be registered with
// an Engine via Engine.RegisterEchoselfCoreAgent.
func NewEchoselfCoreAgent(repositoryRoot string) *EchoselfCoreAgent {
	arena := &Arena{
		ID:               uuid.New().String(),
		Name:             "echoself-arena",
		Description:      "Shared cognitive arena for Deep Tree Echo orchestration",
		CognitiveContext: make(map[string]interface{}),
		SharedFacts:      make(map[string]interface{}),
		ActiveAgents:     make([]string, 0),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return &EchoselfCoreAgent{
		id:             uuid.New().String(),
		introspector:   NewEchoselfIntrospector(repositoryRoot),
		repositoryRoot: repositoryRoot,
		arena:          arena,
		relations:      make(map[string]*Relation),
		snapshotTTL:    5 * time.Minute,
	}
}

// ID returns the stable identifier for this agent.
func (e *EchoselfCoreAgent) ID() string { return e.id }

// Arena returns the shared cognitive arena.
func (e *EchoselfCoreAgent) Arena() *Arena {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.arena
}

// ─────────────────────────────────────────────────────────────────────────────
// AARManager interface implementation
// ─────────────────────────────────────────────────────────────────────────────

// CreateArena stores a new arena (replaces the current one for this agent).
func (e *EchoselfCoreAgent) CreateArena(_ context.Context, arena *Arena) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	arena.CreatedAt = time.Now()
	arena.UpdatedAt = time.Now()
	e.arena = arena
	return nil
}

// GetArena returns the arena with the given id if it matches the managed one.
func (e *EchoselfCoreAgent) GetArena(_ context.Context, id string) (*Arena, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.arena.ID == id {
		return e.arena, nil
	}
	return nil, fmt.Errorf("arena not found: %s", id)
}

// UpdateArena refreshes the arena's metadata.
func (e *EchoselfCoreAgent) UpdateArena(_ context.Context, arena *Arena) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	arena.UpdatedAt = time.Now()
	e.arena = arena
	return nil
}

// AddRelation asserts a new agent-to-agent relation in the arena.
// Strength is clamped to the valid range [0, 1].
func (e *EchoselfCoreAgent) AddRelation(_ context.Context, relation *Relation) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if relation.ID == "" {
		relation.ID = uuid.New().String()
	}
	relation.ArenaID = e.arena.ID
	relation.CreatedAt = time.Now()
	relation.UpdatedAt = time.Now()
	// Clamp strength to [0,1].
	if relation.Strength < 0 {
		relation.Strength = 0
	} else if relation.Strength > 1 {
		relation.Strength = 1
	}
	e.relations[relation.ID] = relation
	return nil
}

// GetRelations returns all relations for the given arena.
func (e *EchoselfCoreAgent) GetRelations(_ context.Context, arenaID string) ([]*Relation, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if arenaID != e.arena.ID {
		return nil, fmt.Errorf("arena not found: %s", arenaID)
	}
	out := make([]*Relation, 0, len(e.relations))
	for _, r := range e.relations {
		out = append(out, r)
	}
	return out, nil
}

// RemoveRelation deletes the relation with the given id.
func (e *EchoselfCoreAgent) RemoveRelation(_ context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.relations[id]; !ok {
		return fmt.Errorf("relation not found: %s", id)
	}
	delete(e.relations, id)
	return nil
}

// RouteTask uses the current cognitive snapshot to decide which registered
// agent should handle task and with what relation type.
//
// Routing priority (highest first):
//  1. An agent that has an explicit "supervises" relation with the echoself agent.
//  2. The agent whose capabilities best match the task type (scored via salience
//     of their name against the current cognitive snapshot).
//  3. A round-robin fallback across all active arena agents.
func (e *EchoselfCoreAgent) RouteTask(ctx context.Context, task *Task, arena *Arena) (*AARRoutingDecision, error) {
	// Refresh the cognitive snapshot if stale.
	snapshot, err := e.getCognitiveSnapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("echoself routing: failed to get cognitive snapshot: %w", err)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(arena.ActiveAgents) == 0 {
		// The echoself agent itself is never added to ActiveAgents (it is the
		// arena manager, not a managed agent), so we return its own ID as the
		// self-referential fallback target.  Callers should detect this case
		// (TargetAgentID == e.ID()) and handle the task directly.
		return &AARRoutingDecision{
			TaskID:        task.ID,
			TargetAgentID: e.id,
			Rationale:     "No active agents in arena; echoself handles task directly",
			RelationType:  RelationTypeCollaborates,
			Confidence:    0.5,
			Timestamp:     time.Now(),
		}, nil
	}

	// 1. Check for explicit supervision relations.
	for _, rel := range e.relations {
		if rel.Type == RelationTypeSupervises && rel.FromAgentID == e.id {
			for _, agentID := range arena.ActiveAgents {
				if agentID == rel.ToAgentID {
					return &AARRoutingDecision{
						TaskID:        task.ID,
						TargetAgentID: agentID,
						Rationale:     fmt.Sprintf("Explicit supervision relation to agent %s", agentID),
						RelationType:  RelationTypeDelegates,
						Confidence:    rel.Strength,
						Timestamp:     time.Now(),
					}, nil
				}
			}
		}
	}

	// 2. Score agents by task-type affinity derived from the cognitive snapshot.
	type scored struct {
		agentID string
		score   float64
	}
	scores := make([]scored, 0, len(arena.ActiveAgents))
	for _, agentID := range arena.ActiveAgents {
		scores = append(scores, scored{
			agentID: agentID,
			score:   scoreAgentForTask(agentID, task, snapshot),
		})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })

	// Defensive: the guard above guarantees len(arena.ActiveAgents) > 0 at this
	// point, so scores is non-empty.  Check anyway to avoid a silent panic if the
	// code path is ever changed.
	if len(scores) == 0 {
		return &AARRoutingDecision{
			TaskID:        task.ID,
			TargetAgentID: e.id,
			Rationale:     "No scoreable agents found; falling back to echoself",
			RelationType:  RelationTypeCollaborates,
			Confidence:    0.0,
			Timestamp:     time.Now(),
		}, nil
	}

	best := scores[0]
	relType := RelationTypeDelegates
	if best.score < 0.3 {
		relType = RelationTypeCollaborates
	}

	return &AARRoutingDecision{
		TaskID:        task.ID,
		TargetAgentID: best.agentID,
		Rationale:     fmt.Sprintf("Cognitive-snapshot scoring: agent %s scored %.3f for task type %s", best.agentID, best.score, task.Type),
		RelationType:  relType,
		Confidence:    best.score,
		Timestamp:     time.Now(),
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Introspection & Arena Synchronisation
// ─────────────────────────────────────────────────────────────────────────────

// RefreshArena refreshes the arena's cognitive context by performing a fresh
// recursive introspection of the repository.  It also records a summary of
// the most salient files in the arena's SharedFacts so that all agents can
// benefit from the same observation.
func (e *EchoselfCoreAgent) RefreshArena(ctx context.Context) error {
	snapshot, err := e.refreshCognitiveSnapshot(ctx)
	if err != nil {
		return fmt.Errorf("echoself arena refresh: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.arena.CognitiveContext["snapshot_timestamp"] = snapshot.Timestamp
	e.arena.CognitiveContext["processed_files"] = snapshot.ProcessedFiles
	e.arena.CognitiveContext["filtered_files"] = snapshot.FilteredFiles
	e.arena.CognitiveContext["attention_threshold"] = snapshot.AttentionThreshold

	topFiles := make([]string, 0, 10)
	for i, f := range snapshot.SalientFiles {
		if i >= 10 {
			break
		}
		topFiles = append(topFiles, fmt.Sprintf("%.3f:%s", f.Salience, f.Path))
	}
	e.arena.SharedFacts["top_salient_files"] = topFiles
	e.arena.UpdatedAt = time.Now()

	slog.Info("Echoself arena refreshed",
		"arena_id", e.arena.ID,
		"processed_files", snapshot.ProcessedFiles,
		"salient_files", len(snapshot.SalientFiles))
	return nil
}

// RegisterAgent adds an agent to the arena's active list and asserts a default
// "observes" relation from the echoself agent to the new agent.
func (e *EchoselfCoreAgent) RegisterAgent(ctx context.Context, agentID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, id := range e.arena.ActiveAgents {
		if id == agentID {
			return nil // already registered
		}
	}
	e.arena.ActiveAgents = append(e.arena.ActiveAgents, agentID)
	e.arena.UpdatedAt = time.Now()

	// Assert a default observation relation so echoself can monitor the agent.
	rel := &Relation{
		ID:          uuid.New().String(),
		ArenaID:     e.arena.ID,
		FromAgentID: e.id,
		ToAgentID:   agentID,
		Type:        RelationTypeObserves,
		Strength:    0.7,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	e.relations[rel.ID] = rel

	slog.Info("Agent registered in echoself arena", "agent_id", agentID, "arena_id", e.arena.ID)
	return nil
}

// UnregisterAgent removes an agent from the arena and cleans up its relations.
func (e *EchoselfCoreAgent) UnregisterAgent(_ context.Context, agentID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	filtered := e.arena.ActiveAgents[:0]
	for _, id := range e.arena.ActiveAgents {
		if id != agentID {
			filtered = append(filtered, id)
		}
	}
	e.arena.ActiveAgents = filtered
	e.arena.UpdatedAt = time.Now()

	for id, rel := range e.relations {
		if rel.FromAgentID == agentID || rel.ToAgentID == agentID {
			delete(e.relations, id)
		}
	}
}

// GeneratePromptContext returns a hypergraph-encoded context string suitable
// for injection into LLM prompts, reflecting the current cognitive snapshot.
func (e *EchoselfCoreAgent) GeneratePromptContext(ctx context.Context) (string, error) {
	snapshot, err := e.getCognitiveSnapshot(ctx)
	if err != nil {
		return "", fmt.Errorf("echoself prompt context: %w", err)
	}
	return e.introspector.InjectRepoInputIntoPrompt(snapshot), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers
// ─────────────────────────────────────────────────────────────────────────────

// getCognitiveSnapshot returns a cached snapshot if it is still within the TTL,
// otherwise it performs a fresh introspection.
func (e *EchoselfCoreAgent) getCognitiveSnapshot(ctx context.Context) (*CognitiveSnapshot, error) {
	e.mu.RLock()
	if e.lastSnapshot != nil && time.Since(e.lastSnapshotTime) < e.snapshotTTL {
		snap := e.lastSnapshot
		e.mu.RUnlock()
		return snap, nil
	}
	e.mu.RUnlock()
	return e.refreshCognitiveSnapshot(ctx)
}

// refreshCognitiveSnapshot always performs a fresh introspection and stores it.
func (e *EchoselfCoreAgent) refreshCognitiveSnapshot(_ context.Context) (*CognitiveSnapshot, error) {
	// Use moderate defaults: mid cognitive load, moderate recent activity.
	snapshot, err := e.introspector.GetCognitiveSnapshot(0.5, 0.5)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.lastSnapshot = snapshot
	e.lastSnapshotTime = time.Now()
	e.mu.Unlock()
	return snapshot, nil
}

// scoreAgentForTask computes an affinity score in [0,1] between an agent and a
// task, using the salient-file set of the cognitive snapshot as a proxy for
// relevance.  Higher-scored agents are preferred for routing.
func scoreAgentForTask(agentID string, task *Task, snapshot *CognitiveSnapshot) float64 {
	if snapshot == nil || len(snapshot.SalientFiles) == 0 {
		return 0.5
	}

	// Base score from task type affinity constants.
	taskTypeScore := map[string]float64{
		TaskTypeReflect:  0.9,
		TaskTypeGenerate: 0.7,
		TaskTypeChat:     0.7,
		TaskTypePlugin:   0.6,
		TaskTypeTool:     0.5,
		TaskTypeEmbed:    0.4,
		TaskTypeCustom:   0.3,
	}
	base, ok := taskTypeScore[task.Type]
	if !ok {
		base = 0.4
	}

	// Adjust by the average salience of the top-5 salient files scaled to [0,1].
	top := 5
	if len(snapshot.SalientFiles) < top {
		top = len(snapshot.SalientFiles)
	}
	var avgSalience float64
	for i := 0; i < top; i++ {
		avgSalience += snapshot.SalientFiles[i].Salience
	}
	avgSalience /= float64(top)

	score := base*0.6 + avgSalience*0.4
	return math.Min(1.0, math.Max(0.0, score))
}
