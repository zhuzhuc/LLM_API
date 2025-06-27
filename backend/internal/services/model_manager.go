package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"llm-backend/internal/config"
)

type ModelInstance struct {
	Config     config.ModelConfig
	Process    *exec.Cmd
	Port       int
	Status     string // "starting", "running", "stopped", "error"
	StartTime  time.Time
	LastUsed   time.Time
	UsageCount int64
	ctx        context.Context
	cancel     context.CancelFunc
}

type ModelManager struct {
	instances    map[string]*ModelInstance
	config       *config.Config
	modelsConfig *config.ModelsConfig
	portPool     []int
	usedPorts    map[int]bool
	mu           sync.RWMutex
	basePort     int
	registry     *ServiceRegistry
}

func NewModelManager(cfg *config.Config) (*ModelManager, error) {
	modelsConfig, err := config.LoadModelsConfig(cfg.ModelConfigPath)
	if err != nil {
		return nil, fmt.Errorf("加载模型配置失败: %w", err)
	}

	// 创建端口池 (8081-8090)
	portPool := make([]int, 10)
	for i := 0; i < 10; i++ {
		portPool[i] = 8081 + i
	}

	mm := &ModelManager{
		instances:    make(map[string]*ModelInstance),
		config:       cfg,
		modelsConfig: modelsConfig,
		portPool:     portPool,
		usedPorts:    make(map[int]bool),
		basePort:     8081,
		registry:     NewServiceRegistry(),
	}

	return mm, nil
}

