# Uroboro Smart Routing Integration Guide 🔧⚡

**Step-by-step guide to add intelligent AI routing to uroboro**

This guide walks through integrating the smart routing prototype into your actual uroboro codebase for immediate dogfooding.

---

## 🎯 Integration Overview

### **What We're Adding**
- Intelligent prompt analysis that detects complexity, urgency, and content type
- Automatic AI selection (local fast, local quality, cloud fallback)
- Enhanced `capture` command with `--prompt` flag for smart routing
- Cost tracking and optimization based on actual usage

### **Implementation Strategy**
1. **Phase 1**: Add routing package with core logic
2. **Phase 2**: Enhance capture command with smart routing
3. **Phase 3**: Add configuration and cost tracking
4. **Phase 4**: Test and iterate based on real usage

---

## 🏗 Step 1: Add Routing Package

### **Create Directory Structure**
```bash
cd labs/projects/uroboro
mkdir -p internal/routing
mkdir -p internal/ai
```

### **File: `internal/routing/analyzer.go`**
```go
package routing

import (
	"regexp"
	"strings"
)

type PromptAnalysis struct {
	Complexity  string  `json:"complexity"`   // "simple", "medium", "complex"
	Urgency     string  `json:"urgency"`      // "immediate", "normal", "quality"
	ContentType string  `json:"content_type"` // "code", "docs", "general"
	Confidence  float64 `json:"confidence"`   // 0.0 to 1.0
}

type Analyzer struct {
	speedKeywords      []string
	qualityKeywords    []string
	codeKeywords       []string
	docsKeywords       []string
	complexityKeywords []string
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
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

func (a *Analyzer) Analyze(prompt string) PromptAnalysis {
	prompt = strings.ToLower(prompt)
	words := strings.Fields(prompt)

	analysis := PromptAnalysis{
		Complexity:  a.analyzeComplexity(prompt, words),
		Urgency:     a.analyzeUrgency(prompt, words),
		ContentType: a.analyzeContentType(prompt, words),
	}

	analysis.Confidence = a.calculateConfidence(analysis, prompt)
	return analysis
}

func (a *Analyzer) analyzeComplexity(prompt string, words []string) string {
	complexityScore := 0

	for _, word := range words {
		if a.containsWord(a.complexityKeywords, word) {
			complexityScore += 2
		}
	}

	if len(words) > 20 {
		complexityScore += 1
	}

	if strings.Count(prompt, " and ") > 2 {
		complexityScore += 1
	}

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

func (a *Analyzer) analyzeUrgency(prompt string, words []string) string {
	speedScore := 0
	qualityScore := 0

	for _, word := range words {
		if a.containsWord(a.speedKeywords, word) {
			speedScore++
		}
		if a.containsWord(a.qualityKeywords, word) {
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

func (a *Analyzer) analyzeContentType(prompt string, words []string) string {
	codeScore := 0
	docsScore := 0

	for _, word := range words {
		if a.containsWord(a.codeKeywords, word) {
			codeScore++
		}
		if a.containsWord(a.docsKeywords, word) {
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

func (a *Analyzer) calculateConfidence(analysis PromptAnalysis, prompt string) float64 {
	confidence := 0.5

	if analysis.ContentType != "general" {
		confidence += 0.2
	}
	if analysis.Urgency != "normal" {
		confidence += 0.2
	}
	if len(prompt) > 10 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

func (a *Analyzer) containsWord(slice []string, word string) bool {
	for _, s := range slice {
		if s == word {
			return true
		}
	}
	return false
}
```

