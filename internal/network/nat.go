package network

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// NATTraversal NAT穿透管理器
type NATTraversal struct {
	relayServer net.IP
	relayPort   uint16
	connections sync.Map
}

// Connection NAT连接
type Connection struct {
	targetID   string
	targetIP   net.IP
	targetPort uint16
	conn       *net.UDPConn
	lastSeen   time.Time
}

// NewNATTraversal 创建新的NAT穿透管理器
func NewNATTraversal(relayServer net.IP, relayPort uint16) *NATTraversal {
	nat := &NATTraversal{
		relayServer: relayServer,
		relayPort:   relayPort,
	}

	// 启动连接清理
	go nat.cleanConnections()

	return nat
}

// CreateRelayConnection 创建中继连接
func (n *NATTraversal) CreateRelayConnection(targetID string, targetIP net.IP, targetPort uint16) error {
	// 创建 UDP 连接
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("创建 UDP 连接失败: %v", err)
	}

	connection := &Connection{
		targetID:   targetID,
		targetIP:   targetIP,
		targetPort: targetPort,
		conn:       conn,
		lastSeen:   time.Now(),
	}

	n.connections.Store(targetID, connection)
	return nil
}

// SendData 发送数据
func (n *NATTraversal) SendData(targetID string, data []byte) error {
	value, ok := n.connections.Load(targetID)
	if !ok {
		return fmt.Errorf("连接不存在: %s", targetID)
	}

	conn := value.(*Connection)
	conn.lastSeen = time.Now()

	// 发送到中继服务器
	relayAddr := &net.UDPAddr{
		IP:   n.relayServer,
		Port: int(n.relayPort),
	}

	_, err := conn.conn.WriteToUDP(data, relayAddr)
	return err
}

// ReceiveData 接收数据
func (n *NATTraversal) ReceiveData(targetID string) ([]byte, error) {
	value, ok := n.connections.Load(targetID)
	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", targetID)
	}

	conn := value.(*Connection)
	conn.lastSeen = time.Now()

	buf := make([]byte, 1500)
	readLen, _, err := conn.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}

	if readLen <= 0 {
		return nil, fmt.Errorf("读取数据长度为 0")
	}

	return buf[:readLen], nil
}

// CloseConnection 关闭连接
func (n *NATTraversal) CloseConnection(targetID string) error {
	value, ok := n.connections.Load(targetID)
	if !ok {
		return fmt.Errorf("连接不存在: %s", targetID)
	}

	conn := value.(*Connection)
	err := conn.conn.Close()
	n.connections.Delete(targetID)
	return err
}

// Close 关闭所有连接
func (n *NATTraversal) Close() error {
	var lastErr error
	n.connections.Range(func(key, value interface{}) bool {
		conn := value.(*Connection)
		if err := conn.conn.Close(); err != nil {
			lastErr = err
			return false
		}
		return true
	})
	return lastErr
}

// cleanConnections 清理过期连接
func (n *NATTraversal) cleanConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		n.connections.Range(func(key, value interface{}) bool {
			conn := value.(*Connection)
			if now.Sub(conn.lastSeen) > 10*time.Minute {
				conn.conn.Close()
				n.connections.Delete(key)
			}
			return true
		})
	}
}
