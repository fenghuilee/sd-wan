package protocol

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/fenghuilee/sd-wan/pkg/crypto"
)

const (
	// 协议版本
	ProtocolVersion = 1

	// 消息类型
	MsgTypeHandshake = 1
	MsgTypeData      = 2
	MsgTypeKeepAlive = 3
	MsgTypeRoute     = 4
	MsgTypeNAT       = 5

	// 头部长度
	HeaderSize = 12
)

// Message 表示一个网络消息
type Message struct {
	Version uint8
	Type    uint8
	Length  uint16
	Data    []byte
}

// HandshakeMessage 握手消息
type HandshakeMessage struct {
	NodeID      string
	PublicIP    net.IP
	PublicPort  uint16
	PrivateIP   net.IP
	PrivatePort uint16
}

// RouteMessage 路由消息
type RouteMessage struct {
	Destination string
	NextHop     string
	Metric      uint8
}

// NATMessage NAT穿透消息
type NATMessage struct {
	TargetID    string
	TargetIP    net.IP
	TargetPort  uint16
	RelayServer net.IP
	RelayPort   uint16
}

// MessageType 消息类型
type MessageType uint8

const (
	TypeHandshake MessageType = iota
	TypeData
	TypeKeepAlive
	TypeRoute
	TypeNAT
)

// Protocol 协议处理器
type Protocol struct {
	crypto *crypto.Crypto
}

// NewProtocol 创建新的协议处理器
func NewProtocol(crypto *crypto.Crypto) *Protocol {
	return &Protocol{
		crypto: crypto,
	}
}

// Encode 将消息编码为字节流
func (m *Message) Encode() ([]byte, error) {
	buf := make([]byte, HeaderSize+len(m.Data))
	buf[0] = m.Version
	buf[1] = m.Type
	binary.BigEndian.PutUint16(buf[2:4], m.Length)
	copy(buf[HeaderSize:], m.Data)
	return buf, nil
}

// Decode 从字节流解码消息
func DecodeMessage(data []byte) (*Message, error) {
	if len(data) < HeaderSize {
		return nil, nil
	}

	msg := &Message{
		Version: data[0],
		Type:    data[1],
		Length:  binary.BigEndian.Uint16(data[2:4]),
	}

	if len(data) > HeaderSize {
		msg.Data = make([]byte, len(data)-HeaderSize)
		copy(msg.Data, data[HeaderSize:])
	}

	return msg, nil
}

// Encode 编码消息
func (p *Protocol) Encode(msg *Message) ([]byte, error) {
	// 加密负载
	var encryptedPayload []byte
	var err error
	if p.crypto.IsEnabled() {
		encryptedPayload, err = p.crypto.Encrypt(msg.Data)
		if err != nil {
			return nil, err
		}
	} else {
		encryptedPayload = msg.Data
	}

	// 构建消息头
	header := make([]byte, 5)
	header[0] = byte(msg.Type)
	binary.BigEndian.PutUint32(header[1:], uint32(len(encryptedPayload)))

	// 组合消息
	return append(header, encryptedPayload...), nil
}

// Decode 解码消息
func (p *Protocol) Decode(data []byte) (*Message, error) {
	if len(data) < 5 {
		return nil, errors.New("message too short")
	}

	msgType := MessageType(data[0])
	payloadLen := binary.BigEndian.Uint32(data[1:5])
	payload := data[5:]

	if uint32(len(payload)) != payloadLen {
		return nil, errors.New("invalid payload length")
	}

	// 解密负载
	var decryptedPayload []byte
	var err error
	if p.crypto.IsEnabled() {
		decryptedPayload, err = p.crypto.Decrypt(payload)
		if err != nil {
			return nil, err
		}
	} else {
		decryptedPayload = payload
	}

	return &Message{
		Version: ProtocolVersion,
		Type:    uint8(msgType),
		Length:  uint16(len(decryptedPayload)),
		Data:    decryptedPayload,
	}, nil
}
