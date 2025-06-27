package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type LLMService struct {
	baseURL      string
	client       *http.Client
	modelManager *ModelManager
}

func NewLLMService(baseURL string, modelManager *ModelManager) *LLMService {
	if baseURL == "" {
		baseURL = "http://localhost:8081" // 默认llama-cpp-server地址
	}

	return &LLMService{
		baseURL:      baseURL,
		modelManager: modelManager,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type LLMRequest struct {
	Prompt    string   `json:"prompt"`
	MaxTokens int      `json:"n_predict"`
	Temp      float64  `json:"temperature,omitempty"`
	TopP      float64  `json:"top_p,omitempty"`
	Stop      []string `json:"stop,omitempty"`
}

type LLMResponse struct {
	Content         string `json:"content"`
	TokensEvaluated int    `json:"tokens_evaluated"`
	TokensPredicted int    `json:"tokens_predicted"`
	Stopped         bool   `json:"stopped_eos"`
}

func (s *LLMService) GenerateResponse(message string, maxTokens int, model string) (string, int, error) {
	// 如果是模拟模式，返回模拟响应
	if s.baseURL == "mock" {
		return s.mockResponse(message, maxTokens)
	}

	// 如果指定了模型且有模型管理器，使用模型管理器
	var targetURL string
	if model != "" && s.modelManager != nil {
		// 确保模型正在运行
		if err := s.modelManager.StartModel(model); err != nil {
			log.Printf("启动模型失败: %v", err)
			return "", 0, fmt.Errorf("启动模型失败: %w", err)
		}

		// 获取模型实例
		instance, err := s.modelManager.GetModelInstance(model)
		if err != nil {
			return "", 0, fmt.Errorf("获取模型实例失败: %w", err)
		}

		targetURL = fmt.Sprintf("http://localhost:%d", instance.Port)
	} else {
		targetURL = s.baseURL
	}

	// 构建请求
	req := LLMRequest{
		Prompt:    message,
		MaxTokens: maxTokens,
		Temp:      0.7,
		TopP:      0.9,
		Stop:      []string{"\n\n", "用户:", "User:"},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", 0, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送HTTP请求
	resp, err := s.client.Post(targetURL+"/completion", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", 0, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("LLM服务返回错误 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", 0, fmt.Errorf("解析响应失败: %w", err)
	}

	// 计算实际消耗的token
	actualTokens := llmResp.TokensEvaluated + llmResp.TokensPredicted
	if actualTokens == 0 {
		// 如果服务器没有返回token信息，使用估算
		actualTokens = estimateTokens(message) + estimateTokens(llmResp.Content)
	}

	return strings.TrimSpace(llmResp.Content), actualTokens, nil
}

func (s *LLMService) mockResponse(message string, maxTokens int) (string, int, error) {
	// 模拟响应，用于测试
	responses := []string{
		"这是一个模拟的AI响应。您的问题很有趣！",
		"我理解您的问题。让我为您提供一些建议...",
		"根据您的描述，我认为可以从以下几个方面来考虑...",
		"这是一个很好的问题。基于我的理解...",
	}

	// 根据消息长度选择响应
	responseIndex := len(message) % len(responses)
	response := responses[responseIndex]

	// 模拟token消耗
	inputTokens := estimateTokens(message)
	outputTokens := estimateTokens(response)
	totalTokens := inputTokens + outputTokens

	// 添加一些随机性
	if len(message) > 50 {
		response += " 这需要更详细的分析和考虑。"
		totalTokens += 10
	}

	return response, totalTokens, nil
}

func estimateTokens(text string) int {
	// 简单的token估算：中文字符按1个token计算，英文按4个字符1个token计算
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	tokenCount := 0
	for _, char := range text {
		if char > 127 { // 非ASCII字符（主要是中文）
			tokenCount++
		} else {
			tokenCount++ // 简化处理，每个字符都算1个token
		}
	}

	// 最少1个token
	if tokenCount == 0 {
		tokenCount = 1
	}

	return tokenCount / 3 // 平均3个字符1个token
}
