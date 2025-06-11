package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// LowSpecBenchmark represents the main benchmarking framework
type LowSpecBenchmark struct {
	DeviceProfile HardwareProfile     `json:"device_profile"`
	TestResults   []BenchmarkResult   `json:"test_results"`
	Summary       BenchmarkSummary    `json:"summary"`
	Timestamp     time.Time           `json:"timestamp"`
}

// HardwareProfile captures device specifications
type HardwareProfile struct {
	DeviceName      string  `json:"device_name"`
	OS              string  `json:"os"`
	Architecture    string  `json:"architecture"`
	CPUCores        int     `json:"cpu_cores"`
	TotalRAM        int64   `json:"total_ram_mb"`
	AvailableRAM    int64   `json:"available_ram_mb"`
	StorageType     string  `json:"storage_type"`
	ThermalProfile  string  `json:"thermal_profile"` // "laptop", "desktop", "mobile", "sbc"
	PowerProfile    string  `json:"power_profile"`   // "battery", "plugged", "unlimited"
	Notes           string  `json:"notes"`
}

// BenchmarkResult represents a single model test result
type BenchmarkResult struct {
	Model              string        `json:"model"`
	ModelSize          string        `json:"model_size"`          // "3b", "7b", "13b"
	Quantization       string        `json:"quantization"`        // "fp16", "int8", "int4"
	TestCase           string        `json:"test_case"`
	Prompt             string        `json:"prompt"`
	Output             string        `json:"output"`
	Success            bool          `json:"success"`
	Error              string        `json:"error,omitempty"`

	// Performance Metrics
	ResponseTime       time.Duration `json:"response_time"`
	TimeToFirstToken   time.Duration `json:"time_to_first_token"`
	TokensPerSecond    float64       `json:"tokens_per_second"`

	// Resource Usage
	PeakMemoryMB       int64         `json:"peak_memory_mb"`
	AvgCPUPercent      float64       `json:"avg_cpu_percent"`
	MemoryEfficiency   float64       `json:"memory_efficiency"`   // output_quality / memory_used

	// Low-Spec Specific Metrics
	ThermalThrottling  bool          `json:"thermal_throttling"`
	SwapUsed           bool          `json:"swap_used"`
	OOMRisk            string        `json:"oom_risk"`            // "low", "medium", "high"
	BatteryImpact      string        `json:"battery_impact"`      // "minimal", "moderate", "high"

	// Quality Metrics
	OutputLength       int           `json:"output_length"`
	QualityScore       float64       `json:"quality_score"`       // 1-5 automated assessment
	UsabilityScore     float64       `json:"usability_score"`     // response_time vs quality trade-off

	Timestamp          time.Time     `json:"timestamp"`
}

// BenchmarkSummary provides recommendations and insights
type BenchmarkSummary struct {
	OptimalModels      map[string]string `json:"optimal_models"`      // use_case -> model
	MemoryRecommendations []string       `json:"memory_recommendations"`
	PerformanceInsights   []string       `json:"performance_insights"`
	CostEfficiencyScore   float64        `json:"cost_efficiency_score"`
	RecommendedConfig     RecommendedConfig `json:"recommended_config"`
}

// RecommendedConfig provides optimal settings for this device
type RecommendedConfig struct {
	PrimaryModel       string            `json:"primary_model"`
	FallbackModel      string            `json:"fallback_model"`
	ContextLength      int               `json:"context_length"`
	ConcurrentRequests int               `json:"concurrent_requests"`
	MemoryBuffer       int               `json:"memory_buffer_mb"`
	SwapRecommendation string            `json:"swap_recommendation"`
	EnvironmentVars    map[string]string `json:"environment_vars"`
}

// TestScenario defines different testing scenarios for low-spec optimization
type TestScenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UseCase     string `json:"use_case"`
	Prompt      string `json:"prompt"`
	MaxTokens   int    `json:"max_tokens"`
	Priority    string `json:"priority"` // "speed", "quality", "memory"
}

