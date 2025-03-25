package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

	// 创建 TUN 接口
	tun, err := network.NewTUN(cfg.Client.DeviceName, cfg.Client.MTU)
	if err != nil {
		log.Fatalf("创建 TUN 接口失败: %v", err)
	}
	defer tun.Close()

	// 设置 IP 地址
	ip := net.ParseIP(cfg.Network.Subnet[:len(cfg.Network.Subnet)-3])
	if err := tun.SetIP(ip); err != nil {
		log.Fatalf("设置 IP 地址失败: %v", err)
	}

	// 创建 NAT 穿透管理器
	nat := network.NewNATTraversal(
		net.ParseIP(cfg.NAT.RelayServer),
		uint16(cfg.NAT.RelayPort),
	)

	// 创建 UDP 连接
	serverAddr, err := net.ResolveUDPAddr("udp", cfg.Client.ServerAddress)
	if err != nil {
		log.Fatalf("解析服务器地址失败: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatalf("连接服务器失败: %v", err)
	}
	defer conn.Close()

	// 发送握手消息
	if err := sendHandshake(conn, tun); err != nil {
		log.Fatalf("发送握手消息失败: %v", err)
	}

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动保活消息发送
	go sendKeepAlive(conn)

	// 启动数据包处理
	go handlePackets(tun, conn, nat)

	// 等待信号
	<-sigChan
	log.Println("正在关闭客户端...")
}

func sendHandshake(conn *net.UDPConn, tun *network.TUN) error {
	// 获取本地 IP 地址
	localIP, err := tun.GetIP()
	if err != nil {
		return err
	}

	// 构建握手消息
	handshake := protocol.HandshakeMessage{
		NodeID:      generateNodeID(),
		PublicIP:    getPublicIP(),
		PublicPort:  uint16(conn.LocalAddr().(*net.UDPAddr).Port),
		PrivateIP:   localIP,
		PrivatePort: uint16(conn.LocalAddr().(*net.UDPAddr).Port),
	}

	data, err := json.Marshal(handshake)
	if err != nil {
		return err
	}

	msg := &protocol.Message{
		Version: protocol.ProtocolVersion,
		Type:    protocol.MsgTypeHandshake,
		Data:    data,
	}

	encoded, err := msg.Encode()
	if err != nil {
		return err
	}

	_, err = conn.Write(encoded)
	return err
}

func sendKeepAlive(conn *net.UDPConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		msg := &protocol.Message{
			Version: protocol.ProtocolVersion,
			Type:    protocol.MsgTypeKeepAlive,
			Data:    []byte(generateNodeID()),
		}

		data, err := msg.Encode()
		if err != nil {
			log.Printf("编码保活消息失败: %v", err)
			continue
		}

		_, err = conn.Write(data)
		if err != nil {
			log.Printf("发送保活消息失败: %v", err)
		}
	}
}

func handlePackets(tun *network.TUN, conn *net.UDPConn, nat *network.NATTraversal) {
	buf := make([]byte, 1500)
	for {
		// 从 TUN 接口读取数据包
		n, err := tun.Read(buf)
		if err != nil {
			log.Printf("读取数据包失败: %v", err)
			continue
		}

		// 构建数据消息
		msg := &protocol.Message{
			Version: protocol.ProtocolVersion,
			Type:    protocol.MsgTypeData,
			Length:  uint16(n),
			Data:    buf[:n],
		}

		// 编码消息
		data, err := msg.Encode()
		if err != nil {
			log.Printf("编码数据消息失败: %v", err)
			continue
		}

		// 发送数据
		_, err = conn.Write(data)
		if err != nil {
			log.Printf("发送数据失败: %v", err)
		}
	}
}

func generateNodeID() string {
	// 生成唯一的节点 ID
	return fmt.Sprintf("node-%d", time.Now().UnixNano())
}

func getPublicIP() net.IP {
	// 获取公网 IP 地址
	// TODO: 实现公网 IP 获取
	return net.ParseIP("0.0.0.0")
}
