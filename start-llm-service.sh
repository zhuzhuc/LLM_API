#!/bin/bash

# CPU 大模型服务启动脚本
# 作者: AI Assistant
# 用途: 启动基于 llama.cpp 的 CPU 大模型服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go 1.19+"
        exit 1
    fi
    
    # 检查 llama.cpp 是否编译
    if [ ! -f "./llama-cpp/build/bin/llama-server" ]; then
        log_error "llama-server 未找到，请先编译 llama.cpp"
        log_info "运行: cd llama-cpp && mkdir -p build && cd build && cmake .. && make -j\$(nproc)"
        exit 1
    fi
    
    # 检查模型文件
    if [ ! -d "./models" ] || [ -z "$(ls -A ./models/*.gguf 2>/dev/null)" ]; then
        log_error "未找到模型文件，请确保 models 目录中有 .gguf 格式的模型文件"
        exit 1
    fi
    
    log_info "依赖检查完成"
}

# 创建环境配置
setup_environment() {
    log_info "设置环境配置..."
    
    if [ ! -f ".env" ]; then
        log_info "创建 .env 文件..."
        cp .env.example .env
        log_warn "请根据需要修改 .env 文件中的配置"
    fi
    
    # 创建日志目录
    mkdir -p logs
    
    log_info "环境配置完成"
}

# 启动后端服务
start_backend() {
    log_info "启动后端服务..."

    # 安装依赖
    if [ ! -f "backend/go.sum" ]; then
        log_info "初始化 Go 模块..."
        cd backend && go mod tidy && cd ..
    fi

    # 编译并启动 (从项目根目录启动，这样相对路径才正确)
    log_info "编译并启动后端服务..."
    cd backend && go build -o llm-server main.go && cd ..
    ./backend/llm-server &
    BACKEND_PID=$!

    log_info "后端服务已启动 (PID: $BACKEND_PID)"
    echo $BACKEND_PID > logs/backend.pid
}

# 等待服务启动
wait_for_service() {
    log_info "等待服务启动..."
    
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            log_info "服务启动成功!"
            return 0
        fi
        sleep 1
    done
    
    log_error "服务启动超时"
    return 1
}

# 显示服务信息
show_service_info() {
    log_info "=== CPU 大模型服务平台信息 ==="
    echo
    echo "🚀 服务地址:"
    echo "   - API 服务: http://localhost:8080"
    echo "   - 健康检查: http://localhost:8080/health"
    echo "   - 系统状态: http://localhost:8080/status"
    echo
    echo "📊 核心接口:"
    echo "   模型管理:"
    echo "   - GET  /api/v1/models          - 获取可用模型"
    echo "   - GET  /api/v1/models/running  - 获取运行中模型"
    echo "   - POST /api/v1/models/{name}/start - 启动模型"
    echo "   - POST /api/v1/models/{name}/chat  - 与模型对话"
    echo
    echo "   OpenAI 兼容接口:"
    echo "   - POST /api/v1/v1/chat/completions - 聊天完成"
    echo "   - GET  /api/v1/v1/models           - 模型列表"
    echo "   - POST /api/v1/v1/batch            - 批量请求"
    echo
    echo "   监控和管理:"
    echo "   - GET  /api/v1/monitoring/metrics - 系统指标"
    echo "   - GET  /api/v1/logs               - 系统日志"
    echo "   - GET  /api/v1/discovery/services - 服务发现"
    echo "   - GET  /api/v1/cluster/stats      - 集群状态"
    echo
    echo "📁 重要文件:"
    echo "   - 配置文件: .env"
    echo "   - 模型配置: models/model_config.json"
    echo "   - 日志目录: logs/"
    echo "   - 后端服务: backend/llm-server"
    echo
    echo "🛠️  管理命令:"
    echo "   - 停止服务: ./stop-llm-service.sh"
    echo "   - 查看日志: tail -f logs/*.log"
    echo "   - 重启服务: ./restart-llm-service.sh"
    echo "   - 集群管理: curl http://localhost:8080/api/v1/cluster/stats"
    echo
}

# 创建停止脚本
create_stop_script() {
    cat > stop-llm-service.sh << 'EOF'
#!/bin/bash

# 停止 LLM 服务脚本

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 停止后端服务
if [ -f "logs/backend.pid" ]; then
    BACKEND_PID=$(cat logs/backend.pid)
    if kill -0 $BACKEND_PID 2>/dev/null; then
        log_info "停止后端服务 (PID: $BACKEND_PID)..."
        kill $BACKEND_PID
        rm logs/backend.pid
    else
        log_error "后端服务进程不存在"
    fi
else
    log_error "未找到后端服务 PID 文件"
fi

# 停止所有 llama-server 进程
log_info "停止所有模型服务..."
pkill -f "llama-server" || true

log_info "服务已停止"
EOF

    chmod +x stop-llm-service.sh
}

# 主函数
main() {
    log_info "启动 CPU 大模型服务平台..."
    
    check_dependencies
    setup_environment
    create_stop_script
    start_backend
    
    if wait_for_service; then
        show_service_info
        
        log_info "服务启动完成! 按 Ctrl+C 停止服务"
        
        # 等待中断信号
        trap 'log_info "正在停止服务..."; ./stop-llm-service.sh; exit 0' INT
        
        # 保持脚本运行
        while true; do
            sleep 1
        done
    else
        log_error "服务启动失败"
        ./stop-llm-service.sh
        exit 1
    fi
}

# 运行主函数
main "$@"
