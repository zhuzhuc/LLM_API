#!/bin/bash

# 单个API接口测试脚本
# 用法: ./test_single_api.sh [接口类型]
# 接口类型: register, login, convert, homework, models

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 配置
BASE_URL="http://localhost:8080"
USERNAME="singletest_$(date +%s)"
PASSWORD="testpass123"
EMAIL="${USERNAME}@example.com"

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_curl() { echo -e "${CYAN}[CURL]${NC} $1"; }

# 显示响应
show_response() {
    local response="$1"
    echo -e "${YELLOW}=== 响应结果 ===${NC}"
    echo "$response" | jq . 2>/dev/null || echo "$response"
    echo ""
}

# 获取token
get_token() {
    log_info "获取认证token..."
    
    # 先注册
    register_data="{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"email\":\"$EMAIL\"}"
    response=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "$register_data")
    
    if [[ "$response" == *"token"* ]]; then
        TOKEN=$(echo "$response" | jq -r '.token' 2>/dev/null || echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        log_success "获得token: ${TOKEN:0:20}..."
        return 0
    else
        # 如果注册失败，尝试登录
        login_data="{\"username\":\"123456\",\"password\":\"123456\"}"
        response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
            -H "Content-Type: application/json" \
            -d "$login_data")
        
        if [[ "$response" == *"token"* ]]; then
            TOKEN=$(echo "$response" | jq -r '.token' 2>/dev/null || echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
            log_success "使用默认用户登录，获得token: ${TOKEN:0:20}..."
            return 0
        else
            log_error "无法获取token"
            return 1
        fi
    fi
}

# 测试用户注册
test_register() {
    log_info "测试用户注册接口"
    
    local data="{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"email\":\"$EMAIL\"}"
    log_curl "curl -X POST $BASE_URL/api/v1/auth/register -d '$data'"
    
    response=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    show_response "$response"
    
    if [[ "$response" == *"token"* ]]; then
        log_success "用户注册成功"
    else
        log_error "用户注册失败"
    fi
}

# 测试用户登录
test_login() {
    log_info "测试用户登录接口"
    
    local data="{\"username\":\"123456\",\"password\":\"123456\"}"
    log_curl "curl -X POST $BASE_URL/api/v1/auth/login -d '$data'"
    
    response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    show_response "$response"
    
    if [[ "$response" == *"token"* ]]; then
        log_success "用户登录成功"
    else
        log_error "用户登录失败"
    fi
}

# 测试格式转换
test_convert() {
    log_info "测试文件格式转换接口"
    
    get_token || exit 1
    
    # 启动模型
    log_info "启动格式转换模型..."
    curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/v1/models/deepseek-coder-1.3b-format/start" > /dev/null
    
    log_info "等待模型启动..."
    sleep 8
    
    local data='{"source_format":"json","target_format":"yaml","content":"{\"name\":\"单接口测试\",\"value\":123,\"active\":true}"}'
    log_curl "curl -X POST $BASE_URL/api/v1/tasks/convert -H 'Authorization: Bearer \$TOKEN' -d '$data'"
    
    response=$(curl -s -X POST "$BASE_URL/api/v1/tasks/convert" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    show_response "$response"
    
    if [[ "$response" == *"converted_content"* ]]; then
        log_success "格式转换成功"
    else
        log_error "格式转换失败"
    fi
    
    # 停止模型
    log_info "停止模型..."
    curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/v1/models/deepseek-coder-1.3b-format/stop" > /dev/null
}

# 测试作业批改
test_homework() {
    log_info "测试作业批改接口"
    
    get_token || exit 1
    
    # 启动模型
    log_info "启动教师模型..."
    curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/v1/models/qwen2-7b-teacher/start" > /dev/null
    
    log_info "等待模型启动..."
    sleep 10
    
    local data='{"subject":"数学","question":"计算 3×4+2 的结果","answer":"14","grade_level":"小学"}'
    log_curl "curl -X POST $BASE_URL/api/v1/tasks/homework -H 'Authorization: Bearer \$TOKEN' -d '$data'"
    
    response=$(curl -s -X POST "$BASE_URL/api/v1/tasks/homework" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    show_response "$response"
    
    if [[ "$response" == *"score"* ]]; then
        log_success "作业批改成功"
    else
        log_error "作业批改失败"
    fi
    
    # 停止模型
    log_info "停止模型..."
    curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/v1/models/qwen2-7b-teacher/stop" > /dev/null
}

# 测试模型管理
test_models() {
    log_info "测试模型管理接口"
    
    get_token || exit 1
    
    log_info "1. 获取可用模型列表"
    log_curl "curl -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/"
    response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/")
    show_response "$response"
    
    log_info "2. 获取运行中的模型"
    log_curl "curl -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/running"
    response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/running")
    show_response "$response"
    
    log_info "3. 启动模型"
    log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/deepseek-coder-1.3b-format/start"
    response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/v1/models/deepseek-coder-1.3b-format/start")
    show_response "$response"
    
    sleep 5
    
    log_info "4. 再次获取运行中的模型"
    response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/running")
    show_response "$response"
    
    log_info "5. 停止模型"
    log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/deepseek-coder-1.3b-format/stop"
    response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/v1/models/deepseek-coder-1.3b-format/stop")
    show_response "$response"
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [接口类型]"
    echo ""
    echo "可用的接口类型:"
    echo "  register  - 测试用户注册接口"
    echo "  login     - 测试用户登录接口"
    echo "  convert   - 测试文件格式转换接口"
    echo "  homework  - 测试作业批改接口"
    echo "  models    - 测试模型管理接口"
    echo ""
    echo "示例:"
    echo "  $0 register"
    echo "  $0 convert"
    echo "  $0 homework"
}

# 主函数
main() {
    local api_type="$1"
    
    echo "=========================================="
    echo "    单个API接口测试脚本"
    echo "=========================================="
    echo "测试时间: $(date)"
    echo "服务器: $BASE_URL"
    echo ""
    
    case "$api_type" in
        "register")
            test_register
            ;;
        "login")
            test_login
            ;;
        "convert")
            test_convert
            ;;
        "homework")
            test_homework
            ;;
        "models")
            test_models
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
    
    echo ""
    log_success "测试完成！"
}

# 运行主函数
main "$@"
