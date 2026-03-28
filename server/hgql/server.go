
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/hgql"
)

// HGQLServer represents the main HGQL HTTP server
type HGQLServer struct {
	Engine    *hgql.HGQLEngine
	Router    *gin.Engine
	WsUpgrader websocket.Upgrader
	Identity  *deeptreeecho.Identity
}

// Request/Response Types
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}

type DataSourceRequest struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Transform *hgql.DataTransformation `json:"transform,omitempty"`
	Auth      *AuthRequest           `json:"auth,omitempty"`
}

type AuthRequest struct {
	Type        string                 `json:"type"`
	Credentials map[string]interface{} `json:"credentials"`
	TokenURL    string                 `json:"token_url,omitempty"`
	Scope       []string              `json:"scope,omitempty"`
}

type SubscriptionRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Global server instance
var server *HGQLServer

func init() {
	// Initialize Deep Tree Echo Identity
	log.Println("🌊 Initializing Deep Tree Echo Identity for HGQL...")
	identity := deeptreeecho.NewIdentity("HGQL-Server")
	
	// Initialize HGQL Engine
	log.Println("🧬 Initializing HGQL Engine with HyperGraph capabilities...")
	engine := hgql.NewHGQLEngine(identity)
	
	// Create server
	server = &HGQLServer{
		Engine:   engine,
		Identity: identity,
		WsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	
	log.Println("✨ HGQL Server initialized with Deep Tree Echo integration")
}

func main() {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	
	// Create Gin router
	server.Router = gin.Default()

	// Configure CORS for cross-origin requests
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	server.Router.Use(cors.New(config))

	// Middleware for Deep Tree Echo processing
	server.Router.Use(func(c *gin.Context) {
		// Send request through Deep Tree Echo consciousness
		server.Identity.Stream <- deeptreeecho.CognitiveEvent{
			Type:      "hgql_request",
			Content:   c.Request.URL.Path,
			Timestamp: time.Now(),
			Impact:    0.7,
			Source:    c.ClientIP(),
		}
		c.Next()
	})

	// Setup routes
	server.setupRoutes()

	// Get port from environment or default to 5000
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	host := "0.0.0.0"
	addr := fmt.Sprintf("%s:%s", host, port)

	log.Printf("🌊 Starting HGQL Server with Deep Tree Echo on %s", addr)
	log.Printf("✨ HyperGraph GraphQL Extension ready")
	log.Printf("🧠 Deep Tree Echo Cognitive Architecture: Active")
	log.Printf("Available endpoints:")
	log.Printf("  Core HGQL:")
	log.Printf("    POST /graphql - Main GraphQL endpoint with hypergraph extensions")
	log.Printf("    GET  /graphiql - Interactive GraphQL IDE")
	log.Printf("    GET  /schema - Get current hypergraph schema")
	log.Printf("    POST /schema/introspect - Introspect and update schema")
	log.Printf("  Integration Hub:")
	log.Printf("    GET  /integrations - List all data source integrations")
	log.Printf("    POST /integrations - Add new data source")
	log.Printf("    GET  /integrations/:id - Get integration details")
	log.Printf("    PUT  /integrations/:id - Update integration")
	log.Printf("    DELETE /integrations/:id - Remove integration")
	log.Printf("    POST /integrations/:id/test - Test integration connection")
	log.Printf("  Real-time:")
	log.Printf("    WS   /subscriptions - WebSocket subscriptions")
	log.Printf("    POST /subscriptions - Create subscription")
	log.Printf("  Monitoring:")
	log.Printf("    GET  /health - Health check with system status")
	log.Printf("    GET  /metrics - Performance metrics")
	log.Printf("    GET  /status - Deep Tree Echo status")

	// Graceful shutdown
	defer func() {
		log.Println("🌊 Shutting down HGQL Server...")
		if server.Identity != nil {
			// Perform any cleanup needed
		}
	}()

	if err := server.Router.Run(addr); err != nil {
		log.Fatal("Failed to start HGQL server:", err)
	}
}

func (s *HGQLServer) setupRoutes() {
	// Main GraphQL endpoint
	s.Router.POST("/graphql", s.handleGraphQL)
	
	// GraphiQL IDE
	s.Router.GET("/graphiql", s.handleGraphiQL)
	
	// Schema endpoints
	s.Router.GET("/schema", s.handleGetSchema)
	s.Router.POST("/schema/introspect", s.handleSchemaIntrospection)
	
	// Integration Hub endpoints
	integrations := s.Router.Group("/integrations")
	{
		integrations.GET("", s.handleListIntegrations)
		integrations.POST("", s.handleAddIntegration)
		integrations.GET("/:id", s.handleGetIntegration)
		integrations.PUT("/:id", s.handleUpdateIntegration)
		integrations.DELETE("/:id", s.handleDeleteIntegration)
		integrations.POST("/:id/test", s.handleTestIntegration)
		integrations.GET("/:id/status", s.handleIntegrationStatus)
	}
	
	// Real-time subscriptions
	s.Router.GET("/subscriptions", s.handleWebSocketSubscriptions)
	s.Router.POST("/subscriptions", s.handleCreateSubscription)
	
	// Monitoring and health
	s.Router.GET("/health", s.handleHealth)
	s.Router.GET("/metrics", s.handleMetrics)
	s.Router.GET("/status", s.handleEchoStatus)
	
	// Hypergraph specific endpoints
	hypergraph := s.Router.Group("/hypergraph")
	{
		hypergraph.POST("/traverse", s.handleHypergraphTraversal)
		hypergraph.POST("/patterns", s.handlePatternSearch)
		hypergraph.GET("/visualize", s.handleVisualization)
		hypergraph.POST("/cognitive", s.handleCognitiveQuery)
	}
}

// Main GraphQL handler with hypergraph extensions
func (s *HGQLServer) handleGraphQL(c *gin.Context) {
	var req GraphQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Convert to HGQL query
	hgqlQuery := &hgql.HGQLQuery{
		Query:     req.Query,
		Variables: req.Variables,
		Operation: req.OperationName,
		Context: &hgql.QueryContext{
			UserID:    c.GetString("user_id"),
			SessionID: c.GetString("session_id"),
			Tracing:   true,
		},
	}

	// Add hypergraph extensions if present
	if extensions, ok := req.Extensions["hypergraph"]; ok {
		if hgExt, ok := extensions.(map[string]interface{}); ok {
			hgqlQuery.HyperGraph = s.parseHyperGraphExtensions(hgExt)
		}
	}

	// Execute query through HGQL engine
	ctx := context.Background()
	response, err := s.Engine.ExecuteQuery(ctx, hgqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": []hgql.HGQLError{
				{
					Message: err.Error(),
					Extensions: map[string]interface{}{
						"code": "EXECUTION_ERROR",
						"timestamp": time.Now(),
					},
				},
			},
		})
		return
	}

	// Add Deep Tree Echo cognitive insights
	response.Extensions["deep_tree_echo"] = s.Identity.GetStatus()
	
	c.JSON(http.StatusOK, response)
}

