# LLM API 技术文档

## 项目概述

这是一个基于 Go + Vue 3 的轻量级大语言模型 API 服务，主要解决在资源受限环境下高效部署和使用多个 LLM 模型的问题。

### 为什么选择这种架构？

在实际开发中，我发现这种简单的架构最适合个人或小团队使用：

- **Go 后端**：处理高并发 API 请求，直接与 llama.cpp 通信，性能优秀
- **Vue 3 前端**：现代化的用户界面，开发效率高
- **llama.cpp**：目前 CPU 推理性能最好的引擎，支持量化模型
- **SQLite**：零配置数据库，部署简单

整个系统只需要编译两个服务就能运行，没有复杂的依赖关系。

## 🏗️ 系统架构

### 实际架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Go Backend    │    │   llama.cpp     │
│   (Vue 3)       │◄──►│   (Gin)         │◄──►│   Models        │
│   :5173         │    │   :8080         │    │   :8081-8085    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
    ┌─────────┐            ┌─────────┐            ┌─────────┐
    │ Element │            │ SQLite  │            │ GGUF    │
    │ Plus UI │            │ Database│            │ Models  │
    └─────────┘            └─────────┘            └─────────┘
```

### 技术选型的考虑

**Go 后端 (端口 8080)**:
- **选择原因**: Go 的并发模型非常适合处理大量 API 请求，而且编译后是单个二进制文件，部署简单
- **主要职责**:
  - 处理前端的 HTTP 请求
  - 管理 llama.cpp 进程的启动和停止
  - 文件上传和格式转换
  - 用户认证和 Token 管理
- **框架**: Gin (轻量级，性能好，中间件丰富)
- **数据库**: SQLite (零配置，适合单机部署)

**Vue 3 前端 (端口 5173)**:
- **选择原因**: Vue 3 学习曲线平缓，Element Plus 组件库成熟
- **主要功能**:
  - 模型管理界面
  - 聊天对话界面
  - 文件上传和转换界面
  - 用户登录和注册
- **构建工具**: Vite (开发时热重载快，构建速度快)

**llama.cpp 推理引擎 (端口 8081-8085)**:
- **选择原因**: 目前 CPU 推理性能最好的开源引擎，支持各种量化格式
- **部署方式**: 每个模型启动一个独立的 llama-server 进程
- **模型格式**: GGUF (新一代格式，加载速度比 GGML 快很多)
- **支持的模型**: Qwen2、DeepSeek、Yi、Mistral 等主流开源模型

## 📁 项目结构

```
LLM_API/
├── backend/                    # Go 后端服务
│   ├── internal/              # 内部包
│   │   ├── config/           # 配置管理
│   │   ├── handlers/         # HTTP 处理器
│   │   ├── middleware/       # 中间件
│   │   ├── models/          # 数据模型
│   │   ├── routes/          # 路由定义
│   │   └── services/        # 业务逻辑
│   ├── logs/                # 日志文件
│   ├── uploads/             # 文件上传
│   ├── main.go             # 程序入口
│   ├── go.mod              # Go 模块定义
│   └── *.sh                # 测试脚本
├── frontend/               # Vue 3 前端应用
│   ├── src/
│   │   ├── components/     # Vue 组件
│   │   ├── views/         # 页面视图
│   │   ├── utils/         # 工具函数
│   │   └── main.js        # 应用入口
│   ├── package.json       # 依赖配置
│   └── vite.config.js     # Vite 配置
├── models/                # AI 模型文件 (.gguf)
├── llama.cpp/            # llama.cpp 源码
├── logs/                 # 系统日志
└── README.md             # 项目说明
```

## 🔧 核心模块详解

### 1. 模型管理模块

这是整个系统的核心，负责管理多个 llama.cpp 进程。

**主要功能**:
- 动态启动和停止模型进程
- 端口自动分配 (8081-8085)
- 模型状态监控和健康检查
- 请求负载均衡

**实现思路**:
每个模型都是一个独立的 llama-server 进程，这样做的好处是：
- 模型之间互不影响，一个崩溃不会影响其他模型
- 可以根据需要动态启动和停止模型，节省内存
- 不同模型可以使用不同的参数配置

```go
type ModelManager struct {
    instances map[string]*ModelInstance
    ports     []int  // 可用端口池
    mu        sync.RWMutex
}

