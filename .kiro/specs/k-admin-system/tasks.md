# Implementation Plan: K-Admin System

## Overview

This implementation plan breaks down the K-Admin system into incremental, testable tasks. The approach follows a bottom-up strategy: establish core infrastructure first (database, configuration, middleware), then build system modules (user, role, menu), followed by advanced features (DB inspector, code generator), and finally frontend integration.

The backend uses Go with Gin framework, Gorm ORM, and MySQL database. The frontend uses React 18 with TypeScript, Vite, Ant Design 5, and Zustand for state management.

## Tasks

### Phase 1: Backend Foundation

- [ ] 1. Set up backend project structure and core infrastructure
  - [x] 1.1 Initialize Go module and create directory structure (api, config, core, global, middleware, model, router, service, utils)
    - Create main.go entry point
    - Set up go.mod with dependencies: gin, gorm, viper, zap, jwt-go, casbin
    - _Requirements: Project structure from PRD_
  
  - [x] 1.2 Implement configuration management using Viper
    - Create config/config.go with structs for Server, Database, JWT, Redis, Logger
    - Implement config loading from YAML, JSON, and environment variables
    - Add validation for required fields
    - _Requirements: 12.2, 12.3, 12.4_
  
  - [x] 1.3 Write property test for configuration management
    - **Property 38: Configuration Source Priority**
    - **Property 39: Configuration Validation on Startup**
    - **Validates: Requirements 12.2, 12.4, 12.5**
  
  - [x] 1.4 Implement logging system using Zap and Lumberjack
    - Create core/zap.go for logger initialization
    - Configure log levels, output destinations, and rotation
    - Implement structured logging helpers
    - _Requirements: 13.1, 13.2, 13.3, 13.7, 13.8_
  
  - [x] 1.5 Write property tests for logging system
    - **Property 40: Log Level Filtering**
    - **Property 43: Environment-Specific Log Output**
    - **Validates: Requirements 13.3, 13.7, 13.8**


- [x] 2. Set up database connection and base models
  - [x] 2.1 Implement database connection with Gorm
    - Create core/gorm.go for database initialization
    - Configure connection pooling and reconnection logic
    - Implement slow query logging
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.8_
  
  - [x] 2.2 Write property tests for database connection
    - **Property 44: Database Connection Pool Management**
    - **Property 45: Automatic Database Reconnection**
    - **Property 48: Slow Query Logging**
    - **Validates: Requirements 15.3, 15.4, 15.8**
  
  - [x] 2.3 Create base model and common structures
    - Define model/common/base.go with BaseModel (ID, CreatedAt, UpdatedAt, DeletedAt)
    - Define model/common/response.go with Response struct
    - Implement response helper functions (Ok, Fail, OkWithDetailed, FailWithCode)
    - _Requirements: 1.1, 1.4, 15.6_
  
  - [x] 2.4 Write property test for unified response structure
    - **Property 1: Unified Response Structure**
    - **Validates: Requirements 1.1, 1.2, 1.3**
  
  - [x] 2.5 Implement database migration system
    - Create initialization script for AutoMigrate
    - Add migration tracking
    - _Requirements: 15.5_
  
  - [x] 2.6 Write property test for database migrations
    - **Property 46: Database Migration Execution**
    - **Validates: Requirements 15.5**

- [x] 3. Implement authentication and JWT management
  - [x] 3.1 Create JWT utility functions
    - Define utils/jwt.go with JWTClaims struct
    - Implement GenerateToken (access + refresh tokens)
    - Implement ParseToken and RefreshToken
    - Implement token blacklist functions using Redis
    - _Requirements: 2.1, 2.5, 2.7_
  
  - [x] 3.2 Write property tests for JWT functionality
    - **Property 3: Token Generation and Refresh Cycle**
    - **Property 4: Token Blacklist Enforcement**
    - **Validates: Requirements 2.1, 2.5, 2.7**
  
  - [x] 3.3 Implement password hashing utilities
    - Create utils/hash.go with bcrypt functions
    - Implement HashPassword and CheckPassword
    - _Requirements: 2.2_
  
  - [x] 3.4 Write property test for password encryption
    - **Property 2: Password Encryption Round-Trip**
    - **Validates: Requirements 2.2**
  
  - [x] 3.5 Create JWT authentication middleware
    - Implement middleware/jwt.go
    - Extract and validate tokens from Authorization header
    - Set user info in Gin context
    - Handle token expiration and blacklist checking
    - _Requirements: 2.3, 16.1_
  
  - [x] 3.6 Write property test for JWT middleware
    - **Property 49: JWT Middleware Token Validation**
    - **Validates: Requirements 16.1**

- [x] 4. Checkpoint - Verify core infrastructure
  - Ensure all tests pass, ask the user if questions arise.


### Phase 2: Core System Modules

