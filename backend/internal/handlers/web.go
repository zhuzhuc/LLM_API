package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebHandler Web 页面处理器
type WebHandler struct{}

// NewWebHandler 创建 Web 处理器
func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

// Index 首页
func (h *WebHandler) Index(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CPU 大模型服务平台</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .status {
            background: #e8f5e8;
            border: 1px solid #4caf50;
            border-radius: 5px;
            padding: 15px;
            margin-bottom: 30px;
        }
        .status.running {
            background: #e8f5e8;
            border-color: #4caf50;
        }
        .api-section {
            margin-bottom: 30px;
        }
        .api-section h3 {
            color: #555;
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
        }
        .api-item {
            background: #f9f9f9;
            border-left: 4px solid #2196f3;
            padding: 15px;
            margin-bottom: 15px;
        }
        .method {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 3px;
            color: white;
            font-weight: bold;
            margin-right: 10px;
        }
        .get { background: #4caf50; }
        .post { background: #ff9800; }
        .put { background: #2196f3; }
        .delete { background: #f44336; }
        code {
            background: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Monaco', 'Consolas', monospace;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 CPU 大模型服务平台</h1>
        
        <div class="status running">
            <strong>✅ 服务状态：运行中</strong><br>
            API 服务地址：<code>http://localhost:8080</code>
        </div>

        <div class="api-section">
            <h3>🔐 认证接口</h3>
            <div class="api-item">
                <span class="method post">POST</span>
                <code>/api/v1/auth/register</code> - 用户注册
            </div>
            <div class="api-item">
                <span class="method post">POST</span>
                <code>/api/v1/auth/login</code> - 用户登录
            </div>
        </div>

        <div class="api-section">
            <h3>🤖 模型管理</h3>
            <div class="api-item">
                <span class="method get">GET</span>
                <code>/api/v1/models</code> - 获取可用模型列表
            </div>
            <div class="api-item">
                <span class="method get">GET</span>
                <code>/api/v1/models/running</code> - 获取运行中的模型
            </div>
            <div class="api-item">
                <span class="method post">POST</span>
                <code>/api/v1/models/{name}/start</code> - 启动指定模型
            </div>
            <div class="api-item">
                <span class="method post">POST</span>
                <code>/api/v1/models/{name}/chat</code> - 与模型对话
            </div>
        </div>

        <div class="api-section">
            <h3>🔄 OpenAI 兼容接口</h3>
            <div class="api-item">
                <span class="method post">POST</span>
                <code>/api/v1/v1/chat/completions</code> - 聊天完成
            </div>
            <div class="api-item">
                <span class="method get">GET</span>
                <code>/api/v1/v1/models</code> - 模型列表
            </div>
        </div>

        <div class="api-section">
            <h3>📊 监控和管理</h3>
            <div class="api-item">
                <span class="method get">GET</span>
                <code>/health</code> - 健康检查
            </div>
            <div class="api-item">
                <span class="method get">GET</span>
                <code>/status</code> - 系统状态
            </div>
            <div class="api-item">
                <span class="method get">GET</span>
                <code>/api/v1/monitoring/metrics</code> - 系统指标
            </div>
            <div class="api-item">
                <span class="method get">GET</span>
                <code>/api/v1/cluster/stats</code> - 集群状态
            </div>
        </div>

        <div class="api-section">
            <h3>📝 快速开始</h3>
            <div class="api-item">
                <strong>1. 注册用户：</strong><br>
                <code>curl -X POST http://localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d '{"username":"admin","password":"password","email":"admin@example.com"}'</code>
            </div>
            <div class="api-item">
                <strong>2. 登录获取 Token：</strong><br>
                <code>curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}'</code>
            </div>
            <div class="api-item">
                <strong>3. 启动模型：</strong><br>
                <code>curl -X POST -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/v1/models/qwen2-7b-instruct/start</code>
            </div>
        </div>

        <div class="footer">
            <p>CPU 大模型服务平台 - 基于 llama.cpp 的高效 CPU 推理服务</p>
        </div>
    </div>
</body>
</html>
`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
