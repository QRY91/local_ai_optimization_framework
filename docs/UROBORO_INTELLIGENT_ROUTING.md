# Uroboro Intelligent AI Routing Implementation 🧠⚡

**Smart AI orchestration for seamless cost optimization**

Turn uroboro into an intelligent routing layer that automatically selects the best AI (local fast, local quality, or cloud) based on prompt analysis, without manual model selection.

---

## 🎯 Core Concept

### **Current State**
```bash
# Manual model selection
export UROBORO_MODEL="orca-mini:3b"
./uroboro capture "quick bug fix"

export UROBORO_MODEL="llama2:13b"  
./uroboro capture "detailed API docs"
```

### **Target State**
```bash
# Intelligent auto-routing
./uroboro capture --prompt "quick bug fix"
# → Auto-routes to orca-mini:3b (speed detected)

./uroboro capture --prompt "detailed API docs"  
# → Auto-routes to llama2:13b (quality + length detected)

./uroboro capture --prompt "complex architecture analysis"
# → Auto-routes to cloud AI (complexity threshold)
```

---

## 🏗 Implementation Architecture

### **Phase 1: Prompt Analysis Engine**

#### **File: `internal/routing/analyzer.go`**
```go
package routing

import (
    "regexp"
    "strings"
    "time"
)

type PromptAnalysis struct {
    Complexity   ComplexityLevel `json:"complexity"`
    Urgency      UrgencyLevel    `json:"urgency"`
    ContentType  ContentType     `json:"content_type"`
    Length       LengthCategory  `json:"length"`
    QualityNeeds QualityLevel    `json:"quality_needs"`
    Confidence   float64         `json:"confidence"`
}

type ComplexityLevel string
const (
    Simple  ComplexityLevel = "simple"
    Medium  ComplexityLevel = "medium"
    Complex ComplexityLevel = "complex"
)

type UrgencyLevel string
const (
    Immediate      UrgencyLevel = "immediate"
    Normal         UrgencyLevel = "normal"
    QualityFocused UrgencyLevel = "quality-focused"
)

type ContentType string
const (
    Code    ContentType = "code"
    Docs    ContentType = "docs"
    Social  ContentType = "social"
    Analysis ContentType = "analysis"
    General ContentType = "general"
)

type PromptAnalyzer struct {
    // Keyword patterns for classification
    speedKeywords     []string
    qualityKeywords   []string
    codeKeywords      []string
    docsKeywords      []string
    complexityKeywords []string
}

func NewPromptAnalyzer() *PromptAnalyzer {
    return &PromptAnalyzer{
        speedKeywords: []string{
            "quick", "fast", "brief", "summary", "urgent", 
            "immediate", "rapid", "short", "simple",
        },
        qualityKeywords: []string{
            "comprehensive", "detailed", "professional", "thorough",
            "complete", "in-depth", "polished", "client", "presentation",
        },
        codeKeywords: []string{
            "code", "function", "bug", "refactor", "implement",
            "debug", "test", "class", "method", "algorithm",
        },
        docsKeywords: []string{
            "document", "API", "guide", "README", "manual",
            "specification", "tutorial", "explanation", "instructions",
        },
        complexityKeywords: []string{
            "architecture", "system", "design", "analyze", "research",
            "strategy", "planning", "migration", "optimization",
        },
    }
}

func (a *PromptAnalyzer) Analyze(prompt string) PromptAnalysis {
    prompt = strings.ToLower(prompt)
    words := strings.Fields(prompt)
    
    analysis := PromptAnalysis{
        Complexity:   a.analyzeComplexity(prompt, words),
        Urgency:      a.analyzeUrgency(prompt, words),
        ContentType:  a.analyzeContentType(prompt, words),
        Length:       a.estimateOutputLength(prompt, words),
        QualityNeeds: a.analyzeQualityNeeds(prompt, words),
    }
    
    analysis.Confidence = a.calculateConfidence(analysis, prompt)
    return analysis
}

func (a *PromptAnalyzer) analyzeComplexity(prompt string, words []string) ComplexityLevel {
    complexityScore := 0
    
    // Check for complexity indicators
    for _, word := range words {
        if contains(a.complexityKeywords, word) {
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
    
    switch {
    case complexityScore >= 4:
        return Complex
    case complexityScore >= 2:
        return Medium
    default:
        return Simple
    }
}

func (a *PromptAnalyzer) analyzeUrgency(prompt string, words []string) UrgencyLevel {
    speedScore := 0
    qualityScore := 0
    
    for _, word := range words {
        if contains(a.speedKeywords, word) {
            speedScore++
        }
        if contains(a.qualityKeywords, word) {
            qualityScore++
        }
    }
    
    switch {
    case speedScore > qualityScore && speedScore > 0:
        return Immediate
    case qualityScore > speedScore && qualityScore > 0:
        return QualityFocused
    default:
        return Normal
    }
}

func (a *PromptAnalyzer) analyzeContentType(prompt string, words []string) ContentType {
    codeScore := 0
    docsScore := 0
    
    for _, word := range words {
        if contains(a.codeKeywords, word) {
            codeScore++
        }
        if contains(a.docsKeywords, word) {
            docsScore++
        }
    }
    
    switch {
    case codeScore > docsScore && codeScore > 0:
        return Code
    case docsScore > codeScore && docsScore > 0:
        return Docs
    default:
        return General
    }
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

### **Phase 2: AI Target Selection**

#### **File: `internal/routing/selector.go`**
```go
package routing