func main() {
	fmt.Println("🔬 QRY Low-Spec AI Device Benchmark")
	fmt.Println("===================================")
	fmt.Println("Optimizing local AI for resource-constrained devices")
	fmt.Println()

	// Profile the current device
	fmt.Println("📊 Profiling device hardware...")
	profile, err := profileHardware()
	if err != nil {
		log.Fatalf("Failed to profile hardware: %v", err)
	}

	printHardwareProfile(profile)

	// Initialize benchmark
	benchmark := LowSpecBenchmark{
		DeviceProfile: profile,
		Timestamp:     time.Now(),
	}

	// Get available models
	models := getAvailableModels()
	if len(models) == 0 {
		log.Fatal("No models available. Please install some models with: ollama pull <model>")
	}

	fmt.Printf("🤖 Found %d models to test: %v\n", len(models), models)

	// Define test scenarios optimized for low-spec devices
	scenarios := getLowSpecTestScenarios()
	fmt.Printf("🧪 Running %d test scenarios\n", len(scenarios))

	// Run comprehensive benchmarks
	fmt.Println("\n⚡ Starting low-spec optimization tests...")
	benchmark.TestResults = runLowSpecBenchmarks(models, scenarios, profile)

	// Generate summary and recommendations
	benchmark.Summary = generateLowSpecSummary(benchmark.TestResults, profile)

	// Save results
	outputPath := fmt.Sprintf("results/low_spec_benchmark_%s_%s.json",
		sanitizeDeviceName(profile.DeviceName),
		time.Now().Format("2006-01-02_15-04-05"))

	if err := saveBenchmark(benchmark, outputPath); err != nil {
		log.Printf("Failed to save benchmark: %v", err)
	} else {
		fmt.Printf("💾 Results saved to: %s\n", outputPath)
	}

	// Print summary and recommendations
	printLowSpecSummary(benchmark.Summary, profile)

	// Generate uroboro-compatible configuration
	generateUroboroConfig(benchmark.Summary, profile)
}

func profileHardware() (HardwareProfile, error) {
	profile := HardwareProfile{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUCores:     runtime.NumCPU(),
	}

	// Get device name
	if hostname, err := os.Hostname(); err == nil {
		profile.DeviceName = hostname
	} else {
		profile.DeviceName = "unknown"
	}

	// Get memory info
	if runtime.GOOS == "linux" {
		profile.TotalRAM, profile.AvailableRAM = getLinuxMemoryInfo()
		profile.StorageType = getLinuxStorageType()
	} else if runtime.GOOS == "darwin" {
		profile.TotalRAM, profile.AvailableRAM = getMacMemoryInfo()
		profile.StorageType = "SSD" // Most Macs have SSD
	}

	// Infer thermal and power profiles
	profile.ThermalProfile = inferThermalProfile(profile)
	profile.PowerProfile = inferPowerProfile(profile)

	return profile, nil
}

func getLinuxMemoryInfo() (int64, int64) {
	// Read /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(data), "\n")
	var total, available int64

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					total = val / 1024 // Convert KB to MB
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					available = val / 1024 // Convert KB to MB
				}
			}
		}
	}

	return total, available
}

func getMacMemoryInfo() (int64, int64) {
	// Use sysctl to get memory info
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	if total, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
		totalMB := total / 1024 / 1024
		// Estimate available as 70% of total (rough approximation)
		availableMB := int64(float64(totalMB) * 0.7)
		return totalMB, availableMB
	}

	return 0, 0
}

func getLinuxStorageType() string {
	// Check if main storage is SSD
	cmd := exec.Command("lsblk", "-d", "-o", "name,rota")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		if strings.Contains(line, "0") { // ROTA=0 means SSD
			return "SSD"
		}
	}
	return "HDD"
}

func inferThermalProfile(profile HardwareProfile) string {
	// Basic heuristics for thermal profile
	if profile.CPUCores <= 2 && profile.TotalRAM <= 4096 {
		return "mobile" // Phone/tablet
	} else if profile.CPUCores <= 4 && profile.TotalRAM <= 8192 {
		return "laptop" // Laptop/ultrabook
	} else if profile.CPUCores <= 8 && profile.TotalRAM <= 16384 {
		return "desktop" // Desktop/workstation
	} else {
		return "server" // High-end system
	}
}

func inferPowerProfile(profile HardwareProfile) string {
	// Assume battery for mobile/laptop, plugged for others
	if profile.ThermalProfile == "mobile" || profile.ThermalProfile == "laptop" {
		return "battery"
	}
	return "plugged"
}

func getAvailableModels() []string {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(string(output), "\n")
	var models []string

	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			modelName := parts[0]
			// Filter out problematic models for low-spec testing
			if !strings.Contains(modelName, "embedding") {
				models = append(models, modelName)
			}
		}
	}

	return models
}

func getLowSpecTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:        "Ultra-Fast Capture",
			Description: "Minimal latency for real-time development capture",
			UseCase:     "capture",
			Priority:    "speed",
			MaxTokens:   100,
			Prompt:      "Summarize in one sentence: Fixed authentication bug in user service",
		},
		{
			Name:        "Memory-Efficient Summary",
			Description: "Quality summary with minimal memory footprint",
			UseCase:     "summary",
			Priority:    "memory",
			MaxTokens:   200,
			Prompt:      "Create a concise technical summary: Implemented Redis caching layer, reduced database queries by 60%, improved API response time from 800ms to 200ms",
		},
		{
			Name:        "Battery-Friendly Social",
			Description: "Social media content optimized for battery life",
			UseCase:     "social",
			Priority:    "speed",
			MaxTokens:   150,
			Prompt:      "Create engaging social media post: Successfully deployed microservices architecture, achieved 99.9% uptime",
		},
		{
			Name:        "Quality-Focused Devlog",
			Description: "Technical content that balances quality and resources",
			UseCase:     "devlog",
			Priority:    "quality",
			MaxTokens:   400,
			Prompt:      "Write technical development log: Migrated from monolith to microservices. Challenges: data consistency, service discovery, monitoring. Solutions: event sourcing, Consul, Prometheus",
		},
		{
			Name:        "Resource-Constrained Blog",
			Description: "Blog post generation for very low-spec devices",
			UseCase:     "blog",
			Priority:    "memory",
			MaxTokens:   600,
			Prompt:      "Write engaging blog post about: Building resilient distributed systems. Cover: fault tolerance, circuit breakers, graceful degradation. Target: developers learning microservices",
		},
	}
}

func runLowSpecBenchmarks(models []string, scenarios []TestScenario, profile HardwareProfile) []BenchmarkResult {
	var results []BenchmarkResult
	total := len(models) * len(scenarios)
	current := 0

	for _, model := range models {
		fmt.Printf("\n🧪 Testing %s on %s (%s, %dMB RAM)\n",
			model, profile.DeviceName, profile.ThermalProfile, profile.AvailableRAM)

		for _, scenario := range scenarios {
			current++
			fmt.Printf("  📝 %s [%d/%d] ", scenario.Name, current, total)

			result := runLowSpecTest(model, scenario, profile)
			results = append(results, result)

			if result.Success {
				fmt.Printf("✅ %v (%.1f t/s, %dMB)\n",
					result.ResponseTime.Round(time.Millisecond),
					result.TokensPerSecond,
					result.PeakMemoryMB)
			} else {
				fmt.Printf("❌ %s\n", result.Error)
			}

			// Brief pause to let system recover
			time.Sleep(2 * time.Second)
		}
	}

	return results
}

func runLowSpecTest(model string, scenario TestScenario, profile HardwareProfile) BenchmarkResult {
	result := BenchmarkResult{
		Model:     model,
		TestCase:  scenario.Name,
		Prompt:    scenario.Prompt,
		Timestamp: time.Now(),
	}

	// Infer model size from name
	result.ModelSize = inferModelSize(model)
	result.Quantization = inferQuantization(model)

	// Capture initial system state
	initialMemory := getCurrentMemoryUsage()

	// Run the test
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ollama", "run", model)
	cmd.Stdin = strings.NewReader(scenario.Prompt)

	output, err := cmd.Output()
	responseTime := time.Since(start)

	// Capture final system state
	finalMemory := getCurrentMemoryUsage()

	result.ResponseTime = responseTime
	result.PeakMemoryMB = finalMemory - initialMemory

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.OOMRisk = assessOOMRisk(result.PeakMemoryMB, profile.AvailableRAM)
		return result
	}

	result.Success = true
	result.Output = strings.TrimSpace(string(output))
	result.OutputLength = len(result.Output)

	// Calculate performance metrics
	if result.OutputLength > 0 && responseTime > 0 {
		// Rough token estimation (4 chars per token average)
		estimatedTokens := float64(result.OutputLength) / 4.0
		result.TokensPerSecond = estimatedTokens / responseTime.Seconds()
	}

	// Assess low-spec specific metrics
	result.OOMRisk = assessOOMRisk(result.PeakMemoryMB, profile.AvailableRAM)
	result.BatteryImpact = assessBatteryImpact(responseTime, result.PeakMemoryMB, profile.PowerProfile)
	result.SwapUsed = result.PeakMemoryMB > profile.AvailableRAM
	result.ThermalThrottling = assessThermalThrottling(responseTime, profile.ThermalProfile)

	// Calculate efficiency scores
	if result.PeakMemoryMB > 0 {
		result.MemoryEfficiency = float64(result.OutputLength) / float64(result.PeakMemoryMB)
	}

	result.QualityScore = assessOutputQuality(result.Output, scenario.UseCase)
	result.UsabilityScore = calculateUsabilityScore(responseTime, result.QualityScore, scenario.Priority)

	return result
}

