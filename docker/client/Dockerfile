# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git make

# 复制源代码
COPY . .

# 构建参数
ARG VERSION
ARG BUILD_TIME
ARG GIT_COMMIT

# 构建应用
RUN make client LDFLAGS="-s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.GitCommit=${GIT_COMMIT}'"

# 运行阶段
FROM alpine:latest

# 安装运行依赖
RUN apk add --no-cache ca-certificates iptables

# 复制二进制文件和配置
COPY --from=builder /app/build/sd-wan-client /usr/local/bin/
COPY --from=builder /app/configs/config.yaml /etc/sd-wan/

# 设置工作目录
WORKDIR /etc/sd-wan

# 启动命令
ENTRYPOINT ["/usr/local/bin/sd-wan-client"]
CMD ["-c", "/etc/sd-wan/config.yaml"] 