func (mm *ModelManager) StartModel(modelName string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// 检查模型是否已经运行
	if instance, exists := mm.instances[modelName]; exists {
		if instance.Status == "running" {
			instance.LastUsed = time.Now()
			return nil
		}
	}

	// 查找模型配置
	var modelConfig *config.ModelConfig
	for _, model := range mm.modelsConfig.Models {
		if model.ModelName == modelName && model.Active {
			modelConfig = &model
			break
		}
	}

	if modelConfig == nil {
		return fmt.Errorf("模型 %s 未找到或未激活", modelName)
	}

	// 分配端口
	port := mm.allocatePort()
	if port == 0 {
		return fmt.Errorf("无可用端口")
	}

	// 创建模型实例
	ctx, cancel := context.WithCancel(context.Background())
	instance := &ModelInstance{
		Config:    *modelConfig,
		Port:      port,
		Status:    "starting",
		StartTime: time.Now(),
		LastUsed:  time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 启动 llama-cpp-server
	modelPath := fmt.Sprintf("%s/%s", modelConfig.ModelPath, modelConfig.ModelFile)
	args := []string{
		"-m", modelPath,
		"--port", fmt.Sprintf("%d", port),
		"--host", "127.0.0.1",
		"-c", fmt.Sprintf("%d", modelConfig.ContextLength),
		"-t", fmt.Sprintf("%d", modelConfig.Threads),
	}

	// 只有当值不是默认值时才添加这些参数
	if modelConfig.Temperature != 0.8 {
		args = append(args, "--temp", fmt.Sprintf("%.2f", modelConfig.Temperature))
	}
	if modelConfig.TopP != 0.9 {
		args = append(args, "--top-p", fmt.Sprintf("%.2f", modelConfig.TopP))
	}
	if modelConfig.RepeatPenalty != 1.0 {
		args = append(args, "--repeat-penalty", fmt.Sprintf("%.2f", modelConfig.RepeatPenalty))
	}

	if modelConfig.GPULayers > 0 {
		args = append(args, "-ngl", fmt.Sprintf("%d", modelConfig.GPULayers))
	}

	// 添加调试日志
	log.Printf("启动模型 %s，命令: %s %v", modelName, mm.config.LlamaCppPath, args)

	cmd := exec.CommandContext(ctx, mm.config.LlamaCppPath, args...)
	instance.Process = cmd

	// 启动进程
	if err := cmd.Start(); err != nil {
		mm.releasePort(port)
		cancel()
		return fmt.Errorf("启动模型进程失败: %w", err)
	}

	mm.instances[modelName] = instance

	// 注册服务到服务注册中心
	serviceInstance := &ServiceInstance{
		Name: fmt.Sprintf("llm-model-%s", modelName),
		Host: "127.0.0.1",
		Port: port,
		Metadata: map[string]string{
			"model_name":     modelName,
			"model_file":     modelConfig.ModelFile,
			"context_length": fmt.Sprintf("%d", modelConfig.ContextLength),
			"threads":        fmt.Sprintf("%d", modelConfig.Threads),
		},
	}

	if err := mm.registry.Register(serviceInstance); err != nil {
		log.Printf("注册服务失败: %v", err)
	}

	// 异步监控进程状态
	go mm.monitorInstance(modelName, instance)

	log.Printf("模型 %s 正在启动，端口: %d", modelName, port)
	return nil
}

func (mm *ModelManager) StopModel(modelName string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	instance, exists := mm.instances[modelName]
	if !exists {
		return fmt.Errorf("模型 %s 未运行", modelName)
	}

	instance.cancel()
	mm.releasePort(instance.Port)

	// 从服务注册中心注销
	serviceName := fmt.Sprintf("llm-model-%s", modelName)
	if err := mm.registry.Deregister(serviceName, instance.Config.ModelName); err != nil {
		log.Printf("注销服务失败: %v", err)
	}

	delete(mm.instances, modelName)

	log.Printf("模型 %s 已停止", modelName)
	return nil
}

func (mm *ModelManager) GetModelInstance(modelName string) (*ModelInstance, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	instance, exists := mm.instances[modelName]
	if !exists || instance.Status != "running" {
		return nil, fmt.Errorf("模型 %s 未运行", modelName)
	}

	instance.LastUsed = time.Now()
	instance.UsageCount++
	return instance, nil
}

func (mm *ModelManager) ListRunningModels() map[string]*ModelInstance {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	result := make(map[string]*ModelInstance)
	for name, instance := range mm.instances {
		if instance.Status == "running" {
			result[name] = instance
		}
	}
	return result
}

func (mm *ModelManager) GetAvailableModels() []config.ModelConfig {
	var models []config.ModelConfig
	for _, model := range mm.modelsConfig.Models {
		if model.Active {
			models = append(models, model)
		}
	}
	return models
}

func (mm *ModelManager) allocatePort() int {
	for _, port := range mm.portPool {
		if !mm.usedPorts[port] {
			mm.usedPorts[port] = true
			return port
		}
	}
	return 0
}

func (mm *ModelManager) releasePort(port int) {
	delete(mm.usedPorts, port)
}

func (mm *ModelManager) monitorInstance(modelName string, instance *ModelInstance) {
	// 等待进程启动
	time.Sleep(3 * time.Second)

	// 检查进程是否还在运行
	if instance.Process.ProcessState == nil {
		instance.Status = "running"
		log.Printf("模型 %s 启动成功，端口: %d", modelName, instance.Port)
	} else {
		instance.Status = "error"
		log.Printf("模型 %s 启动失败", modelName)
		mm.mu.Lock()
		mm.releasePort(instance.Port)
		delete(mm.instances, modelName)
		mm.mu.Unlock()
		return
	}

	// 等待进程结束
	err := instance.Process.Wait()
	if err != nil {
		log.Printf("模型 %s 进程异常退出: %v", modelName, err)
	}

	mm.mu.Lock()
	instance.Status = "stopped"
	mm.releasePort(instance.Port)
	delete(mm.instances, modelName)
	mm.mu.Unlock()

	log.Printf("模型 %s 已停止", modelName)
}

func (mm *ModelManager) GetServiceRegistry() *ServiceRegistry {
	return mm.registry
}

func (mm *ModelManager) Cleanup() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for modelName, instance := range mm.instances {
		instance.cancel()
		log.Printf("清理模型实例: %s", modelName)
	}
}

// ChatWithModel 与指定模型进行对话
func (mm *ModelManager) ChatWithModel(ctx context.Context, modelName, prompt string, maxTokens int) (string, error) {
	// 获取模型实例
	instance, err := mm.GetModelInstance(modelName)
	if err != nil {
		return "", fmt.Errorf("获取模型实例失败: %w", err)
	}

	// 构建请求
	requestBody := map[string]interface{}{
		"prompt":      prompt,
		"n_predict":   maxTokens,
		"stream":      false,
		"temperature": 0.7,
		"top_p":       0.9,
		"stop":        []string{"\n\n", "###", "---"},
	}

	// 发送HTTP请求到模型服务
	url := fmt.Sprintf("http://127.0.0.1:%d/completion", instance.Port)
	return mm.sendRequest(ctx, url, requestBody)
}

// sendRequest 发送HTTP请求到模型服务
func (mm *ModelManager) sendRequest(ctx context.Context, url string, requestBody map[string]interface{}) (string, error) {
	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("模型服务返回错误 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		// 如果解析失败，尝试直接返回响应内容
		log.Printf("解析响应失败，返回原始内容: %s", string(body))
		return string(body), nil
	}

	return response.Content, nil
}