func getCurrentMemoryUsage() int64 {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/meminfo")
		if err != nil {
			return 0
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemAvailable:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if val, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
						return val / 1024 // Convert KB to MB
					}
				}
			}
		}
	}
	return 0
}

func inferModelSize(model string) string {
	model = strings.ToLower(model)
	if strings.Contains(model, "3b") {
		return "3b"
	} else if strings.Contains(model, "7b") {
		return "7b"
	} else if strings.Contains(model, "13b") {
		return "13b"
	} else if strings.Contains(model, "30b") || strings.Contains(model, "33b") {
		return "30b+"
	}
	return "unknown"
}

func inferQuantization(model string) string {
	model = strings.ToLower(model)
	if strings.Contains(model, "q4") || strings.Contains(model, "int4") {
		return "int4"
	} else if strings.Contains(model, "q8") || strings.Contains(model, "int8") {
		return "int8"
	} else if strings.Contains(model, "fp16") {
		return "fp16"
	}
	return "fp16" // Default assumption
}

func assessOOMRisk(memoryUsed, availableMemory int64) string {
	if memoryUsed <= 0 || availableMemory <= 0 {
		return "unknown"
	}

	usage := float64(memoryUsed) / float64(availableMemory)
	if usage < 0.5 {
		return "low"
	} else if usage < 0.8 {
		return "medium"
	}
	return "high"
}

func assessBatteryImpact(responseTime time.Duration, memoryUsed int64, powerProfile string) string {
	if powerProfile != "battery" {
		return "n/a"
	}

	// Battery impact based on response time and memory usage
	impact := responseTime.Seconds() + float64(memoryUsed)/1000.0

	if impact < 5.0 {
		return "minimal"
	} else if impact < 15.0 {
		return "moderate"
	}
	return "high"
}

func assessThermalThrottling(responseTime time.Duration, thermalProfile string) bool {
	// Heuristic: if response time is unusually long for thermal profile
	var expectedTime time.Duration

	switch thermalProfile {
	case "mobile":
		expectedTime = 10 * time.Second
	case "laptop":
		expectedTime = 8 * time.Second
	case "desktop":
		expectedTime = 5 * time.Second
	default:
		expectedTime = 3 * time.Second
	}

	return responseTime > expectedTime*2
}

func assessOutputQuality(output, useCase string) float64 {
	// Simple quality assessment based on length and use case
	length := len(output)

	var expectedLength int
	switch useCase {
	case "capture":
		expectedLength = 100
	case "summary":
		expectedLength = 200
	case "social":
		expectedLength = 150
	case "devlog":
		expectedLength = 400
	case "blog":
		expectedLength = 600
	default:
		expectedLength = 200
	}

	// Quality score from 1-5 based on appropriate length
	ratio := float64(length) / float64(expectedLength)
	if ratio < 0.3 {
		return 1.0 // Too short
	} else if ratio < 0.7 {
		return 2.5 // Somewhat short
	} else if ratio <= 1.5 {
		return 4.0 // Good length
	} else if ratio <= 2.0 {
		return 3.0 // Bit long
	}
	return 2.0 // Too long
}

func calculateUsabilityScore(responseTime time.Duration, qualityScore float64, priority string) float64 {
	// Usability score balances speed and quality based on priority
	timeScore := 5.0 - (responseTime.Seconds() / 10.0) // Max 5, decreases with time
	if timeScore < 1.0 {
		timeScore = 1.0
	}

	switch priority {
	case "speed":
		return 0.7*timeScore + 0.3*qualityScore
	case "quality":
		return 0.3*timeScore + 0.7*qualityScore
	case "memory":
		return 0.5*timeScore + 0.5*qualityScore
	default:
		return 0.5*timeScore + 0.5*qualityScore
	}
}

