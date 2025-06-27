#!/bin/bash

# CPU 大模型服务平台自动化部署脚本
# 支持单机部署、集群部署和 Docker 部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# 显示帮助信息
show_help() {
    cat << EOF
CPU 大模型服务平台部署脚本

用法: $0 [选项]

选项:
    -t, --type TYPE         部署类型: standalone, cluster, docker (默认: standalone)
    -m, --models PATH       模型文件目录 (默认: ./models)
    -p, --port PORT         服务端口 (默认: 8080)
    -w, --workers NUM       工作节点数量 (集群模式, 默认: 2)
    --master-host HOST      主节点地址 (工作节点模式)
    --master-port PORT      主节点端口 (工作节点模式)
    --node-role ROLE        节点角色: master, worker (默认: master)
    --skip-deps             跳过依赖检查
    --skip-build            跳过编译步骤
    --config-only           仅生成配置文件
    -h, --help              显示帮助信息

示例:
    # 单机部署
    $0 --type standalone

    # 集群主节点部署
    $0 --type cluster --node-role master

    # 集群工作节点部署
    $0 --type cluster --node-role worker --master-host 192.168.1.100

    # Docker 部署
    $0 --type docker --workers 3

    # 仅生成配置
    $0 --config-only
EOF
}

# 默认配置
DEPLOY_TYPE="standalone"
MODELS_PATH="./models"
SERVER_PORT="8080"
WORKERS_COUNT="2"
MASTER_HOST=""
MASTER_PORT="8080"
NODE_ROLE="master"
SKIP_DEPS=false
SKIP_BUILD=false
CONFIG_ONLY=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            DEPLOY_TYPE="$2"
            shift 2
            ;;
        -m|--models)
            MODELS_PATH="$2"
            shift 2
            ;;
        -p|--port)
            SERVER_PORT="$2"
            shift 2
            ;;
        -w|--workers)
            WORKERS_COUNT="$2"
            shift 2
            ;;
        --master-host)
            MASTER_HOST="$2"
            shift 2
            ;;
        --master-port)
            MASTER_PORT="$2"
            shift 2
            ;;
        --node-role)
            NODE_ROLE="$2"
            shift 2
            ;;
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --config-only)
            CONFIG_ONLY=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证部署类型
case $DEPLOY_TYPE in
    standalone|cluster|docker)
        ;;
    *)
        log_error "无效的部署类型: $DEPLOY_TYPE"
        exit 1
        ;;
esac

# 检查系统依赖
check_dependencies() {
    if [[ "$SKIP_DEPS" == "true" ]]; then
        log_info "跳过依赖检查"
        return
    fi

    log_info "检查系统依赖..."

    # 检查基本工具
    for cmd in git wget curl; do
        if ! command -v $cmd &> /dev/null; then
            log_error "$cmd 未安装"
            exit 1
        fi
    done

    # 检查 Go
    if [[ "$DEPLOY_TYPE" != "docker" ]]; then
        if ! command -v go &> /dev/null; then
            log_error "Go 未安装，请先安装 Go 1.19+"
            exit 1
        fi
    fi

    # 检查 Docker (如果需要)
    if [[ "$DEPLOY_TYPE" == "docker" ]]; then
        if ! command -v docker &> /dev/null; then
            log_error "Docker 未安装"
            exit 1
        fi
        if ! command -v docker-compose &> /dev/null; then
            log_error "Docker Compose 未安装"
            exit 1
        fi
    fi

    log_info "依赖检查完成"
}

# 生成配置文件
generate_config() {
    log_info "生成配置文件..."

    # 生成 .env 文件
    cat > .env << EOF
# 数据库配置
DATABASE_URL=sqlite3://./llm.db

# JWT 配置
JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "your-secret-key-change-in-production")

# 环境配置
ENVIRONMENT=production

# Token 配置
TOKEN_RATE=0.001
DEFAULT_TOKENS=1000

# 服务器配置
SERVER_PORT=$SERVER_PORT
LLAMA_CPP_PORT=8081

# llama.cpp 配置
LLAMA_CPP_PATH=./llama-cpp/build/bin/llama-server
MODELS_PATH=$MODELS_PATH
MODEL_CONFIG_PATH=$MODELS_PATH/model_config.json

