package network

import (
	"net"
	"sync"
	"time"
)

// Node 表示网络中的一个节点
type Node struct {
	ID          string
	PublicIP    net.IP
	PublicPort  uint16
	PrivateIP   net.IP
	PrivatePort uint16
	LastSeen    time.Time
	Routes      []Route
}

// Route 表示一条路由
type Route struct {
	Destination string
	NextHop     string
	Metric      uint8
}

// Discovery 节点发现管理器
type Discovery struct {
	nodes    map[string]*Node
	mutex    sync.RWMutex
	interval time.Duration
}

// NewDiscovery 创建新的节点发现管理器
func NewDiscovery(interval time.Duration) *Discovery {
	return &Discovery{
		nodes:    make(map[string]*Node),
		interval: interval,
	}
}

// AddNode 添加新节点
func (d *Discovery) AddNode(node *Node) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.nodes[node.ID] = node
}

// RemoveNode 移除节点
func (d *Discovery) RemoveNode(nodeID string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.nodes, nodeID)
}

// GetNode 获取节点信息
func (d *Discovery) GetNode(nodeID string) *Node {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.nodes[nodeID]
}

// UpdateNode 更新节点信息
func (d *Discovery) UpdateNode(node *Node) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if existing, ok := d.nodes[node.ID]; ok {
		existing.PublicIP = node.PublicIP
		existing.PublicPort = node.PublicPort
		existing.PrivateIP = node.PrivateIP
		existing.PrivatePort = node.PrivatePort
		existing.LastSeen = time.Now()
		existing.Routes = node.Routes
	}
}

// GetNodes 获取所有节点
func (d *Discovery) GetNodes() []*Node {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	nodes := make([]*Node, 0, len(d.nodes))
	for _, node := range d.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// Cleanup 清理过期节点
func (d *Discovery) Cleanup(timeout time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	now := time.Now()
	for id, node := range d.nodes {
		if now.Sub(node.LastSeen) > timeout {
			delete(d.nodes, id)
		}
	}
}

// AddRoute 添加路由
func (d *Discovery) AddRoute(nodeID string, route Route) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if node, ok := d.nodes[nodeID]; ok {
		node.Routes = append(node.Routes, route)
	}
}

// GetRoutes 获取节点的路由
func (d *Discovery) GetRoutes(nodeID string) []Route {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if node, ok := d.nodes[nodeID]; ok {
		return node.Routes
	}
	return nil
}

// FindRoute 查找到目标的路由
func (d *Discovery) FindRoute(destination string) *Route {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	var bestRoute *Route
	bestMetric := uint8(255)

	for _, node := range d.nodes {
		for _, route := range node.Routes {
			if route.Destination == destination && route.Metric < bestMetric {
				bestRoute = &route
				bestMetric = route.Metric
			}
		}
	}

	return bestRoute
}

// Start 启动节点发现服务
func (d *Discovery) Start() {
	go func() {
		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()

		for range ticker.C {
			d.Cleanup(d.interval * 3)
		}
	}()
}
