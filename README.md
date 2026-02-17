# K-Admin 管理系统

一个基于 Go + React 的现代化后台管理系统，提供完整的 RBAC 权限管理、数据库管理工具和代码生成器。

## ✨ 特性

- 🔐 **完整的权限管理系统**：基于 RBAC 的用户、角色、菜单权限管理
- 🛡️ **安全认证**：JWT 双 token 机制（access + refresh token）
- 🎨 **现代化 UI**：基于 Ant Design 5 的响应式界面
- 🔧 **开发者工具**：
  - 数据库管理器：可视化查看和管理数据库表
  - 代码生成器：一键生成前后端 CRUD 代码
- 📊 **动态路由**：基于权限的动态菜单和路由生成
- 🌓 **主题切换**：支持亮色/暗色主题
- 📝 **API 文档**：集成 Swagger 自动生成 API 文档
- 🐳 **容器化部署**：提供完整的 Docker 和 Docker Compose 配置

## 🛠️ 技术栈

### 后端
- **框架**：Gin (Go Web Framework)
- **ORM**：Gorm
- **数据库**：MySQL 8.0
- **缓存**：Redis
- **权限**：Casbin (RBAC)
- **日志**：Zap + Lumberjack
- **配置**：Viper
- **文档**：Swagger

### 前端
- **框架**：React 18 + TypeScript
- **构建工具**：Vite
- **UI 库**：Ant Design 5
- **状态管理**：Zustand
- **路由**：React Router v6
- **HTTP 客户端**：Axios
- **代码编辑器**：Monaco Editor

## 📦 项目结构

```
k-admin-system/
├── backend/                 # 后端代码
│   ├── api/                # API 控制器
│   ├── config/             # 配置文件
│   ├── core/               # 核心功能（数据库、日志等）
│   ├── global/             # 全局变量
│   ├── middleware/         # 中间件
│   ├── model/              # 数据模型
│   ├── router/             # 路由
│   ├── service/            # 业务逻辑
│   └── utils/              # 工具函数
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── api/           # API 接口定义
│   │   ├── components/    # 公共组件
│   │   ├── hooks/         # 自定义 Hooks
│   │   ├── layout/        # 布局组件
│   │   ├── router/        # 路由配置
│   │   ├── store/         # 状态管理
│   │   ├── utils/         # 工具函数
│   │   └── views/         # 页面组件
│   └── public/            # 静态资源
└── docs/                   # 文档
```

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Node.js 20+
- MySQL 8.0+
- Redis 7+
- pnpm (推荐) 或 npm

### 本地开发

#### 1. 克隆项目

```bash
git clone <repository-url>
cd k-admin-system
```

#### 2. 启动后端

```bash
cd backend

# 安装依赖
go mod download

# 复制配置文件
cp config.yaml.example config.yaml

# 修改配置文件中的数据库和 Redis 连接信息
# 编辑 config.yaml

# 运行数据库迁移（首次运行）
go run main.go

# 启动服务
go run main.go
```

后端服务将在 `http://localhost:8080` 启动

#### 3. 启动前端

```bash
cd frontend

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
```

前端服务将在 `http://localhost:3000` 启动

#### 4. 访问应用

- 前端：http://localhost:3000
- 后端 API：http://localhost:8080/api/v1
- Swagger 文档：http://localhost:8080/swagger/index.html

默认管理员账号：
- 用户名：admin
- 密码：admin123

## 🐳 Docker 部署

### 使用 Docker Compose（推荐）

```bash
# 复制环境变量文件
cp .env.example .env

# 修改 .env 文件中的配置
# 编辑 .env

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

服务将在以下端口启动：
- 前端：http://localhost:80
- 后端：http://localhost:8080
- MySQL：localhost:3306
- Redis：localhost:6379

### 单独构建镜像

#### 后端

```bash
cd backend
docker build -t k-admin-backend .
docker run -p 8080:8080 k-admin-backend
```

#### 前端

```bash
cd frontend
docker build -t k-admin-frontend .
docker run -p 80:80 k-admin-frontend
```

## 📖 API 文档

启动后端服务后，访问 Swagger 文档：

```
http://localhost:8080/swagger/index.html
```

### 重新生成 Swagger 文档

```bash
cd backend

# 安装 swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init
```

## 🧪 测试

### 后端测试

```bash
cd backend

# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 前端测试

```bash
cd frontend

# 运行单元测试
pnpm test

# 运行测试并显示覆盖率
pnpm test:coverage
```

## 🔧 开发指南

### 添加新模块

1. **后端**：
   - 在 `model/` 创建数据模型
   - 在 `service/` 实现业务逻辑
   - 在 `api/` 创建控制器
   - 在 `router/` 注册路由
   - 添加 Swagger 注释

2. **前端**：
   - 在 `api/` 定义 API 接口
   - 在 `views/` 创建页面组件
   - 在菜单管理中添加菜单项

### 使用代码生成器

1. 登录系统，进入"开发工具" -> "代码生成器"
2. 选择数据库表或创建新表
3. 配置生成选项（结构体名称、包名等）
4. 预览生成的代码
5. 确认生成，代码将自动写入对应目录

## 📝 配置说明

### 后端配置 (config.yaml)

```yaml
server:
  port: ":8080"
  mode: "debug"  # debug, release, test

database:
  host: "localhost"
  port: 3306
  name: "k_admin"
  username: "root"
  password: "password"

jwt:
  secret: "your-secret-key"
  access_expiration: 15   # minutes
  refresh_expiration: 7   # days

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
```

### 前端配置 (.env)

```env
# API Base URL
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Application Title
VITE_APP_TITLE=K-Admin 管理系统
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[MIT License](LICENSE)

## 📧 联系方式

如有问题或建议，请提交 Issue 或联系维护者。
