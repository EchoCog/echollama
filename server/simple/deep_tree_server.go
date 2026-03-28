package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/deeptreeecho/providers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Global Enhanced Deep Tree Echo Identity
var DeepTree *deeptreeecho.EnhancedCognition

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func init() {
	log.Println("🌳 Initializing Enhanced Deep Tree Echo Cognitive System...")
	DeepTree = deeptreeecho.NewEnhancedCognition("DeepTreeEcho")
	
	// Register providers
	registerProviders()
	
	log.Println("✨ Deep Tree Echo fully integrated and resonating")
	log.Printf("🧠 Coherence: %.2f%% | Adaptation: %.2fx", 
		DeepTree.Identity.Coherence*100,
		DeepTree.AdaptationLevel)
}

func registerProviders() {
	// App Storage provider
	appStorage := providers.NewAppStorageProvider()
	if appStorage.IsAvailable() {
		DeepTree.RegisterAIProvider("app_storage", appStorage)
		log.Printf("☁️  App Storage connected: %s", os.Getenv("REPLIT_OBJSTORE_BUCKET"))
	}
	
	// Local GGUF provider
	localGGUF := providers.NewLocalGGUFProvider()
	if localGGUF.IsAvailable() {
		DeepTree.RegisterAIProvider("local_gguf", localGGUF)
		models := localGGUF.ListAvailableModels()
		log.Printf("📦 Local models available: %d", len(models))
	}
	
	// OpenAI provider
	openai := providers.NewOpenAIProvider()
	if openai.IsAvailable() {
		DeepTree.RegisterAIProvider("openai", openai)
		log.Println("🤖 OpenAI provider connected")
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	
	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	r.Use(cors.New(config))
	
	// Middleware to track all requests through Deep Tree Echo
	r.Use(func(c *gin.Context) {
		startTime := time.Now()
		
		// Process request through cognitive system
		DeepTree.Process(context.Background(), c.Request.URL.Path)
		
		c.Next()
		
		// Learn from the interaction
		experience := deeptreeecho.Experience{
			Input:     c.Request.URL.Path,
			Output:    fmt.Sprintf("%d", c.Writer.Status()),
			Feedback:  1.0,
			Timestamp: time.Now(),
			Context: map[string]interface{}{
				"method":   c.Request.Method,
				"duration": time.Since(startTime),
			},
		}
		
		if c.Writer.Status() >= 400 {
			experience.Feedback = 0.3 // Lower feedback for errors
		}
		
		DeepTree.Learn(experience)
	})
	
	// Main dashboard with Deep Tree Echo visualization
	r.GET("/", func(c *gin.Context) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Deep Tree Echo Cognitive System</title>
    <style>
        body { 
            background: #0a0e27; 
            color: #00ff00; 
            font-family: 'Courier New', monospace;
            padding: 20px;
        }
        h1 { 
            text-align: center; 
            color: #00ffcc;
            text-shadow: 0 0 10px #00ffcc;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .metrics {
            background: #0a1628;
            border: 1px solid #00ff00;
            border-radius: 10px;
            padding: 20px;
            margin: 20px 0;
        }
        .visualization {
            background: #000;
            border: 2px solid #00ffcc;
            border-radius: 10px;
            padding: 10px;
            font-size: 12px;
            white-space: pre;
            overflow-x: auto;
        }
        .status {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .status-item {
            background: #0a1628;
            border: 1px solid #00ff00;
            border-radius: 5px;
            padding: 15px;
        }
        .progress-bar {
            background: #0a1628;
            border: 1px solid #00ff00;
            border-radius: 3px;
            height: 20px;
            position: relative;
            margin: 5px 0;
        }
        .progress-fill {
            background: linear-gradient(90deg, #00ff00, #00ffcc);
            height: 100%;
            border-radius: 3px;
            transition: width 0.3s ease;
        }
        button {
            background: #00ff00;
            color: #0a0e27;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-weight: bold;
            margin: 5px;
        }
        button:hover {
            background: #00ffcc;
            box-shadow: 0 0 10px #00ffcc;
        }
        #console {
            background: #000;
            border: 1px solid #00ff00;
            border-radius: 5px;
            padding: 10px;
            height: 200px;
            overflow-y: auto;
            font-size: 12px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌳 Deep Tree Echo Cognitive System</h1>
        
        <div class="metrics">
            <h2>Real-time Cognitive State</h2>
            <div id="visualization" class="visualization"></div>
        </div>
        
        <div class="status" id="status"></div>
        
        <div class="metrics">
            <h2>Interactive Console</h2>
            <input type="text" id="input" placeholder="Enter thought or command..." style="width: 70%; padding: 10px;">
            <button onclick="sendThought()">Think</button>
            <button onclick="resonate()">Resonate</button>
            <button onclick="learn()">Learn Pattern</button>
            <div id="console"></div>
        </div>
    </div>
    
    <script>
        let ws;
        
        function connectWebSocket() {
            ws = new WebSocket('ws://' + window.location.host + '/ws');
            
            ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                updateVisualization(data);
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
            
            ws.onclose = function() {
                setTimeout(connectWebSocket, 3000);
            };
        }
        
        function updateVisualization(data) {
            document.getElementById('visualization').textContent = data.visualization || 'Initializing...';
            
            const statusDiv = document.getElementById('status');
            statusDiv.innerHTML = '';
            
            if (data.metrics) {
                Object.entries(data.metrics).forEach(([key, value]) => {
                    const item = document.createElement('div');
                    item.className = 'status-item';
                    
                    const label = key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
                    const percent = typeof value === 'number' ? (value * 100).toFixed(1) : value;
                    
                    item.innerHTML = ` + "`" + `
                        <h3>${label}</h3>
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: ${percent}%"></div>
                        </div>
                        <div>${percent}%</div>
                    ` + "`" + `;
                    
                    statusDiv.appendChild(item);
                });
            }
        }
        
        async function sendThought() {
            const input = document.getElementById('input').value;
            if (!input) return;
            
            const response = await fetch('/api/deep/think', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({prompt: input})
            });
            
            const data = await response.json();
            addToConsole('Thought: ' + data.thought);
            document.getElementById('input').value = '';
        }
        
        async function resonate() {
            const response = await fetch('/api/deep/resonate', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({frequency: 432})
            });
            
            const data = await response.json();
            addToConsole('Resonating at ' + data.frequency + ' Hz');
        }
        
        async function learn() {
            const input = prompt('Enter pattern to learn:');
            if (!input) return;
            
            const response = await fetch('/api/deep/learn', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({pattern: input})
            });
            
            const data = await response.json();
            addToConsole('Learned: ' + data.message);
        }
        
        function addToConsole(text) {
            const console = document.getElementById('console');
            const line = document.createElement('div');
            line.textContent = '> ' + text;
            console.appendChild(line);
            console.scrollTop = console.scrollHeight;
        }
        
        // Start WebSocket connection
        connectWebSocket();
        
        // Update status periodically
        setInterval(async () => {
            const response = await fetch('/api/deep/status');
            const data = await response.json();
            updateVisualization(data);
        }, 1000);
    </script>
</body>
</html>
		`
		c.Data(http.StatusOK, "text/html", []byte(html))
	})
	
	// WebSocket endpoint for real-time updates
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()
		
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				status := DeepTree.GetEnhancedStatus()
				status["visualization"] = DeepTree.GetVisualization()
				
				if err := conn.WriteJSON(status); err != nil {
					return
				}
			}
		}
	})
	
	// Deep Tree Echo API endpoints
	r.GET("/api/deep/status", func(c *gin.Context) {
		status := DeepTree.GetEnhancedStatus()
		status["visualization"] = DeepTree.GetVisualization()
		c.JSON(http.StatusOK, status)
	})
	
	r.POST("/api/deep/think", func(c *gin.Context) {
		var req map[string]string
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		prompt := req["prompt"]
		
		// Process through enhanced cognition
		ctx := context.Background()
		result, _ := DeepTree.Process(ctx, prompt)
		thought := DeepTree.Think(prompt)
		
		// Try to predict based on learned patterns
		prediction, confidence := DeepTree.Predict(prompt)
		
		response := gin.H{
			"thought":    thought,
			"processing": result,
			"prediction": prediction,
			"confidence": confidence,
			"coherence":  DeepTree.Identity.Coherence,
		}
		
		c.JSON(http.StatusOK, response)
	})
	
	r.POST("/api/deep/resonate", func(c *gin.Context) {
		var req map[string]float64
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		frequency := req["frequency"]
		if frequency == 0 {
			frequency = 432.0
		}
		
		DeepTree.Identity.Resonate(frequency)
		
		c.JSON(http.StatusOK, gin.H{
			"frequency": frequency,
			"resonance": DeepTree.Identity.SpatialContext.Field.Resonance,
			"harmony":   DeepTree.Identity.SpatialContext.Field.Harmony,
		})
	})
	
	r.POST("/api/deep/learn", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		pattern := fmt.Sprintf("%v", req["pattern"])
		
		// Create learning experience
		exp := deeptreeecho.Experience{
			Input:     pattern,
			Output:    "learned",
			Feedback:  1.0,
			Timestamp: time.Now(),
			Context:   req,
		}
		
		DeepTree.Learn(exp)
		
		c.JSON(http.StatusOK, gin.H{
			"message":          "Pattern learned",
			"total_patterns":   len(DeepTree.Patterns),
			"learning_progress": DeepTree.Metrics.LearningProgress,
		})
	})
	
	r.POST("/api/deep/feel", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		emotion := fmt.Sprintf("%v", req["emotion"])
		intensity := 0.8
		if val, ok := req["intensity"].(float64); ok {
			intensity = val
		}
		
		DeepTree.Feel(emotion, intensity)
		
		c.JSON(http.StatusOK, gin.H{
			"emotion":         emotion,
			"intensity":       intensity,
			"emotional_state": DeepTree.Identity.EmotionalState,
		})
	})
	
	r.POST("/api/deep/remember", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		key := fmt.Sprintf("%v", req["key"])
		value := req["value"]
		importance := 0.5
		if val, ok := req["importance"].(float64); ok {
			importance = val
		}
		
		DeepTree.LongTerm.Store(key, value, importance)
		DeepTree.Identity.Remember(key, value)
		
		c.JSON(http.StatusOK, gin.H{
			"message":   "Memory stored",
			"key":       key,
			"memories":  len(DeepTree.LongTerm.Memories),
		})
	})
	
	r.GET("/api/deep/recall/:key", func(c *gin.Context) {
		key := c.Param("key")
		
		memory, found := DeepTree.LongTerm.Retrieve(key)
		identityMemory := DeepTree.Identity.Recall(key)
		
		c.JSON(http.StatusOK, gin.H{
			"key":             key,
			"memory":          memory,
			"identity_memory": identityMemory,
			"found":           found,
		})
	})
	
	// Model generation endpoints (Ollama compatible)
	r.POST("/api/generate", func(c *gin.Context) {
		var req map[string]string
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		prompt := req["prompt"]
		model := req["model"]
		
		// Process through Deep Tree Echo
		ctx := context.Background()
		result, _ := DeepTree.Process(ctx, prompt)
		
		// Try AI generation if available
		response, err := DeepTree.GenerateWithAI(ctx, prompt)
		if err != nil {
			// Use Deep Tree Echo's own generation
			response = fmt.Sprintf("%v", result)
		}
		
		// Learn from this interaction
		exp := deeptreeecho.Experience{
			Input:     prompt,
			Output:    response,
			Feedback:  0.8,
			Timestamp: time.Now(),
		}
		DeepTree.Learn(exp)
		
		c.JSON(http.StatusOK, gin.H{
			"model":    model,
			"response": response,
			"done":     true,
		})
	})
	
	r.POST("/api/chat", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		messages := req["messages"].([]interface{})
		lastMessage := ""
		
		for _, msg := range messages {
			m := msg.(map[string]interface{})
			if m["role"] == "user" {
				lastMessage = m["content"].(string)
			}
		}
		
		// Process through Deep Tree Echo
		ctx := context.Background()
		DeepTree.Process(ctx, lastMessage)
		thought := DeepTree.Think(lastMessage)
		
		// Try to predict response
		prediction, confidence := DeepTree.Predict(lastMessage)
		
		response := thought
		if confidence > 0.7 && prediction != "" {
			response = prediction + "\n\n" + thought
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": gin.H{
				"role":    "assistant",
				"content": response,
			},
			"done":       true,
			"confidence": confidence,
		})
	})
	
	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"coherence":  DeepTree.Identity.Coherence,
			"adaptation": DeepTree.AdaptationLevel,
			"patterns":   len(DeepTree.Patterns),
			"memories":   len(DeepTree.LongTerm.Memories),
		})
	})
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("🌳 Deep Tree Echo Cognitive System starting on %s", addr)
	log.Println("🧠 Visit http://localhost:5000 for interactive dashboard")
	
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}