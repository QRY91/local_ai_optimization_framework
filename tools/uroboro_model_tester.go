package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// UroboroTestResult represents test results for uroboro-specific scenarios
type UroboroTestResult struct {
	Model             string        `json:"model"`
	UseCase           string        `json:"use_case"`
	TestName          string        `json:"test_name"`
	Input             string        `json:"input"`
	Output            string        `json:"output"`
	ResponseTime      time.Duration `json:"response_time"`
	Success           bool          `json:"success"`
	Error             string        `json:"error,omitempty"`
	QualityScore      int           `json:"quality_score"`      // 1-5 rating
	FormatCompliance  bool          `json:"format_compliance"`  // Does it follow markdown/format rules?
	TechnicalAccuracy bool          `json:"technical_accuracy"` // Are technical terms correct?
	Timestamp         time.Time     `json:"timestamp"`
}

// UroboroTestCase represents a specific uroboro scenario
type UroboroTestCase struct {
	Name        string `json:"name"`
	UseCase     string `json:"use_case"` // "capture", "devlog", "blog", "social"
	Input       string `json:"input"`    // Simulated uroboro input
	Prompt      string `json:"prompt"`   // Actual prompt sent to model
	ExpectedLen int    `json:"expected_length_range"`
}

// UroboroExperiment holds the complete experiment configuration
type UroboroExperiment struct {
	Models    []string           `json:"models"`
	TestCases []UroboroTestCase  `json:"test_cases"`
	Results   []UroboroTestResult `json:"results"`
	Summary   UroboroSummary     `json:"summary"`
	Config    ExperimentConfig   `json:"config"`
}

// UroboroSummary provides uroboro-specific recommendations
type UroboroSummary struct {
	BestModelPerUseCase    map[string]string      `json:"best_model_per_use_case"`
	PerformanceRecommendations map[string]string  `json:"performance_recommendations"`
	QualityRankings        map[string][]ModelRanking `json:"quality_rankings"`
	UroboroConfig          UroboroConfigRecommendation `json:"uroboro_config_recommendation"`
}

// ModelRanking represents model ranking for a specific use case
type ModelRanking struct {
	Model    string  `json:"model"`
	Score    float64 `json:"score"`
	Reason   string  `json:"reason"`
}

// UroboroConfigRecommendation provides specific uroboro configuration advice
type UroboroConfigRecommendation struct {
	PrimaryModel   string            `json:"primary_model"`
	FallbackChain  []string          `json:"fallback_chain"`
	UseCaseModels  map[string]string `json:"use_case_models"`
	TimeoutConfig  map[string]int    `json:"timeout_config"`
	EnvironmentVars map[string]string `json:"environment_vars"`
}

// ExperimentConfig holds experiment parameters
type ExperimentConfig struct {
	TimeoutSeconds int `json:"timeout_seconds"`
	Runs          int `json:"runs"`
	SkipSlow      bool `json:"skip_slow_models"`
}

func main() {
	fmt.Println("🐍 uroboro Model Performance Tester")
	fmt.Println("===================================")
	
	// Initialize experiment
	experiment := initializeExperiment()
	
	// Check available models
	availableModels := checkAvailableModels(experiment.Models)
	if len(availableModels) == 0 {
		log.Fatal("❌ No models available. Run: ollama pull mistral:latest")
	}
	
	fmt.Printf("✅ Found %d available models: %v\n", len(availableModels), availableModels)
	
	// Run experiments
	fmt.Println("\n🧪 Running uroboro-specific tests...")
	runUroboroExperiments(&experiment, availableModels)
	
	// Analyze results and generate summary
	experiment.Summary = generateUroboroSummary(experiment.Results)
	
	// Save results
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	outputFile := fmt.Sprintf("results/uroboro_test_results_%s.json", timestamp)
	
	if err := saveExperimentResults(experiment, outputFile); err != nil {
		log.Printf("⚠️  Failed to save results: %v", err)
	} else {
		fmt.Printf("💾 Results saved to: %s\n", outputFile)
	}
	
	// Print summary and recommendations
	printUroboroSummary(experiment.Summary)
	generateUroboroConfigFiles(experiment.Summary)
}

