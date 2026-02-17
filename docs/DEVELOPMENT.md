# K-Admin 开发指南

本文档提供 K-Admin 系统的开发指南，包括代码结构、开发规范、测试指南等。

## 目录

- [项目结构](#项目结构)
- [开发环境设置](#开发环境设置)
- [添加新模块](#添加新模块)
- [代码生成器使用](#代码生成器使用)
- [测试指南](#测试指南)
- [代码规范](#代码规范)
- [常见问题](#常见问题)

## 项目结构

### 后端结构

```
backend/
├── api/                    # API 控制器层
│   └── v1/
│       ├── system/        # 系统模块 API
│       └── tools/         # 工具模块 API
├── config/                # 配置文件和配置管理
│   ├── config.go         # 配置结构定义
│   └── casbin_model.conf # Casbin 权限模型
├── core/                  # 核心功能
│   ├── gorm.go           # 数据库初始化
│   ├── redis.go          # Redis 初始化
│   ├── zap.go            # 日志初始化
│   ├── casbin.go         # Casbin 初始化
│   └── migration.go      # 数据库迁移
├── global/               # 全局变量
│   └── global.go         # 全局实例（DB, Redis, Logger 等）
├── middleware/           # 中间件
│   ├── jwt.go           # JWT 认证中间件
│   ├── casbin.go        # Casbin 权限中间件
│   ├── cors.go          # CORS 中间件
│   ├── rate_limit.go    # 限流中间件
│   ├── logger.go        # 日志中间件
│   └── recovery.go      # 恢复中间件
├── model/               # 数据模型
│   ├── common/         # 公共模型
│   │   ├── base.go     # 基础模型
│   │   └── response.go # 响应结构
│   └── system/         # 系统模块模型
│       ├── sys_user.go
│       ├── sys_role.go
│       └── sys_menu.go
├── router/             # 路由
│   ├── system/        # 系统模块路由
│   └── tools/         # 工具模块路由
├── service/           # 业务逻辑层
│   ├── system/       # 系统模块服务
│   └── tools/        # 工具模块服务
├── utils/            # 工具函数
│   ├── jwt.go       # JWT 工具
│   ├── hash.go      # 密码哈希工具
│   └── casbin.go    # Casbin 管理工具
├── docs/            # Swagger 文档（自动生成）
├── logs/            # 日志文件
├── config.yaml      # 开发环境配置
├── config.prod.yaml # 生产环境配置
├── Dockerfile       # Docker 镜像构建文件
└── main.go          # 程序入口
```

### 前端结构

```
frontend/
├── src/
│   ├── api/              # API 接口定义
│   │   ├── user.ts
│   │   ├── role.ts
│   │   └── menu.ts
│   ├── assets/           # 静态资源
│   ├── components/       # 公共组件
│   │   ├── Auth/        # 权限组件
│   │   ├── ProTable/    # 表格组件
│   │   ├── Theme/       # 主题组件
│   │   └── ErrorBoundary/ # 错误边界
│   ├── hooks/           # 自定义 Hooks
│   ├── layout/          # 布局组件
│   │   ├── Sidebar/    # 侧边栏
│   │   ├── Header/     # 头部
│   │   ├── Tabs/       # 标签页
│   │   └── index.tsx   # 主布局
│   ├── router/         # 路由配置
│   │   ├── generator.ts # 动态路由生成
│   │   ├── guards.ts   # 路由守卫
│   │   └── index.tsx   # 路由配置
│   ├── store/          # 状态管理
│   │   ├── userStore.ts # 用户状态
│   │   └── appStore.ts  # 应用状态
│   ├── utils/          # 工具函数
│   │   ├── request.ts  # HTTP 请求封装
│   │   ├── storage.ts  # 本地存储
│   │   ├── format.ts   # 格式化工具
│   │   └── validator.ts # 验证工具
│   ├── views/          # 页面组件
│   │   ├── login/     # 登录页
│   │   ├── dashboard/ # 仪表板
│   │   ├── system/    # 系统管理页面
│   │   │   ├── user/
│   │   │   ├── role/
│   │   │   └── menu/
│   │   ├── tools/     # 开发工具页面
│   │   │   ├── db-inspector/
│   │   │   └── code-generator/
│   │   └── error/     # 错误页面
│   ├── App.tsx        # 根组件
│   └── main.tsx       # 入口文件
├── public/            # 公共静态资源
├── .env.development   # 开发环境变量
├── .env.production    # 生产环境变量
├── Dockerfile         # Docker 镜像构建文件
├── nginx.conf         # Nginx 配置
├── package.json       # 依赖配置
└── vite.config.ts     # Vite 配置
```

## 开发环境设置

### 后端开发环境

1. **安装 Go**（1.21+）

2. **安装依赖**：
   ```bash
   cd backend
   go mod download
   ```

3. **配置数据库**：
   - 创建 MySQL 数据库：`CREATE DATABASE k_admin;`
   - 修改 `config.yaml` 中的数据库配置

4. **配置 Redis**：
   - 启动 Redis 服务
   - 修改 `config.yaml` 中的 Redis 配置

5. **运行项目**：
   ```bash
   go run main.go
   ```

6. **生成 Swagger 文档**：
   ```bash
   # 安装 swag
   go install github.com/swaggo/swag/cmd/swag@latest
   
   # 生成文档
   swag init
   ```

### 前端开发环境

1. **安装 Node.js**（20+）

2. **安装 pnpm**：
   ```bash
   npm install -g pnpm
   ```

3. **安装依赖**：
   ```bash
   cd frontend
   pnpm install
   ```

4. **配置环境变量**：
   - 复制 `.env.development` 并修改 API 地址

5. **运行项目**：
   ```bash
   pnpm dev
   ```

## 添加新模块

### 后端模块开发流程

以添加"文章管理"模块为例：

#### 1. 创建数据模型

在 `model/system/sys_article.go`：

```go
package system

import "k-admin/model/common"

type SysArticle struct {
	common.BaseModel
	Title      string `json:"title" gorm:"type:varchar(200);not null;comment:标题"`
	Content    string `json:"content" gorm:"type:text;comment:内容"`
	AuthorID   uint   `json:"author_id" gorm:"not null;comment:作者ID"`
	CategoryID uint   `json:"category_id" gorm:"comment:分类ID"`
	Status     int    `json:"status" gorm:"default:1;comment:状态 1-草稿 2-已发布"`
	ViewCount  int    `json:"view_count" gorm:"default:0;comment:浏览次数"`
}

func (SysArticle) TableName() string {
	return "sys_article"
}
```

#### 2. 实现服务层

在 `service/system/article_service.go`：

```go
package system

import (
	"k-admin/global"
	"k-admin/model/system"
)

type ArticleService struct{}

func (s *ArticleService) CreateArticle(article *system.SysArticle) error {
	return global.DB.Create(article).Error
}

func (s *ArticleService) GetArticleByID(id uint) (*system.SysArticle, error) {
	var article system.SysArticle
	err := global.DB.First(&article, id).Error
	return &article, err
}

func (s *ArticleService) GetArticleList(page, pageSize int) ([]system.SysArticle, int64, error) {
	var articles []system.SysArticle
	var total int64
	
	offset := (page - 1) * pageSize
	err := global.DB.Model(&system.SysArticle{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = global.DB.Offset(offset).Limit(pageSize).Find(&articles).Error
	return articles, total, err
}

func (s *ArticleService) UpdateArticle(article *system.SysArticle) error {
	return global.DB.Save(article).Error
}

func (s *ArticleService) DeleteArticle(id uint) error {
	return global.DB.Delete(&system.SysArticle{}, id).Error
}
```

#### 3. 创建 API 控制器

在 `api/v1/system/article.go`：

```go
package system

import (
	"net/http"
	"strconv"

	"k-admin/model/common"
	"k-admin/model/system"
	systemService "k-admin/service/system"

	"github.com/gin-gonic/gin"
)

var articleService = new(systemService.ArticleService)

// CreateArticle godoc
// @Summary 创建文章
// @Description 创建新文章
// @Tags Article
// @Accept json
// @Produce json
// @Param article body system.SysArticle true "文章信息"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response
// @Security ApiKeyAuth
// @Router /article [post]
func CreateArticle(c *gin.Context) {
	var article system.SysArticle
	if err := c.ShouldBindJSON(&article); err != nil {
		common.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := articleService.CreateArticle(&article); err != nil {
		common.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}

	common.OkWithData(c, article)
}

// GetArticleList godoc
// @Summary 获取文章列表
// @Description 分页获取文章列表
// @Tags Article
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} common.Response
// @Security ApiKeyAuth
// @Router /article/list [get]
func GetArticleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	articles, total, err := articleService.GetArticleList(page, pageSize)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}

	common.OkWithData(c, gin.H{
		"list":      articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 其他 API 方法...
```

#### 4. 注册路由

在 `router/system/article.go`：

```go
package system

import (
	"k-admin/api/v1/system"
	"k-admin/middleware"

	"github.com/gin-gonic/gin"
)

func InitArticleRouter(router *gin.RouterGroup) {
	articleRouter := router.Group("/article")
	articleRouter.Use(middleware.JWT())
	articleRouter.Use(middleware.Casbin())
	{
		articleRouter.POST("", system.CreateArticle)
		articleRouter.GET("/list", system.GetArticleList)
		articleRouter.GET("/:id", system.GetArticle)
		articleRouter.PUT("/:id", system.UpdateArticle)
		articleRouter.DELETE("/:id", system.DeleteArticle)
	}
}
```

在 `main.go` 中注册：

```go
systemRouter.InitArticleRouter(apiV1)
```

#### 5. 添加数据库迁移

在 `core/migration.go` 的 `AutoMigrate` 函数中添加：

```go
&system.SysArticle{},
```

#### 6. 生成 Swagger 文档

```bash
swag init
```

### 前端模块开发流程

#### 1. 创建 API 接口

在 `src/api/article.ts`：

```typescript
import request from '@/utils/request';

export interface Article {
  id: number;
  title: string;
  content: string;
  author_id: number;
  category_id: number;
  status: number;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface ArticleListParams {
  page: number;
  page_size: number;
}

export const getArticleList = (params: ArticleListParams) => {
  return request.get<{
    list: Article[];
    total: number;
    page: number;
    page_size: number;
  }>('/article/list', { params });
};

export const createArticle = (data: Partial<Article>) => {
  return request.post('/article', data);
};

export const updateArticle = (id: number, data: Partial<Article>) => {
  return request.put(`/article/${id}`, data);
};

export const deleteArticle = (id: number) => {
  return request.delete(`/article/${id}`);
};
```

#### 2. 创建页面组件

在 `src/views/system/article/index.tsx`：

```tsx
import React, { useRef } from 'react';
import { Button, Space, Tag } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import ProTable, { ProTableRef } from '@/components/ProTable';
import { getArticleList, deleteArticle, Article } from '@/api/article';
import ArticleModal from './components/ArticleModal';

const ArticlePage: React.FC = () => {
  const tableRef = useRef<ProTableRef>(null);
  const [modalVisible, setModalVisible] = React.useState(false);
  const [currentArticle, setCurrentArticle] = React.useState<Article | undefined>();

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '标题',
      dataIndex: 'title',
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: number) => (
        <Tag color={status === 2 ? 'green' : 'default'}>
          {status === 2 ? '已发布' : '草稿'}
        </Tag>
      ),
    },
    {
      title: '浏览次数',
      dataIndex: 'view_count',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
    },
    {
      title: '操作',
      render: (_: any, record: Article) => (
        <Space>
          <Button type="link" onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  const handleEdit = (article: Article) => {
    setCurrentArticle(article);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    await deleteArticle(id);
    tableRef.current?.reload();
  };

  return (
    <>
      <ProTable
        ref={tableRef}
        columns={columns}
        request={getArticleList}
        toolbar={{
          actions: [
            <Button
              key="create"
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setModalVisible(true)}
            >
              新建文章
            </Button>,
          ],
        }}
      />
      <ArticleModal
        visible={modalVisible}
        article={currentArticle}
        onCancel={() => {
          setModalVisible(false);
          setCurrentArticle(undefined);
        }}
        onSuccess={() => {
          setModalVisible(false);
          setCurrentArticle(undefined);
          tableRef.current?.reload();
        }}
      />
    </>
  );
};

export default ArticlePage;
```

#### 3. 添加菜单

在系统的菜单管理中添加新菜单项，或在数据库中插入：

```sql
INSERT INTO sys_menu (parent_id, path, name, component, sort, meta) VALUES
(1, '/system/article', 'Article', 'system/article/index', 3, '{"title":"文章管理","icon":"file-text"}');
```

## 代码生成器使用

K-Admin 提供了强大的代码生成器，可以快速生成 CRUD 代码。

### 使用步骤

1. **登录系统**，进入"开发工具" -> "代码生成器"

2. **选择表**：
   - 从现有表中选择
   - 或创建新表

3. **配置生成选项**：
   - 结构体名称：如 `SysArticle`
   - 包名：如 `system`
   - 前端路径：如 `system/article`
   - 选择要生成的文件类型

4. **预览代码**：
   - 查看生成的代码
   - 确认无误后点击生成

5. **生成代码**：
   - 代码将自动写入对应目录
   - 需要手动在 `main.go` 中注册路由
   - 需要手动在菜单管理中添加菜单

### 生成的文件

- 后端：
  - `model/system/sys_article.go` - 数据模型
  - `service/system/article_service.go` - 业务逻辑
  - `api/v1/system/article.go` - API 控制器
  - `router/system/article.go` - 路由注册

- 前端：
  - `src/api/article.ts` - API 接口
  - `src/views/system/article/index.tsx` - 列表页面
  - `src/views/system/article/components/ArticleModal.tsx` - 表单弹窗

## 测试指南

### 后端测试

#### 单元测试

```go
// api/v1/system/article_test.go
package system

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 初始化测试数据库
	// ...

	router := gin.New()
	router.POST("/article", CreateArticle)

	article := map[string]interface{}{
		"title":   "测试文章",
		"content": "测试内容",
	}
	body, _ := json.Marshal(article)

	req, _ := http.NewRequest("POST", "/article", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
```

#### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./api/v1/system

# 显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 前端测试

```typescript
// src/views/system/article/__tests__/index.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import ArticlePage from '../index';

describe('ArticlePage', () => {
  it('renders article list', () => {
    render(<ArticlePage />);
    expect(screen.getByText('新建文章')).toBeInTheDocument();
  });

  it('opens modal when create button clicked', () => {
    render(<ArticlePage />);
    const createButton = screen.getByText('新建文章');
    fireEvent.click(createButton);
    // 验证弹窗打开
  });
});
```

## 代码规范

### Go 代码规范

1. **命名规范**：
   - 包名：小写，简短，有意义
   - 文件名：小写，下划线分隔
   - 变量名：驼峰命名
   - 常量名：大写，下划线分隔

2. **注释规范**：
   - 所有导出的函数、类型、常量都要有注释
   - 注释以名称开头
   - Swagger 注释要完整

3. **错误处理**：
   - 不要忽略错误
   - 使用有意义的错误信息
   - 适当使用日志记录

### TypeScript 代码规范

1. **命名规范**：
   - 文件名：驼峰命名或短横线分隔
   - 组件名：大驼峰
   - 变量名：小驼峰
   - 常量名：大写，下划线分隔

2. **类型定义**：
   - 优先使用 interface
   - 导出所有公共类型
   - 避免使用 any

3. **组件规范**：
   - 使用函数组件
   - 使用 TypeScript
   - Props 要有类型定义

## 常见问题

### 1. 数据库连接失败

检查 `config.yaml` 中的数据库配置是否正确，确保 MySQL 服务已启动。

### 2. Redis 连接失败

检查 Redis 服务是否启动，配置是否正确。

### 3. 前端请求跨域

开发环境使用 Vite 代理，生产环境使用 Nginx 反向代理。

### 4. Token 过期

前端会自动使用 refresh token 刷新 access token，如果 refresh token 也过期，会跳转到登录页。

### 5. 权限不足

检查用户角色是否有对应的菜单和 API 权限。

## 更多资源

- [Go 官方文档](https://golang.org/doc/)
- [Gin 文档](https://gin-gonic.com/docs/)
- [Gorm 文档](https://gorm.io/docs/)
- [React 文档](https://react.dev/)
- [Ant Design 文档](https://ant.design/)
- [Vite 文档](https://vitejs.dev/)
