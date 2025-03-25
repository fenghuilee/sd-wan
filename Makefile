.PHONY: all build clean test server client install docker docker-push package

# 设置 Go 编译器和标志
GO := go
GOFLAGS := -v
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

# 版本信息
VERSION := 1.0.0
LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.GitBranch=$(GIT_BRANCH)'

# 项目信息
BINARY_NAME := sd-wan
SERVER_BINARY := $(BINARY_NAME)-server
CLIENT_BINARY := $(BINARY_NAME)-client

# Docker 信息
DOCKER_REGISTRY := docker.io
DOCKER_NAMESPACE := fenghuilee
DOCKER_SERVER_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(SERVER_BINARY)
DOCKER_CLIENT_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(CLIENT_BINARY)

# 目标平台
PLATFORMS := linux windows darwin
ARCHITECTURES := amd64 arm64 arm
SUPPORTED_PLATFORMS := \
	linux-amd64 linux-arm64 linux-arm \
	windows-amd64 windows-arm64 \
	darwin-amd64 darwin-arm64

# 源文件目录
CMD_DIR := cmd
BUILD_DIR := build
DIST_DIR := dist

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
	@mkdir -p $(BUILD_DIR)
	@for platform in $(SUPPORTED_PLATFORMS); do \
		IFS=- read -r os arch <<< "$$platform"; \
		echo "构建 $$os/$$arch..."; \
		export GOOS=$$os GOARCH=$$arch; \
		if [ "$$os" = "windows" ]; then \
			$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
				-o $(BUILD_DIR)/$(SERVER_BINARY)-$$os-$$arch.exe $(CMD_DIR)/server/main.go; \
			$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
				-o $(BUILD_DIR)/$(CLIENT_BINARY)-$$os-$$arch.exe $(CMD_DIR)/client/main.go; \
		else \
			$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
				-o $(BUILD_DIR)/$(SERVER_BINARY)-$$os-$$arch $(CMD_DIR)/server/main.go; \
			$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
				-o $(BUILD_DIR)/$(CLIENT_BINARY)-$$os-$$arch $(CMD_DIR)/client/main.go; \
		fi; \
	done

# 打包发布文件
package: cross-build
	@echo "打包发布文件..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(SUPPORTED_PLATFORMS); do \
		IFS=- read -r os arch <<< "$$platform"; \
		echo "打包 $$os/$$arch..."; \
		mkdir -p $(BUILD_DIR)/$(BINARY_NAME)-$$platform; \
		cp README.md LICENSE configs/config.yaml $(BUILD_DIR)/$(BINARY_NAME)-$$platform/; \
		if [ "$$os" = "windows" ]; then \
			cp $(BUILD_DIR)/$(SERVER_BINARY)-$$os-$$arch.exe $(BUILD_DIR)/$(BINARY_NAME)-$$platform/; \
			cp $(BUILD_DIR)/$(CLIENT_BINARY)-$$os-$$arch.exe $(BUILD_DIR)/$(BINARY_NAME)-$$platform/; \
			cd $(BUILD_DIR) && zip -r ../$(DIST_DIR)/$(BINARY_NAME)-$$platform.zip $(BINARY_NAME)-$$platform/; \
		else \
			cp $(BUILD_DIR)/$(SERVER_BINARY)-$$os-$$arch $(BUILD_DIR)/$(BINARY_NAME)-$$platform/; \
			cp $(BUILD_DIR)/$(CLIENT_BINARY)-$$os-$$arch $(BUILD_DIR)/$(BINARY_NAME)-$$platform/; \
			cd $(BUILD_DIR) && tar czf ../$(DIST_DIR)/$(BINARY_NAME)-$$platform.tar.gz $(BINARY_NAME)-$$platform/; \
		fi; \
	done

# Docker 构建
docker-server:
	@echo "构建服务器 Docker 镜像..."
	docker build -t $(DOCKER_SERVER_IMAGE):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-f docker/server/Dockerfile .

docker-client:
	@echo "构建客户端 Docker 镜像..."
	docker build -t $(DOCKER_CLIENT_IMAGE):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-f docker/client/Dockerfile .

docker: docker-server docker-client

# Docker 推送
docker-push:
	@echo "推送 Docker 镜像..."
	docker push $(DOCKER_SERVER_IMAGE):$(VERSION)
	docker push $(DOCKER_CLIENT_IMAGE):$(VERSION)
	docker tag $(DOCKER_SERVER_IMAGE):$(VERSION) $(DOCKER_SERVER_IMAGE):latest
	docker tag $(DOCKER_CLIENT_IMAGE):$(VERSION) $(DOCKER_CLIENT_IMAGE):latest
	docker push $(DOCKER_SERVER_IMAGE):latest
	docker push $(DOCKER_CLIENT_IMAGE):latest

# 运行测试
test:
	@echo "运行测试..."
	$(GO) test -v ./...

# 运行服务器
run-server:
	@echo "运行服务器..."
	sudo $(GO) run -ldflags "$(LDFLAGS)" $(CMD_DIR)/server/main.go

# 运行客户端
run-client:
	@echo "运行客户端..."
	sudo $(GO) run -ldflags "$(LDFLAGS)" $(CMD_DIR)/client/main.go

# 安装二进制文件
install: build
	@echo "安装二进制文件..."
	sudo install -m 755 $(BUILD_DIR)/$(SERVER_BINARY) /usr/local/bin/
	sudo install -m 755 $(BUILD_DIR)/$(CLIENT_BINARY) /usr/local/bin/

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
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

# 显示版本信息
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  make build        - 构建服务器和客户端"
	@echo "  make server      - 仅构建服务器"
	@echo "  make client      - 仅构建客户端"
	@echo "  make cross-build - 交叉编译所有平台"
	@echo "  make package     - 打包发布文件"
	@echo "  make docker      - 构建 Docker 镜像"
	@echo "  make docker-push - 推送 Docker 镜像"
	@echo "  make test        - 运行测试"
	@echo "  make run-server  - 运行服务器"
	@echo "  make run-client  - 运行客户端"
	@echo "  make install     - 安装二进制文件"
	@echo "  make clean       - 清理构建文件"
	@echo "  make deps        - 更新依赖"
	@echo "  make fmt         - 格式化代码"
	@echo "  make lint        - 代码检查"
	@echo "  make version     - 显示版本信息" 