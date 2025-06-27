package handlers

import (
	"net/http"

	"llm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ModelHandler struct {
	modelManager *services.ModelManager
}

func NewModelHandler(modelManager *services.ModelManager) *ModelHandler {
	return &ModelHandler{
		modelManager: modelManager,
	}
}

// GetAvailableModels 获取可用模型列表
func (h *ModelHandler) GetAvailableModels(c *gin.Context) {
	models := h.modelManager.GetAvailableModels()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models,
	})
}

// GetRunningModels 获取正在运行的模型列表
func (h *ModelHandler) GetRunningModels(c *gin.Context) {
	models := h.modelManager.ListRunningModels()

	// 转换为更友好的格式
	result := make([]gin.H, 0)
	for name, instance := range models {
		result = append(result, gin.H{
			"name":        name,
			"port":        instance.Port,
			"status":      instance.Status,
			"start_time":  instance.StartTime,
			"last_used":   instance.LastUsed,
			"usage_count": instance.UsageCount,
			"description": instance.Config.Description,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// StartModel 启动指定模型
func (h *ModelHandler) StartModel(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "模型名称不能为空",
		})
		return
	}

	err := h.modelManager.StartModel(modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模型启动成功",
	})
}

// StopModel 停止指定模型
func (h *ModelHandler) StopModel(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "模型名称不能为空",
		})
		return
	}

	err := h.modelManager.StopModel(modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模型停止成功",
	})
}

// GetModelStatus 获取指定模型状态
func (h *ModelHandler) GetModelStatus(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "模型名称不能为空",
		})
		return
	}

	instance, err := h.modelManager.GetModelInstance(modelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"name":        modelName,
			"port":        instance.Port,
			"status":      instance.Status,
			"start_time":  instance.StartTime,
			"last_used":   instance.LastUsed,
			"usage_count": instance.UsageCount,
			"description": instance.Config.Description,
		},
	})
}

// ChatWithModel 与指定模型对话
func (h *ModelHandler) ChatWithModel(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "模型名称不能为空",
		})
		return
	}

	var req struct {
		Message   string `json:"message" binding:"required"`
		MaxTokens int    `json:"max_tokens"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.MaxTokens <= 0 {
		req.MaxTokens = 200
	}

	// 确保模型正在运行
	err := h.modelManager.StartModel(modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "启动模型失败: " + err.Error(),
		})
		return
	}

	// 获取模型实例
	instance, err := h.modelManager.GetModelInstance(modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取模型实例失败: " + err.Error(),
		})
		return
	}

	// 创建 LLM 服务并发送请求
	llmService := services.NewLLMService("", h.modelManager)
	response, tokens, err := llmService.GenerateResponse(req.Message, req.MaxTokens, modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "生成响应失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"response":    response,
			"tokens_used": tokens,
			"model":       modelName,
			"port":        instance.Port,
		},
	})
}

// GetModelMetrics 获取模型性能指标
func (h *ModelHandler) GetModelMetrics(c *gin.Context) {
	models := h.modelManager.ListRunningModels()

	metrics := make([]gin.H, 0)
	for name, instance := range models {
		metrics = append(metrics, gin.H{
			"name":           name,
			"status":         instance.Status,
			"usage_count":    instance.UsageCount,
			"uptime":         instance.StartTime,
			"last_used":      instance.LastUsed,
			"port":           instance.Port,
			"threads":        instance.Config.Threads,
			"context_length": instance.Config.ContextLength,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// RestartModel 重启指定模型
func (h *ModelHandler) RestartModel(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "模型名称不能为空",
		})
		return
	}

	// 先停止模型
	_ = h.modelManager.StopModel(modelName)

	// 再启动模型
	err := h.modelManager.StartModel(modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "重启模型失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模型重启成功",
	})
}