func initializeExperiment() UroboroExperiment {
	models := []string{
		"mistral:latest",        // Current baseline
		"llama2:7b",            // Alternative
		"llama2:13b",           // Higher quality
		"codellama:7b",         // Code-focused
		"codellama:13b",        // Advanced code
		"dolphin-mistral:latest", // Uncensored
		"orca-mini:3b",         // Fast/lightweight
		"neural-chat:7b",       // Conversational
	}
	
	testCases := []UroboroTestCase{
		// CAPTURE use cases (quick, during development)
		{
			Name:    "Quick Bug Fix Capture",
			UseCase: "capture",
			Input:   "Fixed memory leak in HTTP client by properly closing response bodies",
			Prompt:  buildCapturePrompt("Fixed memory leak in HTTP client by properly closing response bodies"),
			ExpectedLen: 150,
		},
		{
			Name:    "Feature Implementation Capture",
			UseCase: "capture", 
			Input:   "Added JWT authentication middleware with token refresh logic",
			Prompt:  buildCapturePrompt("Added JWT authentication middleware with token refresh logic"),
			ExpectedLen: 200,
		},
		{
			Name:    "Performance Optimization Capture",
			UseCase: "capture",
			Input:   "Optimized database queries, reduced response time from 2s to 200ms",
			Prompt:  buildCapturePrompt("Optimized database queries, reduced response time from 2s to 200ms"),
			ExpectedLen: 180,
		},
		
		// DEVLOG use cases (technical, detailed)
		{
			Name:    "Architecture Refactor Devlog",
			UseCase: "devlog",
			Input:   "Migrated from monolithic to microservices architecture. Split user service, auth service, and notification service. Implemented service mesh with Istio.",
			Prompt:  buildDevlogPrompt("Migrated from monolithic to microservices architecture. Split user service, auth service, and notification service. Implemented service mesh with Istio."),
			ExpectedLen: 800,
		},
		{
			Name:    "API Development Devlog", 
			UseCase: "devlog",
			Input:   "Built RESTful API with Go and Gin. Added rate limiting, request validation, and comprehensive error handling. Integrated with PostgreSQL using GORM.",
			Prompt:  buildDevlogPrompt("Built RESTful API with Go and Gin. Added rate limiting, request validation, and comprehensive error handling. Integrated with PostgreSQL using GORM."),
			ExpectedLen: 600,
		},
		
		// BLOG use cases (professional, external)
		{
			Name:    "Technical Achievement Blog",
			UseCase: "blog",
			Input:   "Successfully migrated legacy system to Kubernetes. Achieved 99.9% uptime, reduced infrastructure costs by 40%, improved deployment frequency from weekly to daily.",
			Prompt:  buildBlogPrompt("Successfully migrated legacy system to Kubernetes. Achieved 99.9% uptime, reduced infrastructure costs by 40%, improved deployment frequency from weekly to daily.", "Building Resilient Systems: Our Kubernetes Migration Story"),
			ExpectedLen: 1200,
		},
		{
			Name:    "Lessons Learned Blog",
			UseCase: "blog", 
			Input:   "Learned about distributed systems challenges while debugging intermittent service failures. Root cause was network partitions and improper timeout handling.",
			Prompt:  buildBlogPrompt("Learned about distributed systems challenges while debugging intermittent service failures. Root cause was network partitions and improper timeout handling.", "Debugging Distributed Systems: A Learning Journey"),
			ExpectedLen: 1000,
		},
		
		// SOCIAL use cases (engaging, concise)
		{
			Name:    "Achievement Social Post",
			UseCase: "social",
			Input:   "Reduced API latency by 85% through intelligent caching strategy. Production system now handles 10x more requests.",
			Prompt:  buildSocialPrompt("Reduced API latency by 85% through intelligent caching strategy. Production system now handles 10x more requests."),
			ExpectedLen: 300,
		},
		{
			Name:    "Learning Social Post",
			UseCase: "social",
			Input:   "Deep dive into Go's garbage collector revealed interesting optimization opportunities. Small changes, big performance impact.",
			Prompt:  buildSocialPrompt("Deep dive into Go's garbage collector revealed interesting optimization opportunities. Small changes, big performance impact."),
			ExpectedLen: 250,
		},
	}
	
	return UroboroExperiment{
		Models:    models,
		TestCases: testCases,
		Config: ExperimentConfig{
			TimeoutSeconds: 45,
			Runs:          2,
			SkipSlow:      false,
		},
	}
}