import (
    "fmt"
    "time"
)

type AITarget struct {
    Type        TargetType `json:"type"`
    Model       string     `json:"model,omitempty"`
    Provider    string     `json:"provider,omitempty"`
    Reason      string     `json:"reason"`
    EstimatedCost float64  `json:"estimated_cost"`
    EstimatedTime time.Duration `json:"estimated_time"`
}

type TargetType string
const (
    Local TargetType = "local"
    Cloud TargetType = "cloud"
)

type ModelSelector struct {
    // Available local models and their characteristics
    localModels map[string]ModelCharacteristics
    // Cloud AI configuration
    cloudConfig CloudConfig
    // Budget constraints
    budget BudgetManager
}

type ModelCharacteristics struct {
    Speed        int     `json:"speed"`         // 1-5 scale
    Quality      int     `json:"quality"`       // 1-5 scale
    MemoryMB     int     `json:"memory_mb"`
    SpecialtyFit float64 `json:"specialty_fit"` // How well suited for content type
}

type BudgetManager struct {
    MonthlyBudget    float64 `json:"monthly_budget"`
    CurrentSpend     float64 `json:"current_spend"`
    LocalCostPerTask float64 `json:"local_cost_per_task"`
    CloudCostPerTask float64 `json:"cloud_cost_per_task"`
}

func NewModelSelector() *ModelSelector {
    return &ModelSelector{
        localModels: map[string]ModelCharacteristics{
            "orca-mini:3b": {
                Speed:        5,
                Quality:      3,
                MemoryMB:     2048,
                SpecialtyFit: 0.7,
            },
            "mistral:7b": {
                Speed:        3,
                Quality:      4,
                MemoryMB:     4096,
                SpecialtyFit: 0.8,
            },
            "codellama:7b": {
                Speed:        3,
                Quality:      4,
                MemoryMB:     4096,
                SpecialtyFit: 0.9, // High for code content
            },
            "llama2:13b": {
                Speed:        2,
                Quality:      5,
                MemoryMB:     8192,
                SpecialtyFit: 0.8,
            },
        },
        budget: BudgetManager{
            LocalCostPerTask: 0.01,
            CloudCostPerTask: 0.50,
        },
    }
}

func (s *ModelSelector) SelectAI(analysis PromptAnalysis) AITarget {
    // Budget check first
    if s.shouldForceLocal() {
        return s.selectBestLocal(analysis)
    }
    
    // Route based on analysis
    switch {
    case analysis.Urgency == Immediate:
        return s.selectFastestLocal(analysis)
        
    case analysis.Complexity == Complex && analysis.QualityNeeds == High:
        return s.selectCloudAI(analysis)
        
    case analysis.ContentType == Code:
        return s.selectCodeSpecialist(analysis)
        
    case analysis.QualityNeeds == High:
        return s.selectHighQualityLocal(analysis)
        
    default:
        return s.selectBalancedLocal(analysis)
    }
}

func (s *ModelSelector) selectFastestLocal(analysis PromptAnalysis) AITarget {
    return AITarget{
        Type:          Local,
        Model:         "orca-mini:3b",
        Reason:        "Speed priority for immediate tasks",
        EstimatedCost: s.budget.LocalCostPerTask,
        EstimatedTime: 3 * time.Second,
    }
}

func (s *ModelSelector) selectCodeSpecialist(analysis PromptAnalysis) AITarget {
    if analysis.Complexity == Complex {
        return AITarget{
            Type:          Local,
            Model:         "codellama:7b",
            Reason:        "Code specialist for technical accuracy",
            EstimatedCost: s.budget.LocalCostPerTask,
            EstimatedTime: 15 * time.Second,
        }
    }
    
    return AITarget{
        Type:          Local,
        Model:         "mistral:7b",
        Reason:        "Balanced model for general code tasks",
        EstimatedCost: s.budget.LocalCostPerTask,
        EstimatedTime: 10 * time.Second,
    }
}

