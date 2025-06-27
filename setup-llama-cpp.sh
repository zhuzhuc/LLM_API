#!/bin/bash

# CPU-optimized LLM deployment setup script
# This script installs and configures llama.cpp for CPU inference

set -e

echo "=== CPU LLM Deployment Setup ==="
echo "Setting up llama.cpp for CPU inference..."

# Check system requirements
echo "Checking system requirements..."
if ! command -v git &> /dev/null; then
    echo "Error: git is required but not installed."
    exit 1
fi

if ! command -v make &> /dev/null; then
    echo "Error: make is required but not installed."
    exit 1
fi

# Create directories
echo "Creating directories..."
mkdir -p llama-cpp
mkdir -p models
mkdir -p logs

# Clone and build llama.cpp
echo "Cloning llama.cpp repository..."
if [ ! -d "llama-cpp/.git" ]; then
    git clone https://github.com/ggerganov/llama.cpp.git llama-cpp
else
    echo "llama.cpp already exists, updating..."
    cd llama-cpp
    git pull
    cd ..
fi

echo "Building llama.cpp with CPU optimizations..."
cd llama-cpp

# Build with CPU optimizations
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS - use Metal acceleration if available
    echo "Building for macOS with Metal acceleration..."
    make clean
    make LLAMA_METAL=1 -j$(sysctl -n hw.ncpu)
else
    # Linux - use OpenBLAS if available
    echo "Building for Linux with OpenBLAS..."
    if command -v pkg-config &> /dev/null && pkg-config --exists openblas; then
        echo "OpenBLAS found, building with BLAS acceleration..."
        make clean
        make LLAMA_OPENBLAS=1 -j$(nproc)
    else
        echo "OpenBLAS not found, building with basic CPU support..."
        make clean
        make -j$(nproc)
    fi
fi

cd ..

echo "Verifying llama.cpp installation..."
if [ -f "llama-cpp/main" ] && [ -f "llama-cpp/server" ]; then
    echo "✓ llama.cpp built successfully"
else
    echo "✗ llama.cpp build failed"
    exit 1
fi

# Download recommended models
echo "Setting up model configurations..."

# Create model configurations for CPU-optimized models
cat > models/model_configs.json << 'EOF'
{
  "models": [
    {
      "model_name": "Qwen2-7B-Instruct",
      "model_file": "qwen2-7b-instruct-q4_k_m.gguf",
      "model_path": "models/qwen2-7b-instruct-q4_k_m.gguf",
      "context_length": 32768,
      "max_tokens": 2048,
      "temperature": 0.7,
      "top_p": 0.8,
      "repeat_penalty": 1.1,
      "threads": 12,
      "gpu_layers": 0,
      "description": "Qwen2 7B model optimized for CPU inference"
    },
    {
      "model_name": "Yi-9B-Chat",
      "model_file": "Yi-9B.Q4_K_M.gguf",
      "model_path": "models/Yi-9B.Q4_K_M.gguf",
      "context_length": 32768,
      "max_tokens": 2048,
      "temperature": 0.7,
      "top_p": 0.8,
      "repeat_penalty": 1.1,
      "threads": 12,
      "gpu_layers": 0,
      "description": "Yi 9B model with excellent Chinese support"
    },
    {
      "model_name": "DeepSeek-Coder-6.7B",
      "model_file": "deepseek-coder-6.7b-instruct-Q4_K_M.gguf",
      "model_path": "models/deepseek-coder-6.7b-instruct-Q4_K_M.gguf",
      "context_length": 16384,
      "max_tokens": 2048,
      "temperature": 0.3,
      "top_p": 0.9,
      "repeat_penalty": 1.1,
      "threads": 12,
      "gpu_layers": 0,
      "description": "DeepSeek Coder model for programming tasks"
    },
    {
      "model_name": "Mistral-7B-Instruct",
      "model_file": "mistral-7b-instruct-v0.1.Q4_K_M.gguf",
      "model_path": "models/mistral-7b-instruct-v0.1.Q4_K_M.gguf",
      "context_length": 32768,
      "max_tokens": 2048,
      "temperature": 0.7,
      "top_p": 0.8,
      "repeat_penalty": 1.1,
      "threads": 12,
      "gpu_layers": 0,
      "description": "Mistral 7B model with Apache 2.0 license"
    }
  ]
}
EOF

echo "Creating test script..."
cat > test-model.sh << 'EOF'
#!/bin/bash

# Test script for model inference
echo "Testing model inference..."

