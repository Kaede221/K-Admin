# K-Admin API 文档

## 概述

K-Admin 系统提供 RESTful API，所有 API 端点都以 `/api/v1` 为前缀。

## 认证

大多数 API 需要 JWT 认证。在请求头中包含 token：

```
Authorization: Bearer <access_token>
```

### Token 刷新

当 access token 过期时，使用 refresh token 获取新的 access token。

## 统一响应格式

所有 API 响应遵循统一格式：

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "error message",
  "data": null
}
```

## API 端点

### 1. 认证模块

#### 1.1 用户登录

```
POST /api/v1/user/login
```

**请求体：**

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应：**

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员",
      "email": "admin@example.com",
      "role_id": 1
    }
  }
}
```

#### 1.2 刷新 Token

```
POST /api/v1/user/refresh
```

**请求头：**

```
Authorization: Bearer <refresh_token>
```

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGc..."
  }
}
```

#### 1.3 用户登出

```
POST /api/v1/user/logout
```

**请求头：**

```
Authorization: Bearer <access_token>
```

### 2. 用户管理

#### 2.1 获取用户列表

```
GET /api/v1/user/list?page=1&page_size=10&username=&role_id=&status=
```

**查询参数：**

- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 10）
- `username`: 用户名（可选）
- `role_id`: 角色 ID（可选）
- `status`: 状态（可选）

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "username": "admin",
        "nickname": "管理员",
        "email": "admin@example.com",
        "phone": "13800138000",
        "role_id": 1,
        "active": true,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

#### 2.2 创建用户

```
POST /api/v1/user
```

**请求体：**

```json
{
  "username": "newuser",
  "password": "password123",
  "nickname": "新用户",
  "email": "user@example.com",
  "phone": "13800138000",
  "role_id": 2
}
```

#### 2.3 更新用户

```
PUT /api/v1/user/:id
```

**请求体：**

```json
{
  "nickname": "更新的昵称",
  "email": "newemail@example.com",
  "phone": "13900139000",
  "role_id": 2
}
```

#### 2.4 删除用户

```
DELETE /api/v1/user/:id
```

#### 2.5 修改密码

```
POST /api/v1/user/change-password
```

**请求体：**

```json
{
  "old_password": "oldpass123",
  "new_password": "newpass123"
}
```

#### 2.6 重置密码

```
POST /api/v1/user/:id/reset-password
```

**请求体：**

```json
{
  "new_password": "resetpass123"
}
```

#### 2.7 切换用户状态

```
POST /api/v1/user/:id/toggle-status
```

### 3. 角色管理

#### 3.1 获取角色列表

```
GET /api/v1/role/list?page=1&page_size=10
```

#### 3.2 创建角色

```
POST /api/v1/role
```

**请求体：**

```json
{
  "role_name": "编辑",
  "role_key": "editor",
  "data_scope": 2,
  "sort": 2,
  "status": true,
  "remark": "编辑角色"
}
```

#### 3.3 更新角色

```
PUT /api/v1/role/:id
```

#### 3.4 删除角色

```
DELETE /api/v1/role/:id
```

#### 3.5 分配菜单权限

```
POST /api/v1/role/:id/menus
```

**请求体：**

```json
{
  "menu_ids": [1, 2, 3, 4, 5]
}
```

#### 3.6 获取角色菜单

```
GET /api/v1/role/:id/menus
```

#### 3.7 分配 API 权限

```
POST /api/v1/role/:id/apis
```

**请求体：**

```json
{
  "policies": [
    {
      "path": "/api/v1/user/*",
      "method": "GET"
    },
    {
      "path": "/api/v1/user/*",
      "method": "POST"
    }
  ]
}
```

#### 3.8 获取角色 API 权限

```
GET /api/v1/role/:id/apis
```

### 4. 菜单管理

#### 4.1 获取菜单树

```
GET /api/v1/menu/tree
```

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "parent_id": 0,
      "path": "/system",
      "name": "System",
      "component": "Layout",
      "sort": 1,
      "meta": {
        "title": "系统管理",
        "icon": "setting",
        "hidden": false
      },
      "children": [
        {
          "id": 2,
          "parent_id": 1,
          "path": "/system/user",
          "name": "User",
          "component": "system/user/index",
          "sort": 1,
          "meta": {
            "title": "用户管理",
            "icon": "user"
          }
        }
      ]
    }
  ]
}
```

#### 4.2 获取所有菜单

```
GET /api/v1/menu/list
```

#### 4.3 创建菜单

```
POST /api/v1/menu
```

**请求体：**