func (s *ModelSelector) selectHighQualityLocal(analysis PromptAnalysis) AITarget {
    return AITarget{
        Type:          Local,
        Model:         "llama2:13b",
        Reason:        "High quality for professional output",
        EstimatedCost: s.budget.LocalCostPerTask,
        EstimatedTime: 25 * time.Second,
    }
}

func (s *ModelSelector) selectCloudAI(analysis PromptAnalysis) AITarget {
    return AITarget{
        Type:          Cloud,
        Provider:      "claude", // or configured cloud provider
        Reason:        "Complex task requiring maximum capability",
        EstimatedCost: s.budget.CloudCostPerTask,
        EstimatedTime: 8 * time.Second,
    }
}

func (s *ModelSelector) shouldForceLocal() bool {
    budgetUsed := s.budget.CurrentSpend / s.budget.MonthlyBudget
    return budgetUsed > 0.8 // Force local if >80% budget used
}
```

### **Phase 3: Enhanced CLI Integration**

#### **File: `cmd/capture_smart.go`**
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/qry91/uroboro/internal/routing"
    "github.com/qry91/uroboro/internal/ai"
)

func newSmartCaptureCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "capture",
        Short: "Intelligent AI-routed capture",
        Long: `Capture content using intelligent AI routing.
        
Analyzes your prompt and automatically selects the best AI:
- Fast local models for quick tasks
- Quality local models for important content  
- Cloud AI fallback for complex analysis

Examples:
  uroboro capture -p "quick bug fix summary"
  uroboro capture -p "comprehensive API documentation"
  uroboro capture -p "complex architecture analysis" --explain`,
        RunE: runSmartCapture,
    }
    
    cmd.Flags().StringP("prompt", "p", "", "Prompt for AI processing")
    cmd.Flags().Bool("explain", false, "Explain routing decision before executing")
    cmd.Flags().String("force-model", "", "Override automatic model selection")
    cmd.Flags().Bool("force-local", false, "Force local AI usage only")
    cmd.Flags().Bool("force-cloud", false, "Force cloud AI usage")
    cmd.Flags().Bool("dry-run", false, "Show routing decision without executing")
    
    return cmd
}

func runSmartCapture(cmd *cobra.Command, args []string) error {
    prompt, _ := cmd.Flags().GetString("prompt")
    explain, _ := cmd.Flags().GetBool("explain")
    forceModel, _ := cmd.Flags().GetString("force-model")
    forceLocal, _ := cmd.Flags().GetBool("force-local")
    forceCloud, _ := cmd.Flags().GetBool("force-cloud")
    dryRun, _ := cmd.Flags().GetBool("dry-run")
    
    if prompt == "" {
        return fmt.Errorf("prompt is required")
    }
    
    // Initialize routing components
    analyzer := routing.NewPromptAnalyzer()
    selector := routing.NewModelSelector()
    
    // Analyze the prompt
    analysis := analyzer.Analyze(prompt)
    
    // Select AI target
    var target routing.AITarget
    switch {
    case forceModel != "":
        target = routing.AITarget{
            Type:   routing.Local,
            Model:  forceModel,
            Reason: "Manual override",
        }
    case forceLocal:
        target = selector.SelectBestLocal(analysis)
    case forceCloud:
        target = selector.SelectCloudAI(analysis)
    default:
        target = selector.SelectAI(analysis)
    }
    
    // Show explanation if requested
    if explain || dryRun {
        showRoutingExplanation(analysis, target)
        if dryRun {
            return nil
        }
        
        if !promptYesNo("Proceed with this routing?") {
            return nil
        }
    }
    
    // Execute the AI request
    return executeAIRequest(target, prompt)
}

func showRoutingExplanation(analysis routing.PromptAnalysis, target routing.AITarget) {
    fmt.Println("🧠 Prompt Analysis:")
    fmt.Printf("   Complexity: %s\n", analysis.Complexity)
    fmt.Printf("   Urgency: %s\n", analysis.Urgency)
    fmt.Printf("   Content Type: %s\n", analysis.ContentType)
    fmt.Printf("   Quality Needs: %s\n", analysis.QualityNeeds)
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
}

func executeAIRequest(target routing.AITarget, prompt string) error {
    switch target.Type {
    case routing.Local:
        return executeLocalAI(target.Model, prompt)
    case routing.Cloud:
        return executeCloudAI(target.Provider, prompt)
    default:
        return fmt.Errorf("unknown AI target type: %s", target.Type)
    }
}

func executeLocalAI(model string, prompt string) error {
    // Integrate with existing uroboro local AI execution
    client := ai.NewOllamaClient()
    response, err := client.Generate(model, prompt)
    if err != nil {
        return fmt.Errorf("local AI generation failed: %w", err)
    }
    
    // Save to uroboro database
    return saveCapture(response, model, "local")
}

func executeCloudAI(provider string, prompt string) error {
    // Integrate with cloud AI providers
    switch provider {
    case "claude":
        return executeClaudeAI(prompt)
    case "openai":
        return executeOpenAI(prompt)
    default:
        return fmt.Errorf("unsupported cloud provider: %s", provider)
    }
}
```