# 集群配置
NODE_ROLE=$NODE_ROLE
MASTER_HOST=$MASTER_HOST
MASTER_PORT=$MASTER_PORT
EOF

    # 创建目录
    mkdir -p logs data $MODELS_PATH

    # 检查模型配置文件
    if [[ ! -f "$MODELS_PATH/model_config.json" ]]; then
        log_warn "模型配置文件不存在，创建默认配置..."
        cat > "$MODELS_PATH/model_config.json" << EOF
{
  "models": [
    {
      "modelName": "qwen2-7b-instruct",
      "modelFile": "qwen2-7b-instruct-q4_k_m.gguf",
      "modelPath": "$MODELS_PATH/",
      "contextLength": 32768,
      "maxTokens": 2048,
      "temperature": 0.7,
      "topP": 0.8,
      "repeatPenalty": 1.1,
      "threads": 8,
      "gpuLayers": 0,
      "active": true,
      "description": "Qwen2 7B 指令微调模型，适合中文对话"
    }
  ]
}
EOF
    fi

    log_info "配置文件生成完成"
}

# 编译项目
build_project() {
    if [[ "$SKIP_BUILD" == "true" ]]; then
        log_info "跳过编译步骤"
        return
    fi

    log_info "编译项目..."

    # 编译 llama.cpp
    if [[ ! -f "llama-cpp/build/bin/llama-server" ]]; then
        log_info "编译 llama.cpp..."
        cd llama-cpp
        mkdir -p build
        cd build
        cmake .. -DLLAMA_OPENBLAS=ON
        make -j$(nproc)
        cd ../..
    fi

    # 编译后端
    log_info "编译后端服务..."
    cd backend
    go mod tidy
    go build -o llm-server main.go
    cd ..

    log_info "编译完成"
}

# 单机部署
deploy_standalone() {
    log_info "开始单机部署..."
    
    generate_config
    build_project
    
    log_info "单机部署完成"
    log_info "启动服务: ./start-llm-service.sh"
}

# 集群部署
deploy_cluster() {
    log_info "开始集群部署 (角色: $NODE_ROLE)..."
    
    generate_config
    build_project
    
    if [[ "$NODE_ROLE" == "master" ]]; then
        log_info "主节点部署完成"
        log_info "启动主节点: ./start-llm-service.sh"
    else
        if [[ -z "$MASTER_HOST" ]]; then
            log_error "工作节点需要指定主节点地址"
            exit 1
        fi
        log_info "工作节点部署完成"
        log_info "启动工作节点: ./start-llm-service.sh --worker --master $MASTER_HOST:$MASTER_PORT"
    fi
}

# Docker 部署
deploy_docker() {
    log_info "开始 Docker 部署..."
    
    # 检查 Docker 文件
    if [[ ! -f "Dockerfile" ]]; then
        log_error "Dockerfile 不存在"
        exit 1
    fi
    
    if [[ ! -f "docker-compose.yml" ]]; then
        log_error "docker-compose.yml 不存在"
        exit 1
    fi
    
    # 创建必要目录
    mkdir -p models logs data ssl
    
    # 构建镜像
    log_info "构建 Docker 镜像..."
    docker-compose build
    
    # 启动服务
    log_info "启动 Docker 服务..."
    docker-compose up -d
    
    log_info "Docker 部署完成"
    log_info "查看状态: docker-compose ps"
    log_info "查看日志: docker-compose logs -f"
}

# 主函数
main() {
    log_info "=== CPU 大模型服务平台部署脚本 ==="
    log_info "部署类型: $DEPLOY_TYPE"
    log_info "模型路径: $MODELS_PATH"
    log_info "服务端口: $SERVER_PORT"
    
    if [[ "$CONFIG_ONLY" == "true" ]]; then
        generate_config
        log_info "配置文件生成完成"
        exit 0
    fi
    
    check_dependencies
    
    case $DEPLOY_TYPE in
        standalone)
            deploy_standalone
            ;;
        cluster)
            deploy_cluster
            ;;
        docker)
            deploy_docker
            ;;
    esac
    
    log_info "部署完成! 🎉"
}

# 运行主函数
main "$@"
