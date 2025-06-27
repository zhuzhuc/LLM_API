package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"llm-backend/internal/middleware"
	"llm-backend/internal/models"
	"llm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *services.TaskService
	userRepo    *models.UserRepository
}

// NewTaskHandler 创建新的任务处理器
func NewTaskHandler(taskService *services.TaskService, userRepo *models.UserRepository) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		userRepo:    userRepo,
	}
}

// ConvertFileFormat 文件格式转换接口
func (h *TaskHandler) ConvertFileFormat(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req services.FileFormatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证必要参数
	if req.SourceFormat == "" || req.TargetFormat == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "源格式、目标格式和内容不能为空"})
		return
	}

	// 估算token消耗（格式转换相对简单，消耗较少）
	tokensNeeded := estimateTaskTokens(req.Content, "convert")

	// 检查用户token余额
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if user.Tokens < tokensNeeded {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":           fmt.Sprintf("token余额不足，当前余额: %d，需要: %d", user.Tokens, tokensNeeded),
			"current_tokens":  user.Tokens,
			"required_tokens": tokensNeeded,
		})
		return
	}

	// 执行格式转换
	response, err := h.taskService.ConvertFileFormat(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "格式转换失败: " + err.Error()})
		return
	}

	// 扣除token
	err = h.userRepo.ConsumeTokens(userID, tokensNeeded)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "扣除token失败: " + err.Error()})
		return
	}

	// 获取更新后的token余额
	updatedUser, _ := h.userRepo.GetByID(userID)
	remainingTokens := 0
	if updatedUser != nil {
		remainingTokens = updatedUser.Tokens
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           response.Success,
		"converted_content": response.ConvertedContent,
		"tokens_consumed":   tokensNeeded,
		"remaining_tokens":  remainingTokens,
	})
}

// estimateTaskTokens 估算任务所需的token数量
func estimateTaskTokens(content string, taskType string) int {
	baseTokens := len(strings.TrimSpace(content)) / 4 // 基础token估算

	switch taskType {
	case "convert":
		// 格式转换相对简单，基础消耗 + 20%
		return int(float64(baseTokens)*1.2) + 5
	case "homework":
		// 作业批改需要分析和生成反馈，消耗较多
		return int(float64(baseTokens)*2.0) + 20
	case "subtitle":
		// 字幕处理复杂，消耗最多
		return int(float64(baseTokens)*3.0) + 50
	default:
		return baseTokens + 10
	}
}

// GradeHomework 作业批改接口
func (h *TaskHandler) GradeHomework(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req services.HomeworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证必要参数
	if req.Subject == "" || req.Question == "" || req.Answer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "科目、题目和答案不能为空"})
		return
	}

	// 设置默认值
	if req.Language == "" {
		req.Language = "中文"
	}
	if req.GradeLevel == "" {
		req.GradeLevel = "中学"
	}

	// 估算token消耗（作业批改需要分析和生成反馈）
	content := req.Question + " " + req.Answer
	tokensNeeded := estimateTaskTokens(content, "homework")

	// 检查用户token余额
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if user.Tokens < tokensNeeded {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":           fmt.Sprintf("token余额不足，当前余额: %d，需要: %d", user.Tokens, tokensNeeded),
			"current_tokens":  user.Tokens,
			"required_tokens": tokensNeeded,
		})
		return
	}

	// 执行作业批改
	response, err := h.taskService.GradeHomework(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "作业批改失败: " + err.Error()})
		return
	}

	// 扣除token
	err = h.userRepo.ConsumeTokens(userID, tokensNeeded)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "扣除token失败: " + err.Error()})
		return
	}

	// 获取更新后的token余额
	updatedUser, _ := h.userRepo.GetByID(userID)
	remainingTokens := 0
	if updatedUser != nil {
		remainingTokens = updatedUser.Tokens
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          response.Success,
		"score":            response.Score,
		"feedback":         response.Feedback,
		"suggestions":      response.Suggestions,
		"tokens_consumed":  tokensNeeded,
		"remaining_tokens": remainingTokens,
	})
}

// ProcessSubtitle 字幕处理接口
func (h *TaskHandler) ProcessSubtitle(c *gin.Context) {
	var req services.SubtitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证必要参数
	if req.VideoPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "视频路径不能为空"})
		return
	}

	// 设置默认值
	if req.SourceLang == "" {
		req.SourceLang = "英文"
	}
	if req.TargetLang == "" {
		req.TargetLang = "中文"
	}
	if req.OutputFormat == "" {
		req.OutputFormat = "srt"
	}

	response, err := h.taskService.ProcessSubtitle(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "字幕处理失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSupportedFormats 获取支持的格式列表
func (h *TaskHandler) GetSupportedFormats(c *gin.Context) {
	formats := map[string][]string{
		"document": {"txt", "md", "html", "json", "xml", "csv", "yaml"},
		"code":     {"py", "js", "go", "java", "cpp", "c", "php", "rb", "rs"},
		"data":     {"json", "xml", "csv", "yaml", "toml", "ini"},
		"markup":   {"html", "xml", "md", "rst", "tex"},
	}

	c.JSON(http.StatusOK, gin.H{
		"supported_formats": formats,
		"message":           "支持的文件格式列表",
	})
}

// GetTaskStatus 获取任务状态
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 这里可以实现任务状态查询逻辑
	// 目前返回模拟数据
	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  "completed",
		"message": "任务已完成",
	})
}

// UploadFile 文件上传接口
func (h *TaskHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}

	// 保存文件到临时目录
	uploadPath := "./uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "文件上传成功",
		"filename":  file.Filename,
		"file_path": uploadPath,
		"size":      file.Size,
	})
}
