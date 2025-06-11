package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// PromptAnalysis represents the analyzed characteristics of a prompt
type PromptAnalysis struct {
	Complexity  string  `json:"complexity"`   // "simple", "medium", "complex"
	Urgency     string  `json:"urgency"`      // "immediate", "normal", "quality"
	ContentType string  `json:"content_type"` // "code", "docs", "general"
	Confidence  float64 `json:"confidence"`   // 0.0 to 1.0
}

// AITarget represents the selected AI and reasoning
type AITarget struct {
	Type          string        `json:"type"`           // "local" or "cloud"
	Model         string        `json:"model"`          // specific model name
	Reason        string        `json:"reason"`         // why this was selected
	EstimatedCost float64       `json:"estimated_cost"` // in dollars
	EstimatedTime time.Duration `json:"estimated_time"` // expected response time
}

// SmartRouter handles prompt analysis and AI selection
type SmartRouter struct {
	speedKeywords      []string
	qualityKeywords    []string
	codeKeywords       []string
	docsKeywords       []string
	complexityKeywords []string
}

// NewSmartRouter creates a new router with predefined keyword patterns
func NewSmartRouter() *SmartRouter {
	return &SmartRouter{
		speedKeywords: []string{
			"quick", "fast", "brief", "summary", "urgent", "immediate",
			"rapid", "short", "simple", "asap",
		},
		qualityKeywords: []string{
			"comprehensive", "detailed", "professional", "thorough",
			"complete", "polished", "client", "presentation", "publish",
		},
		codeKeywords: []string{
			"code", "function", "bug", "refactor", "implement", "debug",
			"test", "class", "method", "algorithm", "fix", "error",
		},
		docsKeywords: []string{
			"document", "api", "guide", "readme", "manual", "specification",
			"tutorial", "explanation", "instructions", "documentation",
		},
		complexityKeywords: []string{
			"architecture", "system", "design", "analyze", "research",
			"strategy", "planning", "migration", "optimization", "complex",
		},
	}
}

// AnalyzePrompt examines a prompt and returns analysis results
func (r *SmartRouter) AnalyzePrompt(prompt string) PromptAnalysis {
	prompt = strings.ToLower(prompt)
	words := strings.Fields(prompt)

	analysis := PromptAnalysis{
		Complexity:  r.analyzeComplexity(prompt, words),
		Urgency:     r.analyzeUrgency(prompt, words),
		ContentType: r.analyzeContentType(prompt, words),
	}

	analysis.Confidence = r.calculateConfidence(analysis, prompt)
	return analysis
}

// analyzeComplexity determines if the prompt suggests a simple, medium, or complex task
func (r *SmartRouter) analyzeComplexity(prompt string, words []string) string {
	complexityScore := 0

	// Check for complexity keywords
	for _, word := range words {
		if r.containsWord(r.complexityKeywords, word) {
			complexityScore += 2
		}
	}

	// Length-based complexity
	if len(words) > 20 {
		complexityScore += 1
	}

	// Multiple concepts indicator
	if strings.Count(prompt, " and ") > 2 {
		complexityScore += 1
	}

	// Look for analysis/architecture terms
	if regexp.MustCompile(`(analyz|architect|design|system|complex)`).MatchString(prompt) {
		complexityScore += 2
	}

	switch {
	case complexityScore >= 4:
		return "complex"
	case complexityScore >= 2:
		return "medium"
	default:
		return "simple"
	}
}

// analyzeUrgency determines if the prompt suggests immediate, normal, or quality-focused needs
func (r *SmartRouter) analyzeUrgency(prompt string, words []string) string {
	speedScore := 0
	qualityScore := 0

	for _, word := range words {
		if r.containsWord(r.speedKeywords, word) {
			speedScore++
		}
		if r.containsWord(r.qualityKeywords, word) {
			qualityScore++
		}
	}

	switch {
	case speedScore > qualityScore && speedScore > 0:
		return "immediate"
	case qualityScore > speedScore && qualityScore > 0:
		return "quality"
	default:
		return "normal"
	}
}

// analyzeContentType determines the type of content being requested
func (r *SmartRouter) analyzeContentType(prompt string, words []string) string {
	codeScore := 0
	docsScore := 0

	for _, word := range words {
		if r.containsWord(r.codeKeywords, word) {
			codeScore++
		}
		if r.containsWord(r.docsKeywords, word) {
			docsScore++
		}
	}

	switch {
	case codeScore > docsScore && codeScore > 0:
		return "code"
	case docsScore > codeScore && docsScore > 0:
		return "docs"
	default:
		return "general"
	}
}

