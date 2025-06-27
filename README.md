# 🤖 LLM Backend - 轻量级大语言模型后端服务

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.0+-green.svg)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

一个基于 Go 和 Vue 3 开发的轻量级大语言模型后端服务，集成多个开源 LLM 模型，提供文件格式转换、作业批改、字幕处理等 AI 功能。

## ✨ 核心特性

- 🚀 **多模型支持** - 集成 Qwen、DeepSeek、Yi、Mistral 等多个轻量级模型
- 🔄 **动态模型管理** - 支持模型的动态启动、停止和切换，节省系统资源
- 🎯 **专用任务处理** - 文件格式转换、作业批改、字幕处理等专门优化的功能
- 🔐 **完整认证系统** - JWT 认证、Token 管理、用户权限控制
- ⚡ **高性能架构** - 负载均衡、连接池、智能缓存机制
- 🌐 **现代化前端** - Vue 3 + Element Plus 响应式 Web 界面
- 📊 **监控与日志** - 完整的请求追踪和性能监控
- 🧪 **完善测试** - 提供多种测试脚本和工具

## 📋 系统要求

### 最低配置
- **CPU**: 8 核心以上
- **内存**: 16GB RAM
- **存储**: 20GB 可用空间
- **操作系统**: macOS 10.15+ / Ubuntu 20.04+ / CentOS 8+

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
cd llm

# 安装系统依赖 (Ubuntu/Debian)
sudo apt update
sudo apt install -y build-essential git wget curl
sudo apt install -y openjdk-11-jdk maven

# 安装系统依赖 (macOS)
brew install git wget curl maven
# Java 11 通过 brew 或 Oracle 官网安装
```

### 2. 设置 llama.cpp

```bash
# 运行自动化设置脚本
chmod +x setup-llama-cpp.sh
./setup-llama-cpp.sh

# 检查系统信息
./system-info.sh
```

### 3. 下载模型

```bash
# 下载推荐的 CPU 优化模型
./download-models.sh

# 或手动下载特定模型
wget https://huggingface.co/Qwen/Qwen2-7B-Instruct-GGUF/resolve/main/qwen2-7b-instruct-q4_k_m.gguf -P models/
```

### 4. 构建和启动服务

```bash
# 构建 Java 应用
mvn clean package -DskipTests

# 启动服务
./start-llm-service.sh

# 或手动启动
java -jar target/*.jar --spring.profiles.active=cpu
```

### 5. 测试部署

```bash
# 测试模型推理
./test-model.sh

# 测试 API 接口
curl -X POST http://localhost:8080/api/v1/chat/generate \
  -H "Content-Type: application/json" \
  -d '{"message": "你好，请介绍一下你自己"}'
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

#### CPU 线程配置
```yaml
# application-cpu.yml
app:
  models:
    inference-mode: direct  # direct 或 server
  llama:
    threads: 12  # 建议设置为 CPU 核心数的 75%
```

#### JVM 优化
```bash
# 内存配置
export JAVA_OPTS="-Xmx8g -Xms4g -XX:+UseG1GC -XX:MaxGCPauseMillis=200"

# 针对大内存服务器
export JAVA_OPTS="-Xmx16g -Xms8g -XX:+UseG1GC -XX:G1HeapRegionSize=16m"
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

### 聊天接口
```bash
POST /api/v1/chat/generate
Content-Type: application/json

{
  "message": "你好，请帮我写一个 Python 函数",
  "modelName": "Qwen2-7B-Instruct",
  "temperature": 0.7,
  "maxTokens": 500,
  "systemPrompt": "你是一个专业的编程助手"
}
```

### 模型管理
```bash
# 获取可用模型
GET /api/v1/models/available

# 获取当前模型
GET /api/v1/models/current

# 切换模型 (需要管理员权限)
POST /api/v1/models/load/Qwen2-7B-Instruct

# 获取系统状态
GET /api/v1/models/status
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

### Dockerfile
```dockerfile
FROM openjdk:11-jre-slim

# 安装构建工具
RUN apt-get update && apt-get install -y \
    build-essential git wget curl \
    && rm -rf /var/lib/apt/lists/*

# 复制应用
COPY target/*.jar app.jar
COPY llama-cpp/ /app/llama-cpp/
COPY models/ /app/models/

WORKDIR /app

EXPOSE 8080

CMD ["java", "-jar", "app.jar", "--spring.profiles.active=cpu"]
```

### Docker Compose
```yaml
version: '3.8'
services:
  llm-app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SPRING_PROFILES_ACTIVE=cpu
      - JAVA_OPTS=-Xmx8g -Xms4g
    volumes:
      - ./models:/app/models
      - ./logs:/app/logs
    depends_on:
      - keycloak
      
  keycloak:
    image: quay.io/keycloak/keycloak:latest
    ports:
      - "8081:8080"
    environment:
      - KEYCLOAK_ADMIN=admin
      - KEYCLOAK_ADMIN_PASSWORD=admin
    command: start-dev
```

## 🔧 故障排除

### 常见问题

1. **内存不足**
```bash
# 检查内存使用
free -h
# 减少 JVM 堆内存或使用更小的模型
export JAVA_OPTS="-Xmx4g -Xms2g"
```

2. **模型加载失败**
```bash
# 检查模型文件
ls -la models/
# 验证文件完整性
file models/*.gguf
```

3. **推理速度慢**
```bash
# 调整线程数
# 在 application-cpu.yml 中设置合适的线程数
threads: 8  # 根据 CPU 核心数调整
```

4. **llama.cpp 编译失败**
```bash
# 清理重新编译
cd llama-cpp
make clean
make -j$(nproc)
```

### 日志分析
```bash
# 查看应用日志
tail -f logs/llm-app.log

# 查看推理性能
grep "inference time" logs/llm-app.log
```

## 📈 监控和运维

### 健康检查
```bash
# 应用健康状态
curl http://localhost:8080/actuator/health

# 系统指标
curl http://localhost:8080/actuator/metrics
```

### 性能监控
```bash
# CPU 使用率
top -p $(pgrep java)

# 内存使用
ps aux | grep java

# 推理统计
curl http://localhost:8080/api/v1/models/status
```

## 🚀 生产部署建议

1. **硬件配置**
   - 使用 SSD 存储模型文件
   - 配置足够的 RAM (模型大小 × 1.5)
   - 选择高频 CPU

2. **系统优化**
   - 关闭不必要的系统服务
   - 设置 CPU 性能模式
   - 优化网络配置

3. **应用配置**
   - 使用生产级数据库 (MySQL/PostgreSQL)
   - 配置日志轮转
   - 设置监控告警

4. **安全加固**
   - 启用 HTTPS
   - 配置防火墙
   - 定期更新依赖

## 📚 参考资源

- [llama.cpp 官方文档](https://github.com/ggerganov/llama.cpp)
- [Qwen2 模型文档](https://huggingface.co/Qwen)
- [Spring Boot 官方文档](https://spring.io/projects/spring-boot)
- [Keycloak 官方文档](https://www.keycloak.org/documentation)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。