### **File: `internal/routing/selector.go`**
```go
package routing

import (
	"time"
)

type AITarget struct {
	Type          string        `json:"type"`           // "local" or "cloud"
	Model         string        `json:"model"`          // specific model name
	Provider      string        `json:"provider"`       // cloud provider if applicable
	Reason        string        `json:"reason"`         // why this was selected
	EstimatedCost float64       `json:"estimated_cost"` // in dollars
	EstimatedTime time.Duration `json:"estimated_time"` // expected response time
}

type Selector struct {
	// Configuration for available models
	speedModel   string
	balanceModel string
	qualityModel string
	codeModel    string
	cloudModel   string
}

func NewSelector() *Selector {
	return &Selector{
		speedModel:   "orca-mini:3b",
		balanceModel: "mistral:7b",
		qualityModel: "llama2:13b",
		codeModel:    "codellama:7b",
		cloudModel:   "claude", // Will need env config
	}
}

func (s *Selector) SelectAI(analysis PromptAnalysis) AITarget {
	// Speed priority routing
	if analysis.Urgency == "immediate" {
		return AITarget{
			Type:          "local",
			Model:         s.speedModel,
			Reason:        "Speed priority for immediate tasks",
			EstimatedCost: 0.01,
			EstimatedTime: 3 * time.Second,
		}
	}

	// Complex + quality = cloud AI
	if analysis.Complexity == "complex" && analysis.Urgency == "quality" {
		return AITarget{
			Type:          "cloud",
			Provider:      s.cloudModel,
			Reason:        "Complex task requiring maximum capability",
			EstimatedCost: 0.50,
			EstimatedTime: 8 * time.Second,
		}
	}

	// Code-specific routing
	if analysis.ContentType == "code" {
		return AITarget{
			Type:          "local",
			Model:         s.codeModel,
			Reason:        "Code specialist for technical accuracy",
			EstimatedCost: 0.02,
			EstimatedTime: 15 * time.Second,
		}
	}

	// Quality documentation
	if analysis.ContentType == "docs" && analysis.Urgency == "quality" {
		return AITarget{
			Type:          "local",
			Model:         s.qualityModel,
			Reason:        "High quality for professional documentation",
			EstimatedCost: 0.03,
			EstimatedTime: 25 * time.Second,
		}
	}

	// Default balanced choice
	return AITarget{
		Type:          "local",
		Model:         s.balanceModel,
		Reason:        "Balanced performance for general tasks",
		EstimatedCost: 0.02,
		EstimatedTime: 12 * time.Second,
	}
}
```

---

## 🔧 Step 2: Enhance Capture Command

### **Modify: `cmd/capture.go`**

Add smart routing functionality to the existing capture command:

```go
// Add these imports
import (
	"github.com/qry91/uroboro/internal/routing"
	// ... existing imports
)

// Add new flags to capture command initialization
func init() {
	captureCmd.Flags().StringP("prompt", "p", "", "Use smart AI routing for prompt")
	captureCmd.Flags().Bool("explain", false, "Explain routing decision")
	captureCmd.Flags().String("force-model", "", "Override model selection")
	captureCmd.Flags().Bool("force-local", false, "Force local AI only")
	captureCmd.Flags().Bool("dry-run", false, "Show routing without executing")
	
	// ... existing flags
}

// Add smart routing function
func runSmartCapture(cmd *cobra.Command, args []string) error {
	prompt, _ := cmd.Flags().GetString("prompt")
	explain, _ := cmd.Flags().GetBool("explain")
	forceModel, _ := cmd.Flags().GetString("force-model")
	forceLocal, _ := cmd.Flags().GetBool("force-local")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if prompt == "" {
		// Fall back to existing capture behavior
		return runExistingCapture(cmd, args)
	}

	// Initialize smart routing
	analyzer := routing.NewAnalyzer()
	selector := routing.NewSelector()

	// Analyze prompt
	analysis := analyzer.Analyze(prompt)
	
	// Select AI target
	var target routing.AITarget
	switch {
	case forceModel != "":
		target = routing.AITarget{
			Type:   "local",
			Model:  forceModel,
			Reason: "Manual override",
		}
	case forceLocal:
		// Force local selection logic
		target = selector.SelectBestLocal(analysis)
	default:
		target = selector.SelectAI(analysis)
	}

	// Show explanation if requested
	if explain || dryRun {
		showRoutingExplanation(analysis, target)
		if dryRun {
			return nil
		}
		
		fmt.Print("Proceed? [Y/n]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "n" {
			return nil
		}
	}

	// Execute the AI request
	return executeSmartCapture(target, prompt)
}

func showRoutingExplanation(analysis routing.PromptAnalysis, target routing.AITarget) {
	fmt.Println("🧠 Prompt Analysis:")
	fmt.Printf("   Complexity: %s\n", analysis.Complexity)
	fmt.Printf("   Urgency: %s\n", analysis.Urgency)
	fmt.Printf("   Content Type: %s\n", analysis.ContentType)
	fmt.Printf("   Confidence: %.1f%%\n", analysis.Confidence*100)

	fmt.Println("\n🎯 AI Selection:")
	fmt.Printf("   Target: %s", target.Type)
	if target.Model != "" {
		fmt.Printf(" (%s)", target.Model)
	}
	if target.Provider != "" {
		fmt.Printf(" (%s)", target.Provider)
	}
	fmt.Printf("\n   Reason: %s\n", target.Reason)
	fmt.Printf("   Estimated Cost: $%.3f\n", target.EstimatedCost)
	fmt.Printf("   Estimated Time: %v\n", target.EstimatedTime)
	fmt.Println()
}

func executeSmartCapture(target routing.AITarget, prompt string) error {
	switch target.Type {
	case "local":
		return executeLocalAI(target.Model, prompt)
	case "cloud":
		return executeCloudAI(target.Provider, prompt)
	default:
		return fmt.Errorf("unknown AI target type: %s", target.Type)
	}
}

func executeLocalAI(model string, prompt string) error {
	// Use existing uroboro Ollama integration
	// This should integrate with your existing capture logic
	// but override the model selection
	
	// Example integration:
	response, err := generateWithOllama(model, prompt)
	if err != nil {
		return fmt.Errorf("local AI failed: %w", err)
	}
	
	// Save using existing uroboro capture logic
	return saveCaptureWithModel(response, model, "smart-routed")
}

func executeCloudAI(provider string, prompt string) error {
	// Implement cloud AI integration
	// This is where you'd add Claude/OpenAI API calls
	switch provider {
	case "claude":
		return executeClaudeAI(prompt)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}
```