- [ ] 5. Implement User module (model, service, API, router)
  - [x] 5.1 Create User model and database table
    - Define model/system/sys_user.go with SysUser struct
    - Include fields: username, password, nickname, header_img, phone, email, role_id, active
    - Add Gorm tags and JSON tags (exclude password from JSON)
    - _Requirements: 4.1, 4.7_
  
  - [x] 5.2 Implement User service layer
    - Create service/system/user_service.go
    - Implement Login (validate credentials, generate tokens)
    - Implement CreateUser, UpdateUser, DeleteUser (soft delete), GetUserByID
    - Implement GetUserList with pagination and filtering
    - Implement ChangePassword, ResetPassword, ToggleUserStatus
    - _Requirements: 4.2, 4.3, 4.4, 4.5, 4.6, 4.8_
  
  - [x] 5.3 Write property tests for User service
    - **Property 9: User CRUD Consistency**
    - **Property 10: Password Field Masking**
    - **Property 11: Username Uniqueness Validation**
    - **Property 12: User List Pagination and Filtering**
    - **Property 47: Soft Delete Behavior**
    - **Validates: Requirements 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 15.7**
  
  - [x] 5.4 Create User API controllers
    - Create api/v1/system/user.go
    - Implement handlers: Login, CreateUser, UpdateUser, DeleteUser, GetUser, GetUserList
    - Implement ChangePassword, ResetPassword, ToggleStatus
    - Add Swagger annotations
    - _Requirements: 4.2, 4.3, 4.4, 4.5, 4.6, 14.3_
  
  - [x] 5.5 Write unit tests for User API
    - Test login with valid/invalid credentials
    - Test user creation with duplicate username
    - Test password masking in responses
    - Test pagination and filtering
    - _Requirements: 4.2, 4.6, 4.7, 4.8_
  
  - [x] 5.6 Register User routes
    - Create router/system/user.go
    - Register routes with appropriate middleware
    - Public routes: /login
    - Protected routes: /user/* (require JWT)
    - _Requirements: 2.3_

- [x] 6. Implement Role module (model, service, API, router)
  - [x] 6.1 Create Role model and database table
    - Define model/system/sys_role.go with SysRole struct
    - Include fields: role_name, role_key, data_scope, sort, status, remark
    - Define many-to-many relationship with SysMenu
    - _Requirements: 5.1_
  
  - [x] 6.2 Implement Role service layer
    - Create service/system/role_service.go
    - Implement CreateRole, UpdateRole, DeleteRole, GetRoleByID, GetRoleList
    - Implement AssignMenus, GetRoleMenus
    - Implement AssignAPIs (Casbin policies), GetRoleAPIs
    - Add validation to prevent deleting roles with users
    - _Requirements: 5.2, 5.3, 5.4, 5.5, 5.6_
  
  - [x] 6.3 Write property tests for Role service
    - **Property 13: Role Deletion Protection**
    - **Property 14: Role Permission Assignment**
    - **Validates: Requirements 5.3, 5.4, 5.6**
  
  - [x] 6.4 Create Role API controllers
    - Create api/v1/system/role.go
    - Implement handlers: CreateRole, UpdateRole, DeleteRole, GetRole, GetRoleList
    - Implement AssignMenus, GetRoleMenus, AssignAPIs, GetRoleAPIs
    - Add Swagger annotations
    - _Requirements: 5.2, 5.3, 5.4, 5.5, 5.6, 14.3_
  
  - [x] 6.5 Write unit tests for Role API
    - Test role creation and updates
    - Test role deletion with associated users
    - Test menu and API permission assignment
    - _Requirements: 5.2, 5.3, 5.4, 5.6_
  
  - [x] 6.6 Register Role routes
    - Create router/system/role.go
    - Register protected routes: /role/* (require JWT + admin permission)
    - _Requirements: 2.3, 3.4_


- [x] 7. Implement Menu module (model, service, API, router)
  - [x] 7.1 Create Menu model and database table
    - Define model/system/sys_menu.go with SysMenu struct
    - Include fields: parent_id, path, name, component, sort, meta (JSON), btn_perms (JSON)
    - Define many-to-many relationship with SysRole
    - _Requirements: 6.1, 6.3_
  
  - [x] 7.2 Implement Menu service layer
    - Create service/system/menu_service.go
    - Implement GetMenuTree (filter by role, build hierarchy)
    - Implement CreateMenu, UpdateMenu, DeleteMenu, GetMenuByID, GetAllMenus
    - Implement BuildMenuTree helper (recursive tree building)
    - _Requirements: 3.2, 6.2, 6.5, 6.6_
  
  - [x] 7.3 Write property tests for Menu service
    - **Property 5: Menu Tree Authorization Filtering**
    - **Property 15: Menu Hierarchy Preservation**
    - **Property 16: Menu Metadata Serialization Round-Trip**
    - **Property 17: Hidden Menu Route Accessibility**
    - **Validates: Requirements 3.2, 6.2, 6.3, 6.6, 6.8**
  
  - [x] 7.4 Create Menu API controllers
    - Create api/v1/system/menu.go
    - Implement handlers: GetMenuTree, CreateMenu, UpdateMenu, DeleteMenu, GetMenu, GetAllMenus
    - Add Swagger annotations
    - _Requirements: 3.2, 6.2, 6.5, 6.6, 14.3_
  
  - [x] 7.5 Write unit tests for Menu API
    - Test menu tree generation for different roles
    - Test menu hierarchy with nested structures
    - Test menu sorting
    - Test hidden menu handling
    - _Requirements: 3.2, 6.2, 6.6, 6.8_
  
  - [x] 7.6 Register Menu routes
    - Create router/system/menu.go
    - Register protected routes: /menu/* (require JWT)
    - _Requirements: 2.3_

- [ ] 8. Implement Casbin authorization system
  - [x] 8.1 Create Casbin model and adapter
    - Define model/system/sys_casbin_rule.go
    - Create Casbin model configuration (RBAC with RESTful path matching)
    - Initialize Casbin enforcer with Gorm adapter
    - _Requirements: 3.4, 3.5_
  
  - [x] 8.2 Implement Casbin manager utility
    - Create utils/casbin.go with CasbinManager
    - Implement Enforce, AddPolicy, RemovePolicy
    - Implement GetPoliciesForRole, UpdatePoliciesForRole
    - _Requirements: 3.4, 5.4_
  
  - [x] 8.3 Create Casbin authorization middleware
    - Implement middleware/casbin.go
    - Extract role from JWT claims
    - Check permission using Casbin enforcer
    - Return 403 for unauthorized requests
    - _Requirements: 3.4, 3.7, 16.2_
  
  - [x] 8.4 Write property tests for Casbin authorization
    - **Property 7: API Authorization Enforcement**
    - **Property 8: Role Permission Inheritance**
    - **Property 50: Casbin Middleware Authorization**
    - **Validates: Requirements 3.7, 3.8, 16.2**
  
  - [x] 8.5 Write unit tests for Casbin middleware
    - Test authorized API access
    - Test unauthorized API access returns 403
    - Test role inheritance
    - _Requirements: 3.7, 3.8_

- [ ] 9. Implement additional middleware components
  - [x] 9.1 Create CORS middleware
    - Implement middleware/cors.go
    - Configure allowed origins, methods, headers
    - _Requirements: 16.3_
  
  - [x] 9.2 Write property test for CORS middleware
    - **Property 51: CORS Header Configuration**
    - **Validates: Requirements 16.3**
  
  - [x] 9.3 Create rate limiting middleware
    - Implement middleware/rate_limit.go
    - Use token bucket or sliding window algorithm
    - Return 429 for rate limit exceeded
    - _Requirements: 16.4_
  
  - [x] 9.4 Write property test for rate limiting
    - **Property 52: Rate Limiting Enforcement**
    - **Validates: Requirements 16.4**
  
  - [x] 9.5 Create request logging middleware
    - Implement middleware/logger.go
    - Log timestamp, method, path, status, latency, client IP
    - _Requirements: 13.4, 16.5_
  
  - [x] 9.6 Write property test for request logging
    - **Property 41: HTTP Request Logging Completeness**
    - **Validates: Requirements 13.4**
  
  - [x] 9.7 Create panic recovery middleware
    - Implement middleware/recovery.go
    - Catch panics, log with stack trace, return 500
    - _Requirements: 11.7, 16.6_
  
  - [x] 9.8 Write property test for panic recovery
    - **Property 36: Panic Recovery Without Crash**
    - **Property 37: Error Logging with Stack Traces**
    - **Validates: Requirements 11.7, 11.8, 16.6**
  
  - [x] 9.9 Configure middleware chain order
    - Update main.go to register middleware in correct order
    - Order: Recovery → CORS → RateLimit → Logger → JWT → Casbin
    - Implement route exclusion mechanism
    - _Requirements: 16.7, 16.8_
  
  - [x] 9.10 Write property tests for middleware system
    - **Property 53: Middleware Execution Order**
    - **Property 54: Middleware Route Exclusion**
    - **Validates: Requirements 16.7, 16.8**

- [x] 10. Checkpoint - Verify core modules
  - Ensure all tests pass, ask the user if questions arise.


### Phase 3: Advanced Backend Features

- [x] 11. Implement Database Inspector module
  - [x] 11.1 Create DB Inspector service layer
    - Create service/tools/db_inspector_service.go
    - Implement GetTables (list all tables in database)
    - Implement GetTableSchema (column info with types, keys, comments)
    - Implement GetTableData with pagination
    - Implement ExecuteSQL with safety checks
    - Implement CreateRecord, UpdateRecord, DeleteRecord
    - Implement ValidateSQL (whitelist/blacklist dangerous commands)
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.7, 7.8, 7.9_
  
  - [x] 11.2 Write property tests for DB Inspector
    - **Property 18: Database Table Listing Completeness**
    - **Property 19: Table Schema Accuracy**
    - **Property 20: DB Inspector CRUD Operations**
    - **Property 21: Dangerous SQL Operation Restriction**
    - **Property 22: Read-Only Mode Enforcement**
    - **Property 23: SQL Error Message Propagation**
    - **Validates: Requirements 7.2, 7.3, 7.4, 7.5, 7.7, 7.8, 7.9**
  
  - [x] 11.3 Create DB Inspector API controllers
    - Create api/v1/tools/db_inspector.go
    - Implement handlers: GetTables, GetTableSchema, GetTableData
    - Implement ExecuteSQL, CreateRecord, UpdateRecord, DeleteRecord
    - Add super admin permission checks for dangerous operations
    - Add Swagger annotations
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.7, 7.8, 14.3_
  
  - [x] 11.4 Write unit tests for DB Inspector API
    - Test table listing
    - Test schema inspection
    - Test CRUD operations
    - Test SQL execution with dangerous commands
    - Test read-only mode
    - _Requirements: 7.2, 7.3, 7.5, 7.7, 7.8_
  
  - [x] 11.5 Register DB Inspector routes
    - Create router/tools/db_inspector.go
    - Register protected routes: /tools/db/* (require JWT + admin permission)
    - _Requirements: 2.3, 3.4_

- [ ] 12. Implement Code Generator module
  - [x] 12.1 Create code generation templates
    - Create resource/template/backend/model.tpl (Gorm model struct)
    - Create resource/template/backend/service.tpl (CRUD service methods)
    - Create resource/template/backend/api.tpl (Gin controller handlers)
    - Create resource/template/backend/router.tpl (route registration)
    - Create resource/template/frontend/api.tpl (Axios API definitions)
    - Create resource/template/frontend/types.tpl (TypeScript interfaces)
    - Create resource/template/frontend/page.tpl (table page component)
    - Create resource/template/frontend/modal.tpl (form modal component)
    - _Requirements: 8.6, 8.7_
  
  - [x] 12.2 Implement Code Generator service layer
    - Create service/tools/code_generator_service.go
    - Implement GetTableMetadata (extract columns, types, constraints)
    - Implement GenerateCode (apply templates, return file map)
    - Implement PreviewCode (generate without writing files)
    - Implement WriteGeneratedCode (write files to disk)
    - Implement CreateTable (create table from field definitions)
    - _Requirements: 8.2, 8.4, 8.5, 8.8, 8.9_
  
  - [x] 12.3 Write property tests for Code Generator
    - **Property 24: Code Generator Metadata Extraction**
    - **Property 25: Backend Code Generation Completeness**
    - **Property 26: Frontend Code Generation Completeness**
    - **Property 27: Code Preview Without Side Effects**
    - **Property 28: Automatic Table Creation**
    - **Validates: Requirements 8.2, 8.4, 8.5, 8.8, 8.9**
  
  - [x] 12.4 Create Code Generator API controllers
    - Create api/v1/tools/code_generator.go
    - Implement handlers: GetTableMetadata, GenerateCode, PreviewCode, CreateTable
    - Add Swagger annotations
    - _Requirements: 8.2, 8.4, 8.5, 8.8, 8.9, 14.3_
  
  - [x] 12.5 Write unit tests for Code Generator API
    - Test metadata extraction
    - Test code generation for sample table
    - Test preview mode
    - Test table creation
    - _Requirements: 8.2, 8.4, 8.5, 8.8, 8.9_
  
  - [x] 12.6 Register Code Generator routes
    - Create router/tools/code_generator.go
    - Register protected routes: /tools/gen/* (require JWT + admin permission)
    - _Requirements: 2.3, 3.4_

- [ ] 13. Generate Swagger documentation
  - [x] 13.1 Add Swagger annotations to all API handlers
    - Add @Summary, @Description, @Tags, @Accept, @Produce
    - Add @Param for path/query/body parameters
    - Add @Success and @Failure responses
    - Add @Security for protected endpoints
    - _Requirements: 14.2, 14.3, 14.5_
  
  - [x] 13.2 Configure Swagger and generate docs
    - Install swag CLI tool
    - Run swag init to generate swagger.json and swagger.yaml
    - Register Swagger UI route at /swagger/index.html
    - _Requirements: 14.1, 14.2_
  
  - [x] 13.3 Write unit test for Swagger endpoint
    - Test that /swagger/index.html returns 200
    - _Requirements: 14.2_

- [x] 14. Checkpoint - Verify advanced backend features
  - Ensure all tests pass, ask the user if questions arise.


### Phase 4: Frontend Foundation

- [-] 15. Set up frontend project structure and core utilities
  - [x] 15.1 Initialize frontend project structure
    - Verify Vite + React + TypeScript setup
    - Create directory structure: api, assets, components, hooks, layout, router, store, utils, views
    - Install dependencies: axios, zustand, react-router-dom, antd, dayjs, lodash
    - Configure TypeScript paths in tsconfig.json
    - _Requirements: Frontend architecture from PRD_
  
  - [x] 15.2 Implement request client with Axios
    - Create utils/request.ts with RequestClient class
    - Define UnifiedResponse interface
    - Implement request/response interceptors
    - Add Authorization header injection
    - Implement automatic token refresh on 401
    - Extract data field from successful responses
    - Display error notifications for failed responses
    - _Requirements: 11.2, 11.3, 11.4, 11.5_
  
  - [x] 15.3 Write property tests for request client
    - **Property 32: Authorization Header Injection**
    - **Property 33: Response Data Extraction**
    - **Property 34: Error Notification Display**
    - **Property 35: Automatic Token Refresh on 401**
    - **Validates: Requirements 11.2, 11.3, 11.4, 11.5**
  
  - [x] 15.4 Implement storage utilities
    - Create utils/storage.ts with localStorage helpers
    - Implement getItem, setItem, removeItem with JSON serialization
    - Add type-safe wrappers for common keys (token, user, theme)
    - _Requirements: 10.2_
  
  - [x] 15.5 Implement common utility functions
    - Create utils/format.ts for date/number formatting
    - Create utils/validator.ts for form validation rules
    - Create utils/helper.ts for common operations
    - _Requirements: General utilities_

- [-] 16. Implement state management with Zustand
  - [x] 16.1 Create User store
    - Create store/userStore.ts with UserState interface
    - Implement state: userInfo, accessToken, refreshToken, permissions, menuTree
    - Implement actions: login, logout, refreshAccessToken, fetchUserMenu, hasPermission, updateUserInfo
    - Add persistence for tokens and userInfo
    - _Requirements: 2.4, 3.2, 9.1_
  
  - [x] 16.2 Create App store
    - Create store/appStore.ts with AppState interface
    - Implement state: theme, collapsed, tabs, activeTab
    - Implement actions: toggleTheme, toggleSidebar, addTab, removeTab, setActiveTab, clearTabs
    - Add persistence for theme and tabs
    - _Requirements: 10.1, 10.2, 10.3_
  
  - [ ] 16.3 Write property test for theme persistence
    - **Property 29: Theme Preference Persistence**
    - **Validates: Requirements 10.2**
  
  - [ ] 16.4 Write unit tests for stores
    - Test user login flow
    - Test token refresh
    - Test permission checking
    - Test theme toggle
    - Test tab management
    - _Requirements: 2.4, 10.1, 10.3_

- [ ] 17. Implement API layer
  - [ ] 17.1 Create User API definitions
    - Create api/user.ts with TypeScript interfaces
    - Implement login, getUserInfo, getUserList, createUser, updateUser, deleteUser
    - Implement changePassword, resetPassword, toggleStatus
    - _Requirements: 4.2, 4.3, 4.4, 4.5, 4.6_
  
  - [ ] 17.2 Create Role API definitions
    - Create api/role.ts with TypeScript interfaces
    - Implement getRoleList, createRole, updateRole, deleteRole
    - Implement assignMenus, getRoleMenus, assignAPIs, getRoleAPIs
    - _Requirements: 5.2, 5.3, 5.4, 5.5_
  
  - [ ] 17.3 Create Menu API definitions
    - Create api/menu.ts with TypeScript interfaces
    - Implement getMenuTree, getAllMenus, createMenu, updateMenu, deleteMenu
    - _Requirements: 3.2, 6.2_
  
  - [ ] 17.4 Create DB Inspector API definitions
    - Create api/dbInspector.ts with TypeScript interfaces
    - Implement getTables, getTableSchema, getTableData, executeSQL
    - Implement createRecord, updateRecord, deleteRecord
    - _Requirements: 7.2, 7.3, 7.4, 7.5_
  
  - [ ] 17.5 Create Code Generator API definitions
    - Create api/codeGenerator.ts with TypeScript interfaces
    - Implement getTableMetadata, generateCode, previewCode, createTable
    - _Requirements: 8.2, 8.4, 8.5, 8.8, 8.9_

- [ ] 18. Checkpoint - Verify frontend foundation
  - Ensure all tests pass, ask the user if questions arise.


### Phase 5: Frontend Components and Routing

- [ ] 19. Implement reusable components
  - [ ] 19.1 Create AuthButton component
    - Create components/Auth/AuthButton.tsx
    - Accept perm prop and check against user permissions
    - Render button only if user has permission
    - Support fallback prop for unauthorized state
    - _Requirements: 3.6_
  
  - [ ] 19.2 Write property test for AuthButton
    - **Property 55: AuthButton Permission Visibility**
    - **Validates: Requirements 3.6**
  
  - [ ] 19.3 Write unit tests for AuthButton
    - Test button renders with permission
    - Test button hidden without permission
    - Test fallback rendering
    - _Requirements: 3.6_
  
  - [ ] 19.4 Create ProTable component
    - Create components/ProTable/index.tsx
    - Integrate search form, table, pagination
    - Implement automatic loading state management
    - Add toolbar with refresh, reset, export actions
    - Support actionRef for external control
    - _Requirements: 11.6_
  
  - [ ] 19.5 Write property test for ProTable loading state
    - **Property 58: Loading State Automation**
    - **Validates: Requirements 11.6**
  
  - [ ] 19.6 Write unit tests for ProTable
    - Test table rendering with data
    - Test pagination
    - Test search and filtering
    - Test loading states
    - _Requirements: 11.6_
  
  - [ ] 19.7 Create ThemeSwitch component
    - Create components/Theme/ThemeSwitch.tsx
    - Toggle between light and dark themes
    - Update Ant Design ConfigProvider token
    - Persist theme preference
    - _Requirements: 10.1, 10.2_
  
  - [ ] 19.8 Write unit test for ThemeSwitch
    - Test theme toggle
    - Test theme persistence
    - _Requirements: 10.1, 10.2_

- [ ] 20. Implement dynamic routing system
  - [ ] 20.1 Create route generator utility
    - Create router/generator.ts
    - Implement generateRoutes function (MenuItem[] → RouteObject[])
    - Implement loadComponent function (dynamic imports with React.lazy)
    - Handle nested routes recursively
    - _Requirements: 3.3, 9.2, 9.3_
  
  - [ ] 20.2 Write property tests for route generation
    - **Property 6: Route Generation from Menu Tree**
    - **Property 56: Dynamic Component Loading**
    - **Validates: Requirements 3.3, 6.7, 9.2, 9.3**
  
  - [ ] 20.3 Write unit tests for route generator
    - Test route generation from menu tree
    - Test nested route handling
    - Test component path mapping
    - _Requirements: 9.2, 9.3, 9.7_
  
  - [ ] 20.4 Create route guards
    - Create router/guards.ts
    - Implement authentication guard (check token)
    - Implement authorization guard (check permissions)
    - Redirect to login for unauthenticated users
    - Redirect to 403 for unauthorized routes
    - _Requirements: 9.6_
  
  - [ ] 20.5 Write unit test for route guards
    - Test redirect to login when not authenticated
    - Test redirect to 403 for unauthorized routes
    - _Requirements: 9.6_
  
  - [ ] 20.6 Configure router with static and dynamic routes
    - Create router/index.tsx
    - Define static routes: login, 404, 403
    - Fetch menu tree on app load
    - Generate and register dynamic routes
    - Apply route guards
    - _Requirements: 9.1, 9.2, 9.4_

- [ ] 21. Implement layout components
  - [ ] 21.1 Create Sidebar component
    - Create layout/Sidebar/index.tsx
    - Render menu tree as Ant Design Menu
    - Support collapsible sidebar
    - Highlight active menu item
    - _Requirements: 6.5, 9.5_
  
  - [ ] 21.2 Write property test for sidebar rendering
    - **Property 57: Sidebar Navigation Rendering**
    - **Validates: Requirements 6.5, 9.5**
  
  - [ ] 21.3 Write unit test for Sidebar
    - Test menu rendering from tree
    - Test menu item click navigation
    - Test sidebar collapse
    - _Requirements: 6.5, 9.5_
  
  - [ ] 21.4 Create Header component
    - Create layout/Header/index.tsx
    - Display user info and avatar
    - Add theme switch
    - Add logout button
    - _Requirements: 10.1_
  
  - [ ] 21.5 Create Tabs component with Keep-Alive
    - Create layout/Tabs/index.tsx
    - Display visited pages as tabs
    - Implement Keep-Alive using CSS display:none
    - Support tab close, close others, close all
    - Support tab refresh
    - _Requirements: 10.3, 10.4, 10.5, 10.6, 10.7_
  
  - [ ] 21.6 Write property tests for tabs system
    - **Property 30: Tab State Preservation**
    - **Property 31: Tab Management Operations**
    - **Validates: Requirements 10.3, 10.4, 10.5, 10.6, 10.7**
  
  - [ ] 21.7 Write unit tests for Tabs
    - Test tab creation on navigation
    - Test tab switching preserves state
    - Test tab close operations
    - Test tab refresh
    - _Requirements: 10.3, 10.4, 10.5, 10.6, 10.7_
  
  - [ ] 21.8 Create main Layout component
    - Create layout/index.tsx
    - Compose Sidebar, Header, Tabs, and content area
    - Handle responsive layout
    - _Requirements: Layout structure_

- [ ] 22. Implement Error Boundary
  - [ ] 22.1 Create Error Boundary component
    - Create components/ErrorBoundary/index.tsx
    - Catch React rendering errors
    - Display fallback UI with error message
    - Provide reload button
    - Log errors to console/error tracking
    - _Requirements: 11.9_
  
  - [ ] 22.2 Write property test for Error Boundary
    - **Property 59: React Error Boundary Catching**
    - **Validates: Requirements 11.9**
  
  - [ ] 22.3 Write unit test for Error Boundary
    - Test error catching
    - Test fallback UI rendering
    - _Requirements: 11.9_

- [ ] 23. Checkpoint - Verify frontend components and routing
  - Ensure all tests pass, ask the user if questions arise.


### Phase 6: Frontend Pages

- [ ] 24. Implement Login page
  - [ ] 24.1 Create Login page component
    - Create views/login/index.tsx
    - Design login form with username and password fields
    - Add form validation
    - Call login API and store tokens
    - Redirect to dashboard on success
    - _Requirements: 2.1, 2.4_
  
  - [ ] 24.2 Write unit tests for Login page
    - Test form validation
    - Test successful login flow
    - Test failed login error display
    - _Requirements: 2.1_

- [ ] 25. Implement Dashboard page
  - [ ] 25.1 Create Dashboard page component
    - Create views/dashboard/index.tsx
    - Display welcome message and user info
    - Add system statistics cards
    - Add recent activity list
    - _Requirements: Dashboard feature_
  
  - [ ] 25.2 Write unit test for Dashboard
    - Test component rendering
    - Test data fetching
    - _Requirements: Dashboard feature_

- [ ] 26. Implement User Management page
  - [ ] 26.1 Create User list page
    - Create views/system/user/index.tsx
    - Use ProTable component for user list
    - Add search form (username, phone, email, role, status)
    - Add action buttons: create, edit, delete, reset password, toggle status
    - Implement AuthButton for permission control
    - _Requirements: 4.2, 4.3, 4.4, 4.5, 4.6_
  
  - [ ] 26.2 Create User form modal
    - Create views/system/user/components/UserModal.tsx
    - Design form with fields: username, password, nickname, phone, email, role, avatar
    - Add form validation
    - Support create and edit modes
    - _Requirements: 4.2, 4.3_
  
  - [ ] 26.3 Write unit tests for User Management
    - Test user list rendering
    - Test search and filtering
    - Test user creation
    - Test user editing
    - Test user deletion
    - Test permission-based button visibility
    - _Requirements: 4.2, 4.3, 4.4, 4.5, 4.6_

- [ ] 27. Implement Role Management page
  - [ ] 27.1 Create Role list page
    - Create views/system/role/index.tsx
    - Use ProTable component for role list
    - Add action buttons: create, edit, delete, assign permissions
    - _Requirements: 5.2, 5.3, 5.4, 5.5, 5.6_
  
  - [ ] 27.2 Create Role form modal
    - Create views/system/role/components/RoleModal.tsx
    - Design form with fields: role_name, role_key, data_scope, sort, status, remark
    - Add form validation
    - _Requirements: 5.2, 5.5_
  
  - [ ] 27.3 Create Permission assignment modal
    - Create views/system/role/components/PermissionModal.tsx
    - Display menu tree with checkboxes
    - Display API permission list
    - Support assigning menus and APIs to role
    - _Requirements: 5.3, 5.4_
  
  - [ ] 27.4 Write unit tests for Role Management
    - Test role list rendering
    - Test role creation
    - Test role editing
    - Test role deletion with users
    - Test permission assignment
    - _Requirements: 5.2, 5.3, 5.4, 5.5, 5.6_

- [ ] 28. Implement Menu Management page
  - [ ] 28.1 Create Menu list page
    - Create views/system/menu/index.tsx
    - Display menu tree in table format
    - Add action buttons: create, edit, delete
    - Support expanding/collapsing tree nodes
    - _Requirements: 6.2, 6.5, 6.6_
  
  - [ ] 28.2 Create Menu form modal
    - Create views/system/menu/components/MenuModal.tsx
    - Design form with fields: parent_id, path, name, component, sort, meta (icon, title, hidden, keep_alive), btn_perms
    - Add form validation
    - Support parent menu selection
    - _Requirements: 6.2, 6.3, 6.4_
  
  - [ ] 28.3 Write unit tests for Menu Management
    - Test menu tree rendering
    - Test menu creation
    - Test menu editing
    - Test menu deletion
    - Test menu sorting
    - _Requirements: 6.2, 6.5, 6.6_

- [ ] 29. Checkpoint - Verify system management pages
  - Ensure all tests pass, ask the user if questions arise.


### Phase 7: Developer Tools Pages

- [ ] 30. Implement Database Inspector page
  - [ ] 30.1 Create DB Inspector layout
    - Create views/tools/db-inspector/index.tsx
    - Design two-column layout: table list (left) and content area (right)
    - Add database connection selector
    - _Requirements: 7.1_
  
  - [ ] 30.2 Create table list component
    - Create views/tools/db-inspector/components/TableList.tsx
    - Display all tables with search filter
    - Highlight selected table
    - _Requirements: 7.2_
  
  - [ ] 30.3 Create table schema viewer
    - Create views/tools/db-inspector/components/SchemaViewer.tsx
    - Display table columns in table format
    - Show column name, type, nullable, key, default, comment
    - _Requirements: 7.3_
  
  - [ ] 30.4 Create table data browser
    - Create views/tools/db-inspector/components/DataBrowser.tsx
    - Display table data with pagination
    - Support inline editing
    - Add action buttons: create, edit, delete
    - _Requirements: 7.4, 7.5_
  
  - [ ] 30.5 Create SQL console component
    - Create views/tools/db-inspector/components/SQLConsole.tsx
    - Integrate Monaco Editor for SQL input
    - Add execute button with read-only mode toggle
    - Display query results in table format
    - Display error messages
    - Add confirmation dialog for dangerous operations
    - _Requirements: 7.6, 7.7, 7.8, 7.9_
  
  - [ ] 30.6 Write unit tests for DB Inspector
    - Test table list rendering
    - Test schema viewer
    - Test data browser CRUD operations
    - Test SQL console execution
    - Test dangerous operation blocking
    - _Requirements: 7.2, 7.3, 7.4, 7.5, 7.7, 7.8_

- [ ] 31. Implement Code Generator page
  - [ ] 31.1 Create Code Generator layout
    - Create views/tools/code-generator/index.tsx
    - Design wizard-style interface with steps
    - Steps: Select Table → Configure → Preview → Generate
    - _Requirements: 8.1, 8.3_
  
  - [ ] 31.2 Create table selection step
    - Create views/tools/code-generator/components/TableSelect.tsx
    - Display table list with radio selection
    - Show table metadata preview
    - _Requirements: 8.1, 8.2_
  
  - [ ] 31.3 Create configuration step
    - Create views/tools/code-generator/components/ConfigForm.tsx
    - Form fields: struct_name, package_name, frontend_path
    - Display field mapping table (editable)
    - Add generation options checkboxes
    - _Requirements: 8.3_
  
  - [ ] 31.4 Create code preview step
    - Create views/tools/code-generator/components/CodePreview.tsx
    - Display generated code in tabs (model, service, API, router, frontend)
    - Use Monaco Editor for syntax highlighting
    - _Requirements: 8.8_
  
  - [ ] 31.5 Create generation confirmation step
    - Create views/tools/code-generator/components/GenerateConfirm.tsx
    - Show summary of files to be generated
    - Add generate button
    - Display success/error messages
    - _Requirements: 8.4, 8.5_
  
  - [ ] 31.6 Add table creation feature
    - Create views/tools/code-generator/components/TableCreate.tsx
    - Form to define table name and fields
    - Field configuration: name, type, nullable, default, comment
    - Call create table API
    - _Requirements: 8.9_
  
  - [ ] 31.7 Write unit tests for Code Generator
    - Test table selection
    - Test configuration form
    - Test code preview
    - Test code generation
    - Test table creation
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.8, 8.9_

- [ ] 32. Checkpoint - Verify developer tools pages
  - Ensure all tests pass, ask the user if questions arise.


### Phase 8: Integration and Deployment

- [ ] 33. Implement error pages
  - [ ] 33.1 Create 404 Not Found page
    - Create views/error/404.tsx
    - Display friendly error message
    - Add button to return to home
    - _Requirements: 9.6_
  
  - [ ] 33.2 Create 403 Forbidden page
    - Create views/error/403.tsx
    - Display permission denied message
    - Add button to return to home
    - _Requirements: 9.6_
  
  - [ ] 33.3 Write unit tests for error pages
    - Test 404 page rendering
    - Test 403 page rendering
    - _Requirements: 9.6_

- [ ] 34. Configure environment and build
  - [ ] 34.1 Create environment configuration files
    - Create .env.development for development
    - Create .env.production for production
    - Define API base URL, app title, etc.
    - _Requirements: 12.2_
  
  - [ ] 34.2 Configure Vite build settings
    - Update vite.config.ts
    - Configure build output directory
    - Configure proxy for development API calls
    - Optimize build for production
    - _Requirements: Frontend build_
  
  - [ ] 34.3 Create backend configuration files
    - Create config/config.yaml for development
    - Create config/config.prod.yaml for production
    - Define server, database, JWT, Redis, logger settings
    - _Requirements: 12.2, 12.3_

- [ ] 35. Create Docker deployment configuration
  - [ ] 35.1 Create backend Dockerfile
    - Create backend/Dockerfile
    - Use multi-stage build (build stage + runtime stage)
    - Copy binary and config files
    - Expose port 8080
    - _Requirements: 17.1_
  
  - [ ] 35.2 Create frontend Dockerfile
    - Create frontend/Dockerfile
    - Use multi-stage build (build stage + nginx stage)
    - Build React app and copy to nginx
    - Configure nginx to serve static files
    - Expose port 80
    - _Requirements: 17.2_
  
  - [ ] 35.3 Create docker-compose configuration
    - Create docker-compose.yml at project root
    - Define services: backend, frontend, mysql, redis
    - Configure environment variables
    - Set up volume mounts for persistent data
    - Configure service dependencies and networking
    - _Requirements: 17.3, 17.4, 17.7_
  
  - [ ] 35.4 Write unit tests for deployment configuration
    - Test that backend exposes port 8080
    - Test that frontend exposes port 80
    - Test health check endpoints
    - _Requirements: 17.5, 17.6_

- [ ] 36. Create health check endpoints
  - [ ] 36.1 Implement backend health check
    - Create api/v1/system/health.go
    - Check database connectivity
    - Check Redis connectivity
    - Return health status JSON
    - _Requirements: 17.6_
  
  - [ ] 36.2 Write unit test for health check
    - Test health endpoint returns 200
    - Test health status includes database and Redis status
    - _Requirements: 17.6_

- [ ] 37. Write project documentation
  - [ ] 37.1 Create README.md
    - Project overview and features
    - Technology stack
    - Installation instructions
    - Development setup
    - Deployment instructions
    - API documentation link
    - _Requirements: Documentation_
  
  - [ ] 37.2 Create API documentation
    - Ensure Swagger documentation is complete
    - Add API usage examples
    - Document authentication flow
    - _Requirements: 14.2, 14.3_
  
  - [ ] 37.3 Create development guide
    - Code structure explanation
    - Adding new modules guide
    - Testing guidelines
    - Code generation workflow
    - _Requirements: Documentation_

- [ ] 38. Final integration testing
  - [ ] 38.1 Run full test suite
    - Execute all unit tests (backend + frontend)
    - Execute all property tests (backend + frontend)
    - Verify test coverage meets goals (80% backend, 75% frontend)
    - _Requirements: All testing requirements_
  
  - [ ] 38.2 Perform end-to-end integration tests
    - Test complete user flows: login → navigate → CRUD operations
    - Test permission system: menu filtering, API authorization, button visibility
    - Test token refresh flow
    - Test DB Inspector operations
    - Test Code Generator workflow
    - _Requirements: Integration testing_
  
  - [ ] 38.3 Test Docker deployment
    - Build Docker images
    - Run docker-compose up
    - Verify all services start correctly
    - Test application functionality in containerized environment
    - _Requirements: 17.1, 17.2, 17.3_

- [ ] 39. Final checkpoint - System ready for deployment
  - Ensure all tests pass, all features working, documentation complete.

## Notes

- Tasks marked with `*` are optional testing tasks and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at major milestones
- Property tests validate universal correctness properties with minimum 100 iterations
- Unit tests validate specific examples, edge cases, and error conditions
- The implementation follows a bottom-up approach: infrastructure → core modules → advanced features → frontend → integration
- Backend and frontend can be developed in parallel after Phase 1 is complete
- Code Generator templates should follow project conventions established in earlier phases
