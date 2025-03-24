# SD-WAN

一个基于 Go 语言开发的轻量级 VPN 解决方案，支持跨平台部署。

## 功能特性

- 支持 Windows、Linux 和 OpenWrt 平台
- 采用 C/S 架构设计
- 支持点对点加密通信（可选）
- 支持自动节点发现
- 支持 NAT 穿透
- 支持自定义路由
- 支持 TUN/TAP 虚拟网卡
- 支持动态路由更新
- 支持节点存活检测
- 支持硬件加速加密（如 AES-NI）

## 系统要求

- Go 1.16 或更高版本
- 支持的操作系统：
  - Windows 10/11
  - Linux (主流发行版)
  - OpenWrt

## 快速开始

### 安装

```bash
go install github.com/yourusername/sd-wan@latest
```

### 配置

1. 创建配置文件 `config.yaml`：
```yaml
server:
  host: "0.0.0.0"
  port: 51820
  cert_file: "certs/server.crt"
  key_file: "certs/server.key"

client:
  server_address: "vpn.example.com:51820"
  device_name: "sd-wan0"
  mtu: 1500

network:
  subnet: "10.0.0.0/24"
  dns: "8.8.8.8"
  keep_alive: 30
  reconnect: 5

nat:
  relay_server: "relay.example.com"
  relay_port: 51821

security:
  encryption: true              # 是否启用加密
  algorithm: "chacha20-poly1305" # 加密算法：chacha20-poly1305 或 aes-256-gcm
  hardware_acceleration: true   # 是否启用硬件加速
  key_size: 32                 # 密钥大小（字节）
```

2. 配置服务器地址和认证信息
3. 启动服务

### 使用

启动服务器：
```bash
sd-wan server -c config.yaml
```

启动客户端：
```bash
sd-wan client -c config.yaml
```

## 项目结构

```
sd-wan/
├── cmd/                          # 主程序入口
│   ├── client/                  # 客户端程序
│   │   └── main.go             # 客户端主程序
│   └── server/                  # 服务器程序
│       └── main.go             # 服务器主程序
├── internal/                    # 内部包
│   ├── config/                 # 配置管理
│   │   └── config.go          # 配置结构定义
│   ├── network/                # 网络相关
│   │   ├── tun.go            # TUN/TAP 接口管理
│   │   ├── discovery.go      # 节点发现
│   │   └── nat.go            # NAT 穿透
│   └── protocol/              # 协议实现
│       └── protocol.go        # 协议定义
├── pkg/                        # 公共包
│   ├── crypto/                # 加密相关
│   │   └── crypto.go         # 加密实现
│   └── utils/                 # 工具函数
├── api/                       # API 定义
├── docs/                      # 文档
└── scripts/                   # 脚本文件
```

## 核心功能说明

### 1. 网络协议层
- 支持多种消息类型：握手、数据、保活、路由、NAT穿透
- 实现了消息的编码和解码
- 支持自定义协议扩展
- 支持消息加密传输

### 2. TUN/TAP 接口管理
- 支持创建和管理虚拟网卡
- 实现了数据包的读写
- 支持 MTU 和 IP 地址配置
- 支持多平台兼容

### 3. 节点发现和路由管理
- 自动发现网络中的节点
- 维护动态路由表
- 支持节点存活检测
- 实现最佳路由查找

### 4. NAT 穿透功能
- 支持通过中继服务器建立连接
- 实现了数据的转发
- 支持连接的管理和清理
- 实现了简单的连接协议

### 5. 加密功能
- 支持可配置的加密开关
- 提供两种加密算法：
  - AES-256-GCM：高性能设备推荐
  - ChaCha20-Poly1305：低性能设备推荐
- 支持硬件加速（如 AES-NI）
- 实现了安全的密钥管理
- 支持消息完整性验证

## 系统架构

### 服务器端
- 处理节点注册和认证
- 维护节点状态和路由表
- 处理数据转发
- 提供 NAT 穿透支持
- 管理加密密钥

### 客户端
- 管理 TUN/TAP 虚拟网卡
- 处理数据包的收发
- 维护与服务器的连接
- 支持自动重连
- 实现数据加密传输

### 中继服务器
- 处理 NAT 穿透请求
- 转发数据包
- 维护连接状态

## 性能优化

### 加密性能
- 支持轻量级加密算法（ChaCha20-Poly1305）
- 支持硬件加速（AES-NI）
- 可配置的加密级别
- 针对低性能设备（如 OpenWrt）的优化：
  - 支持关闭加密
  - 支持降低加密强度
  - 支持使用更轻量的加密算法

### 资源使用
- 内存占用优化
- CPU 使用率控制
- 网络带宽优化
- 支持资源限制配置

## 开发计划

- [ ] 添加更多加密算法支持
- [ ] 实现更多平台支持
- [ ] 添加 Web 管理界面
- [ ] 实现流量统计功能
- [ ] 添加更多路由协议支持
- [ ] 优化加密性能
- [ ] 添加密钥轮换机制

## 注意事项

- 需要 root 权限来创建 TUN 接口
- 确保防火墙允许相关端口
- 建议在生产环境中启用加密通信
- 定期检查节点状态和路由表
- 在低性能设备（如 OpenWrt）上：
  - 建议使用轻量级加密算法
  - 可根据需要关闭加密功能
  - 注意监控系统资源使用情况

## 贡献指南

欢迎提交 Pull Request 或创建 Issue。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件 