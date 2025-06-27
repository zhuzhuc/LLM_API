# CPU 大模型服务平台 Dockerfile
FROM ubuntu:22.04

# 设置环境变量
ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.21.0
ENV NODE_VERSION=18

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    wget \
    curl \
    unzip \
    libopenblas-dev \
    pkg-config \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# 安装 Go
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# 设置 Go 环境
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# 创建工作目录
WORKDIR /app

# 复制项目文件
COPY . .

# 编译 llama.cpp
RUN cd llama-cpp && \
    mkdir -p build && \
    cd build && \
    cmake .. -DLLAMA_OPENBLAS=ON && \
    make -j$(nproc)

# 编译后端服务
RUN cd backend && \
    go mod tidy && \
    go build -o llm-server main.go

# 创建必要的目录
RUN mkdir -p logs models

# 设置权限
RUN chmod +x start-llm-service.sh stop-llm-service.sh

# 暴露端口
EXPOSE 8080 8081-8090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动命令
CMD ["./start-llm-service.sh"]
