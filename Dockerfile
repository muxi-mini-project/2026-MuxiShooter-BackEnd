# 第一阶段：构建
FROM golang:1.25-alpine AS builder

# 设置Go模块代理（在RUN之前设置环境变量）
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 复制依赖文件
COPY src/go.mod src/go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY ./src .

# 构建可执行文件（Linux版）
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 第二阶段：运行
FROM alpine:latest

# 创建应用目录和上传目录
RUN mkdir -p /app/uploads

# 复制默认头像
COPY ./uploads/DefaultHeadImg.jpeg /app/uploads/DefaultHeadImg.jpeg

# 安装必要的运行时（如需要CA证书）
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户并设置 UID (与宿主机用户匹配)
ARG USER_ID=1000
ARG GROUP_ID=1000

RUN addgroup -g ${GROUP_ID} appuser && \
    adduser -u ${USER_ID} -G appuser -s /bin/sh -D appuser

# 目录权限
RUN chown -R appuser:appuser /app

# 切换非 root 用户
USER appuser

WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/main .

# 暴露端口（根据你的main.go中监听的端口）
EXPOSE 8080

# 运行程序
CMD ["./main"]