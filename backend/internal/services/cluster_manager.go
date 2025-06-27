package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusActive   NodeStatus = "active"
	NodeStatusInactive NodeStatus = "inactive"
	NodeStatusJoining  NodeStatus = "joining"
	NodeStatusLeaving  NodeStatus = "leaving"
	NodeStatusFailed   NodeStatus = "failed"
)

// ClusterNode 集群节点
type ClusterNode struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Status      NodeStatus        `json:"status"`
	Role        string            `json:"role"` // "master", "worker", "gateway"
	Metadata    map[string]string `json:"metadata"`
	LastSeen    time.Time         `json:"last_seen"`
	JoinTime    time.Time         `json:"join_time"`
	Load        float64           `json:"load"`         // 负载指标 0-1
	Capacity    int               `json:"capacity"`     // 容量
	ActiveTasks int               `json:"active_tasks"` // 活跃任务数
}

// ClusterManager 集群管理器
type ClusterManager struct {
	nodes       map[string]*ClusterNode
	mu          sync.RWMutex
	nodeID      string
	currentNode *ClusterNode
	registry    *ServiceRegistry
	client      *http.Client
}

// NewClusterManager 创建集群管理器
func NewClusterManager(nodeID, host string, port int, registry *ServiceRegistry) *ClusterManager {
	cm := &ClusterManager{
		nodes:    make(map[string]*ClusterNode),
		nodeID:   nodeID,
		registry: registry,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// 创建当前节点
	cm.currentNode = &ClusterNode{
		ID:       nodeID,
		Name:     fmt.Sprintf("node-%s", nodeID),
		Host:     host,
		Port:     port,
		Status:   NodeStatusActive,
		Role:     "worker",
		Metadata: make(map[string]string),
		JoinTime: time.Now(),
		LastSeen: time.Now(),
		Load:     0.0,
		Capacity: 100,
	}

	cm.nodes[nodeID] = cm.currentNode

	// 启动心跳和健康检查
	go cm.startHeartbeat()
	go cm.startHealthCheck()

	return cm
}

// JoinCluster 加入集群
func (cm *ClusterManager) JoinCluster(masterHost string, masterPort int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.currentNode.Status = NodeStatusJoining

	// 向主节点发送加入请求
	joinURL := fmt.Sprintf("http://%s:%d/api/v1/cluster/join", masterHost, masterPort)

	nodeData, err := json.Marshal(cm.currentNode)
	if err != nil {
		return fmt.Errorf("序列化节点信息失败: %w", err)
	}

	resp, err := cm.client.Post(joinURL, "application/json",
		bytes.NewBuffer(nodeData))
	if err != nil {
		cm.currentNode.Status = NodeStatusFailed
		return fmt.Errorf("发送加入请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cm.currentNode.Status = NodeStatusFailed
		return fmt.Errorf("加入集群失败，状态码: %d", resp.StatusCode)
	}

	cm.currentNode.Status = NodeStatusActive
	log.Printf("成功加入集群: %s:%d", masterHost, masterPort)

	return nil
}

// LeaveCluster 离开集群
func (cm *ClusterManager) LeaveCluster() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.currentNode.Status = NodeStatusLeaving

	// 通知其他节点
	for _, node := range cm.nodes {
		if node.ID != cm.nodeID && node.Status == NodeStatusActive {
			go cm.notifyNodeLeaving(node)
		}
	}

	cm.currentNode.Status = NodeStatusInactive
	log.Printf("节点已离开集群: %s", cm.nodeID)

	return nil
}

// AddNode 添加节点
func (cm *ClusterManager) AddNode(node *ClusterNode) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.nodes[node.ID]; exists {
		return fmt.Errorf("节点已存在: %s", node.ID)
	}

	node.JoinTime = time.Now()
	node.LastSeen = time.Now()
	node.Status = NodeStatusActive

	cm.nodes[node.ID] = node

	log.Printf("节点已添加到集群: %s (%s:%d)", node.ID, node.Host, node.Port)

	return nil
}

// RemoveNode 移除节点
func (cm *ClusterManager) RemoveNode(nodeID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	node, exists := cm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("节点不存在: %s", nodeID)
	}

	delete(cm.nodes, nodeID)

	log.Printf("节点已从集群移除: %s (%s:%d)", node.ID, node.Host, node.Port)

	return nil
}

