# 轻量级大语言模型 API 服务

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.0+-green.svg)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

一个基于 Go + Vue 3 的轻量级大语言模型 API 服务，集成多个开源 LLM 模型，提供文件格式转换、作业批改、字幕处理等 AI 功能。

## ✨ 核心特性

- **多模型支持** - 集成 Qwen、DeepSeek、Yi、Mistral 等多个轻量级模型
- **动态模型管理** - 支持模型的动态启动、停止和切换，节省系统资源
- **专用任务处理** - 文件格式转换、作业批改、字幕处理等专门优化的功能
- **简单认证系统** - JWT 认证、Token 管理、用户权限控制
- **高性能架构** - Go 高性能 API 服务 + llama.cpp 推理引擎
- **现代化前端** - Vue 3 + Element Plus + Vite 响应式 Web 界面
- **监控与日志** - 完整的请求追踪和性能监控
- **完善测试** - 提供多种测试脚本和工具

## 🏗️ 项目架构

### 技术栈
- **后端 API 服务**: Go 1.21+ (Gin 框架)
- **前端界面**: Vue 3 + Element Plus + Vite
- **模型推理**: llama.cpp (GGUF 格式)
- **认证系统**: JWT
- **数据库**: SQLite

### 服务架构
```
Frontend (Vue 3) → Go API Server → llama.cpp
       :5173            :8080        :8081-8085
                           ↓
                      SQLite DB
```

## 📋 系统要求

### 最低配置
- **CPU**: 8 核心以上
- **内存**: 16GB RAM
- **存储**: 20GB 可用空间
- **操作系统**: macOS 10.15+ / Ubuntu 20.04+ / CentOS 8+
- **软件依赖**: Go 1.21+, Node.js 16+

### 推荐配置
- **CPU**: 16 核心以上 (Intel/AMD/Apple Silicon)
- **内存**: 32GB RAM
- **存储**: 50GB SSD
- **网络**: 千兆网络

## 🚀 快速开始

### 1. 环境准备

```bash
# 克隆项目
git clone <your-repo-url>
cd LLM_API

# 安装系统依赖 (Ubuntu/Debian)
sudo apt update
sudo apt install -y build-essential git wget curl nodejs npm

# 安装系统依赖 (macOS)
brew install git wget curl node
```

### 2. 设置 llama.cpp

```bash
# 运行自动化设置脚本
chmod +x setup-llama-cpp.sh
./setup-llama-cpp.sh
```

### 3. 下载模型

```bash
# 手动下载模型到 models/ 目录
# 项目已包含多个预下载的模型文件
ls models/
```

### 4. 构建和启动服务

#### 启动后端服务 (Go)
```bash
cd backend
go mod tidy
go run main.go
# 服务运行在 http://localhost:8080
```

#### 启动前端服务 (Vue 3)
```bash
cd frontend
npm install
npm run dev
# 前端运行在 http://localhost:5173
```

### 5. 测试部署

```bash
# 测试 Go API 服务
cd backend
./test_all_apis.sh

# 测试单个 API
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好，请介绍一下你自己", "model": "qwen2-7b-instruct-q4_k_m"}'
```

## 🔧 配置说明

### 模型配置

项目支持多种 CPU 优化模型：

| 模型 | 参数规模 | 中文能力 | 内存需求 | 推荐用途 |
|------|----------|----------|----------|----------|
| Qwen2-7B-Instruct | 7B | ⭐⭐⭐⭐⭐ | ~8GB | 中文对话、文档处理 |
| Yi-9B-Chat | 9B | ⭐⭐⭐⭐⭐ | ~10GB | 多任务、创作 |
| DeepSeek-Coder-6.7B | 6.7B | ⭐⭐⭐⭐ | ~7GB | 代码生成、编程 |
| Mistral-7B-Instruct | 7B | ⭐⭐⭐ | ~8GB | 英文任务、推理 |

### 性能调优

#### Go 服务配置
```bash
# 设置 Go 运行时参数
export GOMAXPROCS=12  # 建议设置为 CPU 核心数
export GOGC=100       # GC 触发百分比
```



#### llama.cpp 线程配置
```bash
# 在启动 llama-server 时设置线程数
./llama-cpp/llama-server -m models/qwen2-7b-instruct-q4_k_m.gguf -t 12
```

## 📊 性能基准

### 测试环境
- **CPU**: Intel i7-12700K (12核20线程)
- **内存**: 32GB DDR4-3200
- **模型**: Qwen2-7B-Instruct Q4_K_M

### 性能指标
| 任务类型 | 输入长度 | 输出长度 | 响应时间 | 吞吐量 |
|----------|----------|----------|----------|--------|
| 简单问答 | 50 tokens | 100 tokens | 4-6秒 | ~20 tokens/s |
| 文档摘要 | 500 tokens | 200 tokens | 12-15秒 | ~15 tokens/s |
| 代码生成 | 100 tokens | 300 tokens | 18-25秒 | ~12 tokens/s |

## 🔌 API 接口

### Go API 服务 (端口 8080)

#### 聊天接口
```bash
POST /api/chat
Content-Type: application/json

{
  "message": "你好，请帮我写一个 Python 函数",
  "model": "qwen2-7b-instruct-q4_k_m",
  "temperature": 0.7,
  "max_tokens": 500
}
```

