# K-Admin Backend

K-Admin 是一个基于 Go + Gin + Gorm 的后台管理系统后端。

## 技术栈

- **Web框架**: Gin
- **ORM**: Gorm
- **数据库**: MySQL
- **缓存**: Redis
- **配置管理**: Viper
- **日志**: Zap + Lumberjack
- **认证**: JWT
- **权限**: Casbin

## 项目结构

```
backend/
├── api/           # API控制器
├── config/        # 配置管理
├── core/          # 核心功能（数据库、日志、Redis等）
├── global/        # 全局变量
├── middleware/    # 中间件
├── model/         # 数据模型
├── router/        # 路由
├── service/       # 业务逻辑
├── utils/         # 工具函数
├── logs/          # 日志文件
├── config.yaml    # 配置文件
└── main.go        # 入口文件
```

## 已完成功能 (Phase 1)

### 1. 核心基础设施
- ✅ 项目结构初始化
- ✅ 配置管理（支持YAML、JSON、环境变量）
- ✅ 日志系统（Zap + Lumberjack，支持日志轮转）
- ✅ 数据库连接（Gorm + MySQL，连接池管理）
- ✅ Redis连接
- ✅ 数据库迁移系统

### 2. 认证与安全
- ✅ JWT令牌生成与解析
- ✅ 访问令牌和刷新令牌
- ✅ 令牌黑名单（基于Redis）
- ✅ 密码加密（bcrypt）
- ✅ JWT认证中间件

### 3. 通用模块
- ✅ 基础模型（BaseModel）
- ✅ 统一响应结构（Response）
- ✅ 响应辅助函数

## 快速开始

### 前置要求

- Go 1.21+
- MySQL 5.7+
- Redis 6.0+

### 安装依赖

```bash
cd backend
go mod download
```

### 配置

复制并修改配置文件：

```bash
cp config.yaml config.local.yaml
```

编辑 `config.local.yaml`，设置数据库和Redis连接信息。

### 运行

```bash
# 使用默认配置文件 (config.yaml)
go run main.go

# 使用指定配置文件
go run main.go -config=config.local.yaml

# 编译并运行
go build -o k-admin.exe
./k-admin.exe
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行单元测试（排除属性测试）
go test -run "^Test[^P]" ./...

# 查看测试覆盖率
go test -cover ./...
```

## API文档

服务启动后，访问：
- 健康检查: http://localhost:8080/health

## 环境变量

支持通过环境变量覆盖配置：

```bash
export KADMIN_SERVER_PORT=":8080"
export KADMIN_DATABASE_HOST="localhost"
export KADMIN_DATABASE_PASSWORD="your-password"
export KADMIN_JWT_SECRET="your-secret-key"
```

## 开发指南

### 添加新模块

1. 在 `model/` 创建数据模型
2. 在 `service/` 实现业务逻辑
3. 在 `api/` 创建API控制器
4. 在 `router/` 注册路由
5. 在 `core/migration.go` 注册模型迁移

### 日志使用

```go
import "k-admin-system/global"

global.Logger.Info("message", zap.String("key", "value"))
global.Logger.Error("error occurred", zap.Error(err))
```

### JWT使用

```go
import "k-admin-system/utils"

// 生成令牌
accessToken, refreshToken, err := utils.GenerateToken(userID, username, roleID)

// 解析令牌
claims, err := utils.ParseToken(tokenString)

// 刷新令牌
newAccessToken, err := utils.RefreshToken(refreshTokenString)

// 令牌加入黑名单
err := utils.AddTokenToBlacklist(tokenString)
```

## 下一步计划

- [ ] 用户管理模块
- [ ] 角色管理模块
- [ ] 菜单管理模块
- [ ] Casbin权限系统
- [ ] 更多中间件（CORS、限流、日志等）

## License

MIT
