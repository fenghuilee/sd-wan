package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// ClientConfig 客户端配置
type ClientConfig struct {
	ServerAddress string `mapstructure:"server_address"`
	DeviceName    string `mapstructure:"device_name"`
	MTU           int    `mapstructure:"mtu"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Subnet    string `mapstructure:"subnet"`
	DNS       string `mapstructure:"dns"`
	KeepAlive int    `mapstructure:"keep_alive"`
	Reconnect int    `mapstructure:"reconnect"`
}

// NATConfig NAT穿透配置
type NATConfig struct {
	RelayServer string `mapstructure:"relay_server"`
	RelayPort   int    `mapstructure:"relay_port"`
}

// SecurityConfig 加密配置
type SecurityConfig struct {
	Encryption           bool   `yaml:"encryption"`
	Algorithm            string `yaml:"algorithm"`
	HardwareAcceleration bool   `yaml:"hardware_acceleration"`
	KeySize              int    `yaml:"key_size"`
}

// Config 总配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Client   ClientConfig   `mapstructure:"client"`
	Network  NetworkConfig  `mapstructure:"network"`
	NAT      NATConfig      `mapstructure:"nat"`
	Security SecurityConfig `yaml:"security"`
}

// LoadConfig 加载配置
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetRelayAddr 获取中继服务器地址
func (c *Config) GetRelayAddr() string {
	return fmt.Sprintf("%s:%d", c.NAT.RelayServer, c.NAT.RelayPort)
}
