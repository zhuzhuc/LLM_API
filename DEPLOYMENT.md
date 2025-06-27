# CPU 大模型服务平台部署指南

## 概述

这是一个基于 CPU 的开源大模型服务平台，支持多模型管理、负载均衡、服务发现、监控和水平扩展。虽然使用 CPU 性能不如 GPU，但通过合理的架构设计和优化，仍能提供稳定可靠的生产级服务。

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面      │    │   API 网关      │    │   负载均衡器    │
│   (可选)        │◄──►│   认证/限流     │◄──►│   服务发现      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   模型管理器    │◄──►│   集群管理器    │◄──►│   监控系统      │
│   多模型支持    │    │   水平扩展      │    │   日志管理      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  llama.cpp      │    │  llama.cpp      │    │  llama.cpp      │
│  模型实例 1     │    │  模型实例 2     │    │  模型实例 N     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 系统要求

### 最低配置
- **操作系统**: Ubuntu 20.04+ / macOS 10.15+ / Windows 10+
- **CPU**: 8 核心 (推荐 16+ 核心)
- **内存**: 16GB (推荐 32GB+)
- **存储**: 50GB 可用空间
- **网络**: 稳定的网络连接

### 推荐配置
- **CPU**: 16+ 核心，支持 AVX2 指令集
- **内存**: 64GB+
- **存储**: SSD 100GB+
- **网络**: 千兆网络

## 快速开始

### 1. 环境准备

```bash
# 克隆项目
git clone <repository-url>
cd llm

# 检查系统依赖
./start-llm-service.sh --check-deps
```

### 2. 模型准备

将 GGUF 格式的模型文件放入 `models/` 目录：

```bash
# 示例：下载 Qwen 模型
wget https://huggingface.co/Qwen/Qwen2-7B-Instruct-GGUF/resolve/main/qwen2-7b-instruct-q4_k_m.gguf -P models/

# 或使用 huggingface-cli
pip install huggingface_hub
huggingface-cli download Qwen/Qwen2-7B-Instruct-GGUF qwen2-7b-instruct-q4_k_m.gguf --local-dir models/
```

### 3. 配置环境

```bash
# 复制环境配置
cp .env.example .env

# 编辑配置文件
vim .env
```

### 4. 启动服务

```bash
# 一键启动
./start-llm-service.sh

# 或手动启动
cd backend && go run main.go
```

## 详细配置

### 环境变量配置 (.env)

```bash
# 数据库配置
DATABASE_URL=sqlite3://./llm.db

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production

# 服务器配置
SERVER_PORT=8080
LLAMA_CPP_PORT=8081

# llama.cpp 配置
LLAMA_CPP_PATH=./llama-cpp/build/bin/llama-server
MODELS_PATH=./models
MODEL_CONFIG_PATH=./models/model_config.json

# Token 配置
TOKEN_RATE=0.001
DEFAULT_TOKENS=1000
```

### 模型配置 (models/model_config.json)

```json
{
  "models": [
    {
      "modelName": "qwen2-7b-instruct",
      "modelFile": "qwen2-7b-instruct-q4_k_m.gguf",
      "modelPath": "models/",
      "contextLength": 32768,
      "maxTokens": 2048,
      "temperature": 0.7,
      "topP": 0.8,
      "repeatPenalty": 1.1,
      "threads": 8,
      "gpuLayers": 0,
      "active": true,
      "description": "Qwen2 7B 指令微调模型"
    }
  ]
}
```

## API 接口

### 认证接口

```bash
# 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password","email":"admin@example.com"}'

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### 模型管理

```bash
# 获取可用模型
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/models

# 启动模型
curl -X POST -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/models/qwen2-7b-instruct/start

# 与模型对话
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"message":"你好","max_tokens":200}' \
  http://localhost:8080/api/v1/models/qwen2-7b-instruct/chat
```

### OpenAI 兼容接口

```bash
# 聊天完成
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2-7b-instruct",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "max_tokens": 200
  }' \
  http://localhost:8080/api/v1/v1/chat/completions
