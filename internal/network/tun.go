package network

import (
	"net"

	"github.com/songgao/water"
)

// TUN 表示一个 TUN 接口
type TUN struct {
	iface *water.Interface
	mtu   int
}

// NewTUN 创建新的 TUN 接口
func NewTUN(name string, mtu int) (*TUN, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	if name != "" {
		config.Name = name
	}

	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	return &TUN{
		iface: iface,
		mtu:   mtu,
	}, nil
}

// Close 关闭 TUN 接口
func (t *TUN) Close() error {
	return t.iface.Close()
}

// SetIP 设置 TUN 接口的 IP 地址
func (t *TUN) SetIP(ip net.IP) error {
	// 这里需要根据不同的操作系统实现具体的 IP 设置逻辑
	// 暂时返回 nil，表示成功
	return nil
}

// GetIP 获取 TUN 接口的 IP 地址
func (t *TUN) GetIP() (net.IP, error) {
	// 这里需要根据不同的操作系统实现具体的 IP 获取逻辑
	// 暂时返回一个示例 IP
	return net.ParseIP("10.0.0.1"), nil
}

// Read 从 TUN 接口读取数据
func (t *TUN) Read(buf []byte) (int, error) {
	return t.iface.Read(buf)
}

// Write 向 TUN 接口写入数据
func (t *TUN) Write(buf []byte) (int, error) {
	return t.iface.Write(buf)
}

// GetMTU 获取 MTU 值
func (t *TUN) GetMTU() int {
	return t.mtu
}
