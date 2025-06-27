# 🚀 LLM Backend 部署指南

本文档详细介绍了如何在不同环境中部署 LLM Backend 服务。

## 📋 部署概述

LLM Backend 支持多种部署方式：
- **开发环境部署** - 本地开发和测试
- **生产环境部署** - 服务器生产环境
- **Docker 容器部署** - 容器化部署
- **云平台部署** - AWS、阿里云等云平台

## 🔧 环境要求

### 硬件要求

| 配置级别 | CPU | 内存 | 存储 | 适用场景 |
|----------|-----|------|------|----------|
| 最低配置 | 4核 | 8GB | 20GB | 开发测试 |
| 推荐配置 | 8核 | 16GB | 50GB | 小规模生产 |
| 高性能配置 | 16核 | 32GB | 100GB | 大规模生产 |

### 软件要求

- **操作系统**: Linux (Ubuntu 20.04+), macOS (10.15+), Windows 10+
- **Go**: 1.21 或更高版本
- **Node.js**: 18.0 或更高版本
- **Git**: 2.0 或更高版本
- **CMake**: 3.15 或更高版本 (编译 llama.cpp)

## 🏠 开发环境部署

### 1. 克隆项目

```bash
git clone <repository-url>
cd llm
```

### 2. 编译 llama.cpp

```bash
cd llama.cpp
mkdir build && cd build

# 基础编译
cmake ..

# 启用 GPU 支持 (可选)
cmake .. -DLLAMA_CUBLAS=ON

# 启用 Metal 支持 (macOS)
cmake .. -DLLAMA_METAL=ON

# 编译
make -j$(nproc)
cd ../..
```

### 3. 准备模型文件

```bash
mkdir -p models

# 下载推荐模型 (示例)
# 注意: 实际下载链接需要根据模型提供方获取

# 格式转换专用模型 (1.3GB)
wget -O models/deepseek-coder-1.3b-instruct-q4_k_m.gguf \
  "https://example.com/deepseek-coder-1.3b-instruct-q4_k_m.gguf"

# 通用对话模型 (4.2GB)
wget -O models/qwen2-7b-instruct-q4_k_m.gguf \
  "https://example.com/qwen2-7b-instruct-q4_k_m.gguf"

# 验证文件完整性
ls -lh models/
```

### 4. 启动后端服务

```bash
cd backend

# 安装依赖
go mod tidy

# 启动开发服务器
go run main.go

# 或编译后运行
go build -o llm-server main.go
./llm-server
```

### 5. 启动前端服务

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build
```

### 6. 验证部署

```bash
# 检查后端健康状态
curl http://localhost:8080/health

# 检查前端访问
open http://localhost:5173
```

## 🏭 生产环境部署

### 1. 系统准备

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y build-essential cmake git curl

# CentOS/RHEL
sudo yum groupinstall -y "Development Tools"
sudo yum install -y cmake git curl

# 创建专用用户
sudo useradd -m -s /bin/bash llm
sudo usermod -aG sudo llm
```

### 2. 安装 Go

```bash
# 下载并安装 Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 3. 安装 Node.js

```bash
# 使用 NodeSource 仓库
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 或使用 nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 18
nvm use 18
```

### 4. 部署应用

```bash
# 切换到 llm 用户
sudo su - llm

# 克隆项目到生产目录
git clone <repository-url> /opt/llm
cd /opt/llm

# 编译 llama.cpp
cd llama.cpp
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
cd ../..

# 构建后端
cd backend
go mod tidy
go build -o llm-server main.go
cd ..

# 构建前端
cd frontend
npm install
npm run build
cd ..
```

### 5. 配置系统服务

创建 systemd 服务文件：

```bash
sudo tee /etc/systemd/system/llm-backend.service > /dev/null <<EOF
[Unit]
Description=LLM Backend Service
After=network.target

[Service]
Type=simple
User=llm
Group=llm
WorkingDirectory=/opt/llm/backend
ExecStart=/opt/llm/backend/llm-server
Restart=always
RestartSec=5
Environment=LOG_LEVEL=info
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable llm-backend
sudo systemctl start llm-backend
sudo systemctl status llm-backend
```

### 6. 配置 Nginx

```bash
# 安装 Nginx
sudo apt install -y nginx

