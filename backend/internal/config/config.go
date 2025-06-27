package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	Environment      string
	TokenRate        float64 // 每个token的价格
	DefaultTokens    int     // 新用户默认token数量
	LlamaCppPath     string  // llama.cpp 可执行文件路径
	ModelsPath       string  // 模型文件目录
	ModelConfigPath  string  // 模型配置文件路径
	ServerPort       string  // API 服务器端口
	LlamaCppPort     string  // llama.cpp 服务器端口
	KeycloakURL      string  // Keycloak 服务器地址
	KeycloakRealm    string  // Keycloak realm
	KeycloakClientID string  // Keycloak client ID
}

type ModelConfig struct {
	ModelName     string  `json:"modelName"`
	ModelFile     string  `json:"modelFile"`
	ModelPath     string  `json:"modelPath"`
	ContextLength int     `json:"contextLength"`
	MaxTokens     int     `json:"maxTokens"`
	Temperature   float64 `json:"temperature"`
	TopP          float64 `json:"topP"`
	RepeatPenalty float64 `json:"repeatPenalty"`
	Threads       int     `json:"threads"`
	GPULayers     int     `json:"gpuLayers"`
	Active        bool    `json:"active"`
	Description   string  `json:"description"`
}

type ModelsConfig struct {
	Models []ModelConfig `json:"models"`
}

func Load() *Config {
	tokenRate, _ := strconv.ParseFloat(getEnv("TOKEN_RATE", "0.001"), 64)
	defaultTokens, _ := strconv.Atoi(getEnv("DEFAULT_TOKENS", "1000"))

	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "sqlite3://./llm.db"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		TokenRate:        tokenRate,
		DefaultTokens:    defaultTokens,
		LlamaCppPath:     getEnv("LLAMA_CPP_PATH", "../llama.cpp/build/bin/llama-server"),
		ModelsPath:       getEnv("MODELS_PATH", "../models"),
		ModelConfigPath:  getEnv("MODEL_CONFIG_PATH", "../models/model_config.json"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		LlamaCppPort:     getEnv("LLAMA_CPP_PORT", "8081"),
		KeycloakURL:      getEnv("KEYCLOAK_URL", ""),
		KeycloakRealm:    getEnv("KEYCLOAK_REALM", ""),
		KeycloakClientID: getEnv("KEYCLOAK_CLIENT_ID", ""),
	}
}

func LoadModelsConfig(configPath string) (*ModelsConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取模型配置文件失败: %w", err)
	}

	var config ModelsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析模型配置文件失败: %w", err)
	}

	return &config, nil
}

func SaveModelsConfig(configPath string, config *ModelsConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化模型配置失败: %w", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("保存模型配置文件失败: %w", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
