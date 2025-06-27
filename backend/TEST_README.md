# LLM Backend API 测试脚本使用说明

## 📋 概述

本项目提供了两个测试脚本来验证 LLM Backend API 的功能：

1. **`test_all_apis.sh`** - 完整测试脚本，测试所有功能
2. **`quick_test.sh`** - 快速测试脚本，测试核心功能

## 🚀 使用方法

### 前提条件

1. **确保后端服务正在运行**：
   ```bash
   cd backend
   go run main.go
   # 或者
   ./llm-server
   ```

2. **确保模型文件存在**：
   - 检查 `models/` 目录下有相应的 `.gguf` 模型文件
   - 确保 `llama.cpp` 已编译并可用

### 快速测试（推荐）

```bash
cd backend
./quick_test.sh
```

**快速测试包含**：
- ✅ 服务器状态检查
- ✅ 用户注册和登录
- ✅ 模型列表获取
- ✅ 格式转换功能（使用轻量模型）
- ✅ 作业批改功能（使用教师模型）
- ✅ 自动清理（停止所有模型）

**预计耗时**：约 2-3 分钟

### 完整测试

```bash
cd backend
./test_all_apis.sh
```

**完整测试包含**：
- ✅ 所有快速测试功能
- ✅ 测试所有可用模型
- ✅ 字幕处理功能
- ✅ 系统状态检查
- ✅ 模型状态检查
- ✅ 详细的错误报告

**预计耗时**：约 10-15 分钟

## 📊 测试结果解读

### 成功输出示例

```bash
==========================================
       LLM Backend 快速测试
==========================================
[INFO] 1. 检查服务器状态...
[SUCCESS] 服务器运行正常
[INFO] 2. 注册测试用户...
[SUCCESS] 用户注册成功
[INFO] 3. 获取可用模型...
[SUCCESS] 获取模型列表成功
[INFO] 4. 启动格式转换模型...
[SUCCESS] 模型启动成功
[INFO] 测试格式转换...
[SUCCESS] 格式转换测试成功
[SUCCESS] 快速测试完成！
==========================================
```

### 错误处理

如果测试失败，脚本会显示详细的错误信息：

```bash
[ERROR] 服务器未运行
[ERROR] 用户注册失败: {"error":"用户名已存在"}
[ERROR] 模型启动失败: {"error":"模型文件不存在"}
```

## 🔧 测试的功能

### 认证功能
- [x] 用户注册
- [x] 用户登录
- [x] 获取用户信息
- [x] Token 验证

### 模型管理
- [x] 获取可用模型列表
- [x] 获取运行中的模型
- [x] 启动模型
- [x] 停止模型
- [x] 模型状态查询
- [x] 模型对话测试

### 任务处理
- [x] 文件格式转换（JSON ↔ YAML ↔ XML）
- [x] 作业批改（数学、语文等）
- [x] 字幕处理（提取、翻译）
- [x] 支持格式查询

### 系统功能
- [x] 健康检查
- [x] 系统状态
- [x] 错误处理

## 🛠️ 自定义测试

### 修改测试模型

编辑脚本中的模型列表：

```bash
models=(
    "deepseek-coder-1.3b-format"    # 格式转换专用
    "qwen2-7b-teacher"              # 作业批改专用
    "qwen2-7b-instruct"             # 通用对话
    "yi-9b-chat"                    # 通用对话
    "mistral-7b-instruct"           # 通用对话
)
```

### 添加新的测试用例

在脚本中添加新的测试函数：

```bash
test_new_feature() {
    test_start "测试新功能"
    
    local data='{"param1":"value1","param2":"value2"}'
    response=$(make_request "POST" "/api/v1/new/endpoint" "$data" "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"success"* ]]; then
        log_success "新功能测试成功"
        return 0
    else
        log_error "新功能测试失败: $response"
        return 1
    fi
}
```

## 🚨 注意事项

1. **资源管理**：
   - 脚本会自动停止启动的模型
   - 如果脚本被中断（Ctrl+C），会自动清理

2. **模型文件**：
   - 确保模型文件存在且可访问
   - 模型启动需要时间，脚本会自动等待

3. **端口冲突**：
   - 确保 8080 端口可用
   - 确保模型端口（8081-8085）可用

4. **内存使用**：
   - 大模型需要足够的内存
   - 建议一次只运行一个模型

## 📝 日志和调试

### 查看详细日志

```bash
# 运行测试并保存日志
./test_all_apis.sh 2>&1 | tee test_results.log

# 查看后端服务日志
tail -f ../logs/app.log
```

### 手动测试单个API

```bash
# 获取token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"123456","password":"123456"}' | \
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 测试格式转换
curl -X POST http://localhost:8080/api/v1/tasks/convert \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"source_format":"json","target_format":"yaml","content":"{\"test\":123}"}'
```

## 🎯 故障排除

### 常见问题

1. **服务器连接失败**
   - 检查后端服务是否运行
   - 检查端口 8080 是否可用

2. **模型启动失败**
   - 检查模型文件是否存在
   - 检查内存是否足够
   - 检查 llama.cpp 是否正确编译

3. **Token 过期**
   - 脚本会自动处理 token 刷新
   - 如果仍有问题，重新运行脚本

4. **权限问题**
   - 确保脚本有执行权限：`chmod +x *.sh`
   - 确保模型文件有读取权限

### 获取帮助

如果遇到问题，请：

1. 查看测试日志
2. 检查后端服务日志
3. 确认所有依赖都已安装
4. 验证配置文件正确

---

**测试愉快！** 🎉