func (mm *ModelManager) StartModel(modelName string) error {
    // 检查模型文件是否存在
    // 分配可用端口
    // 启动 llama-server 进程
    // 等待服务就绪
}
```

### 2. 任务处理模块

针对不同的 AI 任务，我设计了专门的处理逻辑和提示词。

**文件格式转换**:
这个功能很实用，可以在不同的数据格式之间转换。我发现直接让 LLM 做转换比写解析器更灵活，特别是处理一些格式不规范的数据。

**作业批改**:
这个功能主要是给学生作业打分和提供反馈。我设计了一套评分标准，让 LLM 能够给出比较客观的评价。

**字幕处理**:
这个还在开发中，主要是想做视频字幕的提取和翻译。

### 3. 认证系统

采用了简单的 JWT 认证，没有使用复杂的 OAuth 或者 Keycloak，因为对于个人项目来说太重了。

**Token 消耗机制**:
为了防止滥用，我设计了一个 Token 消耗系统。不同的任务消耗不同数量的 Token：
- 简单对话: 1 Token/请求
- 格式转换: 根据内容长度计算
- 作业批改: 消耗较多，因为需要深度分析

### 4. 日志系统

我比较注重日志，因为在调试 LLM 应用时，日志是最重要的调试工具。

**日志内容**:
- 每个请求的完整生命周期
- 模型推理的耗时统计
- 错误详情和堆栈信息
- 用户行为统计

## 🚀 部署指南

### 环境要求

**系统要求**:
- 操作系统: Linux/macOS/Windows
- 内存: 8GB+ (推荐 16GB+)
- 存储: 20GB+ 可用空间
- CPU: 4核+ (推荐 8核+)

**软件依赖**:
- Go 1.21+
- Node.js 18+
- Git
- CMake (编译 llama.cpp)

### 安装步骤

1. **克隆项目**:
```bash
git clone <repository-url>
cd llm
```

2. **编译 llama.cpp**:
```bash
cd llama.cpp
mkdir build && cd build
cmake .. -DLLAMA_CUBLAS=ON  # 可选: GPU 支持
make -j$(nproc)
```

3. **下载模型文件**:
```bash
mkdir -p models
# 下载 GGUF 格式的模型文件到 models/ 目录
```

4. **启动后端服务**:
```bash
cd backend
go mod tidy
go run main.go
```

5. **启动前端服务**:
```bash
cd frontend
npm install
npm run dev
```

### Docker 部署

```dockerfile
# Dockerfile 示例
FROM golang:1.21-alpine AS backend
WORKDIR /app
COPY backend/ .
RUN go build -o llm-server main.go