### **Modify the main capture command to detect smart routing**
```go
// In the main capture command function, add routing detection
func runCapture(cmd *cobra.Command, args []string) error {
	prompt, _ := cmd.Flags().GetString("prompt")
	
	// If prompt flag is used, switch to smart routing
	if prompt != "" {
		return runSmartCapture(cmd, args)
	}
	
	// Otherwise, use existing capture logic
	return runExistingCapture(cmd, args)
}
```

---

## ⚙️ Step 3: Configuration Setup

### **Create: `internal/config/routing.go`**
```go
package config

import (
	"os"
	"strconv"
)

type RoutingConfig struct {
	// Model preferences
	SpeedModel   string
	BalanceModel string
	QualityModel string
	CodeModel    string
	
	// Cloud AI config
	CloudProvider string
	CloudAPIKey   string
	
	// Budget settings
	MonthlyBudget    float64
	LocalCostPerTask float64
	CloudCostPerTask float64
	
	// Thresholds
	ForceLocalRatio float64 // Force local if budget usage > this
}

func LoadRoutingConfig() RoutingConfig {
	config := RoutingConfig{
		// Defaults
		SpeedModel:       getEnvOrDefault("UROBORO_SPEED_MODEL", "orca-mini:3b"),
		BalanceModel:     getEnvOrDefault("UROBORO_BALANCE_MODEL", "mistral:7b"),
		QualityModel:     getEnvOrDefault("UROBORO_QUALITY_MODEL", "llama2:13b"),
		CodeModel:        getEnvOrDefault("UROBORO_CODE_MODEL", "codellama:7b"),
		CloudProvider:    getEnvOrDefault("UROBORO_CLOUD_PROVIDER", "claude"),
		CloudAPIKey:      os.Getenv("ANTHROPIC_API_KEY"),
		MonthlyBudget:    getEnvFloat("UROBORO_MONTHLY_BUDGET", 25.0),
		LocalCostPerTask: getEnvFloat("UROBORO_LOCAL_COST", 0.01),
		CloudCostPerTask: getEnvFloat("UROBORO_CLOUD_COST", 0.50),
		ForceLocalRatio:  getEnvFloat("UROBORO_FORCE_LOCAL_RATIO", 0.8),
	}
	
	return config
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
```

### **Create: `~/.uroboro/smart-routing.env`**
```bash
# Uroboro Smart Routing Configuration

# Model Selection
UROBORO_SPEED_MODEL=orca-mini:3b
UROBORO_BALANCE_MODEL=mistral:7b
UROBORO_QUALITY_MODEL=llama2:13b
UROBORO_CODE_MODEL=codellama:7b

# Cloud AI (optional)
UROBORO_CLOUD_PROVIDER=claude
# ANTHROPIC_API_KEY=your_key_here

# Budget Management
UROBORO_MONTHLY_BUDGET=25.00
UROBORO_LOCAL_COST=0.01
UROBORO_CLOUD_COST=0.50
UROBORO_FORCE_LOCAL_RATIO=0.8

# Performance Tuning
UROBORO_MAX_RESPONSE_TIME=30s
UROBORO_MIN_QUALITY_THRESHOLD=3.5
```

---

## 🧪 Step 4: Testing & Validation