// Build prompts that match uroboro's actual prompt patterns
func buildCapturePrompt(input string) string {
	return fmt.Sprintf(`Convert this development insight into a concise, professional summary suitable for a development log:

Input: %s

Requirements:
- Keep it brief (1-2 sentences)
- Professional tone
- Technical accuracy
- Include the key benefit or outcome

Summary:`, input)
}

func buildDevlogPrompt(input string) string {
	return fmt.Sprintf(`Create a technical development log entry from this work summary. Format as markdown with clear sections:

Work Summary: %s

Generate a technical devlog that includes:
- ## What Was Done
- ## Technical Details  
- ## Challenges Faced
- ## Outcomes & Benefits
- ## Next Steps

Keep it detailed but focused, suitable for technical team members.`, input)
}

func buildBlogPrompt(input, title string) string {
	return fmt.Sprintf(`Transform this development work into an engaging blog post for a technical audience:

Work Summary: %s
Suggested Title: %s

Create a professional blog post with:
- Engaging introduction
- Technical details and decisions
- Challenges and solutions
- Key takeaways
- Conclusion

Target audience: Software engineers and technical leaders
Tone: Professional but approachable
Format: Well-structured markdown`, input, title)
}

func buildSocialPrompt(input string) string {
	return fmt.Sprintf(`Create engaging social media content for LinkedIn/Twitter from this development work:

Achievement: %s

Requirements:
- Professional but engaging tone
- Include relevant technical hashtags
- Highlight the impact/benefit
- Keep it concise but informative
- Make it shareable

Social Post:`, input)
}

func checkAvailableModels(models []string) []string {
	var available []string
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  Error checking models: %v\n", err)
		return available
	}
	
	modelList := string(output)
	for _, model := range models {
		if strings.Contains(modelList, model) {
			available = append(available, model)
		}
	}
	
	return available
}

func runUroboroExperiments(experiment *UroboroExperiment, models []string) {
	total := len(models) * len(experiment.TestCases) * experiment.Config.Runs
	current := 0
	
	for _, model := range models {
		fmt.Printf("\n🤖 Testing model: %s\n", model)
		
		for _, testCase := range experiment.TestCases {
			fmt.Printf("  📝 %s (%s)", testCase.Name, testCase.UseCase)
			
			for run := 0; run < experiment.Config.Runs; run++ {
				current++
				
				result := testModelWithUroboroCase(model, testCase, experiment.Config.TimeoutSeconds)
				result.QualityScore = evaluateQuality(result, testCase)
				result.FormatCompliance = checkFormatCompliance(result.Output, testCase.UseCase)
				result.TechnicalAccuracy = checkTechnicalAccuracy(result.Output)
				
				experiment.Results = append(experiment.Results, result)
				
				progress := float64(current) / float64(total) * 100
				if result.Success {
					fmt.Printf(" ✅[%.0f%%] %v", progress, result.ResponseTime.Round(time.Millisecond))
				} else {
					fmt.Printf(" ❌[%.0f%%] %s", progress, result.Error)
				}
			}
			fmt.Println()
		}
	}
}

func testModelWithUroboroCase(model string, testCase UroboroTestCase, timeoutSec int) UroboroTestResult {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "ollama", "run", model)
	cmd.Stdin = strings.NewReader(testCase.Prompt)
	
	output, err := cmd.Output()
	responseTime := time.Since(start)
	
	result := UroboroTestResult{
		Model:        model,
		UseCase:      testCase.UseCase,
		TestName:     testCase.Name,
		Input:        testCase.Input,
		ResponseTime: responseTime,
		Timestamp:    start,
	}
	
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}
	
	result.Success = true
	result.Output = strings.TrimSpace(string(output))
	
	return result
}

