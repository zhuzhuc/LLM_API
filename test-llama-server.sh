#!/bin/bash

# 测试 llama.cpp 服务器脚本
# 用途: 测试单个模型的 llama-cpp-server 是否正常工作

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

# 配置
LLAMA_SERVER="./llama-cpp/build/bin/llama-server"
MODEL_DIR="./models"
TEST_PORT=8081

# 检查可用模型
check_models() {
    log_info "检查可用模型..."
    
    if [ ! -d "$MODEL_DIR" ]; then
        log_error "模型目录不存在: $MODEL_DIR"
        exit 1
    fi
    
    MODELS=($(ls $MODEL_DIR/*.gguf 2>/dev/null))
    
    if [ ${#MODELS[@]} -eq 0 ]; then
        log_error "未找到 .gguf 格式的模型文件"
        exit 1
    fi
    
    log_info "找到 ${#MODELS[@]} 个模型文件:"
    for model in "${MODELS[@]}"; do
        echo "  - $(basename $model)"
    done
}

# 选择测试模型
select_model() {
    echo
    log_info "请选择要测试的模型:"
    
    for i in "${!MODELS[@]}"; do
        echo "  $((i+1)). $(basename ${MODELS[$i]})"
    done
    
    echo -n "请输入模型编号 (1-${#MODELS[@]}): "
    read -r choice
    
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#MODELS[@]}" ]; then
        SELECTED_MODEL="${MODELS[$((choice-1))]}"
        log_info "已选择模型: $(basename $SELECTED_MODEL)"
    else
        log_error "无效的选择"
        exit 1
    fi
}

# 启动 llama-server
start_server() {
    log_info "启动 llama-server..."
    
    # 检查端口是否被占用
    if lsof -Pi :$TEST_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_error "端口 $TEST_PORT 已被占用"
        exit 1
    fi
    
    # 启动服务器
    $LLAMA_SERVER \
        -m "$SELECTED_MODEL" \
        --port $TEST_PORT \
        --host 127.0.0.1 \
        -c 2048 \
        -t 8 \
        --temp 0.7 \
        --top-p 0.9 \
        --repeat-penalty 1.1 \
        > logs/llama-server-test.log 2>&1 &
    
    SERVER_PID=$!
    echo $SERVER_PID > logs/llama-server-test.pid
    
    log_info "服务器已启动 (PID: $SERVER_PID)"
    log_info "日志文件: logs/llama-server-test.log"
}

# 等待服务器启动
wait_for_server() {
    log_info "等待服务器启动..."
    
    for i in {1..30}; do
        if curl -s http://localhost:$TEST_PORT/health > /dev/null 2>&1; then
            log_info "服务器启动成功!"
            return 0
        fi
        sleep 2
        echo -n "."
    done
    
    echo
    log_error "服务器启动超时"
    return 1
}

# 测试服务器
test_server() {
    log_info "测试服务器功能..."
    
    # 测试健康检查
    log_info "1. 测试健康检查..."
    if curl -s http://localhost:$TEST_PORT/health | grep -q "ok\|status"; then
        log_info "✓ 健康检查通过"
    else
        log_error "✗ 健康检查失败"
        return 1
    fi
    
    # 测试模型信息
    log_info "2. 测试模型信息..."
    MODEL_INFO=$(curl -s http://localhost:$TEST_PORT/v1/models)
    if [ $? -eq 0 ] && [ -n "$MODEL_INFO" ]; then
        log_info "✓ 模型信息获取成功"
        echo "   模型信息: $MODEL_INFO"
    else
        log_warn "⚠ 模型信息获取失败（某些版本可能不支持此接口）"
    fi
    
    # 测试文本生成
    log_info "3. 测试文本生成..."
    
    TEST_PROMPT="你好，请简单介绍一下自己。"
    
    RESPONSE=$(curl -s -X POST http://localhost:$TEST_PORT/completion \
        -H "Content-Type: application/json" \
        -d "{
            \"prompt\": \"$TEST_PROMPT\",
            \"n_predict\": 100,
            \"temperature\": 0.7,
            \"top_p\": 0.9,
            \"stop\": [\"\\n\\n\"]
        }")
    
    if [ $? -eq 0 ] && echo "$RESPONSE" | grep -q "content"; then
        log_info "✓ 文本生成测试通过"
        echo "   提示: $TEST_PROMPT"
        echo "   响应: $(echo "$RESPONSE" | jq -r '.content' 2>/dev/null || echo "$RESPONSE")"
    else
        log_error "✗ 文本生成测试失败"
        echo "   响应: $RESPONSE"
        return 1
    fi
    
    log_info "所有测试通过! 🎉"
}

# 停止服务器
stop_server() {
    if [ -f "logs/llama-server-test.pid" ]; then
        PID=$(cat logs/llama-server-test.pid)
        if kill -0 $PID 2>/dev/null; then
            log_info "停止测试服务器 (PID: $PID)..."
            kill $PID
            rm logs/llama-server-test.pid
        fi
    fi
}

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    stop_server
}

# 主函数
main() {
    log_info "=== llama.cpp 服务器测试 ==="
    
    # 创建日志目录
    mkdir -p logs
    
    # 设置清理陷阱
    trap cleanup EXIT INT
    
    check_models
    select_model
    start_server
    
    if wait_for_server; then
        test_server
        
        echo
        log_info "测试完成! 服务器将继续运行..."
        log_info "访问地址: http://localhost:$TEST_PORT"
        log_info "按 Ctrl+C 停止服务器"
        
        # 保持运行
        while true; do
            sleep 1
        done
    else
        log_error "服务器启动失败，请检查日志: logs/llama-server-test.log"
        exit 1
    fi
}

# 运行主函数
main "$@"
