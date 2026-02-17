# AGENTS.md - 后端开发指南

## 项目概述

使用 Gin 框架、GORM ORM 和 MySQL 的 Go 后端项目，提供游戏应用 RESTful API，支持 JWT 认证。

## 项目结构

```
src/
├── main.go              # 应用入口
├── controller/          # HTTP 处理器
├── dto/                 # 数据传输对象
├── models/              # GORM 数据模型
├── middleware/          # 中间件（JWT、限流、分页）
├── routes/              # 路由注册
├── config/              # 配置
├── utils/               # 工具函数
├── docs/                # Swagger 文档
└── test/                # 测试工具
```

## 构建与测试

### 构建运行

```bash
cd src
go mod download          # 下载依赖
go build -o main .        # 构建
go run .                  # 运行
```

### Docker 部署

```bash
docker build -t firechickenmp4/backend-app:latest .
docker-compose up -d
```

### 测试

```bash
go test ./...             # 所有测试
go test -v ./controller/  # 单文件测试
go test -v -run TestName ./  # 单函数测试
go test -cover ./...      # 覆盖率
```

### 代码检查

```bash
go vet ./...              # 静态检查
go fmt ./...              # 格式化
```

### Swagger 文档

```bash
swag init -g main.go -o docs
```

## 代码风格规范

### 导入顺序

1. 标准库（fmt, log, errors 等）
2. 第三方包（github.com/...）
3. 项目内部包

```go
import (
    "errors"
    "log"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    config "MuXi/2026-MuxiShooter-Backend/config"
    "MuXi/2026-MuxiShooter-Backend/dto"
)
```

### 命名约定

- 文件：小写 + 下划线（controller.go）
- 函数/变量：驼峰命名法
- 结构体/类型：帕斯卡命名法
- 数据库字段：GORM 标签中用 snake_case

```go
type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"user_id"`
    Username  string    `gorm:"unique;not null" json:"username"`
    Password  string    `gorm:"not null" json:"-"`
}
```

### GORM 模型规范

- 主键：`ID uint \`gorm:"primaryKey;autoIncrement"\``
- 时间戳：使用 `CreatedAt` 和 `UpdatedAt`
- 敏感字段：`json:"-"` 隐藏
- 关联关系：设置外键约束

### DTO 规范

- 请求 DTO：使用 binding 标签验证
- 响应 DTO：统一 `Code`、`Message`、`Data` 结构
- 表单参数用 `form` 标签，JSON 用 `json` 标签

### 错误处理

- 使用 `errors.Is()` 比较错误
- 返回 HTTP 状态码：400、401、403、404、409、500

```go
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusNotFound, dto.Response{...})
    } else {
        c.JSON(http.StatusInternalServerError, dto.Response{...})
    }
    return
}
```

### 事务处理

```go
tx := config.DB.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()
// ... 操作 ...
if err := tx.Commit().Error; err != nil {
    tx.Rollback()
}
```

### Swagger 注释

```go
// @Summary     用户注册
// @Description 注册新用户
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body dto.RegisterRequest true "注册请求"
// @Success     200 {object} dto.Response{data=dto.AuthData}
// @Failure     400 {object} dto.Response
// @Router      /api/auth/register [post]
func Register(c *gin.Context) { ... }
```

## 环境变量

```bash
DB_HOST=localhost
DB_PORT=3306
DB_USER=adminuser
DB_PASSWORD=your_password
DB_NAME=mini
JWT_SECRET=base64_encoded
ADMIN_PASSWORD=admin_pass
GIN_MODE=release
```

## 添加新接口流程

1. 在 `dto/` 添加请求/响应结构
2. 在 `models/` 添加模型（如需）
3. 在 `controller/` 创建处理器并添加 Swagger 注释
4. 在 `routes/` 注册路由
5. 运行 `swag init` 生成文档

## 运行应用

```bash
cd src
go run .
# 服务启动在 http://localhost:8080
# Swagger 文档：http://localhost:8080/swagger/index.html
```
