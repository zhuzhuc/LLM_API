# 🏗️ LLM Backend 系统架构

## 📋 架构概述

LLM Backend 采用现代化的微服务架构设计，前后端分离，支持多模型动态管理和负载均衡。

## 🎯 设计原则

- **模块化设计** - 各功能模块独立，便于维护和扩展
- **高可用性** - 支持负载均衡和故障转移
- **可扩展性** - 支持水平扩展和垂直扩展
- **安全性** - 完整的认证授权机制
- **性能优化** - 缓存、连接池、异步处理

## 🏛️ 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        用户层                                │
├─────────────────────────────────────────────────────────────┤
│  Web Browser  │  Mobile App  │  API Client  │  CLI Tool    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      负载均衡层                              │
├─────────────────────────────────────────────────────────────┤
│              Nginx / HAProxy / Cloud LB                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      前端服务层                              │
├─────────────────────────────────────────────────────────────┤
│  Vue 3 + Vite  │  Element Plus  │  Pinia  │  Vue Router   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ HTTP/HTTPS
┌─────────────────────────────────────────────────────────────┐
│                      API 网关层                             │
├─────────────────────────────────────────────────────────────┤
│    认证中间件   │   限流中间件   │   日志中间件   │   CORS    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      应用服务层                              │
├─────────────────────────────────────────────────────────────┤
│  Auth Service │ Model Service │ Task Service │ User Service │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      模型推理层                              │
├─────────────────────────────────────────────────────────────┤
│   Qwen2-7B    │ DeepSeek-1.3B │   Yi-9B     │  Mistral-7B  │
│   (Port 8081) │  (Port 8082)  │ (Port 8083) │ (Port 8084)  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      数据存储层                              │
├─────────────────────────────────────────────────────────────┤
│   SQLite DB   │  Model Files  │  Log Files  │  Cache Redis │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 核心组件

### 1. 前端架构 (Frontend)

```
frontend/
├── src/
│   ├── components/          # 可复用组件
│   │   ├── ModelCard.vue   # 模型卡片组件
│   │   ├── TaskForm.vue    # 任务表单组件
│   │   └── UserInfo.vue    # 用户信息组件
│   ├── views/              # 页面视图
│   │   ├── Dashboard.vue   # 仪表板
│   │   ├── Models.vue      # 模型管理
│   │   ├── Converter.vue   # 格式转换
│   │   ├── Homework.vue    # 作业批改
│   │   └── Subtitle.vue    # 字幕处理
│   ├── stores/             # 状态管理
│   │   ├── auth.js         # 认证状态
│   │   ├── models.js       # 模型状态
│   │   └── tasks.js        # 任务状态
│   ├── utils/              # 工具函数
│   │   ├── api.js          # API 客户端
│   │   ├── auth.js         # 认证工具
│   │   └── format.js       # 格式化工具
│   └── router/             # 路由配置
│       └── index.js
```

**技术栈**:
- **Vue 3** - 渐进式框架
- **Vite** - 快速构建工具
- **Element Plus** - UI 组件库
- **Pinia** - 状态管理
- **Axios** - HTTP 客户端

### 2. 后端架构 (Backend)

```
backend/
├── internal/
│   ├── handlers/           # HTTP 处理器
│   │   ├── auth.go        # 认证处理器
│   │   ├── models.go      # 模型处理器
│   │   ├── tasks.go       # 任务处理器
│   │   └── llm.go         # LLM 处理器
│   ├── services/          # 业务逻辑层
│   │   ├── auth_service.go
│   │   ├── model_manager.go
│   │   ├── task_service.go
│   │   └── load_balancer.go
│   ├── models/            # 数据模型
│   │   ├── user.go
│   │   ├── api_call.go
│   │   └── model.go
│   ├── middleware/        # 中间件
│   │   ├── auth.go        # JWT 认证
│   │   ├── cors.go        # 跨域处理
│   │   ├── logger.go      # 日志记录
│   │   └── rate_limiter.go # 限流控制
│   ├── config/            # 配置管理
│   │   └── config.go
│   └── routes/            # 路由定义
│       └── routes.go
```

**技术栈**:
- **Go 1.21+** - 高性能语言
- **Gin** - Web 框架
- **SQLite** - 轻量级数据库
- **JWT** - 认证机制
- **llama.cpp** - 模型推理

### 3. 模型推理层

```
模型实例管理:
┌─────────────────┐    ┌─────────────────┐
│   ModelManager  │    │ ServiceRegistry │
│                 │◄──►│                 │
│ - instances     │    │ - services      │
│ - ports         │    │ - health_check  │
│ - configs       │    │ - load_balance  │
└─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│            模型实例池                    │
├─────────────────────────────────────────┤
│ Qwen2-7B:8081    │ DeepSeek:8082       │
│ Yi-9B:8083       │ Mistral:8084        │
└─────────────────────────────────────────┘
```

**特性**:
- **动态加载** - 按需启动/停止模型
- **端口管理** - 自动分配可用端口
- **健康检查** - 监控模型状态
- **负载均衡** - 智能请求分发

## 🔄 数据流架构

### 1. 用户认证流程

```
用户 → 前端 → 后端 → 数据库
 │      │      │      │
 │      │      │      └─ 验证用户凭据
 │      │      └─ 生成 JWT Token
 │      └─ 存储 Token 到本地
 └─ 后续请求携带 Token
```