// calculateConfidence estimates how confident we are in the analysis
func (r *SmartRouter) calculateConfidence(analysis PromptAnalysis, prompt string) float64 {
	confidence := 0.5 // base confidence

	// Boost confidence if we found clear indicators
	if analysis.ContentType != "general" {
		confidence += 0.2
	}
	if analysis.Urgency != "normal" {
		confidence += 0.2
	}
	if len(prompt) > 10 { // longer prompts give more signal
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

// SelectAI chooses the best AI target based on analysis
func (r *SmartRouter) SelectAI(analysis PromptAnalysis) AITarget {
	// Speed priority routing
	if analysis.Urgency == "immediate" {
		return AITarget{
			Type:          "local",
			Model:         "orca-mini:3b",
			Reason:        "Speed priority for immediate tasks",
			EstimatedCost: 0.01,
			EstimatedTime: 3 * time.Second,
		}
	}

	// Complex tasks might need cloud AI
	if analysis.Complexity == "complex" && analysis.Urgency == "quality" {
		return AITarget{
			Type:          "cloud",
			Model:         "claude-3-sonnet",
			Reason:        "Complex task requiring maximum capability",
			EstimatedCost: 0.50,
			EstimatedTime: 8 * time.Second,
		}
	}

	// Code-specific routing
	if analysis.ContentType == "code" {
		return AITarget{
			Type:          "local",
			Model:         "codellama:7b",
			Reason:        "Code specialist for technical accuracy",
			EstimatedCost: 0.02,
			EstimatedTime: 15 * time.Second,
		}
	}

	// Documentation with quality focus
	if analysis.ContentType == "docs" && analysis.Urgency == "quality" {
		return AITarget{
			Type:          "local",
			Model:         "llama2:13b",
			Reason:        "High quality for professional documentation",
			EstimatedCost: 0.03,
			EstimatedTime: 25 * time.Second,
		}
	}

	// Default balanced choice
	return AITarget{
		Type:          "local",
		Model:         "mistral:7b",
		Reason:        "Balanced performance for general tasks",
		EstimatedCost: 0.02,
		EstimatedTime: 12 * time.Second,
	}
}

// containsWord checks if a word exists in a slice of words
func (r *SmartRouter) containsWord(slice []string, word string) bool {
	for _, s := range slice {
		if s == word {
			return true
		}
	}
	return false
}

// ExecuteRouting demonstrates the complete routing process
func (r *SmartRouter) ExecuteRouting(prompt string, explain bool) AITarget {
	analysis := r.AnalyzePrompt(prompt)
	target := r.SelectAI(analysis)

	if explain {
		fmt.Println("🧠 Prompt Analysis:")
		fmt.Printf("   Complexity: %s\n", analysis.Complexity)
		fmt.Printf("   Urgency: %s\n", analysis.Urgency)
		fmt.Printf("   Content Type: %s\n", analysis.ContentType)
		fmt.Printf("   Confidence: %.1f%%\n", analysis.Confidence*100)

		fmt.Println("\n🎯 AI Selection:")
		fmt.Printf("   Target: %s (%s)\n", target.Type, target.Model)
		fmt.Printf("   Reason: %s\n", target.Reason)
		fmt.Printf("   Estimated Cost: $%.3f\n", target.EstimatedCost)
		fmt.Printf("   Estimated Time: %v\n", target.EstimatedTime)
		fmt.Println()
	}

	return target
}

// main function for testing the routing system
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run uroboro_smart_routing_starter.go \"your prompt here\" [--explain]")
		fmt.Println("\nExample prompts to test:")
		fmt.Println("  \"quick bug fix summary\"")
		fmt.Println("  \"comprehensive API documentation for user service\"")
		fmt.Println("  \"complex microservice architecture analysis\"")
		fmt.Println("  \"refactor authentication function\"")
		os.Exit(1)
	}

	prompt := os.Args[1]
	explain := len(os.Args) > 2 && os.Args[2] == "--explain"

	router := NewSmartRouter()
	target := router.ExecuteRouting(prompt, explain)

	if !explain {
		fmt.Printf("Selected: %s (%s) - %s\n", target.Type, target.Model, target.Reason)
	}

	// In real implementation, this would integrate with uroboro's AI execution
	fmt.Printf("→ Would execute with %s: %s\n", target.Model, prompt)
}
