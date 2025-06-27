#!/bin/bash

# LLM Backend API curl 测试脚本
# 使用curl命令测试所有API接口

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
TEST_USERNAME="curltest_$(date +%s)"
TEST_PASSWORD="testpass123"
TEST_EMAIL="${TEST_USERNAME}@example.com"
TOKEN=""

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_curl() { echo -e "${CYAN}[CURL]${NC} $1"; }

# 显示响应
show_response() {
    local response="$1"
    local title="$2"
    echo -e "${YELLOW}=== $title ===${NC}"
    echo "$response" | jq . 2>/dev/null || echo "$response"
    echo ""
}

# 等待函数
wait_seconds() {
    local seconds=$1
    local message="$2"
    log_info "$message"
    for ((i=seconds; i>0; i--)); do
        printf "\r等待中... %d 秒" $i
        sleep 1
    done
    printf "\r等待完成!     \n"
}

echo "=========================================="
echo "    LLM Backend API curl 测试脚本"
echo "=========================================="
echo "测试时间: $(date)"
echo "服务器: $BASE_URL"
echo ""

# 1. 检查服务器状态
log_info "1. 检查服务器状态"
log_curl "curl -s $BASE_URL/health"
response=$(curl -s "$BASE_URL/health" || echo '{"error":"连接失败"}')
show_response "$response" "服务器状态"

if [[ "$response" == *"running_models"* ]]; then
    log_success "服务器运行正常"
else
    log_error "服务器未运行或无响应"
    exit 1
fi

# 2. 用户注册
log_info "2. 用户注册"
register_data="{\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\",\"email\":\"$TEST_EMAIL\"}"
log_curl "curl -X POST $BASE_URL/api/v1/auth/register -d '$register_data'"
response=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "$register_data")
show_response "$response" "用户注册"

if [[ "$response" == *"token"* ]]; then
    TOKEN=$(echo "$response" | jq -r '.token' 2>/dev/null || echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    log_success "用户注册成功，获得token: ${TOKEN:0:20}..."
else
    log_error "用户注册失败"
    exit 1
fi

# 3. 用户登录
log_info "3. 用户登录"
login_data="{\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}"
log_curl "curl -X POST $BASE_URL/api/v1/auth/login -d '$login_data'"
response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "$login_data")
show_response "$response" "用户登录"

