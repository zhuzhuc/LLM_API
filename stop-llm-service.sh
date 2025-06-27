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
