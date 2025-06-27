package services

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Status      string            `json:"status"` // "healthy", "unhealthy", "starting"
	Metadata    map[string]string `json:"metadata"`
	LastCheck   time.Time         `json:"last_check"`
	RegisterTime time.Time        `json:"register_time"`
	FailCount   int               `json:"fail_count"`
}

// ServiceRegistry 服务注册中心
type ServiceRegistry struct {
	services map[string][]*ServiceInstance // serviceName -> instances
	mu       sync.RWMutex
	client   *http.Client
}

// NewServiceRegistry 创建服务注册中心
func NewServiceRegistry() *ServiceRegistry {
	sr := &ServiceRegistry{
		services: make(map[string][]*ServiceInstance),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// 启动健康检查协程
	go sr.startHealthCheck()
	
	return sr
}

// Register 注册服务实例
func (sr *ServiceRegistry) Register(instance *ServiceInstance) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if instance.ID == "" {
		instance.ID = fmt.Sprintf("%s-%s-%d-%d", 
			instance.Name, instance.Host, instance.Port, time.Now().Unix())
	}

	instance.RegisterTime = time.Now()
	instance.LastCheck = time.Now()
	instance.Status = "starting"
	instance.FailCount = 0

	if sr.services[instance.Name] == nil {
		sr.services[instance.Name] = make([]*ServiceInstance, 0)
	}

	// 检查是否已存在相同的实例
	for i, existing := range sr.services[instance.Name] {
		if existing.Host == instance.Host && existing.Port == instance.Port {
			// 更新现有实例
			sr.services[instance.Name][i] = instance
			log.Printf("服务实例已更新: %s (%s:%d)", instance.Name, instance.Host, instance.Port)
			return nil
		}
	}

	// 添加新实例
	sr.services[instance.Name] = append(sr.services[instance.Name], instance)
	log.Printf("服务实例已注册: %s (%s:%d)", instance.Name, instance.Host, instance.Port)
	
	return nil
}

// Deregister 注销服务实例
func (sr *ServiceRegistry) Deregister(serviceName, instanceID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instances := sr.services[serviceName]
	for i, instance := range instances {
		if instance.ID == instanceID {
			// 移除实例
			sr.services[serviceName] = append(instances[:i], instances[i+1:]...)
			log.Printf("服务实例已注销: %s (%s)", serviceName, instanceID)
			
			// 如果没有实例了，删除服务
			if len(sr.services[serviceName]) == 0 {
				delete(sr.services, serviceName)
			}
			return nil
		}
	}

	return fmt.Errorf("服务实例未找到: %s/%s", serviceName, instanceID)
}

// Discover 发现服务实例
func (sr *ServiceRegistry) Discover(serviceName string) ([]*ServiceInstance, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	instances := sr.services[serviceName]
	if len(instances) == 0 {
		return nil, fmt.Errorf("服务未找到: %s", serviceName)
	}

	// 只返回健康的实例
	var healthyInstances []*ServiceInstance
	for _, instance := range instances {
		if instance.Status == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("没有健康的服务实例: %s", serviceName)
	}

	return healthyInstances, nil
}

// GetInstance 获取单个服务实例（负载均衡）
func (sr *ServiceRegistry) GetInstance(serviceName string, strategy string) (*ServiceInstance, error) {
	instances, err := sr.Discover(serviceName)
	if err != nil {
		return nil, err
	}

	switch strategy {
	case "random":
		return sr.randomSelect(instances), nil
	case "round_robin":
		return sr.roundRobinSelect(serviceName, instances), nil
	case "least_connections":
		return sr.leastConnectionsSelect(instances), nil
	default:
		return sr.randomSelect(instances), nil
	}
}

// randomSelect 随机选择
func (sr *ServiceRegistry) randomSelect(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	return instances[rand.Intn(len(instances))]
}

// roundRobinSelect 轮询选择
var roundRobinCounters = make(map[string]int)
var roundRobinMu sync.Mutex