```json
{
  "parent_id": 1,
  "path": "/system/role",
  "name": "Role",
  "component": "system/role/index",
  "sort": 2,
  "meta": {
    "title": "角色管理",
    "icon": "team",
    "hidden": false,
    "keep_alive": true
  },
  "btn_perms": ["create", "edit", "delete"]
}
```

#### 4.4 更新菜单

```
PUT /api/v1/menu/:id
```

#### 4.5 删除菜单

```
DELETE /api/v1/menu/:id
```

### 5. 数据库管理工具

#### 5.1 获取表列表

```
GET /api/v1/tools/db/tables
```

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "name": "sys_user",
      "comment": "用户表",
      "rows": 100,
      "engine": "InnoDB"
    }
  ]
}
```

#### 5.2 获取表结构

```
GET /api/v1/tools/db/table/:tableName/schema
```

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "columns": [
      {
        "name": "id",
        "type": "bigint",
        "nullable": false,
        "key": "PRI",
        "default": null,
        "comment": "主键ID"
      }
    ]
  }
}
```

#### 5.3 获取表数据

```
GET /api/v1/tools/db/table/:tableName/data?page=1&page_size=10
```

#### 5.4 执行 SQL

```
POST /api/v1/tools/db/execute
```

**请求体：**

```json
{
  "sql": "SELECT * FROM sys_user LIMIT 10",
  "read_only": true
}
```

#### 5.5 创建记录

```
POST /api/v1/tools/db/table/:tableName/record
```

#### 5.6 更新记录

```
PUT /api/v1/tools/db/table/:tableName/record/:id
```

#### 5.7 删除记录

```
DELETE /api/v1/tools/db/table/:tableName/record/:id
```

### 6. 代码生成器

#### 6.1 获取表元数据

```
GET /api/v1/tools/gen/metadata/:tableName
```

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "table_name": "sys_user",
    "comment": "用户表",
    "columns": [
      {
        "name": "id",
        "type": "bigint",
        "go_type": "int64",
        "ts_type": "number",
        "comment": "主键ID"
      }
    ]
  }
}
```

#### 6.2 预览生成代码

```
POST /api/v1/tools/gen/preview
```

**请求体：**

```json
{
  "table_name": "sys_user",
  "struct_name": "SysUser",
  "package_name": "system",
  "frontend_path": "system/user",
  "options": {
    "generate_model": true,
    "generate_service": true,
    "generate_api": true,
    "generate_router": true,
    "generate_frontend": true
  }
}
```

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "files": {
      "model/system/sys_user.go": "package system\n\n...",
      "service/system/user_service.go": "package system\n\n...",
      "api/v1/system/user.go": "package system\n\n...",
      "router/system/user.go": "package system\n\n...",
      "frontend/api/user.ts": "import request from '@/utils/request';\n\n...",
      "frontend/types/user.ts": "export interface User {\n...",
      "frontend/views/system/user/index.tsx": "import React from 'react';\n\n..."
    }
  }
}
```

#### 6.3 生成代码

```
POST /api/v1/tools/gen/generate
```

**请求体：** 同预览接口

#### 6.4 创建表

```
POST /api/v1/tools/gen/create-table
```

**请求体：**

```json
{
  "table_name": "sys_log",
  "comment": "系统日志表",
  "fields": [
    {
      "name": "id",
      "type": "bigint",
      "nullable": false,
      "primary": true,
      "auto_increment": true,
      "comment": "主键ID"
    },
    {
      "name": "user_id",
      "type": "bigint",
      "nullable": false,
      "comment": "用户ID"
    },
    {
      "name": "action",
      "type": "varchar(100)",
      "nullable": false,
      "comment": "操作"
    }
  ]
}
```

### 7. 健康检查

#### 7.1 健康检查

```
GET /api/v1/health
```

**响应：**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或 token 无效 |
| 403 | 无权限访问 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

## 使用示例

### JavaScript/TypeScript

```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// 添加请求拦截器
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 登录
const login = async (username: string, password: string) => {
  const response = await api.post('/user/login', { username, password });
  return response.data;
};

// 获取用户列表
const getUserList = async (page: number, pageSize: number) => {
  const response = await api.get('/user/list', {
    params: { page, page_size: pageSize },
  });
  return response.data;
};
```

### cURL

```bash
# 登录
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 获取用户列表
curl -X GET "http://localhost:8080/api/v1/user/list?page=1&page_size=10" \
  -H "Authorization: Bearer <access_token>"

# 创建用户
curl -X POST http://localhost:8080/api/v1/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "username":"newuser",
    "password":"password123",
    "nickname":"新用户",
    "email":"user@example.com",
    "role_id":2
  }'
```

## 更多信息

完整的 API 文档请访问 Swagger UI：

```
http://localhost:8080/swagger/index.html
```