```

### 监控接口

```bash
# 系统指标
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/monitoring/metrics

# 系统日志
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/logs

# 集群状态
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/cluster/stats
```

## 集群部署

### 主节点部署

```bash
# 启动主节点
./start-llm-service.sh

# 主节点会自动成为集群的第一个节点
```

### 工作节点部署

```bash
# 在其他机器上部署工作节点
export MASTER_HOST=192.168.1.100
export MASTER_PORT=8080
export NODE_PORT=8081

# 启动工作节点
./start-llm-service.sh --worker --master $MASTER_HOST:$MASTER_PORT --port $NODE_PORT
```

### 集群管理

```bash
# 查看集群状态
curl http://localhost:8080/api/v1/cluster/stats

# 手动添加节点
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "id": "worker-1",
    "host": "192.168.1.101",
    "port": 8081,
    "role": "worker"
  }' \
  http://localhost:8080/api/v1/cluster/join
```

## 性能优化

### CPU 优化

```bash
# 启用 OpenBLAS 加速
sudo apt install libopenblas-dev

# 重新编译 llama.cpp
cd llama-cpp
make clean
make LLAMA_OPENBLAS=1 -j$(nproc)
```

### 内存优化

```bash
# 调整模型配置
# 减少 context_length 和 threads 数量
# 使用更小的量化模型 (Q4_K_M -> Q4_0)
```

### 网络优化

```bash
# 启用 HTTP/2
# 配置负载均衡
# 使用 CDN 加速
```

## 监控和维护

### 日志管理

```bash
# 查看实时日志
tail -f logs/llm-app.log

# 查看错误日志
grep ERROR logs/*.log

# 日志轮转配置
logrotate -f /etc/logrotate.d/llm-service
```

### 性能监控

```bash
# 系统资源监控
htop
iostat -x 1
free -h

# 服务监控
curl http://localhost:8080/api/v1/monitoring/system
```

### 备份和恢复

```bash
# 备份数据库
cp backend/llm.db backup/llm-$(date +%Y%m%d).db

# 备份配置
tar -czf backup/config-$(date +%Y%m%d).tar.gz .env models/model_config.json

# 恢复
cp backup/llm-20240625.db backend/llm.db
```

## 故障排除

### 常见问题

1. **模型启动失败**
   - 检查模型文件是否存在
   - 确认内存是否充足
   - 查看 llama.cpp 日志

2. **服务无响应**
   - 检查端口是否被占用
   - 确认防火墙设置
   - 查看系统资源使用情况

3. **性能问题**
   - 调整线程数配置
   - 使用更小的模型
   - 启用 CPU 优化

### 调试命令

```bash
# 检查服务状态
./start-llm-service.sh --status

# 测试模型
./test-llama-server.sh

# 查看详细日志
./start-llm-service.sh --debug
```

## 安全配置

### JWT 安全

```bash
# 生成强密钥
openssl rand -base64 32

# 设置环境变量
export JWT_SECRET="your-generated-secret"
```

### 网络安全

```bash
# 配置防火墙
ufw allow 8080/tcp
ufw enable

# 使用 HTTPS
# 配置 SSL 证书
```

### 访问控制

```bash
# 限制 IP 访问
# 配置 API 密钥
# 设置用户权限
```

## 扩展开发

### 添加新模型

1. 将模型文件放入 `models/` 目录
2. 更新 `models/model_config.json`
3. 重启服务

### 自定义中间件

```go
// 在 backend/internal/middleware/ 中添加新中间件
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 自定义逻辑
        c.Next()
    }
}
```

### 集成第三方服务

```go
// 在 backend/internal/services/ 中添加新服务
type CustomService struct {
    // 服务配置
}
```

## 支持和社区

- **文档**: 查看项目 README.md
- **问题反馈**: 提交 GitHub Issues
- **功能请求**: 提交 Feature Request
- **贡献代码**: 提交 Pull Request

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。