func generateLowSpecSummary(results []BenchmarkResult, profile HardwareProfile) BenchmarkSummary {
	summary := BenchmarkSummary{
		OptimalModels: make(map[string]string),
	}

	// Group results by use case
	useCaseResults := make(map[string][]BenchmarkResult)
	for _, result := range results {
		if result.Success {
			useCase := strings.ToLower(result.TestCase)
			useCaseResults[useCase] = append(useCaseResults[useCase], result)
		}
	}

	// Find optimal model for each use case
	for useCase, caseResults := range useCaseResults {
		bestModel := findOptimalModelForUseCase(caseResults, profile)
		if bestModel != "" {
			summary.OptimalModels[useCase] = bestModel
		}
	}

	// Generate recommendations
	summary.MemoryRecommendations = generateMemoryRecommendations(results, profile)
	summary.PerformanceInsights = generatePerformanceInsights(results, profile)
	summary.CostEfficiencyScore = calculateCostEfficiencyScore(results, profile)
	summary.RecommendedConfig = generateRecommendedConfig(results, profile)

	return summary
}

func findOptimalModelForUseCase(results []BenchmarkResult, profile HardwareProfile) string {
	if len(results) == 0 {
		return ""
	}

	bestModel := ""
	bestScore := 0.0

	for _, result := range results {
		score := result.UsabilityScore

		// Penalize high memory usage on low-spec devices
		if profile.AvailableRAM < 8192 && result.OOMRisk == "high" {
			score -= 2.0
		}

		// Penalize battery impact
		if result.BatteryImpact == "high" {
			score -= 1.0
		}

		// Bonus for thermal efficiency
		if !result.ThermalThrottling {
			score += 0.5
		}

		if score > bestScore {
			bestScore = score
			bestModel = result.Model
		}
	}

	return bestModel
}

func generateMemoryRecommendations(results []BenchmarkResult, profile HardwareProfile) []string {
	var recommendations []string

	if profile.AvailableRAM < 4096 {
		recommendations = append(recommendations, "Consider using 3B models only for this device")
		recommendations = append(recommendations, "Enable swap file (4GB minimum) for stability")
		recommendations = append(recommendations, "Close other applications during AI processing")
	} else if profile.AvailableRAM < 8192 {
		recommendations = append(recommendations, "7B models should work well with careful memory management")
		recommendations = append(recommendations, "Monitor memory usage and consider swap for larger models")
	} else {
		recommendations = append(recommendations, "Your device can handle most models efficiently")
		recommendations = append(recommendations, "Consider running multiple models for different use cases")
	}

	return recommendations
}

func generatePerformanceInsights(results []BenchmarkResult, profile HardwareProfile) []string {
	var insights []string

	// Analyze successful results
	successfulResults := make([]BenchmarkResult, 0)
	for _, result := range results {
		if result.Success {
			successfulResults = append(successfulResults, result)
		}
	}

	if len(successfulResults) == 0 {
		insights = append(insights, "No successful tests - device may be too constrained")
		return insights
	}

	// Average response time
	var totalTime time.Duration
	for _, result := range successfulResults {
		totalTime += result.ResponseTime
	}
	avgTime := totalTime / time.Duration(len(successfulResults))

	insights = append(insights, fmt.Sprintf("Average response time: %v", avgTime.Round(time.Millisecond)))

	// Thermal assessment
	throttlingCount := 0
	for _, result := range successfulResults {
		if result.ThermalThrottling {
			throttlingCount++
		}
	}

	if throttlingCount > 0 {
		insights = append(insights, fmt.Sprintf("Thermal throttling detected in %d/%d tests", throttlingCount, len(successfulResults)))
	}

	return insights
}

