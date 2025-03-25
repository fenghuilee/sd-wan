.PHONY: all build clean test server client install

# 设置 Go 编译器和标志
GO := go
GOFLAGS := -v
LDFLAGS := -s -w

# 项目信息
BINARY_NAME := sd-wan
SERVER_BINARY := $(BINARY_NAME)-server
CLIENT_BINARY := $(BINARY_NAME)-client
VERSION := 1.0.0

# 目标平台
PLATFORMS := linux windows
ARCHITECTURES := amd64 arm64

# 源文件目录
CMD_DIR := cmd
BUILD_DIR := build

all: build

# 构建所有目标
build: server client

# 构建服务器
server:
	@echo "构建服务器..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(SERVER_BINARY) $(CMD_DIR)/server/main.go

# 构建客户端
client:
	@echo "构建客户端..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(CLIENT_BINARY) $(CMD_DIR)/client/main.go

# 交叉编译
cross-build:
	@echo "开始交叉编译..."
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			echo "构建 $$platform/$$arch..."; \
			GOOS=$$platform GOARCH=$$arch $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
				-o $(BUILD_DIR)/$(SERVER_BINARY)-$$platform-$$arch $(CMD_DIR)/server/main.go; \
			GOOS=$$platform GOARCH=$$arch $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
				-o $(BUILD_DIR)/$(CLIENT_BINARY)-$$platform-$$arch $(CMD_DIR)/client/main.go; \
		done; \
	done

# 运行测试
test:
	@echo "运行测试..."
	$(GO) test -v ./...

# 运行服务器
run-server:
	@echo "运行服务器..."
	sudo $(GO) run $(CMD_DIR)/server/main.go

# 运行客户端
run-client:
	@echo "运行客户端..."
	sudo $(GO) run $(CMD_DIR)/client/main.go

# 安装二进制文件
install: build
	@echo "安装二进制文件..."
	sudo install -m 755 $(BUILD_DIR)/$(SERVER_BINARY) /usr/local/bin/
	sudo install -m 755 $(BUILD_DIR)/$(CLIENT_BINARY) /usr/local/bin/

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	$(GO) clean

# 生成依赖文件
deps:
	@echo "更新依赖..."
	$(GO) mod tidy
	$(GO) mod verify

# 代码格式化
fmt:
	@echo "格式化代码..."
	$(GO) fmt ./...

# 代码检查
lint:
	@echo "检查代码..."
	$(GO) vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed"; \
	fi

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  make build        - 构建服务器和客户端"
	@echo "  make server      - 仅构建服务器"
	@echo "  make client      - 仅构建客户端"
	@echo "  make cross-build - 交叉编译所有平台"
	@echo "  make test        - 运行测试"
	@echo "  make run-server  - 运行服务器"
	@echo "  make run-client  - 运行客户端"
	@echo "  make install     - 安装二进制文件"
	@echo "  make clean       - 清理构建文件"
	@echo "  make deps        - 更新依赖"
	@echo "  make fmt         - 格式化代码"
	@echo "  make lint        - 代码检查" 