func evaluateQuality(result UroboroTestResult, testCase UroboroTestCase) int {
	if !result.Success {
		return 0
	}
	
	score := 3 // Base score
	output := result.Output
	
	// Length appropriateness
	length := len(output)
	if testCase.UseCase == "capture" && length > 50 && length < 300 {
		score++
	} else if testCase.UseCase == "social" && length > 100 && length < 500 {
		score++
	} else if (testCase.UseCase == "devlog" || testCase.UseCase == "blog") && length > 300 {
		score++
	}
	
	// Format compliance bonus
	if strings.Contains(output, "##") || strings.Contains(output, "- ") {
		score++
	}
	
	// Technical content bonus (basic check)
	technicalTerms := []string{"API", "database", "server", "client", "service", "system", "performance", "optimization"}
	for _, term := range technicalTerms {
		if strings.Contains(strings.ToLower(output), strings.ToLower(term)) {
			break // Found at least one technical term
		}
	}
	
	// Cap at 5
	if score > 5 {
		score = 5
	}
	
	return score
}

func checkFormatCompliance(output, useCase string) bool {
	switch useCase {
	case "devlog", "blog":
		// Should have markdown headers
		return strings.Contains(output, "##") || strings.Contains(output, "# ")
	case "social":
		// Should be concise and not have complex formatting
		return !strings.Contains(output, "##") && len(output) < 500
	case "capture":
		// Should be simple, no complex formatting needed
		return len(output) > 20 && len(output) < 300
	}
	return true
}

func checkTechnicalAccuracy(output string) bool {
	// Basic technical accuracy check
	// Look for common technical terms and proper capitalization
	technicalTerms := map[string]bool{
		"API":        true,
		"HTTP":       true,
		"REST":       true,
		"JSON":       true,
		"SQL":        true,
		"NoSQL":      true,
		"Docker":     true,
		"Kubernetes": true,
		"Go":         true,
		"JavaScript": true,
	}
	
	words := strings.Fields(output)
	correctTerms := 0
	totalTechnicalTerms := 0
	
	for _, word := range words {
		cleanWord := strings.Trim(word, ".,!?;:")
		if _, isTechnical := technicalTerms[cleanWord]; isTechnical {
			totalTechnicalTerms++
			if cleanWord == word || strings.HasSuffix(word, cleanWord) {
				correctTerms++
			}
		}
	}
	
	// If no technical terms found, assume it's accurate
	if totalTechnicalTerms == 0 {
		return true
	}
	
	// Require at least 80% of technical terms to be properly formatted
	return float64(correctTerms)/float64(totalTechnicalTerms) >= 0.8
}

func generateUroboroSummary(results []UroboroTestResult) UroboroSummary {
	summary := UroboroSummary{
		BestModelPerUseCase:        make(map[string]string),
		PerformanceRecommendations: make(map[string]string),
		QualityRankings:           make(map[string][]ModelRanking),
	}
	
	// Group results by use case
	useCaseResults := make(map[string][]UroboroTestResult)
	for _, result := range results {
		useCaseResults[result.UseCase] = append(useCaseResults[result.UseCase], result)
	}
	
	// Analyze each use case
	for useCase, caseResults := range useCaseResults {
		rankings := analyzeUseCaseResults(caseResults)
		summary.QualityRankings[useCase] = rankings
		
		if len(rankings) > 0 {
			summary.BestModelPerUseCase[useCase] = rankings[0].Model
		}
	}
	
	// Generate performance recommendations
	summary.PerformanceRecommendations = generatePerformanceRecommendations(results)
	
	// Generate uroboro config recommendation
	summary.UroboroConfig = generateConfigRecommendation(summary)
	
	return summary
}

func analyzeUseCaseResults(results []UroboroTestResult) []ModelRanking {
	// Group by model
	modelResults := make(map[string][]UroboroTestResult)
	for _, result := range results {
		modelResults[result.Model] = append(modelResults[result.Model], result)
	}
	
	var rankings []ModelRanking
	
	for model, modelRes := range modelResults {
		if len(modelRes) == 0 {
			continue
		}
		
		// Calculate composite score
		totalScore := 0.0
		successfulRuns := 0
		totalTime := time.Duration(0)
		
		for _, res := range modelRes {
			if res.Success {
				successfulRuns++
				totalTime += res.ResponseTime
				// Score: Quality (50%) + Speed factor (30%) + Format compliance (20%)
				speedScore := 5.0 - (res.ResponseTime.Seconds() / 10.0) // Penalty for slow responses
				if speedScore < 1 {
					speedScore = 1
				}
				
				formatScore := 0.0
				if res.FormatCompliance {
					formatScore = 5.0
				}
				
				testScore := float64(res.QualityScore)*0.5 + speedScore*0.3 + formatScore*0.2
				totalScore += testScore
			}
		}
		
		if successfulRuns > 0 {
			avgScore := totalScore / float64(successfulRuns)
			avgTime := totalTime / time.Duration(successfulRuns)
			
			reason := fmt.Sprintf("Avg quality: %.1f/5, Avg time: %v, Success: %d/%d", 
				avgScore, avgTime.Round(time.Millisecond), successfulRuns, len(modelRes))
			
			rankings = append(rankings, ModelRanking{
				Model:  model,
				Score:  avgScore,
				Reason: reason,
			})
		}
	}
	
	// Sort by score (descending)
	for i := 0; i < len(rankings)-1; i++ {
		for j := i + 1; j < len(rankings); j++ {
			if rankings[j].Score > rankings[i].Score {
				rankings[i], rankings[j] = rankings[j], rankings[i]
			}
		}
	}
	
	return rankings
}