MODEL_PATH="models/qwen2-7b-instruct-q4_k_m.gguf"
if [ ! -f "$MODEL_PATH" ]; then
    echo "Model file not found: $MODEL_PATH"
    echo "Please download the model file first."
    exit 1
fi

echo "Running test inference..."
./llama-cpp/main -m "$MODEL_PATH" -p "你好，请介绍一下你自己。" -n 100 -t 8

echo "Test completed."
EOF

chmod +x test-model.sh

echo "Creating download script for models..."
cat > download-models.sh << 'EOF'
#!/bin/bash

# Download script for recommended CPU models
echo "Downloading recommended models for CPU inference..."

# Function to download with progress
download_model() {
    local url=$1
    local filename=$2
    
    if [ -f "models/$filename" ]; then
        echo "Model $filename already exists, skipping..."
        return
    fi
    
    echo "Downloading $filename..."
    curl -L --progress-bar -o "models/$filename" "$url"
    
    if [ $? -eq 0 ]; then
        echo "✓ Downloaded $filename successfully"
    else
        echo "✗ Failed to download $filename"
    fi
}

# Download Qwen2-7B (recommended for Chinese tasks)
echo "Downloading Qwen2-7B-Instruct (Q4_K_M)..."
download_model "https://huggingface.co/Qwen/Qwen2-7B-Instruct-GGUF/resolve/main/qwen2-7b-instruct-q4_k_m.gguf" "qwen2-7b-instruct-q4_k_m.gguf"

# Note: Other models are already present in the models directory
echo "Model download completed."
echo "Available models:"
ls -la models/*.gguf
EOF

chmod +x download-models.sh

echo "Creating startup script..."
cat > start-llm-service.sh << 'EOF'
#!/bin/bash

# Startup script for LLM service
echo "Starting LLM Service..."

# Check if Java is installed
if ! command -v java &> /dev/null; then
    echo "Error: Java is required but not installed."
    exit 1
fi

# Set JVM options for CPU optimization
export JAVA_OPTS="-Xmx8g -Xms4g -XX:+UseG1GC -XX:MaxGCPauseMillis=200"

# Start the Spring Boot application
echo "Starting Spring Boot application with CPU profile..."
java $JAVA_OPTS -jar target/*.jar --spring.profiles.active=cpu
EOF

chmod +x start-llm-service.sh

echo "Creating system info script..."
cat > system-info.sh << 'EOF'
#!/bin/bash

# System information for LLM deployment
echo "=== System Information for LLM Deployment ==="
echo

echo "CPU Information:"
if [[ "$OSTYPE" == "darwin"* ]]; then
    sysctl -n machdep.cpu.brand_string
    echo "CPU Cores: $(sysctl -n hw.ncpu)"
    echo "Memory: $(( $(sysctl -n hw.memsize) / 1024 / 1024 / 1024 ))GB"
else
    lscpu | grep "Model name" | cut -d: -f2 | xargs
    echo "CPU Cores: $(nproc)"
    echo "Memory: $(free -h | awk '/^Mem:/ {print $2}')"
fi

echo
echo "Available Models:"
ls -la models/*.gguf 2>/dev/null || echo "No GGUF models found in models directory"

echo
echo "llama.cpp Status:"
if [ -f "llama-cpp/main" ]; then
    echo "✓ llama.cpp main executable found"
else
    echo "✗ llama.cpp main executable not found"
fi

if [ -f "llama-cpp/server" ]; then
    echo "✓ llama.cpp server executable found"
else
    echo "✗ llama.cpp server executable not found"
fi

echo
echo "Recommended Settings:"
echo "- Threads: $(( $(nproc 2>/dev/null || sysctl -n hw.ncpu) * 3 / 4 ))"
echo "- Memory: Ensure at least 16GB for 7B models"
echo "- Model: Qwen2-7B-Instruct for Chinese tasks"
EOF

chmod +x system-info.sh

echo
echo "=== Setup Complete ==="
echo "✓ llama.cpp installed and built"
echo "✓ Model configurations created"
echo "✓ Helper scripts created"
echo
echo "Next steps:"
echo "1. Run './download-models.sh' to download recommended models"
echo "2. Run './system-info.sh' to check system compatibility"
echo "3. Run './test-model.sh' to test model inference"
echo "4. Build the Java application: mvn clean package"
echo "5. Start the service: ./start-llm-service.sh"
echo
echo "For CPU optimization tips:"
echo "- Use Q4_K_M quantized models for best speed/quality balance"
echo "- Set threads to 75% of your CPU cores"
echo "- Ensure sufficient RAM (16GB+ for 7B models)"
echo "- Consider using smaller models (1.8B-3B) for faster inference"