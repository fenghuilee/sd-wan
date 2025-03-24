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
	connections map[string]*net.UDPConn
	mutex       sync.RWMutex
}

// NewNATTraversal 创建新的NAT穿透管理器
func NewNATTraversal(relayServer net.IP, relayPort uint16) *NATTraversal {
	return &NATTraversal{
		relayServer: relayServer,
		relayPort:   relayPort,
		connections: make(map[string]*net.UDPConn),
	}
}

// CreateRelayConnection 创建中继连接
func (n *NATTraversal) CreateRelayConnection(targetID string, targetIP net.IP, targetPort uint16) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// 检查是否已存在连接
	if _, exists := n.connections[targetID]; exists {
		return fmt.Errorf("连接已存在: %s", targetID)
	}

	// 创建本地UDP连接
	localAddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return fmt.Errorf("创建UDP连接失败: %v", err)
	}

	// 发送连接请求到中继服务器
	relayAddr := &net.UDPAddr{
		IP:   n.relayServer,
		Port: int(n.relayPort),
	}

	// 构建连接请求消息
	request := fmt.Sprintf("CONNECT:%s:%s:%d", targetID, targetIP.String(), targetPort)
	_, err = conn.WriteToUDP([]byte(request), relayAddr)
	if err != nil {
		conn.Close()
		return fmt.Errorf("发送连接请求失败: %v", err)
	}

	// 等待连接建立
	response := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err = conn.ReadFromUDP(response)
	if err != nil {
		conn.Close()
		return fmt.Errorf("等待连接响应超时: %v", err)
	}

	// 存储连接
	n.connections[targetID] = conn
	return nil
}

// SendData 通过中继发送数据
func (n *NATTraversal) SendData(targetID string, data []byte) error {
	n.mutex.RLock()
	conn, exists := n.connections[targetID]
	n.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("连接不存在: %s", targetID)
	}

	// 构建数据包
	packet := append([]byte("DATA:"), data...)
	_, err := conn.Write(packet)
	return err
}

// ReceiveData 接收数据
func (n *NATTraversal) ReceiveData(targetID string) ([]byte, error) {
	n.mutex.RLock()
	conn, exists := n.connections[targetID]
	n.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("连接不存在: %s", targetID)
	}

	buf := make([]byte, 1500)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}

	// 解析数据包
	if n < 5 || string(buf[:5]) != "DATA:" {
		return nil, fmt.Errorf("无效的数据包格式")
	}

	return buf[5:n], nil
}

// CloseConnection 关闭连接
func (n *NATTraversal) CloseConnection(targetID string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	conn, exists := n.connections[targetID]
	if !exists {
		return fmt.Errorf("连接不存在: %s", targetID)
	}

	err := conn.Close()
	delete(n.connections, targetID)
	return err
}

// Close 关闭所有连接
func (n *NATTraversal) Close() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for _, conn := range n.connections {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	n.connections = make(map[string]*net.UDPConn)
	return nil
}
