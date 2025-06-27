package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"llm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type GatewayHandler struct {
	modelManager *services.ModelManager
	llmService   *services.LLMService
}

func NewGatewayHandler(modelManager *services.ModelManager, llmService *services.LLMService) *GatewayHandler {
	return &GatewayHandler{
		modelManager: modelManager,
		llmService:   llmService,
	}
}

// ProxyRequest 代理请求结构
type ProxyRequest struct {
	Model       string                 `json:"model"`
	Messages    []ChatMessage          `json:"messages"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float64                `json:"temperature"`
	TopP        float64                `json:"top_p"`
	Stream      bool                   `json:"stream"`
	Stop        []string               `json:"stop"`
	Extra       map[string]interface{} `json:"-"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI 兼容的聊天完成接口
func (h *GatewayHandler) ChatCompletions(c *gin.Context) {
	var req ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request format: " + err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	// 设置默认值
	if req.Model == "" {
		req.Model = "qwen2-7b-instruct" // 默认模型
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 200
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}
	if req.TopP <= 0 {
		req.TopP = 0.9
	}

	// 构建提示词
	prompt := h.buildPromptFromMessages(req.Messages)

	// 确保模型正在运行
	if err := h.modelManager.StartModel(req.Model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Failed to start model: " + err.Error(),
				"type":    "internal_error",
				"code":    "model_unavailable",
			},
		})
		return
	}

	// 生成响应
	response, tokens, err := h.llmService.GenerateResponse(prompt, req.MaxTokens, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Failed to generate response: " + err.Error(),
				"type":    "internal_error",
				"code":    "generation_failed",
			},
		})
		return
	}

	// 返回 OpenAI 兼容格式
	c.JSON(http.StatusOK, gin.H{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   req.Model,
		"choices": []gin.H{
			{
				"index": 0,
				"message": gin.H{
					"role":    "assistant",
					"content": response,
				},
				"finish_reason": "stop",
			},
		},
		"usage": gin.H{
			"prompt_tokens":     tokens / 2, // 估算
			"completion_tokens": tokens / 2, // 估算
			"total_tokens":      tokens,
		},
	})
}

// 获取可用模型列表（OpenAI 兼容）
func (h *GatewayHandler) ListModels(c *gin.Context) {
	models := h.modelManager.GetAvailableModels()

	var modelList []gin.H
	for _, model := range models {
		modelList = append(modelList, gin.H{
			"id":       model.ModelName,
			"object":   "model",
			"created":  time.Now().Unix(),
			"owned_by": "local",
			"permission": []gin.H{
				{
					"id":                   fmt.Sprintf("modelperm-%s", model.ModelName),
					"object":               "model_permission",
					"created":              time.Now().Unix(),
					"allow_create_engine":  false,
					"allow_sampling":       true,
					"allow_logprobs":       false,
					"allow_search_indices": false,
					"allow_view":           true,
					"allow_fine_tuning":    false,
					"organization":         "*",
					"group":                nil,
					"is_blocking":          false,
				},
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   modelList,
	})
}

// 代理到 llama.cpp 服务器
func (h *GatewayHandler) ProxyToLlamaCpp(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Model name is required",
		})
		return
	}

	// 确保模型正在运行
	if err := h.modelManager.StartModel(modelName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start model: " + err.Error(),
		})
		return
	}

	// 获取模型实例
	instance, err := h.modelManager.GetModelInstance(modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get model instance: " + err.Error(),
		})
		return
	}

	// 创建反向代理
	targetURL := fmt.Sprintf("http://localhost:%d", instance.Port)
	target, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid target URL: " + err.Error(),
		})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 修改请求路径
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = strings.Replace(originalPath, "/api/v1/proxy/"+modelName, "", 1)

	proxy.ServeHTTP(c.Writer, c.Request)
}

// 健康检查
func (h *GatewayHandler) HealthCheck(c *gin.Context) {
	runningModels := h.modelManager.ListRunningModels()

	status := "healthy"
	if len(runningModels) == 0 {
		status = "no_models_running"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         status,
		"timestamp":      time.Now().Unix(),
		"running_models": len(runningModels),
		"version":        "1.0.0",
	})
}

// 获取系统状态
func (h *GatewayHandler) SystemStatus(c *gin.Context) {
	runningModels := h.modelManager.ListRunningModels()
	availableModels := h.modelManager.GetAvailableModels()

	var modelStatus []gin.H
	for name, instance := range runningModels {
		modelStatus = append(modelStatus, gin.H{
			"name":        name,
			"status":      instance.Status,
			"port":        instance.Port,
			"uptime":      time.Since(instance.StartTime).Seconds(),
			"usage_count": instance.UsageCount,
			"last_used":   instance.LastUsed.Unix(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"system": gin.H{
				"status":    "running",
				"timestamp": time.Now().Unix(),
				"uptime":    time.Since(time.Now()).Seconds(), // 这里应该是服务启动时间
			},
			"models": gin.H{
				"available": len(availableModels),
				"running":   len(runningModels),
				"details":   modelStatus,
			},
		},
	})
}

// buildPromptFromMessages 从消息列表构建提示词
func (h *GatewayHandler) buildPromptFromMessages(messages []ChatMessage) string {
	var prompt strings.Builder

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			prompt.WriteString("System: " + msg.Content + "\n")
		case "user":
			prompt.WriteString("User: " + msg.Content + "\n")
		case "assistant":
			prompt.WriteString("Assistant: " + msg.Content + "\n")
		}
	}

	prompt.WriteString("Assistant: ")
	return prompt.String()
}

// 批量请求处理
func (h *GatewayHandler) BatchRequest(c *gin.Context) {
	var requests []ProxyRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid batch request format: " + err.Error(),
		})
		return
	}

	if len(requests) > 10 { // 限制批量请求数量
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Batch size too large, maximum 10 requests allowed",
		})
		return
	}

	var responses []gin.H
	for i, req := range requests {
		// 设置默认值
		if req.Model == "" {
			req.Model = "qwen2-7b-instruct"
		}
		if req.MaxTokens <= 0 {
			req.MaxTokens = 200
		}

		prompt := h.buildPromptFromMessages(req.Messages)

		// 启动模型
		if err := h.modelManager.StartModel(req.Model); err != nil {
			responses = append(responses, gin.H{
				"index": i,
				"error": "Failed to start model: " + err.Error(),
			})
			continue
		}

		// 生成响应
		response, tokens, err := h.llmService.GenerateResponse(prompt, req.MaxTokens, req.Model)
		if err != nil {
			responses = append(responses, gin.H{
				"index": i,
				"error": "Failed to generate response: " + err.Error(),
			})
			continue
		}

		responses = append(responses, gin.H{
			"index": i,
			"response": gin.H{
				"id":      fmt.Sprintf("batch-%d-%d", time.Now().Unix(), i),
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   req.Model,
				"choices": []gin.H{
					{
						"index": 0,
						"message": gin.H{
							"role":    "assistant",
							"content": response,
						},
						"finish_reason": "stop",
					},
				},
				"usage": gin.H{
					"total_tokens": tokens,
				},
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"object":    "batch.completion",
		"created":   time.Now().Unix(),
		"responses": responses,
	})
}
