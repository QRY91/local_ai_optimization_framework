package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ModelResult represents the performance and output of a single model test
type ModelResult struct {
	Model          string        `json:"model"`
	Prompt         string        `json:"prompt"`
	Output         string        `json:"output"`
	ResponseTime   time.Duration `json:"response_time"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
	OutputLength   int           `json:"output_length"`
	Timestamp      time.Time     `json:"timestamp"`
}

// TestCase represents a scenario to test across models
type TestCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	UseCase     string `json:"use_case"` // "devlog", "blog", "social", "capture"
}

// ExperimentConfig holds the experiment configuration
type ExperimentConfig struct {
	Models    []string   `json:"models"`
	TestCases []TestCase `json:"test_cases"`
	TimeoutSec int       `json:"timeout_sec"`
	Runs      int        `json:"runs"` // Number of times to run each test
}

// ExperimentResults holds all results from the experiment
type ExperimentResults struct {
	Config      ExperimentConfig `json:"config"`
	Results     []ModelResult    `json:"results"`
	Summary     ResultSummary    `json:"summary"`
	GeneratedAt time.Time        `json:"generated_at"`
}

// ResultSummary provides aggregate statistics
type ResultSummary struct {
	ModelStats map[string]ModelStats `json:"model_stats"`
	BestModel  map[string]string     `json:"best_model"` // use_case -> model
}

// ModelStats contains aggregate statistics for a model
type ModelStats struct {
	SuccessRate     float64       `json:"success_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	AvgOutputLength float64       `json:"avg_output_length"`
	TotalTests      int           `json:"total_tests"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		printHelp()
		return
	}

	fmt.Println("🧪 uroboro Model Comparison Experiment")
	fmt.Println("=====================================")

	config := getDefaultConfig()
	
	// Allow custom config file
	if len(os.Args) > 1 {
		configPath := os.Args[1]
		if customConfig, err := loadConfig(configPath); err == nil {
			config = customConfig
			fmt.Printf("📋 Loaded custom config from %s\n", configPath)
		} else {
			fmt.Printf("⚠️  Failed to load config from %s, using defaults: %v\n", configPath, err)
		}
	}

	fmt.Printf("🎯 Testing %d models across %d test cases\n", len(config.Models), len(config.TestCases))
	fmt.Printf("⏱️  Timeout: %d seconds per test\n", config.TimeoutSec)
	fmt.Printf("🔄 Runs per test: %d\n", config.Runs)
	
	// Verify models are available
	fmt.Println("\n🔍 Checking model availability...")
	availableModels := checkModelAvailability(config.Models)
	if len(availableModels) == 0 {
		log.Fatal("❌ No models available. Please install models with: ollama pull <model-name>")
	}
	fmt.Printf("✅ Found %d available models: %v\n", len(availableModels), availableModels)

	// Run experiments
	fmt.Println("\n🚀 Starting experiments...")
	results := runExperiments(config, availableModels)
	
	// Generate summary
	summary := generateSummary(results)
	
	// Save results
	experimentResults := ExperimentResults{
		Config:      config,
		Results:     results,
		Summary:     summary,
		GeneratedAt: time.Now(),
	}
	
	outputPath := fmt.Sprintf("model_comparison_results_%s.json", time.Now().Format("2006-01-02_15-04-05"))
	if err := saveResults(experimentResults, outputPath); err != nil {
		log.Printf("⚠️  Failed to save results: %v", err)
	} else {
		fmt.Printf("💾 Results saved to: %s\n", outputPath)
	}
	
	// Print summary
	printSummary(summary)
}

func getDefaultConfig() ExperimentConfig {
	return ExperimentConfig{
		Models: []string{
			"mistral:latest",
			"llama2:7b",
			"codellama:7b", 
			"dolphin-mistral:latest",
			"orca-mini:3b",
		},
		TestCases: []TestCase{
			{
				Name:        "Quick Capture",
				Description: "Fast insight capture during development",
				UseCase:     "capture",
				Prompt:      "Convert this development insight into a concise, professional summary: 'Fixed memory leak in connection pool by properly closing idle connections after 30s timeout'",
			},
			{
				Name:        "Technical Devlog",
				Description: "Detailed technical development log",
				UseCase:     "devlog",
				Prompt:      "Create a technical development log entry from these insights: 'Implemented OAuth2 JWT authentication', 'Added rate limiting middleware', 'Optimized database queries reducing response time from 500ms to 50ms'. Format as markdown with technical details.",
			},
			{
				Name:        "Professional Blog Post",
				Description: "High-quality blog content for external sharing",
				UseCase:     "blog",
				Prompt:      "Transform these development activities into an engaging blog post: 'Built microservice architecture', 'Implemented event-driven communication', 'Achieved 99.9% uptime'. Target audience: technical professionals. Include lessons learned.",
			},
			{
				Name:        "Social Media Content",
				Description: "Engaging social media posts about development work",
				UseCase:     "social",
				Prompt:      "Create engaging social media content from: 'Reduced API latency by 80% through caching strategy', 'Deployed to production with zero downtime', 'Team delivered major feature ahead of schedule'. Keep it professional but engaging.",
			},
			{
				Name:        "Code Analysis",
				Description: "Technical analysis of code changes",
				UseCase:     "devlog",
				Prompt:      "Analyze and explain this development work: 'Refactored authentication service from monolithic to microservice architecture. Extracted user management, session handling, and permission services. Implemented service mesh for inter-service communication.' Focus on technical decisions and benefits.",
			},
		},
		TimeoutSec: 45,
		Runs:       2,
	}
}

func loadConfig(path string) (ExperimentConfig, error) {
	var config ExperimentConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	return config, json.Unmarshal(data, &config)
}

func checkModelAvailability(models []string) []string {
	var available []string
	for _, model := range models {
		if isModelAvailable(model) {
			available = append(available, model)
			fmt.Printf("  ✅ %s\n", model)
		} else {
			fmt.Printf("  ❌ %s (not installed)\n", model)
		}
	}
	return available
}

func isModelAvailable(model string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.Contains(string(output), model)
}

func runExperiments(config ExperimentConfig, models []string) []ModelResult {
	var results []ModelResult
	total := len(models) * len(config.TestCases) * config.Runs
	current := 0
	
	for _, model := range models {
		fmt.Printf("\n🤖 Testing model: %s\n", model)
		
		for _, testCase := range config.TestCases {
			fmt.Printf("  📝 %s", testCase.Name)
			
			for run := 0; run < config.Runs; run++ {
				current++
				if config.Runs > 1 {
					fmt.Printf(" (run %d/%d)", run+1, config.Runs)
				}
				
				result := testModel(model, testCase, config.TimeoutSec)
				results = append(results, result)
				
				// Progress indicator
				progress := float64(current) / float64(total) * 100
				fmt.Printf(" [%.1f%%]", progress)
				
				if result.Success {
					fmt.Printf(" ✅ %v", result.ResponseTime.Round(time.Millisecond))
				} else {
					fmt.Printf(" ❌ %s", result.Error)
				}
			}
			fmt.Println()
		}
	}
	
	return results
}

func testModel(model string, testCase TestCase, timeoutSec int) ModelResult {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "ollama", "run", model)
	cmd.Stdin = strings.NewReader(testCase.Prompt)
	
	output, err := cmd.Output()
	responseTime := time.Since(start)
	
	result := ModelResult{
		Model:        model,
		Prompt:       testCase.Prompt,
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
	result.OutputLength = len(result.Output)
	
	return result
}

func generateSummary(results []ModelResult) ResultSummary {
	summary := ResultSummary{
		ModelStats: make(map[string]ModelStats),
		BestModel:  make(map[string]string),
	}
	
	// Calculate per-model statistics
	modelData := make(map[string][]ModelResult)
	for _, result := range results {
		modelData[result.Model] = append(modelData[result.Model], result)
	}
	
	for model, modelResults := range modelData {
		stats := calculateModelStats(modelResults)
		summary.ModelStats[model] = stats
	}
	
	// Find best models per use case (by success rate, then by speed)
	useCaseResults := make(map[string][]ModelResult)
	for _, result := range results {
		// We'll need to track use case somehow - for now, infer from prompt content
		useCase := inferUseCase(result.Prompt)
		useCaseResults[useCase] = append(useCaseResults[useCase], result)
	}
	
	for useCase, caseResults := range useCaseResults {
		bestModel := findBestModel(caseResults)
		if bestModel != "" {
			summary.BestModel[useCase] = bestModel
		}
	}
	
	return summary
}

func calculateModelStats(results []ModelResult) ModelStats {
	if len(results) == 0 {
		return ModelStats{}
	}
	
	successCount := 0
	var totalResponseTime time.Duration
	var totalOutputLength int
	
	for _, result := range results {
		if result.Success {
			successCount++
			totalResponseTime += result.ResponseTime
			totalOutputLength += result.OutputLength
		}
	}
	
	stats := ModelStats{
		TotalTests:  len(results),
		SuccessRate: float64(successCount) / float64(len(results)),
	}
	
	if successCount > 0 {
		stats.AvgResponseTime = totalResponseTime / time.Duration(successCount)
		stats.AvgOutputLength = float64(totalOutputLength) / float64(successCount)
	}
	
	return stats
}

func inferUseCase(prompt string) string {
	prompt = strings.ToLower(prompt)
	if strings.Contains(prompt, "blog post") || strings.Contains(prompt, "engaging blog") {
		return "blog"
	}
	if strings.Contains(prompt, "social media") || strings.Contains(prompt, "engaging social") {
		return "social"
	}
	if strings.Contains(prompt, "devlog") || strings.Contains(prompt, "development log") {
		return "devlog"
	}
	if strings.Contains(prompt, "summary") || strings.Contains(prompt, "concise") {
		return "capture"
	}
	return "general"
}

func findBestModel(results []ModelResult) string {
	if len(results) == 0 {
		return ""
	}
	
	// Group by model and calculate average performance
	modelPerf := make(map[string]struct {
		successRate    float64
		avgResponseTime time.Duration
		count          int
	})
	
	for _, result := range results {
		perf := modelPerf[result.Model]
		perf.count++
		if result.Success {
			perf.successRate += 1.0
			perf.avgResponseTime += result.ResponseTime
		}
		modelPerf[result.Model] = perf
	}
	
	// Finalize averages
	for model, perf := range modelPerf {
		perf.successRate /= float64(perf.count)
		if perf.successRate > 0 {
			perf.avgResponseTime /= time.Duration(perf.successRate * float64(perf.count))
		}
		modelPerf[model] = perf
	}
	
	// Find best model (highest success rate, then fastest)
	bestModel := ""
	bestScore := -1.0
	
	for model, perf := range modelPerf {
		// Score = success_rate - (response_time_seconds / 100)
		// This prioritizes success rate but considers speed
		score := perf.successRate - (perf.avgResponseTime.Seconds() / 100.0)
		if score > bestScore {
			bestScore = score
			bestModel = model
		}
	}
	
	return bestModel
}

func saveResults(results ExperimentResults, path string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func printSummary(summary ResultSummary) {
	fmt.Println("\n📊 EXPERIMENT SUMMARY")
	fmt.Println("====================")
	
	fmt.Println("\n🏆 Best Models by Use Case:")
	for useCase, model := range summary.BestModel {
		fmt.Printf("  %s: %s\n", useCase, model)
	}
	
	fmt.Println("\n📈 Model Performance Stats:")
	for model, stats := range summary.ModelStats {
		fmt.Printf("\n  🤖 %s:\n", model)
		fmt.Printf("    Success Rate: %.1f%% (%d/%d tests)\n", 
			stats.SuccessRate*100, int(stats.SuccessRate*float64(stats.TotalTests)), stats.TotalTests)
		if stats.SuccessRate > 0 {
			fmt.Printf("    Avg Response Time: %v\n", stats.AvgResponseTime.Round(time.Millisecond))
			fmt.Printf("    Avg Output Length: %.0f chars\n", stats.AvgOutputLength)
		}
	}
	
	fmt.Println("\n💡 Recommendations:")
	fastestModel := findFastestModel(summary.ModelStats)
	mostReliableModel := findMostReliableModel(summary.ModelStats)
	
	if fastestModel != "" {
		fmt.Printf("  ⚡ Fastest: %s\n", fastestModel)
	}
	if mostReliableModel != "" {
		fmt.Printf("  🛡️  Most Reliable: %s\n", mostReliableModel)
	}
}

func findFastestModel(stats map[string]ModelStats) string {
	fastest := ""
	fastestTime := time.Hour
	
	for model, stat := range stats {
		if stat.SuccessRate > 0.5 && stat.AvgResponseTime < fastestTime {
			fastestTime = stat.AvgResponseTime
			fastest = model
		}
	}
	return fastest
}

func findMostReliableModel(stats map[string]ModelStats) string {
	mostReliable := ""
	bestRate := 0.0
	
	for model, stat := range stats {
		if stat.SuccessRate > bestRate {
			bestRate = stat.SuccessRate
			mostReliable = model
		}
	}
	return mostReliable
}

func printHelp() {
	fmt.Println("🧪 uroboro Model Comparison Tool")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run model_comparison.go [config.json]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  config.json    Optional JSON config file (uses defaults if not provided)")
	fmt.Println("  --help         Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run model_comparison.go")
	fmt.Println("  go run model_comparison.go custom_config.json")
	fmt.Println()
	fmt.Println("The tool will:")
	fmt.Println("  1. Check which models are available via Ollama")
	fmt.Println("  2. Run test prompts across all available models")
	fmt.Println("  3. Measure response times and success rates")
	fmt.Println("  4. Generate recommendations for different use cases")
	fmt.Println("  5. Save detailed results to JSON file")
	fmt.Println()
	fmt.Println("Make sure you have Ollama installed and models pulled:")
	fmt.Println("  ollama pull mistral:latest")
	fmt.Println("  ollama pull llama2:7b")
	fmt.Println("  ollama pull codellama:7b")
}