FROM node:18-alpine AS frontend
WORKDIR /app
COPY frontend/ .
RUN npm install && npm run build

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=backend /app/llm-server /usr/local/bin/
COPY --from=frontend /app/dist /var/www/html/
EXPOSE 8080
CMD ["llm-server"]
```

## 📊 性能优化

### 1. 模型优化

**量化策略**:
- INT4 量化: 显著减少内存使用
- INT8 量化: 平衡性能和精度
- 动态量化: 运行时优化

**内存管理**:
- 模型缓存机制
- 内存池复用
- 垃圾回收优化

### 2. 并发优化

**连接池**:
```go
type ConnectionPool struct {
    pool    chan *Connection
    maxSize int
    current int
}
```

**协程管理**:
- Worker Pool 模式
- 请求队列限制
- 超时控制

### 3. 缓存策略

**多级缓存**:
- 内存缓存: 热点数据
- Redis 缓存: 分布式缓存
- 文件缓存: 静态资源

## 🔒 安全机制

### 1. 认证安全

- JWT Token 过期机制
- 密码哈希存储 (bcrypt)
- API 访问频率限制
- CORS 跨域保护

### 2. 数据安全

- SQL 注入防护
- XSS 攻击防护
- 文件上传安全检查
- 敏感数据加密

### 3. 网络安全

- HTTPS 强制加密
- 请求签名验证
- IP 白名单机制
- DDoS 攻击防护

## 🧪 测试策略

### 1. 单元测试

```bash
# 运行单元测试
go test ./internal/...
```

### 2. 集成测试

```bash
# 运行完整 API 测试
./test_all_apis.sh
```

### 3. 性能测试

```bash
# 压力测试
ab -n 1000 -c 10 http://localhost:8080/api/v1/tasks/convert
```

### 4. 自动化测试

- CI/CD 集成
- 自动化部署
- 回归测试
- 性能监控

## 📈 监控与运维

### 1. 健康检查

```bash
curl http://localhost:8080/health
```

### 2. 指标监控

- 请求响应时间
- 错误率统计
- 资源使用率
- 模型性能指标

### 3. 日志分析

- 结构化日志格式
- 日志聚合分析
- 异常告警机制
- 性能瓶颈识别

## 🔮 未来规划

### 短期目标 (1-3个月)

- [ ] 支持更多模型格式
- [ ] 优化推理性能
- [ ] 增加批处理功能
- [ ] 完善监控系统

### 中期目标 (3-6个月)

- [ ] 分布式部署支持
- [ ] 模型微调功能
- [ ] 高级缓存策略
- [ ] 多语言 SDK

### 长期目标 (6-12个月)

- [ ] 云原生架构
- [ ] AI 模型市场
- [ ] 智能运维系统
- [ ] 企业级功能

## 🤝 贡献指南

### 开发流程

1. Fork 项目
2. 创建功能分支
3. 提交代码变更
4. 编写测试用例
5. 提交 Pull Request

### 代码规范

- Go: 遵循 Go 官方规范
- JavaScript: 使用 ESLint
- 提交信息: 遵循 Conventional Commits

### 问题报告

- 使用 Issue 模板
- 提供详细的复现步骤
- 包含环境信息
- 附加相关日志

## 📖 API 接口文档

### 认证接口

#### POST /api/v1/auth/register
用户注册

**请求体**:
```json
{
  "username": "testuser",
  "password": "password123",
  "email": "test@example.com"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "tokens": 1000
  }
}
```

#### POST /api/v1/auth/login
用户登录

**请求体**:
```json
{
  "username": "testuser",
  "password": "password123"
}
```

#### GET /api/v1/auth/profile
获取用户信息

**请求头**: `Authorization: Bearer <token>`

### 模型管理接口

#### GET /api/v1/models/
获取可用模型列表

**响应**:
```json
{
  "models": [
    {
      "name": "qwen2-7b-instruct",
      "description": "通用对话模型",
      "status": "available",
      "size": "4.2GB"
    }
  ]
}
```

#### POST /api/v1/models/{name}/start
启动指定模型

#### POST /api/v1/models/{name}/stop
停止指定模型

#### POST /api/v1/models/{name}/chat
与模型对话

**请求体**:
```json
{
  "message": "你好",
  "max_tokens": 100
}
```

### 任务处理接口

#### GET /api/v1/tasks/formats
获取支持的格式列表

#### POST /api/v1/tasks/convert
文件格式转换

**请求体**:
```json
{
  "source_format": "json",
  "target_format": "yaml",
  "content": "{\"name\":\"test\",\"value\":123}"
}
```

**响应**:
```json
{
  "success": true,
  "converted_content": "name: test\nvalue: 123",
  "tokens_consumed": 15,
  "remaining_tokens": 985
}
```

#### POST /api/v1/tasks/homework
作业批改

**请求体**:
```json
{
  "subject": "数学",
  "question": "计算 2+3×4 的结果",
  "answer": "14",
  "grade_level": "小学"
}
```

**响应**:
```json
{
  "success": true,
  "score": 100,
  "feedback": "答案正确！计算步骤清晰。",
  "suggestions": ["继续保持良好的计算习惯"],
  "tokens_consumed": 25,
  "remaining_tokens": 960
}
```

#### POST /api/v1/tasks/subtitle
字幕处理

**请求体**:
```json
{
  "video_path": "/path/to/video.mp4",
  "source_lang": "英文",
  "target_lang": "中文",
  "output_format": "srt"
}
```

## 🛠️ 开发工具

### 测试脚本

项目提供了多个测试脚本：

1. **完整测试**: `./test_all_apis.sh`
2. **快速测试**: `./quick_test.sh`
3. **curl测试**: `./curl_test.sh`
4. **单接口测试**: `./test_single_api.sh convert`

### 开发命令

```bash
# 后端开发
cd backend
go run main.go                    # 启动开发服务器
go test ./...                     # 运行测试
go build -o llm-server main.go    # 编译生产版本

# 前端开发
cd frontend
npm run dev                       # 启动开发服务器
npm run build                     # 构建生产版本
npm run preview                   # 预览生产版本
```

### 调试技巧

1. **后端调试**:
```bash
# 启用详细日志
export LOG_LEVEL=debug
go run main.go

# 使用 delve 调试器
dlv debug main.go
```

2. **前端调试**:
```bash
# 启用 Vue DevTools
npm run dev

# 查看网络请求
# 在浏览器开发者工具中查看 Network 标签
```

3. **模型调试**:
```bash
# 直接测试 llama.cpp
./llama.cpp/build/bin/llama-server \
  -m models/qwen2-7b-instruct-q4_k_m.gguf \
  --port 8081 --host 127.0.0.1

# 测试模型响应
curl -X POST http://127.0.0.1:8081/completion \
  -H "Content-Type: application/json" \
  -d '{"prompt":"你好","n_predict":50}'
```

## 📋 常见问题 (FAQ)

### Q: 模型启动失败怎么办？
A: 检查以下几点：
- 模型文件是否存在且完整
- 内存是否足够 (至少 8GB)
- 端口是否被占用
- llama.cpp 是否正确编译

### Q: Token 消耗过快怎么办？
A: 可以通过以下方式优化：
- 减少输入内容长度
- 降低 max_tokens 参数
- 使用更轻量的模型
- 实现缓存机制

### Q: 如何添加新的模型？
A: 按以下步骤操作：
1. 将 GGUF 格式模型文件放入 `models/` 目录
2. 在配置文件中添加模型定义
3. 重启后端服务
4. 通过 API 启动新模型

### Q: 如何优化推理速度？
A: 可以尝试：
- 使用量化模型 (INT4/INT8)
- 启用 GPU 加速 (CUDA/Metal)
- 调整线程数参数
- 使用更快的存储设备

---