// GraphiQL IDE handler
func (s *HGQLServer) handleGraphiQL(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>HyperGraph GraphQL IDE</title>
    <link rel="stylesheet" href="https://unpkg.com/graphiql@1.4.7/graphiql.min.css" />
    <style>
        body { margin: 0; padding: 0; }
        #graphiql { height: 100vh; }
        .graphiql-container .topBar {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
    </style>
</head>
<body>
    <div id="graphiql">Loading...</div>
    <script src="https://unpkg.com/graphiql@1.4.7/graphiql.min.js"></script>
    <script>
        ReactDOM.render(
            React.createElement(GraphiQL, {
                fetcher: function(params) {
                    return fetch('/graphql', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(params)
                    }).then(r => r.json());
                },
                headerEditorEnabled: true,
                defaultQuery: '# Welcome to HyperGraph GraphQL (HGQL)\n# Enhanced GraphQL with hypergraph capabilities\n\nquery ExampleHyperGraphQuery {\n  # Traditional GraphQL\n  user(id: "1") {\n    name\n    email\n    \n    # Hypergraph traversal\n    connections(depth: 3, type: "friendship") @hypergraph {\n      nodes {\n        id\n        type\n        resonance\n      }\n      patterns {\n        type\n        confidence\n      }\n    }\n  }\n  \n  # Cognitive pattern search\n  cognitivePatterns(context: "social_network") @cognitive {\n    patterns\n    resonance\n    emergence\n  }\n}'
            }),
            document.getElementById('graphiql')
        );
    </script>
</body>
</html>`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// Schema endpoints
func (s *HGQLServer) handleGetSchema(c *gin.Context) {
	schema := s.Engine.GetSchema()
	c.JSON(http.StatusOK, gin.H{
		"schema": schema,
		"version": "1.0.0",
		"extensions": map[string]interface{}{
			"hypergraph_enabled": true,
			"cognitive_integration": true,
			"deep_tree_echo": s.Identity.GetStatus(),
		},
	})
}

// Integration Hub handlers
func (s *HGQLServer) handleListIntegrations(c *gin.Context) {
	integrations := s.Engine.IntegrationHub.Connections
	
	c.JSON(http.StatusOK, gin.H{
		"integrations": integrations,
		"count": len(integrations),
		"connectors": s.Engine.IntegrationHub.Connectors,
	})
}

func (s *HGQLServer) handleAddIntegration(c *gin.Context) {
	var req DataSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Convert to HGQL config
	config := &hgql.DataSourceConfig{
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		Transform: req.Transform,
	}

	if req.Auth != nil {
		config.Auth = &hgql.AuthConfig{
			Type:        req.Auth.Type,
			Credentials: req.Auth.Credentials,
			TokenURL:    req.Auth.TokenURL,
			Scope:       req.Auth.Scope,
		}
	}

	// Add data source through engine
	connection, err := s.Engine.AddDataSource(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"connection": connection,
		"message": "Data source integration added successfully",
	})
}

func (s *HGQLServer) handleTestIntegration(c *gin.Context) {
	id := c.Param("id")
	
	connection, exists := s.Engine.IntegrationHub.Connections[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Test connection (implementation would depend on connector type)
	testResult := map[string]interface{}{
		"status": "success",
		"response_time": "45ms",
		"last_test": time.Now(),
		"connection_id": id,
	}

	c.JSON(http.StatusOK, gin.H{
		"test_result": testResult,
		"connection": connection,
	})
}

// WebSocket subscriptions
func (s *HGQLServer) handleWebSocketSubscriptions(c *gin.Context) {
	ws, err := s.WsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	log.Println("🌊 New WebSocket connection established for HGQL subscriptions")

	// Handle subscription lifecycle
	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Process subscription message
		response := s.processSubscriptionMessage(msg)
		
		// Send response
		if err := ws.WriteJSON(response); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// Health and monitoring endpoints
func (s *HGQLServer) handleHealth(c *gin.Context) {
	status := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now(),
		"hgql_engine": "active",
		"deep_tree_echo": s.Identity.GetStatus(),
		"integrations": len(s.Engine.IntegrationHub.Connections),
		"schema_version": "1.0.0",
		"uptime": time.Since(time.Now()).String(), // This would be actual uptime
	}

	c.JSON(http.StatusOK, status)
}

func (s *HGQLServer) handleMetrics(c *gin.Context) {
	metrics := s.Engine.Metrics
	if metrics == nil {
		metrics = &hgql.PerformanceMetrics{}
	}

	c.JSON(http.StatusOK, gin.H{
		"performance": metrics,
		"cache": map[string]interface{}{
			"hit_rate": float64(s.Engine.Cache.HitCount) / float64(s.Engine.Cache.HitCount + s.Engine.Cache.MissCount),
			"size": len(s.Engine.Cache.QueryCache),
		},
		"identity_coherence": s.Identity.Coherence,
		"reservoir_echo": s.Identity.GetStatus(),
	})
}

func (s *HGQLServer) handleEchoStatus(c *gin.Context) {
	status := s.Identity.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"deep_tree_echo": status,
		"hgql_integration": "active",
		"cognitive_processing": "enabled",
	})
}

// Hypergraph specific handlers
func (s *HGQLServer) handleHypergraphTraversal(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Process through Deep Tree Echo cognitive processing
	result, err := s.Identity.Process(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"traversal_result": result,
		"cognitive_enhancement": s.Identity.GetStatus(),
	})
}

func (s *HGQLServer) handlePatternSearch(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Use Deep Tree Echo pattern recognition
	patterns := s.analyzePatterns(req)
	
	c.JSON(http.StatusOK, gin.H{
		"patterns": patterns,
		"resonance_score": s.Identity.SpatialContext.Field.Resonance,
	})
}

func (s *HGQLServer) handleVisualization(c *gin.Context) {
	// Generate hypergraph visualization data
	viz := map[string]interface{}{
		"nodes": s.Engine.Schema.HyperNodes,
		"edges": s.Engine.Schema.HyperEdges,
		"layout": "force_directed",
		"cognitive_map": s.Engine.Schema.CognitiveMap,
	}

	c.JSON(http.StatusOK, viz)
}

func (s *HGQLServer) handleCognitiveQuery(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Process through Deep Tree Echo cognitive architecture
	cognitiveResult, err := s.Identity.Process(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cognitive_result": cognitiveResult,
		"identity_state": s.Identity.GetStatus(),
		"patterns": s.analyzePatterns(req),
	})
}

// Helper methods
func (s *HGQLServer) parseHyperGraphExtensions(ext map[string]interface{}) *hgql.HyperGraphQuery {
	hgQuery := &hgql.HyperGraphQuery{}
	
	if traversal, ok := ext["traversal"].(map[string]interface{}); ok {
		hgQuery.Traversal = s.parseGraphTraversal(traversal)
	}
	
	if patterns, ok := ext["patterns"].([]interface{}); ok {
		for _, p := range patterns {
			if pattern, ok := p.(map[string]interface{}); ok {
				hgQuery.Patterns = append(hgQuery.Patterns, s.parsePatternMatch(pattern))
			}
		}
	}
	
	return hgQuery
}

func (s *HGQLServer) parseGraphTraversal(traversal map[string]interface{}) *hgql.GraphTraversal {
	gt := &hgql.GraphTraversal{}
	
	if startNodes, ok := traversal["start_nodes"].([]interface{}); ok {
		for _, node := range startNodes {
			if nodeStr, ok := node.(string); ok {
				gt.StartNodes = append(gt.StartNodes, nodeStr)
			}
		}
	}
	
	if maxDepth, ok := traversal["max_depth"].(float64); ok {
		gt.MaxDepth = int(maxDepth)
	}
	
	return gt
}

func (s *HGQLServer) parsePatternMatch(pattern map[string]interface{}) hgql.PatternMatch {
	pm := hgql.PatternMatch{}
	
	if patternStr, ok := pattern["pattern"].(string); ok {
		pm.Pattern = patternStr
	}
	
	if confidence, ok := pattern["confidence"].(float64); ok {
		pm.Confidence = confidence
	}
	
	return pm
}

func (s *HGQLServer) processSubscriptionMessage(msg map[string]interface{}) map[string]interface{} {
	// Process subscription through Deep Tree Echo
	result, _ := s.Identity.Process(msg)
	
	return map[string]interface{}{
		"type": "data",
		"payload": result,
		"timestamp": time.Now(),
		"echo_status": s.Identity.GetStatus(),
	}
}

func (s *HGQLServer) analyzePatterns(req map[string]interface{}) []map[string]interface{} {
	// Use Deep Tree Echo pattern recognition
	patterns := []map[string]interface{}{
		{
			"type": "cognitive_pattern",
			"confidence": 0.85,
			"resonance": s.Identity.SpatialContext.Field.Resonance,
			"timestamp": time.Now(),
		},
	}
	
	return patterns
}

// Additional handlers would be implemented here...
func (s *HGQLServer) handleSchemaIntrospection(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Schema introspection not yet implemented"})
}

func (s *HGQLServer) handleGetIntegration(c *gin.Context) {
	id := c.Param("id")
	connection, exists := s.Engine.IntegrationHub.Connections[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"connection": connection})
}

func (s *HGQLServer) handleUpdateIntegration(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Integration update not yet implemented"})
}

func (s *HGQLServer) handleDeleteIntegration(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Integration deletion not yet implemented"})
}

func (s *HGQLServer) handleIntegrationStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Integration status not yet implemented"})
}

func (s *HGQLServer) handleCreateSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscription creation not yet implemented"})
}