// GetNodes 获取所有节点
func (cm *ClusterManager) GetNodes() map[string]*ClusterNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*ClusterNode)
	for id, node := range cm.nodes {
		result[id] = node
	}
	return result
}

// GetActiveNodes 获取活跃节点
func (cm *ClusterManager) GetActiveNodes() []*ClusterNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var activeNodes []*ClusterNode
	for _, node := range cm.nodes {
		if node.Status == NodeStatusActive {
			activeNodes = append(activeNodes, node)
		}
	}
	return activeNodes
}

// SelectNode 选择节点（负载均衡）
func (cm *ClusterManager) SelectNode(strategy string) (*ClusterNode, error) {
	activeNodes := cm.GetActiveNodes()
	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("没有可用的活跃节点")
	}

	switch strategy {
	case "least_load":
		return cm.selectLeastLoadNode(activeNodes), nil
	case "least_tasks":
		return cm.selectLeastTasksNode(activeNodes), nil
	case "round_robin":
		return cm.selectRoundRobinNode(activeNodes), nil
	default:
		return cm.selectLeastLoadNode(activeNodes), nil
	}
}

// selectLeastLoadNode 选择负载最低的节点
func (cm *ClusterManager) selectLeastLoadNode(nodes []*ClusterNode) *ClusterNode {
	if len(nodes) == 0 {
		return nil
	}

	selected := nodes[0]
	for _, node := range nodes[1:] {
		if node.Load < selected.Load {
			selected = node
		}
	}
	return selected
}

// selectLeastTasksNode 选择任务最少的节点
func (cm *ClusterManager) selectLeastTasksNode(nodes []*ClusterNode) *ClusterNode {
	if len(nodes) == 0 {
		return nil
	}

	selected := nodes[0]
	for _, node := range nodes[1:] {
		if node.ActiveTasks < selected.ActiveTasks {
			selected = node
		}
	}
	return selected
}

// selectRoundRobinNode 轮询选择节点
var (
	roundRobinIndex int
	roundRobinMutex sync.Mutex
)

func (cm *ClusterManager) selectRoundRobinNode(nodes []*ClusterNode) *ClusterNode {
	if len(nodes) == 0 {
		return nil
	}

	roundRobinMutex.Lock()
	defer roundRobinMutex.Unlock()

	selected := nodes[roundRobinIndex%len(nodes)]
	roundRobinIndex++
	return selected
}

// UpdateNodeLoad 更新节点负载
func (cm *ClusterManager) UpdateNodeLoad(nodeID string, load float64, activeTasks int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	node, exists := cm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("节点不存在: %s", nodeID)
	}

	node.Load = load
	node.ActiveTasks = activeTasks
	node.LastSeen = time.Now()

	return nil
}

// startHeartbeat 启动心跳
func (cm *ClusterManager) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cm.sendHeartbeat()
	}
}

// sendHeartbeat 发送心跳
func (cm *ClusterManager) sendHeartbeat() {
	cm.mu.RLock()
	currentNode := cm.currentNode
	nodes := make([]*ClusterNode, 0, len(cm.nodes))
	for _, node := range cm.nodes {
		if node.ID != cm.nodeID && node.Status == NodeStatusActive {
			nodes = append(nodes, node)
		}
	}
	cm.mu.RUnlock()

	// 更新当前节点的负载信息
	currentNode.LastSeen = time.Now()
	// 这里可以添加实际的负载计算逻辑

	// 向其他节点发送心跳
	for _, node := range nodes {
		go cm.sendHeartbeatToNode(node, currentNode)
	}
}

// sendHeartbeatToNode 向指定节点发送心跳
func (cm *ClusterManager) sendHeartbeatToNode(targetNode, currentNode *ClusterNode) {
	heartbeatURL := fmt.Sprintf("http://%s:%d/api/v1/cluster/heartbeat",
		targetNode.Host, targetNode.Port)

	nodeData, err := json.Marshal(currentNode)
	if err != nil {
		log.Printf("序列化心跳数据失败: %v", err)
		return
	}

	resp, err := cm.client.Post(heartbeatURL, "application/json",
		bytes.NewBuffer(nodeData))
	if err != nil {
		log.Printf("发送心跳到节点 %s 失败: %v", targetNode.ID, err)
		cm.markNodeUnhealthy(targetNode.ID)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("节点 %s 心跳响应异常: %d", targetNode.ID, resp.StatusCode)
		cm.markNodeUnhealthy(targetNode.ID)
	}
}

