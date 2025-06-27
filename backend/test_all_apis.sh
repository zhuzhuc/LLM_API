#!/bin/bash

# LLM Backend API 完整测试脚本
# 测试注册、登录和所有API功能
# 每个模型测试完成后会自动关闭，然后启动下一个

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BASE_URL="http://localhost:8080"
TEST_USERNAME="testuser_$(date +%s)"
TEST_PASSWORD="testpass123"
TEST_EMAIL="${TEST_USERNAME}@example.com"

# 全局变量
TOKEN=""
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 测试计数
test_start() {
    ((TOTAL_TESTS++))
    log_info "测试 $TOTAL_TESTS: $1"
}

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

# 等待函数
wait_with_progress() {
    local seconds=$1
    local message=$2
    log_info "$message"
    for ((i=seconds; i>0; i--)); do
        printf "\r等待中... %d 秒" $i
        sleep 1
    done
    printf "\r等待完成!     \n"
}

# 检查服务器状态
check_server() {
    test_start "检查服务器状态"
    
    response=$(curl -s "$BASE_URL/health" || echo "ERROR")
    if [[ "$response" == *"running_models"* ]]; then
        log_success "服务器运行正常"
        return 0
    else
        log_error "服务器未运行或无响应"
        return 1
    fi
}

# 测试用户注册
test_register() {
    test_start "用户注册"
    
    local data="{\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\",\"email\":\"$TEST_EMAIL\"}"
    response=$(make_request "POST" "/api/v1/auth/register" "$data")
    
    if [[ "$response" == *"token"* ]]; then
        TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        log_success "用户注册成功，获得token: ${TOKEN:0:20}..."
        return 0
    else
        log_error "用户注册失败: $response"
        return 1
    fi
}

# 测试用户登录
test_login() {
    test_start "用户登录"
    
    local data="{\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}"
    response=$(make_request "POST" "/api/v1/auth/login" "$data")
    
    if [[ "$response" == *"token"* ]]; then
        TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        log_success "用户登录成功，获得新token: ${TOKEN:0:20}..."
        return 0
    else
        log_error "用户登录失败: $response"
        return 1
    fi
}

# 测试获取用户信息
test_profile() {
    test_start "获取用户信息"
    
    response=$(make_request "GET" "/api/v1/auth/profile" "" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"username"* ]]; then
        log_success "获取用户信息成功"
        return 0
    else
        log_error "获取用户信息失败: $response"
        return 1
    fi
}

# 测试获取可用模型列表
test_available_models() {
    test_start "获取可用模型列表"
    
    response=$(make_request "GET" "/api/v1/models/" "" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"models"* ]]; then
        log_success "获取可用模型列表成功"
        echo "$response" | jq '.models[].name' 2>/dev/null || echo "模型列表: $response"
        return 0
    else
        log_error "获取可用模型列表失败: $response"
        return 1
    fi
}

# 测试获取运行中的模型
test_running_models() {
    test_start "获取运行中的模型"
    
    response=$(make_request "GET" "/api/v1/models/running" "" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"models"* ]] || [[ "$response" == "[]" ]]; then
        log_success "获取运行中的模型成功"
        return 0
    else
        log_error "获取运行中的模型失败: $response"
        return 1
    fi
}

# 启动模型
start_model() {
    local model_name=$1
    test_start "启动模型: $model_name"
    
    response=$(make_request "POST" "/api/v1/models/$model_name/start" "" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"success"* ]] || [[ "$response" == *"启动成功"* ]]; then
        log_success "模型 $model_name 启动成功"
        return 0
    else
        log_error "模型 $model_name 启动失败: $response"
        return 1
    fi
}

# 停止模型
stop_model() {
    local model_name=$1
    test_start "停止模型: $model_name"
    
    response=$(make_request "POST" "/api/v1/models/$model_name/stop" "" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"success"* ]] || [[ "$response" == *"停止成功"* ]]; then
        log_success "模型 $model_name 停止成功"
        return 0
    else
        log_warning "模型 $model_name 停止失败或已停止: $response"
        return 0  # 不算作错误，可能模型本来就没运行
    fi
}

# 测试模型对话
test_model_chat() {
    local model_name=$1
    local message=$2
    test_start "测试模型对话: $model_name"
    
    local data="{\"message\":\"$message\",\"max_tokens\":100}"
    response=$(make_request "POST" "/api/v1/models/$model_name/chat" "$data" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"response"* ]] || [[ "$response" == *"content"* ]]; then
        log_success "模型 $model_name 对话测试成功"
        echo "回复: $(echo "$response" | head -c 100)..."
        return 0
    else
        log_error "模型 $model_name 对话测试失败: $response"
        return 1
    fi
}

# 测试支持的格式列表
test_supported_formats() {
    test_start "获取支持的格式列表"
    
    response=$(make_request "GET" "/api/v1/tasks/formats" "" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"supported_formats"* ]]; then
        log_success "获取支持的格式列表成功"
        return 0
    else
        log_error "获取支持的格式列表失败: $response"
        return 1
    fi
}

