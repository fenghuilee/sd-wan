package network

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"

	"github.com/songgao/water"
)

// TUN 接口管理
type TUN struct {
	*water.Interface
	config *water.Config
}

// NewTUN 创建新的 TUN 接口
func NewTUN(name string, mtu int) (*TUN, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}

	// 根据操作系统设置接口名称
	if runtime.GOOS == "linux" {
		config.Name = name
	}

	// 创建 TUN 接口
	iface, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("创建 TUN 接口失败: %v", err)
	}

	tun := &TUN{
		Interface: iface,
		config:    &config,
	}

	// 设置 MTU
	if err := tun.setMTU(mtu); err != nil {
		iface.Close()
		return nil, err
	}

	return tun, nil
}

// setMTU 设置接口 MTU
func (t *TUN) setMTU(mtu int) error {
	if runtime.GOOS == "linux" {
		cmd := fmt.Sprintf("ip link set %s mtu %d up", t.Name(), mtu)
		return exec.Command("sh", "-c", cmd).Run()
	}
	return nil
}

// ReadPacket 读取数据包
func (t *TUN) ReadPacket() ([]byte, error) {
	buf := make([]byte, 1500)
	n, err := t.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// WritePacket 写入数据包
func (t *TUN) WritePacket(packet []byte) error {
	_, err := t.Write(packet)
	return err
}

// Close 关闭接口
func (t *TUN) Close() error {
	return t.Interface.Close()
}

// GetIP 获取接口 IP 地址
func (t *TUN) GetIP() (net.IP, error) {
	if runtime.GOOS == "linux" {
		cmd := fmt.Sprintf("ip addr show %s", t.Name())
		output, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			return nil, err
		}
		// 解析 IP 地址
		// TODO: 实现 IP 地址解析
		return nil, nil
	}
	return nil, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
}

// SetIP 设置接口 IP 地址
func (t *TUN) SetIP(ip net.IP) error {
	if runtime.GOOS == "linux" {
		cmd := fmt.Sprintf("ip addr add %s/24 dev %s", ip.String(), t.Name())
		return exec.Command("sh", "-c", cmd).Run()
	}
	return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
}