# 创建配置文件
sudo tee /etc/nginx/sites-available/llm-backend > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /opt/llm/frontend/dist;
        try_files \$uri \$uri/ /index.html;
    }

    # 后端 API
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # 健康检查
    location /health {
        proxy_pass http://localhost:8080;
    }
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/llm-backend /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## 🐳 Docker 部署

### 1. 创建 Dockerfile

```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app
COPY backend/ .
RUN go mod tidy && go build -o llm-server main.go

FROM node:18-alpine AS frontend-builder

WORKDIR /app
COPY frontend/ .
RUN npm install && npm run build

FROM ubuntu:22.04

# 安装运行时依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 创建应用用户
RUN useradd -m -s /bin/bash llm

# 复制编译好的文件
COPY --from=backend-builder /app/llm-server /usr/local/bin/
COPY --from=frontend-builder /app/dist /var/www/html/
COPY llama.cpp/build/bin/llama-server /usr/local/bin/

# 创建必要目录
RUN mkdir -p /opt/llm/models /opt/llm/logs
RUN chown -R llm:llm /opt/llm

USER llm
WORKDIR /opt/llm

EXPOSE 8080

CMD ["llm-server"]
```

### 2. 创建 docker-compose.yml

```yaml
version: '3.8'

services:
  llm-backend:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./models:/opt/llm/models:ro
      - ./logs:/opt/llm/logs
    environment:
      - LOG_LEVEL=info
      - PORT=8080
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - llm-backend
    restart: unless-stopped
```

### 3. 部署命令

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## ☁️ 云平台部署

### AWS 部署

1. **创建 EC2 实例**
```bash
# 选择 Ubuntu 22.04 LTS
# 实例类型: t3.large 或更高
# 存储: 50GB GP3
# 安全组: 开放 80, 443, 22 端口
```

2. **配置 ELB 负载均衡器**
```bash
# 创建 Application Load Balancer
# 配置健康检查: /health
# 配置 SSL 证书
```

3. **使用 RDS 数据库** (可选)
```bash
# 创建 RDS MySQL 实例
# 配置数据库连接
```

### 阿里云部署

1. **创建 ECS 实例**
```bash
# 选择 Ubuntu 20.04
# 规格: ecs.c6.2xlarge 或更高
# 系统盘: 100GB ESSD
```

2. **配置 SLB 负载均衡**
```bash
# 创建应用型负载均衡
# 配置监听器和后端服务器
```

## 🔒 安全配置

### 1. 防火墙配置

```bash
# Ubuntu UFW
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable

# CentOS firewalld
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 2. SSL 证书配置

```bash
# 使用 Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

### 3. 安全加固

```bash
# 禁用 root 登录
sudo sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config

# 配置密钥认证
ssh-keygen -t rsa -b 4096
ssh-copy-id user@server

# 更新系统
sudo apt update && sudo apt upgrade -y
```

## 📊 监控和维护

### 1. 日志管理

```bash
# 配置日志轮转
sudo tee /etc/logrotate.d/llm-backend > /dev/null <<EOF
/opt/llm/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    copytruncate
}
EOF
```

### 2. 性能监控

```bash
# 安装监控工具
sudo apt install htop iotop nethogs

# 监控系统资源
htop
iotop
nethogs
```

### 3. 备份策略

```bash
# 创建备份脚本
#!/bin/bash
BACKUP_DIR="/backup/llm-$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# 备份数据库
cp /opt/llm/backend/database.db $BACKUP_DIR/

# 备份配置文件
cp -r /opt/llm/backend/config $BACKUP_DIR/

# 压缩备份
tar -czf $BACKUP_DIR.tar.gz $BACKUP_DIR
rm -rf $BACKUP_DIR
```

## 🚨 故障排除

### 常见问题

1. **模型启动失败**
```bash
# 检查内存使用
free -h
# 检查模型文件
ls -lh models/
# 查看错误日志
tail -f logs/app.log
```

2. **端口占用**
```bash
# 查看端口使用
sudo netstat -tlnp | grep :8080
# 杀死占用进程
sudo kill -9 <PID>
```

3. **权限问题**
```bash
# 修复文件权限
sudo chown -R llm:llm /opt/llm
sudo chmod +x /opt/llm/backend/llm-server
```

---

**部署完成后，请访问应用并进行功能测试！** 🎉