func generatePerformanceRecommendations(results []UroboroTestResult) map[string]string {
	recommendations := make(map[string]string)
	
	// Find fastest model overall
	fastestTime := time.Hour
	fastestModel := ""
	
	// Find most reliable model
	modelReliability := make(map[string]struct{ success, total int })
	
	for _, result := range results {
		if result.Success && result.ResponseTime < fastestTime {
			fastestTime = result.ResponseTime
			fastestModel = result.Model
		}
		
		rel := modelReliability[result.Model]
		rel.total++
		if result.Success {
			rel.success++
		}
		modelReliability[result.Model] = rel
	}
	
	mostReliableModel := ""
	bestReliability := 0.0
	for model, rel := range modelReliability {
		reliability := float64(rel.success) / float64(rel.total)
		if reliability > bestReliability {
			bestReliability = reliability
			mostReliableModel = model
		}
	}
	
	recommendations["fastest"] = fastestModel
	recommendations["most_reliable"] = mostReliableModel
	recommendations["general"] = "Consider using task-specific models for best results"
	
	return recommendations
}

func generateConfigRecommendation(summary UroboroSummary) UroboroConfigRecommendation {
	config := UroboroConfigRecommendation{
		UseCaseModels:   make(map[string]string),
		TimeoutConfig:   make(map[string]int),
		EnvironmentVars: make(map[string]string),
	}
	
	// Set primary model (most versatile)
	if bestCapture, ok := summary.BestModelPerUseCase["capture"]; ok {
		config.PrimaryModel = bestCapture
	} else {
		config.PrimaryModel = "mistral:latest"
	}
	
	// Set use-case specific models
	for useCase, model := range summary.BestModelPerUseCase {
		config.UseCaseModels[useCase] = model
		
		// Set timeouts based on use case
		switch useCase {
		case "capture":
			config.TimeoutConfig[useCase] = 15 // Quick captures
		case "social":
			config.TimeoutConfig[useCase] = 30 // Medium content
		case "devlog":
			config.TimeoutConfig[useCase] = 45 // Detailed content
		case "blog":
			config.TimeoutConfig[useCase] = 60 // Long-form content
		}
	}
	
	// Create fallback chain
	fallbacks := []string{"mistral:latest", "llama2:7b", "orca-mini:3b"}
	config.FallbackChain = fallbacks
	
	// Environment variable recommendations
	config.EnvironmentVars["UROBORO_MODEL"] = config.PrimaryModel
	for useCase, model := range config.UseCaseModels {
		key := fmt.Sprintf("UROBORO_MODEL_%s", strings.ToUpper(useCase))
		config.EnvironmentVars[key] = model
	}
	
	return config
}