---

## 🚀 Implementation Plan

### **Phase 1: Core Routing (Week 1)**
1. **Day 1-2**: Implement `PromptAnalyzer` with basic keyword detection
2. **Day 3-4**: Build `ModelSelector` with simple routing rules
3. **Day 5-7**: Create enhanced capture command with routing

### **Phase 2: Integration (Week 2)**
1. **Day 1-3**: Wire up with existing uroboro database and AI clients
2. **Day 4-5**: Add cloud AI fallback integration
3. **Day 6-7**: Testing and debugging

### **Phase 3: Optimization (Week 3)**
1. **Day 1-3**: Add learning from user feedback
2. **Day 4-5**: Budget management and cost tracking
3. **Day 6-7**: Performance tuning and edge case handling

---

## 🧪 Testing Strategy

### **Unit Tests**
```go
func TestPromptAnalyzer(t *testing.T) {
    analyzer := routing.NewPromptAnalyzer()
    
    tests := []struct {
        prompt   string
        expected routing.PromptAnalysis
    }{
        {
            prompt: "quick bug fix summary",
            expected: routing.PromptAnalysis{
                Complexity: routing.Simple,
                Urgency:    routing.Immediate,
                ContentType: routing.Code,
            },
        },
        {
            prompt: "comprehensive API documentation for user service",
            expected: routing.PromptAnalysis{
                Complexity:   routing.Medium,
                Urgency:      routing.QualityFocused,
                ContentType:  routing.Docs,
            },
        },
    }
    
    for _, tt := range tests {
        result := analyzer.Analyze(tt.prompt)
        assert.Equal(t, tt.expected.Complexity, result.Complexity)
        assert.Equal(t, tt.expected.Urgency, result.Urgency)
        assert.Equal(t, tt.expected.ContentType, result.ContentType)
    }
}
```

### **Integration Tests**
```bash
# Test real routing decisions
./uroboro capture -p "quick summary" --dry-run
# Should route to orca-mini:3b

./uroboro capture -p "comprehensive documentation" --dry-run  
# Should route to llama2:13b

./uroboro capture -p "complex system architecture analysis" --dry-run
# Should route to cloud AI
```

### **Dogfooding Tests**
```bash
# Your actual workflow tests
./uroboro capture -p "Fixed memory leak in connection pool"
./uroboro capture -p "Document new authentication middleware"
./uroboro capture -p "Analyze microservice communication patterns"

# Measure actual costs and satisfaction
```

---

## 🎯 Success Metrics

### **Immediate (Week 1)**
- [ ] Routing works for basic speed vs quality decisions
- [ ] Can detect code vs docs vs general content
- [ ] Manual override options work
- [ ] Integration with existing uroboro capture

### **Short-term (Week 2-3)**
- [ ] 90%+ routing decisions feel "right" for your workflow
- [ ] Cost tracking shows actual savings vs manual selection
- [ ] Cloud fallback works for complex tasks
- [ ] No significant workflow slowdown

### **Medium-term (Month 1)**
- [ ] Learning improves routing accuracy over time
- [ ] Budget management prevents cost overruns
- [ ] You prefer smart routing over manual model selection
- [ ] Ready to share as part of optimization framework

---

## 🔧 Configuration

### **Default Config File: `~/.uroboro/smart-routing.yaml`**
```yaml
routing:
  default_strategy: "balanced"  # "speed", "quality", "balanced", "cost"
  
  budget:
    monthly_limit: 25.00
    cloud_ai_ratio: 0.1  # Max 10% of tasks to cloud
    
  models:
    speed_priority: "orca-mini:3b"
    balanced: "mistral:7b" 
    quality_priority: "llama2:13b"
    code_specialist: "codellama:7b"
    
  cloud:
    provider: "claude"  # "claude", "openai", "disabled"
    api_key_env: "ANTHROPIC_API_KEY"
    
  thresholds:
    complexity_for_cloud: 0.8
    quality_for_cloud: 0.9
    speed_requirement_ms: 5000
```

---

## 🚀 Immediate Next Steps

1. **Create routing package structure** in uroboro
2. **Implement basic PromptAnalyzer** with your common use cases
3. **Add smart capture command** to uroboro CLI
4. **Test with your actual daily prompts**
5. **Iterate based on what feels right vs wrong**

This gives you intelligent routing you can start dogfooding immediately while running your optimization framework testing. The routing will get smarter as you use it and tune it to your actual workflow patterns.

Ready to start with the PromptAnalyzer implementation?