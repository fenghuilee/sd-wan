package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fenghuilee/sd-wan/internal/config"
	"github.com/fenghuilee/sd-wan/internal/network"
	"github.com/fenghuilee/sd-wan/internal/protocol"
)

var (
	configFile = flag.String("config", "config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建节点发现管理器
	discovery := network.NewDiscovery(30 * time.Second)
	discovery.Start()

	// 创建 NAT 穿透管理器
	nat := network.NewNATTraversal(
		net.ParseIP(cfg.NAT.RelayServer),
		uint16(cfg.NAT.RelayPort),
	)

	// 创建 UDP 服务器
	addr, err := net.ResolveUDPAddr("udp", cfg.GetServerAddr())
	if err != nil {
		log.Fatalf("解析地址失败: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("创建 UDP 服务器失败: %v", err)
	}
	defer conn.Close()

	log.Printf("服务器启动在 %s", addr.String())

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动消息处理循环
	go handleMessages(conn, discovery, nat)

	// 等待信号
	<-sigChan
	log.Println("正在关闭服务器...")
}

func handleMessages(conn *net.UDPConn, discovery *network.Discovery, nat *network.NATTraversal) {
	buf := make([]byte, 1500)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("读取数据失败: %v", err)
			continue
		}

		// 解码消息
		msg, err := protocol.DecodeMessage(buf[:n])
		if err != nil {
			log.Printf("解码消息失败: %v", err)
			continue
		}

		// 处理不同类型的消息
		switch msg.Type {
		case protocol.MsgTypeHandshake:
			handleHandshake(conn, remoteAddr, msg, discovery)
		case protocol.MsgTypeData:
			handleData(conn, remoteAddr, msg, discovery, nat)
		case protocol.MsgTypeKeepAlive:
			handleKeepAlive(conn, remoteAddr, msg, discovery)
		case protocol.MsgTypeRoute:
			handleRoute(conn, remoteAddr, msg, discovery)
		case protocol.MsgTypeNAT:
			handleNAT(conn, remoteAddr, msg, nat)
		default:
			log.Printf("未知消息类型: %d", msg.Type)
		}
	}
}

func handleHandshake(conn *net.UDPConn, remoteAddr *net.UDPAddr, msg *protocol.Message, discovery *network.Discovery) {
	var handshake protocol.HandshakeMessage
	if err := json.Unmarshal(msg.Data, &handshake); err != nil {
		log.Printf("解析握手消息失败: %v", err)
		return
	}

	// 创建新节点
	node := &network.Node{
		ID:          handshake.NodeID,
		PublicIP:    handshake.PublicIP,
		PublicPort:  handshake.PublicPort,
		PrivateIP:   handshake.PrivateIP,
		PrivatePort: handshake.PrivatePort,
		LastSeen:    time.Now(),
	}

	// 添加或更新节点
	discovery.AddNode(node)

	// 发送响应
	response := &protocol.Message{
		Version: protocol.ProtocolVersion,
		Type:    protocol.MsgTypeHandshake,
		Data:    []byte("OK"),
	}

	data, err := response.Encode()
	if err != nil {
		log.Printf("编码响应失败: %v", err)
		return
	}

	_, err = conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		log.Printf("发送响应失败: %v", err)
	}
}

func handleData(conn *net.UDPConn, remoteAddr *net.UDPAddr, msg *protocol.Message, discovery *network.Discovery, nat *network.NATTraversal) {
	// 查找目标节点
	route := discovery.FindRoute(string(msg.Data[:4]))
	if route == nil {
		log.Printf("未找到路由: %s", string(msg.Data[:4]))
		return
	}

	// 获取目标节点信息
	targetNode := discovery.GetNode(route.NextHop)
	if targetNode == nil {
		log.Printf("未找到目标节点: %s", route.NextHop)
		return
	}

	// 尝试直接发送
	_, err := conn.WriteToUDP(msg.Data, &net.UDPAddr{
		IP:   targetNode.PublicIP,
		Port: int(targetNode.PublicPort),
	})

	// 如果直接发送失败，使用 NAT 穿透
	if err != nil {
		err = nat.SendData(targetNode.ID, msg.Data)
		if err != nil {
			log.Printf("发送数据失败: %v", err)
		}
	}
}

func handleKeepAlive(conn *net.UDPConn, remoteAddr *net.UDPAddr, msg *protocol.Message, discovery *network.Discovery) {
	// 更新节点最后可见时间
	node := discovery.GetNode(string(msg.Data))
	if node != nil {
		node.LastSeen = time.Now()
	}

	// 发送响应
	response := &protocol.Message{
		Version: protocol.ProtocolVersion,
		Type:    protocol.MsgTypeKeepAlive,
		Data:    []byte("OK"),
	}

	data, err := response.Encode()
	if err != nil {
		log.Printf("编码响应失败: %v", err)
		return
	}

	_, err = conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		log.Printf("发送响应失败: %v", err)
	}
}

func handleRoute(conn *net.UDPConn, remoteAddr *net.UDPAddr, msg *protocol.Message, discovery *network.Discovery) {
	var route protocol.RouteMessage
	if err := json.Unmarshal(msg.Data, &route); err != nil {
		log.Printf("解析路由消息失败: %v", err)
		return
	}

	// 添加路由
	discovery.AddRoute(string(msg.Data[:4]), network.Route{
		Destination: route.Destination,
		NextHop:     route.NextHop,
		Metric:      route.Metric,
	})

	// 发送响应
	response := &protocol.Message{
		Version: protocol.ProtocolVersion,
		Type:    protocol.MsgTypeRoute,
		Data:    []byte("OK"),
	}

	data, err := response.Encode()
	if err != nil {
		log.Printf("编码响应失败: %v", err)
		return
	}

	_, err = conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		log.Printf("发送响应失败: %v", err)
	}
}

func handleNAT(conn *net.UDPConn, remoteAddr *net.UDPAddr, msg *protocol.Message, nat *network.NATTraversal) {
	var natMsg protocol.NATMessage
	if err := json.Unmarshal(msg.Data, &natMsg); err != nil {
		log.Printf("解析 NAT 消息失败: %v", err)
		return
	}

	// 创建中继连接
	err := nat.CreateRelayConnection(natMsg.TargetID, natMsg.TargetIP, natMsg.TargetPort)
	if err != nil {
		log.Printf("创建中继连接失败: %v", err)
		return
	}

	// 发送响应
	response := &protocol.Message{
		Version: protocol.ProtocolVersion,
		Type:    protocol.MsgTypeNAT,
		Data:    []byte("OK"),
	}

	data, err := response.Encode()
	if err != nil {
		log.Printf("编码响应失败: %v", err)
		return
	}

	_, err = conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		log.Printf("发送响应失败: %v", err)
	}
}