func calculateCostEfficiencyScore(results []BenchmarkResult, profile HardwareProfile) float64 {
	// Calculate cost efficiency as quality/resource_usage
	var totalScore float64
	var count int

	for _, result := range results {
		if result.Success && result.PeakMemoryMB > 0 {
			efficiency := (result.QualityScore * result.TokensPerSecond) / float64(result.PeakMemoryMB)
			totalScore += efficiency
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalScore / float64(count)
}

func generateRecommendedConfig(results []BenchmarkResult, profile HardwareProfile) RecommendedConfig {
	config := RecommendedConfig{
		EnvironmentVars: make(map[string]string),
	}

	// Find the most successful model
	bestModel := ""
	bestScore := 0.0

	for _, result := range results {
		if result.Success && result.UsabilityScore > bestScore {
			bestScore = result.UsabilityScore
			bestModel = result.Model
		}
	}

	config.PrimaryModel = bestModel

	// Memory management settings
	if profile.AvailableRAM < 8192 {
		config.ContextLength = 2048
		config.ConcurrentRequests = 1
		config.MemoryBuffer = 1024
		config.SwapRecommendation = "Enable 4GB swap file"
	} else {
		config.ContextLength = 4096
		config.ConcurrentRequests = 2
		config.MemoryBuffer = 2048
		config.SwapRecommendation = "Optional"
	}

	// Environment variables for optimization
	config.EnvironmentVars["OLLAMA_MAX_LOADED_MODELS"] = "1"
	config.EnvironmentVars["OLLAMA_NUM_PARALLEL"] = "1"

	if profile.ThermalProfile == "mobile" || profile.ThermalProfile == "laptop" {
		config.EnvironmentVars["OLLAMA_FLASH_ATTENTION"] = "1"
	}

	return config
}

func saveBenchmark(benchmark LowSpecBenchmark, path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(benchmark, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func printHardwareProfile(profile HardwareProfile) {
	fmt.Printf("📱 Device: %s\n", profile.DeviceName)
	fmt.Printf("💻 OS: %s/%s\n", profile.OS, profile.Architecture)
	fmt.Printf("🔧 CPU Cores: %d\n", profile.CPUCores)
	fmt.Printf("💾 RAM: %dMB total, %dMB available\n", profile.TotalRAM, profile.AvailableRAM)
	fmt.Printf("💽 Storage: %s\n", profile.StorageType)
	fmt.Printf("🌡️  Thermal Profile: %s\n", profile.ThermalProfile)
	fmt.Printf("🔋 Power Profile: %s\n", profile.PowerProfile)
	fmt.Println()
}

func printLowSpecSummary(summary BenchmarkSummary, profile HardwareProfile) {
	fmt.Println("\n🎯 LOW-SPEC OPTIMIZATION RESULTS")
	fmt.Println("=================================")

	fmt.Println("\n🏆 Optimal Models by Use Case:")
	for useCase, model := range summary.OptimalModels {
		fmt.Printf("  %s: %s\n", useCase, model)
	}

	fmt.Println("\n💾 Memory Recommendations:")
	for _, rec := range summary.MemoryRecommendations {
		fmt.Printf("  • %s\n", rec)
	}

	fmt.Println("\n⚡ Performance Insights:")
	for _, insight := range summary.PerformanceInsights {
		fmt.Printf("  • %s\n", insight)
	}

	fmt.Printf("\n💰 Cost Efficiency Score: %.2f\n", summary.CostEfficiencyScore)

	fmt.Println("\n⚙️ Recommended Configuration:")
	config := summary.RecommendedConfig
	fmt.Printf("  Primary Model: %s\n", config.PrimaryModel)
	if config.FallbackModel != "" {
		fmt.Printf("  Fallback Model: %s\n", config.FallbackModel)
	}
	fmt.Printf("  Context Length: %d tokens\n", config.ContextLength)
	fmt.Printf("  Concurrent Requests: %d\n", config.ConcurrentRequests)
	fmt.Printf("  Memory Buffer: %dMB\n", config.MemoryBuffer)
	fmt.Printf("  Swap Recommendation: %s\n", config.SwapRecommendation)

	if len(config.EnvironmentVars) > 0 {
		fmt.Println("\n  Environment Variables:")
		for key, value := range config.EnvironmentVars {
			fmt.Printf("    %s=%s\n", key, value)
		}
	}
}

func generateUroboroConfig(summary BenchmarkSummary, profile HardwareProfile) {
	// Generate uroboro-compatible configuration files
	config := summary.RecommendedConfig

	// Shell script for environment setup
	shellScript := fmt.Sprintf(`#!/bin/bash
# QRY Low-Spec AI Configuration for %s
# Generated: %s
# Device Profile: %s, %dMB RAM, %s

echo "🔧 Configuring QRY AI for low-spec device optimization..."

# Ollama optimizations
export OLLAMA_MAX_LOADED_MODELS=1
export OLLAMA_NUM_PARALLEL=1
export OLLAMA_FLASH_ATTENTION=1

# Primary model for uroboro
export UROBORO_DEFAULT_MODEL="%s"
export UROBORO_CONTEXT_LENGTH=%d
export UROBORO_MAX_CONCURRENT=%d

# Memory management
export UROBORO_MEMORY_BUFFER=%d

echo "✅ Low-spec optimization applied"
echo "📊 Device Profile: %s"
echo "🤖 Primary Model: %s"
echo "💾 Memory Buffer: %dMB"
`,
		profile.DeviceName,
		time.Now().Format("2006-01-02 15:04:05"),
		profile.ThermalProfile,
		profile.AvailableRAM,
		profile.StorageType,
		config.PrimaryModel,
		config.ContextLength,
		config.ConcurrentRequests,
		config.MemoryBuffer,
		profile.ThermalProfile,
		config.PrimaryModel,
		config.MemoryBuffer,
	)

	// Save shell script
	scriptPath := fmt.Sprintf("results/uroboro_lowspec_config_%s.sh", sanitizeDeviceName(profile.DeviceName))
	if err := os.WriteFile(scriptPath, []byte(shellScript), 0755); err != nil {
		log.Printf("Failed to save shell config: %v", err)
	} else {
		fmt.Printf("🔧 Shell config saved to: %s\n", scriptPath)
	}

	// Generate JSON config for uroboro
	jsonConfig := map[string]interface{}{
		"device_profile":    profile,
		"optimal_models":    summary.OptimalModels,
		"recommended_model": config.PrimaryModel,
		"context_length":    config.ContextLength,
		"memory_buffer":     config.MemoryBuffer,
		"environment_vars":  config.EnvironmentVars,
		"performance_notes": summary.PerformanceInsights,
		"generated_at":      time.Now(),
	}

	jsonData, err := json.MarshalIndent(jsonConfig, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal JSON config: %v", err)
		return
	}

	jsonPath := fmt.Sprintf("results/uroboro_lowspec_config_%s.json", sanitizeDeviceName(profile.DeviceName))
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		log.Printf("Failed to save JSON config: %v", err)
	} else {
		fmt.Printf("📄 JSON config saved to: %s\n", jsonPath)
	}

	// Generate usage instructions
	instructions := fmt.Sprintf(`# QRY Low-Spec AI Setup Instructions

## Device: %s (%s, %dMB RAM)

### Quick Setup
1. Apply configuration:
   ```bash
   source %s
   ```

2. Test with uroboro:
   ```bash
   cd ../../labs/projects/uroboro
   ./uroboro capture "Testing low-spec optimization"
   ```

### Recommended Models
- Primary: %s
- Use cases: %s

### Memory Management
- Available RAM: %dMB
- Recommended buffer: %dMB
- Swap recommendation: %s

### Performance Expectations
%s

### Troubleshooting
- If you get OOM errors, try: `sudo swapon -s` to check swap
- Monitor memory: `htop` or `free -m`
- If thermal throttling occurs, reduce concurrent requests to 1

Generated: %s
`,
		profile.DeviceName,
		profile.ThermalProfile,
		profile.AvailableRAM,
		scriptPath,
		config.PrimaryModel,
		formatOptimalModels(summary.OptimalModels),
		profile.AvailableRAM,
		config.MemoryBuffer,
		config.SwapRecommendation,
		strings.Join(summary.PerformanceInsights, "\n- "),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	instructionsPath := fmt.Sprintf("results/LOWSPEC_SETUP_%s.md", sanitizeDeviceName(profile.DeviceName))
	if err := os.WriteFile(instructionsPath, []byte(instructions), 0644); err != nil {
		log.Printf("Failed to save instructions: %v", err)
	} else {
		fmt.Printf("📋 Setup instructions saved to: %s\n", instructionsPath)
	}
}

func sanitizeDeviceName(name string) string {
	// Replace spaces and special characters with underscores
	sanitized := strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return strings.ToLower(sanitized)
}

func formatOptimalModels(models map[string]string) string {
	var parts []string
	for useCase, model := range models {
		parts = append(parts, fmt.Sprintf("%s (%s)", useCase, model))
	}
	return strings.Join(parts, ", ")
}
