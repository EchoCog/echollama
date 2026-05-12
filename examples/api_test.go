//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🧪 Deep Tree Echo API Test")
	fmt.Println("==========================")
	fmt.Println()

	baseURL := "http://localhost:8080/api"
	
	// Wait for server to be ready
	fmt.Println("⏳ Waiting for server to be ready...")
	if !waitForServer(baseURL, 30*time.Second) {
		fmt.Println("❌ Server not ready after 30 seconds")
		fmt.Println("💡 Please start the API server first: go run examples/api_server.go")
		return
	}
	fmt.Println("✅ Server is ready!")
	fmt.Println()

	// Test Deep Tree Echo status
	fmt.Println("📊 Testing Deep Tree Echo status...")
	if testDTEStatus(baseURL) {
		fmt.Println("✅ DTE status test passed")
	} else {
		fmt.Println("❌ DTE status test failed")
	}

	// Test initialization
	fmt.Println("🧠 Testing DTE initialization...")
	if testDTEInitialize(baseURL) {
		fmt.Println("✅ DTE initialization test passed")
	} else {
		fmt.Println("❌ DTE initialization test failed")
	}

	// Test diagnostics
	fmt.Println("🔧 Testing DTE diagnostics...")
	if testDTEDiagnostics(baseURL) {
		fmt.Println("✅ DTE diagnostics test passed")
	} else {
		fmt.Println("❌ DTE diagnostics test failed")
	}

	// Test dashboard data
	fmt.Println("📈 Testing dashboard data...")
	if testDTEDashboard(baseURL) {
		fmt.Println("✅ Dashboard data test passed")
	} else {
		fmt.Println("❌ Dashboard data test failed")
	}

	// Test agent listing
	fmt.Println("🤖 Testing agent listing...")
	if testAgentListing(baseURL) {
		fmt.Println("✅ Agent listing test passed")
	} else {
		fmt.Println("❌ Agent listing test failed")
	}

	// Test tools listing
	fmt.Println("🔧 Testing tools listing...")
	if testToolsListing(baseURL) {
		fmt.Println("✅ Tools listing test passed")
	} else {
		fmt.Println("❌ Tools listing test failed")
	}

	fmt.Println()
	fmt.Println("🎉 API tests completed!")
	fmt.Println("🌐 Open http://localhost:8080/dashboard.html to view the dashboard")
}

func waitForServer(baseURL string, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(baseURL + "/deep-tree-echo/status")
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func testDTEStatus(baseURL string) bool {
	resp, err := http.Get(baseURL + "/deep-tree-echo/status")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("   HTTP %d\n", resp.StatusCode)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   JSON decode error: %v\n", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok || status != "success" {
		fmt.Printf("   Unexpected status: %v\n", result["status"])
		return false
	}

	fmt.Printf("   Status: %s\n", status)
	return true
}

func testDTEInitialize(baseURL string) bool {
	resp, err := http.Post(baseURL+"/deep-tree-echo/initialize", "application/json", nil)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("   HTTP %d\n", resp.StatusCode)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   JSON decode error: %v\n", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok || status != "success" {
		fmt.Printf("   Unexpected status: %v\n", result["status"])
		return false
	}

	fmt.Printf("   Message: %v\n", result["message"])
	return true
}

func testDTEDiagnostics(baseURL string) bool {
	resp, err := http.Post(baseURL+"/deep-tree-echo/diagnostics", "application/json", nil)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("   HTTP %d\n", resp.StatusCode)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   JSON decode error: %v\n", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok || status != "success" {
		fmt.Printf("   Unexpected status: %v\n", result["status"])
		return false
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		fmt.Printf("   No data in response\n")
		return false
	}

	overallHealth, ok := data["overall_health"].(string)
	if !ok {
		fmt.Printf("   No overall_health in data\n")
		return false
	}

	fmt.Printf("   Overall Health: %s\n", overallHealth)
	return true
}

func testDTEDashboard(baseURL string) bool {
	resp, err := http.Get(baseURL + "/deep-tree-echo/dashboard")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("   HTTP %d\n", resp.StatusCode)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   JSON decode error: %v\n", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok || status != "success" {
		fmt.Printf("   Unexpected status: %v\n", result["status"])
		return false
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		fmt.Printf("   No data in response\n")
		return false
	}

	// Check for expected sections
	expectedSections := []string{"system_metrics", "identity_coherence", "memory_resonance", "echo_patterns"}
	for _, section := range expectedSections {
		if _, exists := data[section]; !exists {
			fmt.Printf("   Missing section: %s\n", section)
			return false
		}
	}

	fmt.Printf("   Dashboard sections: %d\n", len(data))
	return true
}

func testAgentListing(baseURL string) bool {
	resp, err := http.Get(baseURL + "/agents")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("   HTTP %d\n", resp.StatusCode)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   JSON decode error: %v\n", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok || status != "success" {
		fmt.Printf("   Unexpected status: %v\n", result["status"])
		return false
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		fmt.Printf("   No agents data in response\n")
		return false
	}

	fmt.Printf("   Agents count: %d\n", len(data))
	return true
}

func testToolsListing(baseURL string) bool {
	resp, err := http.Get(baseURL + "/orchestration/tools")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("   HTTP %d\n", resp.StatusCode)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   JSON decode error: %v\n", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok || status != "success" {
		fmt.Printf("   Unexpected status: %v\n", result["status"])
		return false
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		fmt.Printf("   No tools data in response\n")
		return false
	}

	fmt.Printf("   Tools count: %d\n", len(data))
	return true
}
