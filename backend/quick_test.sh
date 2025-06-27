#!/bin/bash

# LLM Backend API 快速测试脚本
# 快速测试核心功能，适合日常验证

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
BASE_URL="http://localhost:8080"
TEST_USERNAME="quicktest_$(date +%s)"
TEST_PASSWORD="testpass123"
TEST_EMAIL="${TEST_USERNAME}@example.com"
TOKEN=""

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# HTTP请求函数
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4
    
    if [ -n "$headers" ]; then
        curl -s -X "$method" "$BASE_URL$endpoint" \
             -H "Content-Type: application/json" \
             -H "$headers" \
             -d "$data"
    else
        curl -s -X "$method" "$BASE_URL$endpoint" \
             -H "Content-Type: application/json" \
             -d "$data"
    fi
}

# 快速测试
quick_test() {
    echo "=========================================="
    echo "       LLM Backend 快速测试"
    echo "=========================================="
    
    # 1. 检查服务器
    log_info "1. 检查服务器状态..."
    response=$(curl -s "$BASE_URL/health" || echo "ERROR")
    if [[ "$response" == *"running_models"* ]]; then
        log_success "服务器运行正常"
    else
        log_error "服务器未运行"
        exit 1
    fi
    
    # 2. 注册用户
    log_info "2. 注册测试用户..."
    local data="{\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\",\"email\":\"$TEST_EMAIL\"}"
    response=$(make_request "POST" "/api/v1/auth/register" "$data")
    if [[ "$response" == *"token"* ]]; then
        TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        log_success "用户注册成功"
    else
        log_error "用户注册失败: $response"
        exit 1
    fi
    
    # 3. 测试模型列表
    log_info "3. 获取可用模型..."
    response=$(make_request "GET" "/api/v1/models/" "" "Authorization: Bearer $TOKEN")
    if [[ "$response" == *"models"* ]]; then
        log_success "获取模型列表成功"
    else
        log_error "获取模型列表失败"
        exit 1
    fi
    
    # 4. 测试格式转换（使用最轻量的模型）
    log_info "4. 启动格式转换模型..."
    response=$(make_request "POST" "/api/v1/models/deepseek-coder-1.3b-format/start" "" "Authorization: Bearer $TOKEN")
    if [[ "$response" == *"success"* ]] || [[ "$response" == *"启动成功"* ]]; then
        log_success "模型启动成功"
        
        log_info "等待模型启动..."
        sleep 8
        
        log_info "测试格式转换..."
        local convert_data='{"source_format":"json","target_format":"yaml","content":"{\"test\":\"快速测试\",\"value\":123}"}'
        response=$(make_request "POST" "/api/v1/tasks/convert" "$convert_data" "Authorization: Bearer $TOKEN")
        if [[ "$response" == *"converted_content"* ]]; then
            log_success "格式转换测试成功"
        else
            log_error "格式转换测试失败: $response"
        fi
        
        log_info "停止模型..."
        make_request "POST" "/api/v1/models/deepseek-coder-1.3b-format/stop" "" "Authorization: Bearer $TOKEN" >/dev/null
        sleep 3
    else
        log_error "模型启动失败: $response"
    fi
    
    # 5. 测试作业批改
    log_info "5. 启动教师模型..."
    response=$(make_request "POST" "/api/v1/models/qwen2-7b-teacher/start" "" "Authorization: Bearer $TOKEN")
    if [[ "$response" == *"success"* ]] || [[ "$response" == *"启动成功"* ]]; then
        log_success "教师模型启动成功"
        
        log_info "等待模型启动..."
        sleep 10
        
        log_info "测试作业批改..."
        local homework_data='{"subject":"数学","question":"1+1=?","answer":"2","grade_level":"小学"}'
        response=$(make_request "POST" "/api/v1/tasks/homework" "$homework_data" "Authorization: Bearer $TOKEN")
        if [[ "$response" == *"score"* ]]; then
            log_success "作业批改测试成功"
        else
            log_error "作业批改测试失败: $response"
        fi
        
        log_info "停止模型..."
        make_request "POST" "/api/v1/models/qwen2-7b-teacher/stop" "" "Authorization: Bearer $TOKEN" >/dev/null
        sleep 3
    else
        log_error "教师模型启动失败: $response"
    fi
    
    echo ""
    log_success "快速测试完成！"
    echo "=========================================="
}

# 清理函数
cleanup() {
    log_info "清理环境..."
    if [ -n "$TOKEN" ]; then
        make_request "POST" "/api/v1/models/deepseek-coder-1.3b-format/stop" "" "Authorization: Bearer $TOKEN" >/dev/null 2>&1 || true
        make_request "POST" "/api/v1/models/qwen2-7b-teacher/stop" "" "Authorization: Bearer $TOKEN" >/dev/null 2>&1 || true
    fi
}

# 信号处理
trap cleanup EXIT INT TERM

# 运行测试
quick_test "$@"
