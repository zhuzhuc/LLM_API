package services

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricType 指标类型
type MetricType string

const (
	MetricCounter   MetricType = "counter"
	MetricGauge     MetricType = "gauge"
	MetricHistogram MetricType = "histogram"
	MetricSummary   MetricType = "summary"
)

// Metric 指标结构
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels"`
	Timestamp time.Time              `json:"timestamp"`
	Help      string                 `json:"help"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		metrics: make(map[string]*Metric),
	}
	
	// 启动系统指标收集
	go mc.collectSystemMetrics()
	
	return mc
}

// RecordMetric 记录指标
func (mc *MetricsCollector) RecordMetric(name string, metricType MetricType, value float64, labels map[string]string, help string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.generateKey(name, labels)
	
	metric := &Metric{
		Name:      name,
		Type:      metricType,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
		Help:      help,
	}

	// 对于计数器类型，累加值
	if metricType == MetricCounter {
		if existing, exists := mc.metrics[key]; exists {
			metric.Value = existing.Value + value
		}
	}

	mc.metrics[key] = metric
}

// IncrementCounter 增加计数器
func (mc *MetricsCollector) IncrementCounter(name string, labels map[string]string, help string) {
	mc.RecordMetric(name, MetricCounter, 1, labels, help)
}

// SetGauge 设置仪表盘值
func (mc *MetricsCollector) SetGauge(name string, value float64, labels map[string]string, help string) {
	mc.RecordMetric(name, MetricGauge, value, labels, help)
}

// RecordHistogram 记录直方图
func (mc *MetricsCollector) RecordHistogram(name string, value float64, labels map[string]string, help string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.generateKey(name, labels)
	
	metric := &Metric{
		Name:      name,
		Type:      MetricHistogram,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
		Help:      help,
		Extra:     make(map[string]interface{}),
	}

	// 如果已存在，更新统计信息
	if existing, exists := mc.metrics[key]; exists {
		if existing.Extra == nil {
			existing.Extra = make(map[string]interface{})
		}
		
		count := existing.Extra["count"].(float64) + 1
		sum := existing.Extra["sum"].(float64) + value
		
		existing.Extra["count"] = count
		existing.Extra["sum"] = sum
		existing.Extra["avg"] = sum / count
		existing.Value = value
		existing.Timestamp = time.Now()
		
		mc.metrics[key] = existing
	} else {
		metric.Extra["count"] = float64(1)
		metric.Extra["sum"] = value
		metric.Extra["avg"] = value
		mc.metrics[key] = metric
	}
}

// GetMetrics 获取所有指标
func (mc *MetricsCollector) GetMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*Metric)
	for key, metric := range mc.metrics {
		result[key] = metric
	}
	return result
}

// GetMetricsByName 根据名称获取指标
func (mc *MetricsCollector) GetMetricsByName(name string) []*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var result []*Metric
	for _, metric := range mc.metrics {
		if metric.Name == name {
			result = append(result, metric)
		}
	}
	return result
}

// generateKey 生成指标键
func (mc *MetricsCollector) generateKey(name string, labels map[string]string) string {
	key := name
	if labels != nil {
		labelBytes, _ := json.Marshal(labels)
		key += string(labelBytes)
	}
	return key
}

// collectSystemMetrics 收集系统指标
func (mc *MetricsCollector) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// 内存指标
		mc.SetGauge("system_memory_alloc_bytes", float64(m.Alloc), nil, "当前分配的内存字节数")
		mc.SetGauge("system_memory_total_alloc_bytes", float64(m.TotalAlloc), nil, "总分配的内存字节数")
		mc.SetGauge("system_memory_sys_bytes", float64(m.Sys), nil, "系统内存字节数")
		mc.SetGauge("system_memory_heap_alloc_bytes", float64(m.HeapAlloc), nil, "堆分配的内存字节数")
		mc.SetGauge("system_memory_heap_sys_bytes", float64(m.HeapSys), nil, "堆系统内存字节数")

		// GC 指标
		mc.SetGauge("system_gc_num", float64(m.NumGC), nil, "GC 次数")
		mc.SetGauge("system_gc_pause_total_ns", float64(m.PauseTotalNs), nil, "GC 暂停总时间（纳秒）")

		// Goroutine 指标
		mc.SetGauge("system_goroutines", float64(runtime.NumGoroutine()), nil, "当前 Goroutine 数量")
		mc.SetGauge("system_cpu_num", float64(runtime.NumCPU()), nil, "CPU 核心数")
	}
}

// MonitoringMiddleware 监控中间件
func (mc *MetricsCollector) MonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 记录请求开始
		mc.IncrementCounter("http_requests_total", map[string]string{
			"method": c.Request.Method,
			"path":   c.FullPath(),
		}, "HTTP 请求总数")

		c.Next()

		// 记录请求完成
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		mc.RecordHistogram("http_request_duration_seconds", duration, map[string]string{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"status": status,
		}, "HTTP 请求持续时间")

		mc.IncrementCounter("http_responses_total", map[string]string{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"status": status,
		}, "HTTP 响应总数")

		// 记录错误
		if c.Writer.Status() >= 400 {
			mc.IncrementCounter("http_errors_total", map[string]string{
				"method": c.Request.Method,
				"path":   c.FullPath(),
				"status": status,
			}, "HTTP 错误总数")
		}
	}
}

// MonitoringHandler 监控处理器
type MonitoringHandler struct {
	collector *MetricsCollector
}

// NewMonitoringHandler 创建监控处理器
func NewMonitoringHandler(collector *MetricsCollector) *MonitoringHandler {
	return &MonitoringHandler{
		collector: collector,
	}
}

// GetMetrics 获取指标接口
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	metrics := h.collector.GetMetrics()
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    metrics,
		"count":   len(metrics),
	})
}

// GetMetricsByName 根据名称获取指标
func (h *MonitoringHandler) GetMetricsByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "指标名称不能为空",
		})
		return
	}

	metrics := h.collector.GetMetricsByName(name)
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    metrics,
		"count":   len(metrics),
	})
}

// GetSystemStats 获取系统统计信息
func (h *MonitoringHandler) GetSystemStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := gin.H{
		"memory": gin.H{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"heap_alloc":  m.HeapAlloc,
			"heap_sys":    m.HeapSys,
		},
		"gc": gin.H{
			"num_gc":         m.NumGC,
			"pause_total_ns": m.PauseTotalNs,
			"last_gc":        time.Unix(0, int64(m.LastGC)),
		},
		"runtime": gin.H{
			"goroutines": runtime.NumGoroutine(),
			"cpu_num":    runtime.NumCPU(),
			"version":    runtime.Version(),
		},
		"timestamp": time.Now(),
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    stats,
	})
}

// RecordCustomMetric 记录自定义指标
func (h *MonitoringHandler) RecordCustomMetric(c *gin.Context) {
	var req struct {
		Name   string            `json:"name" binding:"required"`
		Type   string            `json:"type" binding:"required"`
		Value  float64           `json:"value" binding:"required"`
		Labels map[string]string `json:"labels"`
		Help   string            `json:"help"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	metricType := MetricType(req.Type)
	validTypes := map[MetricType]bool{
		MetricCounter:   true,
		MetricGauge:     true,
		MetricHistogram: true,
		MetricSummary:   true,
	}

	if !validTypes[metricType] {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "无效的指标类型",
			"valid_types": []string{"counter", "gauge", "histogram", "summary"},
		})
		return
	}

	h.collector.RecordMetric(req.Name, metricType, req.Value, req.Labels, req.Help)

	c.JSON(200, gin.H{
		"success": true,
		"message": "指标记录成功",
	})
}

// 全局指标收集器
var globalMetricsCollector *MetricsCollector
var metricsOnce sync.Once

// GetGlobalMetricsCollector 获取全局指标收集器
func GetGlobalMetricsCollector() *MetricsCollector {
	metricsOnce.Do(func() {
		globalMetricsCollector = NewMetricsCollector()
		log.Println("指标收集器已初始化")
	})
	return globalMetricsCollector
}