// startHealthCheck 启动健康检查
func (cm *ClusterManager) startHealthCheck() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cm.performHealthCheck()
	}
}

// performHealthCheck 执行健康检查
func (cm *ClusterManager) performHealthCheck() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for nodeID, node := range cm.nodes {
		if nodeID == cm.nodeID {
			continue
		}

		// 检查节点是否超时
		if now.Sub(node.LastSeen) > 2*time.Minute {
			log.Printf("节点 %s 超时，标记为不健康", nodeID)
			node.Status = NodeStatusFailed
		}
	}
}

// markNodeUnhealthy 标记节点为不健康
func (cm *ClusterManager) markNodeUnhealthy(nodeID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if node, exists := cm.nodes[nodeID]; exists {
		node.Status = NodeStatusFailed
		log.Printf("节点 %s 已标记为失败", nodeID)
	}
}

// notifyNodeLeaving 通知节点离开
func (cm *ClusterManager) notifyNodeLeaving(targetNode *ClusterNode) {
	leaveURL := fmt.Sprintf("http://%s:%d/api/v1/cluster/leave",
		targetNode.Host, targetNode.Port)

	leaveData := map[string]string{
		"node_id": cm.nodeID,
	}

	data, _ := json.Marshal(leaveData)
	resp, err := cm.client.Post(leaveURL, "application/json",
		bytes.NewBuffer(data))
	if err != nil {
		log.Printf("通知节点 %s 离开失败: %v", targetNode.ID, err)
		return
	}
	defer resp.Body.Close()
}

// GetClusterStats 获取集群统计信息
func (cm *ClusterManager) GetClusterStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	totalNodes := len(cm.nodes)
	activeNodes := 0
	totalLoad := 0.0
	totalTasks := 0
	totalCapacity := 0

	for _, node := range cm.nodes {
		if node.Status == NodeStatusActive {
			activeNodes++
			totalLoad += node.Load
			totalTasks += node.ActiveTasks
			totalCapacity += node.Capacity
		}
	}

	avgLoad := 0.0
	if activeNodes > 0 {
		avgLoad = totalLoad / float64(activeNodes)
	}

	return map[string]interface{}{
		"total_nodes":    totalNodes,
		"active_nodes":   activeNodes,
		"failed_nodes":   totalNodes - activeNodes,
		"total_capacity": totalCapacity,
		"total_tasks":    totalTasks,
		"average_load":   avgLoad,
		"cluster_health": float64(activeNodes) / float64(totalNodes),
	}
}

// ClusterHandler 集群处理器
type ClusterHandler struct {
	manager *ClusterManager
}

// NewClusterHandler 创建集群处理器
func NewClusterHandler(manager *ClusterManager) *ClusterHandler {
	return &ClusterHandler{
		manager: manager,
	}
}

// JoinCluster 加入集群
func (h *ClusterHandler) JoinCluster(c *gin.Context) {
	var node ClusterNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.manager.AddNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "添加节点失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点加入成功",
		"node_id": node.ID,
	})
}

// LeaveCluster 离开集群
func (h *ClusterHandler) LeaveCluster(c *gin.Context) {
	var req struct {
		NodeID string `json:"node_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.manager.RemoveNode(req.NodeID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点离开成功",
	})
}

// Heartbeat 心跳处理
func (h *ClusterHandler) Heartbeat(c *gin.Context) {
	var node ClusterNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新节点信息
	if err := h.manager.UpdateNodeLoad(node.ID, node.Load, node.ActiveTasks); err != nil {
		// 如果节点不存在，尝试添加
		if err := h.manager.AddNode(&node); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "更新节点失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "心跳接收成功",
	})
}

// GetNodes 获取集群节点
func (h *ClusterHandler) GetNodes(c *gin.Context) {
	nodes := h.manager.GetNodes()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nodes,
		"count":   len(nodes),
	})
}

// GetClusterStats 获取集群统计
func (h *ClusterHandler) GetClusterStats(c *gin.Context) {
	stats := h.manager.GetClusterStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// SelectNode 选择节点
func (h *ClusterHandler) SelectNode(c *gin.Context) {
	strategy := c.DefaultQuery("strategy", "least_load")

	node, err := h.manager.SelectNode(strategy)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    node,
	})
}