### **Create: `test/routing_test.go`**
```go
package test

import (
	"testing"
	"github.com/qry91/uroboro/internal/routing"
)

func TestPromptAnalysis(t *testing.T) {
	analyzer := routing.NewAnalyzer()
	
	tests := []struct {
		prompt      string
		expectedComplexity string
		expectedUrgency    string
		expectedContentType string
	}{
		{
			prompt: "quick bug fix summary",
			expectedComplexity: "simple",
			expectedUrgency: "immediate",
			expectedContentType: "code",
		},
		{
			prompt: "comprehensive API documentation",
			expectedComplexity: "simple",
			expectedUrgency: "quality",
			expectedContentType: "docs",
		},
		{
			prompt: "complex system architecture analysis",
			expectedComplexity: "complex",
			expectedUrgency: "normal",
			expectedContentType: "general",
		},
	}
	
	for _, tt := range tests {
		result := analyzer.Analyze(tt.prompt)
		
		if result.Complexity != tt.expectedComplexity {
			t.Errorf("Complexity for '%s': expected %s, got %s", 
				tt.prompt, tt.expectedComplexity, result.Complexity)
		}
		
		if result.Urgency != tt.expectedUrgency {
			t.Errorf("Urgency for '%s': expected %s, got %s", 
				tt.prompt, tt.expectedUrgency, result.Urgency)
		}
		
		if result.ContentType != tt.expectedContentType {
			t.Errorf("ContentType for '%s': expected %s, got %s", 
				tt.prompt, tt.expectedContentType, result.ContentType)
		}
	}
}

func TestModelSelection(t *testing.T) {
	selector := routing.NewSelector()
	
	// Test speed priority
	speedAnalysis := routing.PromptAnalysis{
		Complexity: "simple",
		Urgency: "immediate",
		ContentType: "code",
	}
	
	target := selector.SelectAI(speedAnalysis)
	if target.Model != "orca-mini:3b" {
		t.Errorf("Expected orca-mini:3b for speed priority, got %s", target.Model)
	}
	
	// Test quality priority
	qualityAnalysis := routing.PromptAnalysis{
		Complexity: "medium",
		Urgency: "quality",
		ContentType: "docs",
	}
	
	target = selector.SelectAI(qualityAnalysis)
	if target.Model != "llama2:13b" {
		t.Errorf("Expected llama2:13b for quality docs, got %s", target.Model)
	}
}
```

### **Manual Testing Commands**
```bash
# Test the enhanced capture command
./uroboro capture -p "quick bug fix summary" --explain
./uroboro capture -p "comprehensive API documentation" --explain
./uroboro capture -p "complex architecture analysis" --explain

# Test overrides
./uroboro capture -p "any prompt" --force-model mistral:7b
./uroboro capture -p "complex task" --force-local

# Test dry run
./uroboro capture -p "test prompt" --dry-run
```

---

## 🚀 Step 5: Implementation Checklist

### **Phase 1: Core Implementation**
- [ ] Create `internal/routing/` package with analyzer and selector
- [ ] Add routing logic with keyword-based analysis  
- [ ] Test routing decisions with various prompts
- [ ] Ensure basic model selection works correctly

### **Phase 2: CLI Integration**
- [ ] Add `--prompt` flag to capture command
- [ ] Implement smart routing in capture workflow
- [ ] Add explanation and override flags
- [ ] Test CLI integration with existing functionality

### **Phase 3: Configuration**
- [ ] Add configuration loading for model preferences
- [ ] Create environment variable setup
- [ ] Add budget tracking capabilities
- [ ] Test configuration override functionality

### **Phase 4: Advanced Features**
- [ ] Add cloud AI integration (if needed)
- [ ] Implement cost tracking and reporting
- [ ] Add learning from user feedback
- [ ] Create usage analytics and optimization

---

## 🎯 Quick Start for Dogfooding

### **Minimal Implementation (30 minutes)**
1. Copy the routing package files into uroboro
2. Add basic `--prompt` flag to capture command
3. Test with your actual daily prompts
4. Iterate based on routing accuracy

### **Daily Testing Prompts**
```bash
# Your typical workflow prompts
./uroboro capture -p "Fixed authentication timeout issue in user service"
./uroboro capture -p "Document the new rate limiting middleware implementation"  
./uroboro capture -p "Analyze performance implications of current database queries"
./uroboro capture -p "Quick summary of today's development progress"
```

### **Tuning the Routing**
- Adjust keyword lists based on your terminology
- Modify complexity thresholds based on your task types
- Update cost estimates based on actual model performance
- Add custom routing rules for your specific workflows

---

## 💡 Expected Benefits

### **Immediate (Week 1)**
- No more manual model selection decisions
- Automatic cost optimization based on task complexity
- Consistent routing for similar prompt patterns
- Reduced cognitive overhead in daily workflows

### **Short-term (Month 1)**
- Learned preferences improve routing accuracy
- Measurable cost savings from optimized model usage
- Faster development velocity from seamless AI integration
- Data-driven insights into AI usage patterns

### **Long-term (Month 3+)**
- Fully optimized AI workflow with minimal manual intervention
- Substantial cost savings compared to manual cloud AI usage
- Foundation for advanced features like batch processing and learning
- Community-shareable optimization strategies

---

**Ready to implement? Start with Phase 1 and test the basic routing logic with your actual prompts. The routing will get smarter as you use it and tune it to your workflow.**