package services

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	registry *ServiceRegistry
	strategy string // "random", "round_robin", "least_connections"
	proxies  map[string]*httputil.ReverseProxy
	mu       sync.RWMutex
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(registry *ServiceRegistry, strategy string) *LoadBalancer {
	if strategy == "" {
		strategy = "round_robin"
	}

	return &LoadBalancer{
		registry: registry,
		strategy: strategy,
		proxies:  make(map[string]*httputil.ReverseProxy),
	}
}

// ProxyRequest 代理请求到服务实例
func (lb *LoadBalancer) ProxyRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取服务实例
		instance, err := lb.registry.GetInstance(serviceName, lb.strategy)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error":   fmt.Sprintf("服务不可用: %s - %v", serviceName, err),
				"code":    "SERVICE_UNAVAILABLE",
			})
			return
		}

		// 创建或获取代理
		proxyKey := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
		proxy := lb.getOrCreateProxy(proxyKey, instance)

		// 设置请求头
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Host)
		c.Request.Header.Set("X-Forwarded-Proto", "http")
		c.Request.Header.Set("X-Real-IP", c.ClientIP())

		// 代理请求
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// getOrCreateProxy 获取或创建代理
func (lb *LoadBalancer) getOrCreateProxy(proxyKey string, instance *ServiceInstance) *httputil.ReverseProxy {
	lb.mu.RLock()
	proxy, exists := lb.proxies[proxyKey]
	lb.mu.RUnlock()

	if exists {
		return proxy
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	// 双重检查
	if proxy, exists := lb.proxies[proxyKey]; exists {
		return proxy
	}

	// 创建新代理
	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", instance.Host, instance.Port),
	}

	proxy = httputil.NewSingleHostReverseProxy(targetURL)
	
	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("代理错误: %s -> %s: %v", r.URL.Path, targetURL.String(), err)
		
		// 标记实例为不健康
		lb.registry.markInstanceUnhealthy(instance.Name, instance, err)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"success": false, "error": "服务暂时不可用", "code": "PROXY_ERROR"}`))
	}

	// 自定义请求修改
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// 添加追踪头
		req.Header.Set("X-Request-ID", generateRequestID())
		req.Header.Set("X-Forwarded-Time", time.Now().Format(time.RFC3339))
		req.Header.Set("X-Service-Instance", instance.ID)
	}

	lb.proxies[proxyKey] = proxy
	return proxy
}

// HealthCheckMiddleware 健康检查中间件
func (lb *LoadBalancer) HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			// 检查所有服务的健康状态
			stats := lb.registry.GetServiceStats()
			
			status := "healthy"
			if stats["healthy_instances"].(int) == 0 {
				status = "unhealthy"
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": status,
					"stats":  stats,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": status,
				"stats":  stats,
			})
			return
		}
		
		c.Next()
	}
}

// ServiceDiscoveryHandler 服务发现处理器
type ServiceDiscoveryHandler struct {
	registry     *ServiceRegistry
	loadBalancer *LoadBalancer
}

// NewServiceDiscoveryHandler 创建服务发现处理器
func NewServiceDiscoveryHandler(registry *ServiceRegistry, loadBalancer *LoadBalancer) *ServiceDiscoveryHandler {
	return &ServiceDiscoveryHandler{
		registry:     registry,
		loadBalancer: loadBalancer,
	}
}

// RegisterService 注册服务
func (h *ServiceDiscoveryHandler) RegisterService(c *gin.Context) {
	var instance ServiceInstance
	if err := c.ShouldBindJSON(&instance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.registry.Register(&instance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "注册服务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务注册成功",
		"instance_id": instance.ID,
	})
}

// DeregisterService 注销服务
func (h *ServiceDiscoveryHandler) DeregisterService(c *gin.Context) {
	serviceName := c.Param("service")
	instanceID := c.Param("instance")

	if err := h.registry.Deregister(serviceName, instanceID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务注销成功",
	})
}

// DiscoverServices 发现服务
func (h *ServiceDiscoveryHandler) DiscoverServices(c *gin.Context) {
	serviceName := c.Param("service")
	
	if serviceName == "" {
		// 返回所有服务
		services := h.registry.GetAllServices()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    services,
		})
		return
	}

	// 返回指定服务的实例
	instances, err := h.registry.Discover(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    instances,
	})
}

// GetServiceStats 获取服务统计
func (h *ServiceDiscoveryHandler) GetServiceStats(c *gin.Context) {
	stats := h.registry.GetServiceStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// SetLoadBalancingStrategy 设置负载均衡策略
func (h *ServiceDiscoveryHandler) SetLoadBalancingStrategy(c *gin.Context) {
	var req struct {
		Strategy string `json:"strategy" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	validStrategies := map[string]bool{
		"random":            true,
		"round_robin":       true,
		"least_connections": true,
	}

	if !validStrategies[req.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的负载均衡策略",
			"valid_strategies": []string{"random", "round_robin", "least_connections"},
		})
		return
	}

	h.loadBalancer.strategy = req.Strategy

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "负载均衡策略已更新",
		"strategy": req.Strategy,
	})
}

// GetLoadBalancingStrategy 获取当前负载均衡策略
func (h *ServiceDiscoveryHandler) GetLoadBalancingStrategy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategy": h.loadBalancer.strategy,
		},
	})
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// CleanupInactiveProxies 清理不活跃的代理
func (lb *LoadBalancer) CleanupInactiveProxies() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// 获取当前活跃的实例
	activeProxies := make(map[string]bool)
	for _, instances := range lb.registry.GetAllServices() {
		for _, instance := range instances {
			if instance.Status == "healthy" {
				proxyKey := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
				activeProxies[proxyKey] = true
			}
		}
	}

	// 删除不活跃的代理
	for proxyKey := range lb.proxies {
		if !activeProxies[proxyKey] {
			delete(lb.proxies, proxyKey)
			log.Printf("清理不活跃的代理: %s", proxyKey)
		}
	}
}

// StartCleanupRoutine 启动清理协程
func (lb *LoadBalancer) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			lb.CleanupInactiveProxies()
		}
	}()
}