if [[ "$response" == *"token"* ]]; then
    TOKEN=$(echo "$response" | jq -r '.token' 2>/dev/null || echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    log_success "用户登录成功，更新token: ${TOKEN:0:20}..."
else
    log_error "用户登录失败"
    exit 1
fi

# 4. 获取用户信息
log_info "4. 获取用户信息"
log_curl "curl -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/auth/profile"
response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/auth/profile")
show_response "$response" "用户信息"

if [[ "$response" == *"username"* ]]; then
    log_success "获取用户信息成功"
else
    log_error "获取用户信息失败"
fi

# 5. 获取可用模型列表
log_info "5. 获取可用模型列表"
log_curl "curl -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/"
response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/")
show_response "$response" "可用模型列表"

if [[ "$response" == *"models"* ]]; then
    log_success "获取模型列表成功"
else
    log_error "获取模型列表失败"
fi

# 6. 获取运行中的模型
log_info "6. 获取运行中的模型"
log_curl "curl -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/running"
response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/running")
show_response "$response" "运行中的模型"

# 7. 获取支持的格式列表
log_info "7. 获取支持的格式列表"
log_curl "curl -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/tasks/formats"
response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/tasks/formats")
show_response "$response" "支持的格式列表"

if [[ "$response" == *"supported_formats"* ]]; then
    log_success "获取格式列表成功"
else
    log_error "获取格式列表失败"
fi

echo ""
log_info "开始测试模型功能..."
echo ""

# 8. 启动格式转换模型并测试
log_info "8. 启动格式转换模型"
model_name="deepseek-coder-1.3b-format"
log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/$model_name/start"
response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/$model_name/start")
show_response "$response" "启动格式转换模型"

if [[ "$response" == *"success"* ]] || [[ "$response" == *"启动成功"* ]]; then
    log_success "格式转换模型启动成功"
    
    wait_seconds 8 "等待模型完全启动..."
    
    # 测试格式转换
    log_info "测试文件格式转换"
    convert_data='{"source_format":"json","target_format":"yaml","content":"{\"name\":\"curl测试\",\"value\":123,\"items\":[\"a\",\"b\",\"c\"]}"}'
    log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/tasks/convert -d '$convert_data'"
    response=$(curl -s -X POST "$BASE_URL/api/v1/tasks/convert" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$convert_data")
    show_response "$response" "格式转换结果"
    
    if [[ "$response" == *"converted_content"* ]]; then
        log_success "格式转换测试成功"
    else
        log_error "格式转换测试失败"
    fi
    
    # 停止模型
    log_info "停止格式转换模型"
    response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/$model_name/stop")
    show_response "$response" "停止模型"
    wait_seconds 3 "等待模型停止..."
else
    log_error "格式转换模型启动失败"
fi

# 9. 启动教师模型并测试
log_info "9. 启动教师模型"
model_name="qwen2-7b-teacher"
log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/$model_name/start"
response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/$model_name/start")
show_response "$response" "启动教师模型"

if [[ "$response" == *"success"* ]] || [[ "$response" == *"启动成功"* ]]; then
    log_success "教师模型启动成功"
    
    wait_seconds 10 "等待模型完全启动..."
    
    # 测试作业批改
    log_info "测试作业批改"
    homework_data='{"subject":"数学","question":"计算 (2+3)×4-5 的结果","answer":"15","grade_level":"小学"}'
    log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/tasks/homework -d '$homework_data'"
    response=$(curl -s -X POST "$BASE_URL/api/v1/tasks/homework" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$homework_data")
    show_response "$response" "作业批改结果"
    
    if [[ "$response" == *"score"* ]]; then
        log_success "作业批改测试成功"
    else
        log_error "作业批改测试失败"
    fi
    
    # 测试模型对话
    log_info "测试模型对话"
    chat_data='{"message":"你好，请简单介绍一下你自己","max_tokens":100}'
    log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/models/$model_name/chat -d '$chat_data'"
    response=$(curl -s -X POST "$BASE_URL/api/v1/models/$model_name/chat" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$chat_data")
    show_response "$response" "模型对话结果"
    
    if [[ "$response" == *"response"* ]]; then
        log_success "模型对话测试成功"
    else
        log_error "模型对话测试失败"
    fi
    
    # 停止模型
    log_info "停止教师模型"
    response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/models/$model_name/stop")
    show_response "$response" "停止模型"
    wait_seconds 3 "等待模型停止..."
else
    log_error "教师模型启动失败"
fi

# 10. 测试字幕处理（可能失败，因为需要实际视频文件）
log_info "10. 测试字幕处理"
subtitle_data='{"video_path":"/test/sample.mp4","source_lang":"英文","target_lang":"中文","output_format":"srt"}'
log_curl "curl -X POST -H 'Authorization: Bearer \$TOKEN' $BASE_URL/api/v1/tasks/subtitle -d '$subtitle_data'"
response=$(curl -s -X POST "$BASE_URL/api/v1/tasks/subtitle" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$subtitle_data")
show_response "$response" "字幕处理结果"

if [[ "$response" == *"success"* ]]; then
    log_success "字幕处理测试成功"
else
    log_warning "字幕处理测试失败（可能需要实际视频文件）"
fi

# 11. 获取最终用户信息（查看token变化）
log_info "11. 获取最终用户信息（查看token消耗）"
response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/auth/profile")
show_response "$response" "最终用户信息"

echo ""
echo "=========================================="
echo "           curl 测试完成"
echo "=========================================="
echo "结束时间: $(date)"
echo ""
log_success "所有curl测试已完成！"
echo ""
echo "💡 提示："
echo "- 如果某些测试失败，请检查后端服务是否正常运行"
echo "- 模型启动需要时间，请耐心等待"
echo "- token会在每次使用后减少，这是正常现象"
echo "- 可以重复运行此脚本进行测试"
