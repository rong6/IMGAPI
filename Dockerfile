# 使用官方Go镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖（curl用于美团图床）
RUN apk add --no-cache curl

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o imgapi .

# 使用更小的alpine镜像作为运行环境
FROM alpine:latest

# 安装curl（美团图床需要）和ca-certificates（HTTPS请求需要）
RUN apk --no-cache add curl ca-certificates

# 创建非root用户
RUN addgroup -g 1001 -S imgapi && \
    adduser -S imgapi -u 1001 -G imgapi

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/imgapi .

# 复制配置文件
COPY config.yaml ./

# 更改文件所有者
RUN chown -R imgapi:imgapi /app

# 切换到非root用户
USER imgapi

# 暴露端口
EXPOSE 8080

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/getinfo || exit 1

# 启动应用
CMD ["./imgapi"]