func (sr *ServiceRegistry) roundRobinSelect(serviceName string, instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	roundRobinMu.Lock()
	defer roundRobinMu.Unlock()

	counter := roundRobinCounters[serviceName]
	instance := instances[counter%len(instances)]
	roundRobinCounters[serviceName] = counter + 1

	return instance
}

// leastConnectionsSelect 最少连接选择
func (sr *ServiceRegistry) leastConnectionsSelect(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	// 简化实现：选择失败次数最少的实例
	minFails := instances[0].FailCount
	selected := instances[0]

	for _, instance := range instances[1:] {
		if instance.FailCount < minFails {
			minFails = instance.FailCount
			selected = instance
		}
	}

	return selected
}

// GetAllServices 获取所有服务
func (sr *ServiceRegistry) GetAllServices() map[string][]*ServiceInstance {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make(map[string][]*ServiceInstance)
	for name, instances := range sr.services {
		result[name] = make([]*ServiceInstance, len(instances))
		copy(result[name], instances)
	}

	return result
}

// startHealthCheck 启动健康检查
func (sr *ServiceRegistry) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sr.performHealthCheck()
	}
}

// performHealthCheck 执行健康检查
func (sr *ServiceRegistry) performHealthCheck() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for serviceName, instances := range sr.services {
		for _, instance := range instances {
			go sr.checkInstanceHealth(serviceName, instance)
		}
	}
}

// checkInstanceHealth 检查实例健康状态
func (sr *ServiceRegistry) checkInstanceHealth(serviceName string, instance *ServiceInstance) {
	healthURL := fmt.Sprintf("http://%s:%d/health", instance.Host, instance.Port)
	
	resp, err := sr.client.Get(healthURL)
	if err != nil {
		sr.markInstanceUnhealthy(serviceName, instance, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		sr.markInstanceHealthy(serviceName, instance)
	} else {
		sr.markInstanceUnhealthy(serviceName, instance, fmt.Errorf("健康检查返回状态码: %d", resp.StatusCode))
	}
}

// markInstanceHealthy 标记实例为健康
func (sr *ServiceRegistry) markInstanceHealthy(serviceName string, instance *ServiceInstance) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if instance.Status != "healthy" {
		log.Printf("服务实例恢复健康: %s (%s:%d)", serviceName, instance.Host, instance.Port)
	}

	instance.Status = "healthy"
	instance.LastCheck = time.Now()
	instance.FailCount = 0
}

// markInstanceUnhealthy 标记实例为不健康
func (sr *ServiceRegistry) markInstanceUnhealthy(serviceName string, instance *ServiceInstance, err error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instance.FailCount++
	instance.LastCheck = time.Now()

	if instance.FailCount >= 3 {
		if instance.Status != "unhealthy" {
			log.Printf("服务实例标记为不健康: %s (%s:%d) - %v", serviceName, instance.Host, instance.Port, err)
		}
		instance.Status = "unhealthy"
	}
}

// GetServiceStats 获取服务统计信息
func (sr *ServiceRegistry) GetServiceStats() map[string]interface{} {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	stats := make(map[string]interface{})
	totalServices := len(sr.services)
	totalInstances := 0
	healthyInstances := 0

	serviceDetails := make(map[string]interface{})

	for serviceName, instances := range sr.services {
		totalInstances += len(instances)
		
		healthy := 0
		unhealthy := 0
		starting := 0

		for _, instance := range instances {
			switch instance.Status {
			case "healthy":
				healthy++
				healthyInstances++
			case "unhealthy":
				unhealthy++
			case "starting":
				starting++
			}
		}

		serviceDetails[serviceName] = map[string]interface{}{
			"total":     len(instances),
			"healthy":   healthy,
			"unhealthy": unhealthy,
			"starting":  starting,
		}
	}

	stats["total_services"] = totalServices
	stats["total_instances"] = totalInstances
	stats["healthy_instances"] = healthyInstances
	stats["unhealthy_instances"] = totalInstances - healthyInstances
	stats["services"] = serviceDetails

	return stats
}
