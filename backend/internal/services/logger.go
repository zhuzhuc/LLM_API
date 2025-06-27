package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String 返回日志级别字符串
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Source    string                 `json:"source,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// Logger 日志器
type Logger struct {
	level      LogLevel
	outputs    []io.Writer
	mu         sync.Mutex
	fields     map[string]interface{}
	source     string
	bufferSize int
	buffer     []*LogEntry
	bufferMu   sync.Mutex
}

// NewLogger 创建新的日志器
func NewLogger(level LogLevel, outputs ...io.Writer) *Logger {
	if len(outputs) == 0 {
		outputs = []io.Writer{os.Stdout}
	}

	return &Logger{
		level:      level,
		outputs:    outputs,
		fields:     make(map[string]interface{}),
		bufferSize: 1000,
		buffer:     make([]*LogEntry, 0, 1000),
	}
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := *l
	newLogger.fields = make(map[string]interface{})
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return &newLogger
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := *l
	newLogger.fields = make(map[string]interface{})
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return &newLogger
}

// WithSource 设置来源
func (l *Logger) WithSource(source string) *Logger {
	newLogger := *l
	newLogger.source = source
	return &newLogger
}

// log 记录日志
func (l *Logger) log(level LogLevel, message string) {
	if level < l.level {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    l.fields,
		Source:    l.source,
	}

	// 添加到缓冲区
	l.bufferMu.Lock()
	if len(l.buffer) >= l.bufferSize {
		// 移除最老的条目
		l.buffer = l.buffer[1:]
	}
	l.buffer = append(l.buffer, entry)
	l.bufferMu.Unlock()

	// 输出日志
	l.mu.Lock()
	defer l.mu.Unlock()

	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	for _, output := range l.outputs {
		fmt.Fprintln(output, string(jsonData))
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(message string) {
	l.log(LogLevelDebug, message)
}

// Info 记录信息日志
func (l *Logger) Info(message string) {
	l.log(LogLevelInfo, message)
}

// Warn 记录警告日志
func (l *Logger) Warn(message string) {
	l.log(LogLevelWarn, message)
}

// Error 记录错误日志
func (l *Logger) Error(message string) {
	l.log(LogLevelError, message)
}

// Fatal 记录致命错误日志
func (l *Logger) Fatal(message string) {
	l.log(LogLevelFatal, message)
	os.Exit(1)
}

// GetRecentLogs 获取最近的日志
func (l *Logger) GetRecentLogs(limit int) []*LogEntry {
	l.bufferMu.Lock()
	defer l.bufferMu.Unlock()

	if limit <= 0 || limit > len(l.buffer) {
		limit = len(l.buffer)
	}

	start := len(l.buffer) - limit
	result := make([]*LogEntry, limit)
	copy(result, l.buffer[start:])
	return result
}

// LogManager 日志管理器
type LogManager struct {
	loggers map[string]*Logger
	mu      sync.RWMutex
	logDir  string
}

// NewLogManager 创建日志管理器
func NewLogManager(logDir string) *LogManager {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
	}

	return &LogManager{
		loggers: make(map[string]*Logger),
		logDir:  logDir,
	}
}

// GetLogger 获取或创建日志器
func (lm *LogManager) GetLogger(name string, level LogLevel) *Logger {
	lm.mu.RLock()
	logger, exists := lm.loggers[name]
	lm.mu.RUnlock()

	if exists {
		return logger
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 双重检查
	if logger, exists := lm.loggers[name]; exists {
		return logger
	}

	// 创建日志文件
	logFile := filepath.Join(lm.logDir, fmt.Sprintf("%s.log", name))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file %s: %v", logFile, err)
		// 回退到标准输出
		logger = NewLogger(level, os.Stdout)
	} else {
		// 同时输出到文件和标准输出
		logger = NewLogger(level, file, os.Stdout)
	}

	logger = logger.WithSource(name)
	lm.loggers[name] = logger
	return logger
}

// LoggingMiddleware 日志中间件
func (lm *LogManager) LoggingMiddleware() gin.HandlerFunc {
	logger := lm.GetLogger("http", LogLevelInfo)

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 生成请求ID
		requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
		c.Set("request_id", requestID)

		// 记录请求开始
		logger.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("HTTP request started")

		c.Next()

		// 记录请求完成
		latency := time.Since(start)
		status := c.Writer.Status()
		
		logLevel := LogLevelInfo
		if status >= 400 {
			logLevel = LogLevelError
		}

		logger.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"status":     status,
			"latency":    latency.String(),
			"size":       c.Writer.Size(),
		}).log(logLevel, "HTTP request completed")
	}
}

// LogHandler 日志处理器
type LogHandler struct {
	manager *LogManager
}

// NewLogHandler 创建日志处理器
func NewLogHandler(manager *LogManager) *LogHandler {
	return &LogHandler{
		manager: manager,
	}
}

// GetLogs 获取日志
func (h *LogHandler) GetLogs(c *gin.Context) {
	loggerName := c.DefaultQuery("logger", "http")
	limitStr := c.DefaultQuery("limit", "100")
	
	limit := 100
	if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
		limit = 100
	}

	logger := h.manager.GetLogger(loggerName, LogLevelDebug)
	logs := logger.GetRecentLogs(limit)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"logger": loggerName,
			"logs":   logs,
			"count":  len(logs),
		},
	})
}

// GetLoggers 获取所有日志器
func (h *LogHandler) GetLoggers(c *gin.Context) {
	h.manager.mu.RLock()
	defer h.manager.mu.RUnlock()

	loggers := make([]string, 0, len(h.manager.loggers))
	for name := range h.manager.loggers {
		loggers = append(loggers, name)
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    loggers,
	})
}

// WriteLog 写入日志
func (h *LogHandler) WriteLog(c *gin.Context) {
	var req struct {
		Logger  string                 `json:"logger" binding:"required"`
		Level   string                 `json:"level" binding:"required"`
		Message string                 `json:"message" binding:"required"`
		Fields  map[string]interface{} `json:"fields"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析日志级别
	var level LogLevel
	switch req.Level {
	case "DEBUG":
		level = LogLevelDebug
	case "INFO":
		level = LogLevelInfo
	case "WARN":
		level = LogLevelWarn
	case "ERROR":
		level = LogLevelError
	case "FATAL":
		level = LogLevelFatal
	default:
		c.JSON(400, gin.H{
			"success": false,
			"error":   "无效的日志级别",
			"valid_levels": []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		})
		return
	}

	logger := h.manager.GetLogger(req.Logger, LogLevelDebug)
	if req.Fields != nil {
		logger = logger.WithFields(req.Fields)
	}

	logger.log(level, req.Message)

	c.JSON(200, gin.H{
		"success": true,
		"message": "日志写入成功",
	})
}

// 全局日志管理器
var globalLogManager *LogManager
var logManagerOnce sync.Once

// GetGlobalLogManager 获取全局日志管理器
func GetGlobalLogManager() *LogManager {
	logManagerOnce.Do(func() {
		globalLogManager = NewLogManager("./logs")
		log.Println("日志管理器已初始化")
	})
	return globalLogManager
}
