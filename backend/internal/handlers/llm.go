package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"llm-backend/internal/middleware"
	"llm-backend/internal/models"
	"llm-backend/internal/services"
)

type LLMHandler struct {
	userRepo    *models.UserRepository
	apiCallRepo *models.APICallRepository
	llmService  *services.LLMService
}

func NewLLMHandler(userRepo *models.UserRepository, apiCallRepo *models.APICallRepository, llmService *services.LLMService) *LLMHandler {
	return &LLMHandler{
		userRepo:    userRepo,
		apiCallRepo: apiCallRepo,
		llmService:  llmService,
	}
}

type ChatRequest struct {
	Message   string `json:"message" binding:"required"`
	MaxTokens int    `json:"max_tokens,omitempty"`
	Model     string `json:"model,omitempty"`
}

type ChatResponse struct {
	Response       string `json:"response"`
	TokensConsumed int    `json:"tokens_consumed"`
	RemainingTokens int   `json:"remaining_tokens"`
}

func (h *LLMHandler) Chat(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置默认值
	if req.MaxTokens == 0 {
		req.MaxTokens = 100
	}
	if req.Model == "" {
		req.Model = "default"
	}

	// 估算需要消耗的token数量（简单估算：输入token + 输出token）
	inputTokens := estimateTokens(req.Message)
	outputTokens := req.MaxTokens
	totalTokens := inputTokens + outputTokens

	// 检查用户token余额
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if user.Tokens < totalTokens {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error": fmt.Sprintf("token余额不足，当前余额: %d，需要: %d", user.Tokens, totalTokens),
			"current_tokens": user.Tokens,
			"required_tokens": totalTokens,
		})
		return
	}

	// 调用LLM服务
	response, actualTokens, err := h.llmService.GenerateResponse(req.Message, req.MaxTokens, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM服务调用失败: " + err.Error()})
		return
	}

	// 扣除实际消耗的token
	err = h.userRepo.ConsumeTokens(userID, actualTokens)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "扣除token失败: " + err.Error()})
		return
	}

	// 记录API调用
	requestData, _ := json.Marshal(req)
	responseData, _ := json.Marshal(map[string]interface{}{
		"response": response,
		"tokens_consumed": actualTokens,
	})

	_, err = h.apiCallRepo.Create(userID, "/api/chat", actualTokens, string(requestData), string(responseData))
	if err != nil {
		// 记录日志但不影响响应
		fmt.Printf("Failed to record API call: %v\n", err)
	}

	// 获取更新后的token余额
	updatedUser, _ := h.userRepo.GetByID(userID)
	remainingTokens := 0
	if updatedUser != nil {
		remainingTokens = updatedUser.Tokens
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response:        response,
		TokensConsumed:  actualTokens,
		RemainingTokens: remainingTokens,
	})
}

func (h *LLMHandler) GetHistory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取分页参数
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	calls, err := h.apiCallRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"calls": calls,
		"limit": limit,
		"offset": offset,
	})
}

func (h *LLMHandler) GetStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	stats, err := h.apiCallRepo.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	// 获取当前用户信息
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	stats["current_tokens"] = user.Tokens
	c.JSON(http.StatusOK, stats)
}

// 简单的token估算函数（实际应用中应该使用更精确的方法）
func estimateTokens(text string) int {
	// 粗略估算：平均4个字符为1个token
	return len(strings.TrimSpace(text)) / 4
}