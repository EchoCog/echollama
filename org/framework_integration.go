
package org

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

// GlobalIdentityFramework holds the singleton instance
var (
	GlobalIdentityFramework *OrganizationalIdentityFramework
	frameworkOnce          sync.Once
)

// InitializeGlobalFramework initializes the global identity framework
func InitializeGlobalFramework() *OrganizationalIdentityFramework {
	frameworkOnce.Do(func() {
		log.Println("🌳 Initializing Global Deep Tree Echo Identity Framework...")
		
		GlobalIdentityFramework = NewOrganizationalIdentityFramework()
		
		ctx := context.Background()
		if err := GlobalIdentityFramework.Initialize(ctx); err != nil {
			log.Printf("Error initializing framework: %v", err)
		}
		
		log.Println("✨ Global Identity Framework initialized successfully")
	})
	
	return GlobalIdentityFramework
}

// GetGlobalFramework returns the global framework instance
func GetGlobalFramework() *OrganizationalIdentityFramework {
	if GlobalIdentityFramework == nil {
		return InitializeGlobalFramework()
	}
	return GlobalIdentityFramework
}

// IntegrateWithExistingIdentity integrates the framework with existing identity
func IntegrateWithExistingIdentity(existingIdentity *deeptreeecho.Identity) {
	framework := GetGlobalFramework()
	
	if framework.CoreIdentity != nil && existingIdentity != nil {
		// Sync states between identities
		syncIdentityStates(framework.CoreIdentity, existingIdentity)
		
		// Update framework with existing identity insights
		updateFrameworkFromIdentity(framework, existingIdentity)
		
		log.Println("🔗 Identity framework integrated with existing identity system")
	}
}

// syncIdentityStates synchronizes state between identities
func syncIdentityStates(frameworkIdentity, existingIdentity *deeptreeecho.Identity) {
	// Sync spatial context
	if existingIdentity.SpatialContext != nil {
		frameworkIdentity.SpatialContext = existingIdentity.SpatialContext
	}
	
	// Sync emotional state
	if existingIdentity.EmotionalState != nil {
		frameworkIdentity.EmotionalState = existingIdentity.EmotionalState
	}
	
	// Sync memory patterns
	if existingIdentity.Memory != nil {
		// Merge memory systems
		for nodeID, node := range existingIdentity.Memory.Nodes {
			frameworkIdentity.Memory.Nodes[nodeID] = node
		}
		
		for edgeID, edge := range existingIdentity.Memory.Edges {
			frameworkIdentity.Memory.Edges[edgeID] = edge
		}
	}
	
	// Sync patterns
	for patternName, pattern := range existingIdentity.Patterns {
		frameworkIdentity.Patterns[patternName] = pattern
	}
}

// updateFrameworkFromIdentity updates framework based on existing identity
func updateFrameworkFromIdentity(framework *OrganizationalIdentityFramework, identity *deeptreeecho.Identity) {
	// Update persona model based on identity characteristics
	if framework.PersonaModel != nil && identity.EmotionalState != nil {
		// Update emotional profile
		framework.PersonaModel.EmotionalProfile.PrimaryEmotions[identity.EmotionalState.Primary.Type] = 
			identity.EmotionalState.Primary.Strength
		framework.PersonaModel.EmotionalProfile.EmotionalIntensity = identity.EmotionalState.Intensity
	}
	
	// Update adaptation metrics based on identity performance
	if framework.AdaptationMetrics != nil {
		framework.AdaptationMetrics.FlexibilityScore = identity.Coherence
		framework.AdaptationMetrics.ConsistencyMaintenance = identity.Coherence
	}
	
	framework.LastUpdated = time.Now()
}

// ProcessThroughFramework processes input through the integrated framework
func ProcessThroughFramework(input string) (string, error) {
	framework := GetGlobalFramework()
	return framework.ProcessWithIdentity(input)
}

// GetFrameworkStatus returns the current framework status
func GetFrameworkStatus() map[string]interface{} {
	framework := GetGlobalFramework()
	return framework.GetFrameworkStatus()
}

// SaveFrameworkState saves the current framework state
func SaveFrameworkState() error {
	framework := GetGlobalFramework()
	return framework.SaveFramework()
}

// LoadFrameworkState loads framework state from disk
func LoadFrameworkState() error {
	framework := GetGlobalFramework()
	return framework.LoadFramework()
}