#### 模型管理
```bash
# 获取可用模型
GET /api/models

# 获取模型状态
GET /api/models/status

# 切换模型
POST /api/models/switch
{
  "model": "qwen2-7b-instruct-q4_k_m"
}
```

#### 文件处理
```bash
# 文件上传和处理
POST /api/upload
Content-Type: multipart/form-data

# 文件格式转换
POST /api/convert
```

#### 认证接口
```bash
# 用户注册
POST /api/auth/register

# 用户登录
POST /api/auth/login

# 获取用户信息
GET /api/auth/profile
Authorization: Bearer <token>
```

## 🔐 安全配置

### Keycloak 集成

1. **安装 Keycloak**
```bash
# Docker 方式
docker run -p 8080:8080 -e KEYCLOAK_ADMIN=admin -e KEYCLOAK_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:latest start-dev
```

2. **配置 Realm 和 Client**
- 创建 Realm: `llm-realm`
- 创建 Client: `llm-app`
- 配置用户角色: `USER`, `ADMIN`

3. **更新配置**
```yaml
keycloak:
  auth-server-url: http://localhost:8080/auth
  realm: llm-realm
  resource: llm-app
```

## 🐳 Docker 部署

项目提供了完整的 Docker 部署方案，包含所有服务组件。

### 使用 Docker Compose (推荐)
```bash
# 一键启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 服务端口映射
- **前端服务**: http://localhost:5173
- **Go API 服务**: http://localhost:8080
- **llama.cpp 服务**: http://localhost:8081-8085

### 自定义 Dockerfile

#### Go 服务 Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/ .
RUN go mod tidy && go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

#### 前端 Dockerfile
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
```

## 🔧 故障排除

### 常见问题

1. **Go 服务启动失败**
```bash
# 检查端口占用
lsof -i :8080
# 检查 Go 版本
go version
# 重新编译
cd backend && go build -o main .
```

2. **内存不足**
```bash
# 检查内存使用
free -h
# 使用更小的模型或减少并发数
```

3. **模型加载失败**
```bash
# 检查模型文件
ls -la models/
# 验证文件完整性
file models/*.gguf
# 检查 llama-server 进程
ps aux | grep llama-server
```

4. **前端构建失败**
```bash
# 清理依赖重新安装
cd frontend
rm -rf node_modules package-lock.json
npm install
```

5. **llama.cpp 编译失败**
```bash
# 清理重新编译
cd llama-cpp
make clean
make -j$(nproc)
```

### 日志分析
```bash
# 查看 Go 服务日志
tail -f logs/http.log

# 查看 llama-server 日志
tail -f logs/llama-server.log
```

## 📈 监控和运维

### 健康检查
```bash
# Go API 服务健康检查
curl http://localhost:8080/api/health

# 前端服务检查
curl http://localhost:5173
```

### 性能监控
```bash
# 检查所有服务进程
ps aux | grep -E "(main|java|node)"

# Go 服务内存使用
ps aux | grep main

# 模型推理统计
curl http://localhost:8080/api/models/status
```

### 服务管理脚本
```bash
# 启动所有服务
./start-llm-service.sh

# 停止所有服务
./stop-llm-service.sh

# 重启服务
./stop-llm-service.sh && ./start-llm-service.sh
```

## 🚀 生产部署建议

1. **硬件配置**
   - 使用 SSD 存储模型文件
   - 配置足够的 RAM (模型大小 × 1.5)
   - 选择高频 CPU (推荐 16+ 核心)

2. **系统优化**
   - 关闭不必要的系统服务
   - 设置 CPU 性能模式
   - 优化网络配置和文件描述符限制

3. **应用配置**
   - 使用生产级数据库 (MySQL/PostgreSQL)
   - 配置 Nginx 反向代理和负载均衡
   - 设置日志轮转和监控告警
   - 配置 HTTPS 和 SSL 证书

4. **服务部署**
   - 使用 Docker Compose 或 Kubernetes
   - 配置服务自动重启和健康检查
   - 设置资源限制和环境变量

5. **安全加固**
   - 启用 JWT 认证和授权
   - 配置防火墙和网络安全组
   - 定期更新依赖和安全补丁
   - 设置 API 限流和访问控制

## 📚 参考资源

### 核心技术文档
- [Go 官方文档](https://golang.org/doc/)
- [Gin Web 框架](https://gin-gonic.com/docs/)
- [Vue 3 官方文档](https://vuejs.org/guide/)
- [Element Plus 组件库](https://element-plus.org/)
- [Spring Boot 官方文档](https://spring.io/projects/spring-boot)

### AI 模型相关
- [llama.cpp 官方文档](https://github.com/ggerganov/llama.cpp)
- [Qwen2 模型文档](https://huggingface.co/Qwen)
- [GGUF 格式说明](https://github.com/ggerganov/ggml/blob/master/docs/gguf.md)

### 认证和安全
- [JWT 认证指南](https://jwt.io/introduction)

### 部署和运维
- [Docker 官方文档](https://docs.docker.com/)
- [Docker Compose 指南](https://docs.docker.com/compose/)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。