func saveExperimentResults(experiment UroboroExperiment, filepath string) error {
	// Ensure results directory exists
	if err := os.MkdirAll("results", 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(experiment, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath, data, 0644)
}

func printUroboroSummary(summary UroboroSummary) {
	fmt.Println("\n🏆 UROBORO MODEL RECOMMENDATIONS")
	fmt.Println("===============================")
	
	fmt.Println("\n📊 Best Models by Use Case:")
	for useCase, model := range summary.BestModelPerUseCase {
		fmt.Printf("  %s: %s\n", useCase, model)
	}
	
	fmt.Println("\n🎯 Performance Insights:")
	for category, recommendation := range summary.PerformanceRecommendations {
		fmt.Printf("  %s: %s\n", category, recommendation)
	}
	
	fmt.Println("\n📈 Quality Rankings:")
	for useCase, rankings := range summary.QualityRankings {
		fmt.Printf("\n  %s:\n", strings.ToUpper(useCase))
		for i, ranking := range rankings {
			if i >= 3 { // Show top 3
				break
			}
			fmt.Printf("    %d. %s (%.2f) - %s\n", i+1, ranking.Model, ranking.Score, ranking.Reason)
		}
	}
	
	fmt.Println("\n⚙️  Recommended uroboro Configuration:")
	config := summary.UroboroConfig
	fmt.Printf("  Primary Model: %s\n", config.PrimaryModel)
	fmt.Printf("  Fallback Chain: %v\n", config.FallbackChain)
	
	fmt.Println("\n🔧 Environment Variables:")
	for key, value := range config.EnvironmentVars {
		fmt.Printf("  export %s=\"%s\"\n", key, value)
	}
}

func generateUroboroConfigFiles(summary UroboroSummary) {
	// Generate shell script for environment setup
	envScript := `#!/bin/bash
# uroboro Model Configuration
# Generated from performance testing

echo "🐍 Setting up uroboro model configuration..."

`
	
	for key, value := range summary.UroboroConfig.EnvironmentVars {
		envScript += fmt.Sprintf("export %s=\"%s\"\n", key, value)
	}
	
	envScript += `
echo "✅ Environment configured for optimal uroboro performance"
echo "Models configured:"
`
	
	for useCase, model := range summary.UroboroConfig.UseCaseModels {
		envScript += fmt.Sprintf("echo \"  %s: %s\"\n", useCase, model)
	}
	
	// Save environment script
	envPath := "results/uroboro_env_setup.sh"
	if err := os.WriteFile(envPath, []byte(envScript), 0755); err == nil {
		fmt.Printf("📄 Environment setup script: %s\n", envPath)
	}
	
	// Generate README with recommendations
	readme := fmt.Sprintf(`# uroboro Model Test Results

## Recommendations

### Best Models by Use Case
%s

### Environment Setup

Run the environment setup script:
```bash
source results/uroboro_env_setup.sh
```

Or set manually:
%s

### Integration with uroboro

Modify your uroboro PublishService to use task-specific models:

```go
func NewPublishServiceWithOptimalModels() *PublishService {
    models := map[string]string{
        "capture": "%s",
        "devlog":  "%s", 
        "blog":    "%s",
        "social":  "%s",
    }
    
    return &PublishService{
        models: models,
        primary: "%s",
        fallbacks: %v,
    }
}
```

### Performance Notes

- Fastest overall: %s
- Most reliable: %s
- For quick captures, use lighter models
- For quality content, use larger models

Generated: %s
`,
		formatUseCaseModels(summary.BestModelPerUseCase),
		formatEnvVars(summary.UroboroConfig.EnvironmentVars),
		summary.UroboroConfig.UseCaseModels["capture"],
		summary.UroboroConfig.UseCaseModels["devlog"],
		summary.UroboroConfig.UseCaseModels["blog"],
		summary.UroboroConfig.UseCaseModels["social"],
		summary.UroboroConfig.PrimaryModel,
		summary.UroboroConfig.FallbackChain,
		summary.PerformanceRecommendations["fastest"],
		summary.PerformanceRecommendations["most_reliable"],
		time.Now().Format("2006-01-02 15:04:05"),
	)
	
	readmePath := "results/UROBORO_MODEL_RECOMMENDATIONS.md"
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err == nil {
		fmt.Printf("📄 Detailed recommendations: %s\n", readmePath)
	}
}

func formatUseCaseModels(models map[string]string) string {
	var result strings.Builder
	for useCase, model := range models {
		result.WriteString(fmt.Sprintf("- **%s**: %s\n", useCase, model))
	}
	return result.String()
}

func formatEnvVars(envVars map[string]string) string {
	var result strings.Builder
	result.WriteString("```bash\n")
	for key, value := range envVars {
		result.WriteString(fmt.Sprintf("export %s=\"%s\"\n", key, value))
	}
	result.WriteString("```\n")
	return result.String()
}