### 2. 模型调用流程

```
用户请求 → API网关 → 认证中间件 → 业务处理器
    │         │         │           │
    │         │         │           ├─ Token扣除
    │         │         │           ├─ 模型选择
    │         │         │           └─ 请求转发
    │         │         │
    │         │         └─ JWT验证
    │         │
    │         └─ 限流检查
    │
    ▼
模型实例 → 推理计算 → 返回结果 → 响应用户
    │         │         │         │
    │         │         │         └─ 更新Token余额
    │         │         └─ 格式化输出
    │         └─ llama.cpp处理
    └─ 负载均衡选择
```

### 3. 任务处理流程

```
任务请求 → 参数验证 → 模型选择 → 提示词构建
    │         │         │         │
    │         │         │         └─ 专用提示词模板
    │         │         └─ 根据任务类型选择
    │         └─ 检查必要参数
    │
    ▼
模型推理 → 结果解析 → 格式化 → 返回响应
    │         │         │       │
    │         │         │       └─ 包含Token消耗信息
    │         │         └─ JSON/文本格式化
    │         └─ 智能解析模型输出
    └─ 调用专用模型
```

## 🔐 安全架构

### 1. 认证授权

```
┌─────────────────┐    ┌─────────────────┐
│   JWT Token     │    │  用户权限管理    │
│                 │    │                 │
│ - user_id       │◄──►│ - 角色定义      │
│ - username      │    │ - 权限矩阵      │
│ - expiration    │    │ - 资源访问控制   │
└─────────────────┘    └─────────────────┘
```

### 2. 数据安全

- **传输加密** - HTTPS/TLS
- **存储加密** - 敏感数据加密存储
- **访问控制** - 基于角色的权限控制
- **审计日志** - 完整的操作记录

### 3. API 安全

- **输入验证** - 严格的参数校验
- **SQL 注入防护** - 参数化查询
- **XSS 防护** - 输出编码
- **CSRF 防护** - Token 验证

## ⚡ 性能架构

### 1. 缓存策略

```
┌─────────────────┐    ┌─────────────────┐
│   内存缓存       │    │   分布式缓存     │
│                 │    │                 │
│ - 热点数据      │◄──►│ - Redis集群     │
│ - 用户会话      │    │ - 模型结果缓存   │
│ - 配置信息      │    │ - 跨实例共享     │
└─────────────────┘    └─────────────────┘
```

### 2. 连接池管理

```
┌─────────────────┐
│   连接池管理     │
├─────────────────┤
│ - 数据库连接池   │
│ - HTTP连接池     │
│ - 模型连接池     │
│ - 资源复用      │
└─────────────────┘
```

### 3. 异步处理

- **协程池** - Go routine 管理
- **任务队列** - 异步任务处理
- **事件驱动** - 非阻塞 I/O
- **流式处理** - 大文件处理

## 📊 监控架构

### 1. 指标收集

```
应用指标 → 系统指标 → 业务指标
    │         │         │
    │         │         └─ 用户活跃度、任务成功率
    │         └─ CPU、内存、磁盘、网络
    └─ 请求量、响应时间、错误率
```

### 2. 日志系统

```
┌─────────────────┐    ┌─────────────────┐
│   结构化日志     │    │   日志聚合      │
│                 │    │                 │
│ - 请求日志      │───►│ - ELK Stack     │
│ - 错误日志      │    │ - 日志分析      │
│ - 性能日志      │    │ - 告警通知      │
└─────────────────┘    └─────────────────┘
```

### 3. 健康检查

- **服务健康检查** - /health 端点
- **模型健康检查** - 模型响应测试
- **依赖健康检查** - 数据库、缓存连接
- **自动故障转移** - 异常实例隔离

## 🚀 扩展架构

### 1. 水平扩展

```
┌─────────────────┐    ┌─────────────────┐
│   负载均衡器     │    │   服务实例      │
│                 │    │                 │
│ - 请求分发      │───►│ - 实例1:8080    │
│ - 健康检查      │    │ - 实例2:8081    │
│ - 故障转移      │    │ - 实例N:808N    │
└─────────────────┘    └─────────────────┘
```

### 2. 垂直扩展

- **资源升级** - CPU、内存、存储
- **模型优化** - 量化、剪枝、蒸馏
- **算法优化** - 缓存、索引、并行

### 3. 云原生架构

```
┌─────────────────┐    ┌─────────────────┐
│   Kubernetes    │    │   微服务拆分     │
│                 │    │                 │
│ - Pod管理       │◄──►│ - 认证服务      │
│ - 服务发现      │    │ - 模型服务      │
│ - 自动扩缩容     │    │ - 任务服务      │
└─────────────────┘    └─────────────────┘
```

## 🔮 未来架构演进

### 短期目标
- **服务网格** - Istio/Linkerd
- **配置中心** - Consul/etcd
- **消息队列** - RabbitMQ/Kafka

### 中期目标
- **分布式追踪** - Jaeger/Zipkin
- **服务治理** - 熔断、降级、限流
- **多云部署** - 跨云平台支持

### 长期目标
- **边缘计算** - 边缘节点部署
- **联邦学习** - 分布式模型训练
- **智能运维** - AIOps 自动化

---

**架构设计遵循业界最佳实践，确保系统的可靠性、可扩展性和可维护性。** 🏗️
