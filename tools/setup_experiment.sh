#!/bin/bash

# uroboro Model Experiment Setup Script
# ====================================
# Sets up the environment for running local LLM model comparisons

set -e  # Exit on any error

echo "🧪 uroboro Model Experiment Setup"
echo "=================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Ollama is installed
check_ollama() {
    echo "🔍 Checking Ollama installation..."
    if ! command -v ollama &> /dev/null; then
        echo -e "${RED}❌ Ollama not found${NC}"
        echo "Please install Ollama first:"
        echo "  curl -fsSL https://ollama.ai/install.sh | sh"
        echo "Or visit: https://ollama.ai"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Ollama found${NC}"
    
    # Check if Ollama service is running
    if ! ollama list &> /dev/null; then
        echo -e "${YELLOW}⚠️  Ollama service not running. Starting...${NC}"
        ollama serve &
        sleep 3
    fi
}

# Install recommended models
install_models() {
    echo -e "\n🤖 Installing recommended models..."
    
    # Core models for comparison
    MODELS=(
        "mistral:latest"      # Current baseline
        "llama2:7b"          # Popular alternative
        "codellama:7b"       # Code-specialized
        "orca-mini:3b"       # Speed variant
    )
    
    # Optional models (larger, slower)
    OPTIONAL_MODELS=(
        "llama2:13b"         # Better reasoning
        "codellama:13b"      # Advanced code tasks
        "dolphin-mistral:latest"  # Uncensored variant
        "neural-chat:7b"     # Conversational
    )
    
    echo "Installing core models (required):"
    for model in "${MODELS[@]}"; do
        echo -e "${BLUE}📥 Pulling $model...${NC}"
        if ollama pull "$model"; then
            echo -e "${GREEN}✅ $model installed${NC}"
        else
            echo -e "${RED}❌ Failed to install $model${NC}"
        fi
    done
    
    echo -e "\n${YELLOW}Optional models (install if you have time/bandwidth):${NC}"
    for model in "${OPTIONAL_MODELS[@]}"; do
        echo "  ollama pull $model"
    done
    
    echo -e "\nTo install optional models now, run:"
    echo "  ./setup_experiment.sh --install-optional"
}

# Install optional models
install_optional_models() {
    echo -e "\n🔄 Installing optional models..."
    
    OPTIONAL_MODELS=(
        "llama2:13b"
        "codellama:13b"
        "dolphin-mistral:latest"
        "neural-chat:7b"
    )
    
    for model in "${OPTIONAL_MODELS[@]}"; do
        echo -e "${BLUE}📥 Pulling $model...${NC}"
        if ollama pull "$model"; then
            echo -e "${GREEN}✅ $model installed${NC}"
        else
            echo -e "${RED}❌ Failed to install $model${NC}"
        fi
    done
}

# Setup experiment directory structure
setup_directories() {
    echo -e "\n📁 Setting up experiment directories..."
    
    mkdir -p results
    mkdir -p configs
    mkdir -p datasets
    
    echo -e "${GREEN}✅ Directory structure created${NC}"
}

# Check system requirements
check_requirements() {
    echo -e "\n⚙️  Checking system requirements..."
    
    # Check available memory
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        TOTAL_RAM=$(free -g | awk '/^Mem:/{print $2}')
        echo "Available RAM: ${TOTAL_RAM}GB"
        
        if [ "$TOTAL_RAM" -lt 8 ]; then
            echo -e "${YELLOW}⚠️  Warning: Less than 8GB RAM detected${NC}"
            echo "Some larger models may run slowly or fail"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        TOTAL_RAM=$(sysctl -n hw.memsize | awk '{print int($1/1024/1024/1024)}')
        echo "Available RAM: ${TOTAL_RAM}GB"
        
        if [ "$TOTAL_RAM" -lt 8 ]; then
            echo -e "${YELLOW}⚠️  Warning: Less than 8GB RAM detected${NC}"
            echo "Some larger models may run slowly or fail"
        fi
    fi
    
    # Check Go installation
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        echo -e "${GREEN}✅ Go found: $GO_VERSION${NC}"
    else
        echo -e "${YELLOW}⚠️  Go not found${NC}"
        echo "Install Go to run the comparison script"
    fi
}

# List available models
list_models() {
    echo -e "\n📋 Currently installed models:"
    if ollama list 2>/dev/null | grep -v "NAME"; then
        echo -e "${GREEN}✅ Models available for testing${NC}"
    else
        echo -e "${RED}❌ No models installed${NC}"
        echo "Run: ./setup_experiment.sh --install-models"
    fi
}

# Show usage instructions
show_usage() {
    echo -e "\n🚀 Usage Instructions:"
    echo "====================="
    echo
    echo "1. Run basic experiment:"
    echo "   go run model_comparison.go"
    echo
    echo "2. Use custom configuration:"
    echo "   go run model_comparison.go sample_config.json"
    echo
    echo "3. Quick test with just installed models:"
    echo "   go run model_comparison.go --quick"
    echo
    echo "4. View results:"
    echo "   ls -la results/"
    echo "   cat results/model_comparison_results_*.json"
    echo
    echo -e "${BLUE}💡 Pro Tips:${NC}"
    echo "• Start with core models, add optional ones later"
    echo "• Run experiments during off-hours (CPU intensive)"
    echo "• Results are saved in timestamped JSON files"
    echo "• Compare multiple runs to account for variability"
}

# Main execution
main() {
    case "${1:-}" in
        --install-optional)
            check_ollama
            install_optional_models
            ;;
        --install-models)
            check_ollama
            install_models
            ;;
        --list-models)
            list_models
            ;;
        --requirements)
            check_requirements
            ;;
        --help|-h)
            echo "Usage: $0 [option]"
            echo
            echo "Options:"
            echo "  --install-models    Install core models for experiments"
            echo "  --install-optional  Install optional (larger) models"
            echo "  --list-models       List currently installed models"
            echo "  --requirements      Check system requirements"
            echo "  --help             Show this help"
            echo
            echo "Default (no option): Full setup process"
            ;;
        *)
            # Full setup process
            check_ollama
            check_requirements
            setup_directories
            install_models
            list_models
            show_usage
            ;;
    esac
}

main "$@"

echo
echo -e "${GREEN}🎉 Setup complete!${NC}"
echo -e "Ready to run: ${BLUE}go run model_comparison.go${NC}"