# 测试文件格式转换
test_format_conversion() {
    test_start "测试文件格式转换"
    
    local data='{"source_format":"json","target_format":"yaml","content":"{\"name\":\"测试\",\"value\":123,\"active\":true}"}'
    response=$(make_request "POST" "/api/v1/tasks/convert" "$data" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"converted_content"* ]] && [[ "$response" == *"success"* ]]; then
        log_success "文件格式转换测试成功"
        echo "转换结果: $(echo "$response" | grep -o '"converted_content":"[^"]*"' | cut -d'"' -f4 | head -c 50)..."
        return 0
    else
        log_error "文件格式转换测试失败: $response"
        return 1
    fi
}

# 测试作业批改
test_homework_grading() {
    test_start "测试作业批改"
    
    local data='{"subject":"数学","question":"计算 2+3×4 的结果","answer":"14","grade_level":"小学"}'
    response=$(make_request "POST" "/api/v1/tasks/homework" "$data" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"score"* ]] && [[ "$response" == *"success"* ]]; then
        log_success "作业批改测试成功"
        echo "批改结果: $(echo "$response" | head -c 100)..."
        return 0
    else
        log_error "作业批改测试失败: $response"
        return 1
    fi
}

# 测试字幕处理
test_subtitle_processing() {
    test_start "测试字幕处理"

    local data='{"video_path":"/test/sample.mp4","source_lang":"英文","target_lang":"中文","output_format":"srt"}'
    response=$(make_request "POST" "/api/v1/tasks/subtitle" "$data" "Authorization: Bearer $TOKEN")

    if [[ "$response" == *"success"* ]] || [[ "$response" == *"subtitle_path"* ]]; then
        log_success "字幕处理测试成功"
        return 0
    else
        log_warning "字幕处理测试失败（可能需要实际视频文件）: $response"
        return 0  # 不算作错误，因为需要实际视频文件
    fi
}

# 测试系统状态
test_system_status() {
    test_start "获取系统状态"

    response=$(make_request "GET" "/status" "" "Authorization: Bearer $TOKEN")

    if [[ "$response" == *"status"* ]] || [[ "$response" == *"version"* ]]; then
        log_success "获取系统状态成功"
        return 0
    else
        log_warning "获取系统状态失败: $response"
        return 0
    fi
}

# 测试模型状态
test_model_status() {
    local model_name=$1
    test_start "获取模型状态: $model_name"

    response=$(make_request "GET" "/api/v1/models/$model_name/status" "" "Authorization: Bearer $TOKEN")

    if [[ "$response" == *"status"* ]]; then
        log_success "获取模型状态成功"
        return 0
    else
        log_warning "获取模型状态失败: $response"
        return 0
    fi
}

# 清理函数 - 确保所有模型都停止
cleanup() {
    log_info "清理环境，停止所有运行中的模型..."

    models=(
        "deepseek-coder-1.3b-format"
        "qwen2-7b-teacher"
        "qwen2-7b-instruct"
        "yi-9b-chat"
        "mistral-7b-instruct"
    )

    for model in "${models[@]}"; do
        stop_model "$model" >/dev/null 2>&1 || true
    done

    log_info "清理完成"
}

# 信号处理 - Ctrl+C 时清理
trap cleanup EXIT INT TERM

# 主测试函数
main() {
    echo "=========================================="
    echo "    LLM Backend API 完整测试脚本"
    echo "=========================================="
    echo "开始时间: $(date)"
    echo ""
    
    # 基础测试
    check_server || exit 1
    test_register || exit 1
    test_login || exit 1
    test_profile || exit 1
    test_available_models || exit 1
    test_running_models || exit 1
    test_system_status || true
    test_supported_formats || exit 1
    
    echo ""
    log_info "开始模型功能测试..."
    echo ""
    
    # 定义要测试的模型列表
    models=(
        "deepseek-coder-1.3b-format"
        "qwen2-7b-teacher"
        "qwen2-7b-instruct"
        "yi-9b-chat"
        "mistral-7b-instruct"
    )
    
    # 测试每个模型
    for model in "${models[@]}"; do
        echo "----------------------------------------"
        log_info "测试模型: $model"
        echo "----------------------------------------"
        
        # 启动模型
        if start_model "$model"; then
            wait_with_progress 8 "等待模型完全启动..."

            # 测试模型状态
            test_model_status "$model"

            # 测试模型对话
            test_model_chat "$model" "你好，请简单介绍一下你自己"

            # 根据模型类型进行专门测试
            if [[ "$model" == *"format"* ]]; then
                log_info "测试格式转换专用功能..."
                test_format_conversion
            elif [[ "$model" == *"teacher"* ]]; then
                log_info "测试作业批改专用功能..."
                test_homework_grading
                test_subtitle_processing
            else
                log_info "测试通用对话功能..."
                test_model_chat "$model" "请用中文回答：什么是人工智能？"
            fi

            # 停止模型
            stop_model "$model"
            wait_with_progress 3 "等待模型完全停止..."
        fi
        
        echo ""
    done
    
    # 测试总结
    echo "=========================================="
    echo "           测试完成总结"
    echo "=========================================="
    echo "总测试数: $TOTAL_TESTS"
    echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"
    echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    echo "结束时间: $(date)"
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "所有测试通过！🎉"
        exit 0
    else
        log_error "有 $FAILED_TESTS 个测试失败"
        exit 1
    fi
}

# 运行主函数
main "$@"
