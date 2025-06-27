# LLM Backend API curl 测试指南

## 📋 概述

本文档提供了使用 curl 命令测试 LLM Backend API 的完整指南，包括所有接口的详细用法和示例。

## 🚀 快速开始

### 运行自动化测试脚本

```bash
cd backend
chmod +x curl_test.sh
./curl_test.sh
```

### 手动测试步骤

## 1. 基础配置

```bash
# 设置基础URL
BASE_URL="http://localhost:8080"

# 测试用户信息
USERNAME="testuser"
PASSWORD="testpass123"
EMAIL="testuser@example.com"
```

## 2. 认证相关接口

### 2.1 用户注册

```bash
curl -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "'$USERNAME'",
    "password": "'$PASSWORD'",
    "email": "'$EMAIL'"
  }'
```

**预期响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "testuser@example.com",
    "tokens": 1000
  }
}
```

### 2.2 用户登录

```bash
curl -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "'$USERNAME'",
    "password": "'$PASSWORD'"
  }'
```

### 2.3 获取用户信息

```bash
# 先获取token
TOKEN=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"'$USERNAME'","password":"'$PASSWORD'"}' | \
  jq -r '.token')

# 获取用户信息
curl -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/auth/profile
```

## 3. 模型管理接口

### 3.1 获取可用模型列表

```bash
curl -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/models/
```

**预期响应**:
```json
{
  "models": [
    {
      "name": "deepseek-coder-1.3b-format",
      "description": "格式转换专用模型",
      "status": "available"
    },
    {
      "name": "qwen2-7b-teacher",
      "description": "作业批改专用模型",
      "status": "available"
    }
  ]
}
```

### 3.2 获取运行中的模型

```bash
curl -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/models/running
```

### 3.3 启动模型

```bash
# 启动格式转换模型
curl -X POST -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/models/deepseek-coder-1.3b-format/start

# 启动教师模型
curl -X POST -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/models/qwen2-7b-teacher/start
```

### 3.4 停止模型

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/models/deepseek-coder-1.3b-format/stop
```

### 3.5 模型对话

```bash
curl -X POST $BASE_URL/api/v1/models/qwen2-7b-teacher/chat \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，请介绍一下你自己",
    "max_tokens": 100
  }'
```

## 4. 任务处理接口

### 4.1 获取支持的格式列表

```bash
curl -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/tasks/formats
```

**预期响应**:
```json
{
  "supported_formats": [
    "json", "yaml", "xml", "csv", "toml"
  ]
}
```

### 4.2 文件格式转换

```bash
# JSON 转 YAML
curl -X POST $BASE_URL/api/v1/tasks/convert \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_format": "json",
    "target_format": "yaml",
    "content": "{\"name\":\"测试\",\"value\":123,\"items\":[\"a\",\"b\",\"c\"]}"
  }'
```

**预期响应**:
```json
{
  "success": true,
  "converted_content": "name: 测试\nvalue: 123\nitems:\n- a\n- b\n- c",
  "tokens_consumed": 15,
  "remaining_tokens": 985
}
```

### 4.3 作业批改

```bash
curl -X POST $BASE_URL/api/v1/tasks/homework \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "数学",
    "question": "计算 (2+3)×4-5 的结果",
    "answer": "15",
    "grade_level": "小学"
  }'
```

**预期响应**:
```json
{
  "success": true,
  "score": 100,
  "feedback": "答案正确！计算步骤：(2+3)=5, 5×4=20, 20-5=15",
  "suggestions": "很好地掌握了运算顺序",
  "tokens_consumed": 25,
  "remaining_tokens": 960
}
```

### 4.4 字幕处理

```bash
curl -X POST $BASE_URL/api/v1/tasks/subtitle \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "video_path": "/path/to/video.mp4",
    "source_lang": "英文",
    "target_lang": "中文",
    "output_format": "srt"
  }'
```

## 5. 系统状态接口

### 5.1 健康检查

```bash
curl $BASE_URL/health
```

**预期响应**:
```json
{
  "status": "healthy",
  "timestamp": "2025-06-26T13:00:00Z",
  "running_models": ["qwen2-7b-teacher"]
}
```

## 6. 完整测试流程示例

### 6.1 完整的格式转换测试

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"
USERNAME="testuser_$(date +%s)"
PASSWORD="testpass123"
EMAIL="${USERNAME}@example.com"

echo "1. 注册用户..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"email\":\"$EMAIL\"}")

TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.token')
echo "获得Token: ${TOKEN:0:20}..."

echo "2. 启动格式转换模型..."
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/models/deepseek-coder-1.3b-format/start

echo "3. 等待模型启动..."
sleep 8

echo "4. 测试格式转换..."
CONVERT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/tasks/convert \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_format": "json",
    "target_format": "yaml",
    "content": "{\"name\":\"测试\",\"value\":123}"
  }')

echo "转换结果:"
echo $CONVERT_RESPONSE | jq .

echo "5. 停止模型..."
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  $BASE_URL/api/v1/models/deepseek-coder-1.3b-format/stop

echo "测试完成！"
```

## 7. 错误处理

### 7.1 常见错误响应

**Token不足**:
```json
{
  "error": "token余额不足，当前余额: 10，需要: 25",
  "current_tokens": 10,
  "required_tokens": 25
}
```

**未认证**:
```json
{
  "error": "未认证"
}
```

**模型未运行**:
```json
{
  "error": "模型 deepseek-coder-1.3b-format 未运行"
}
```

### 7.2 调试技巧

1. **查看详细响应**:
   ```bash
   curl -v $BASE_URL/api/v1/auth/login ...
   ```

2. **保存响应到文件**:
   ```bash
   curl ... > response.json
   cat response.json | jq .
   ```

3. **检查HTTP状态码**:
   ```bash
   curl -w "%{http_code}" -s -o /dev/null $BASE_URL/health
   ```

## 8. 性能测试

### 8.1 并发测试

```bash
# 并发发送10个请求
for i in {1..10}; do
  curl -X POST $BASE_URL/api/v1/tasks/convert \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"source_format":"json","target_format":"yaml","content":"{}"}' &
done
wait
```

### 8.2 压力测试

```bash
# 使用 ab 工具进行压力测试
ab -n 100 -c 10 -H "Authorization: Bearer $TOKEN" \
  -p convert_data.json -T application/json \
  $BASE_URL/api/v1/tasks/convert
```

## 9. 自动化脚本

项目提供了以下自动化测试脚本：

- `curl_test.sh` - 完整的curl测试脚本
- `quick_test.sh` - 快速测试脚本  
- `test_all_apis.sh` - 全面的API测试脚本

运行方式：
```bash
chmod +x *.sh
./curl_test.sh
```

---

**Happy Testing!** 🎉
