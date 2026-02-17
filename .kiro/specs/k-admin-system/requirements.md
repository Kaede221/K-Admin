# Requirements Document: K-Admin System

## Introduction

K-Admin is a modern, high-performance, full-stack admin management system built with Go (Gin + Gorm) backend and React 18 (Vite + Ant Design 5) frontend. The system provides a comprehensive RBAC permission framework, visual database management tools, and intelligent code generation capabilities to accelerate enterprise application development.

## Glossary

- **System**: The K-Admin application (frontend + backend)
- **Backend**: The Go-based server application using Gin and Gorm
- **Frontend**: The React-based web application using Vite and Ant Design
- **User**: An authenticated person using the system
- **Administrator**: A user with elevated privileges to manage system configuration
- **Access_Token**: Short-lived JWT token (15 minutes) for API authentication
- **Refresh_Token**: Long-lived JWT token (7 days) for obtaining new Access_Tokens
- **RBAC**: Role-Based Access Control system
- **Casbin**: Authorization library for API-level permission enforcement
- **Menu_Tree**: Hierarchical structure of navigation menus and routes
- **Code_Generator**: Tool that generates CRUD code from database table schemas
- **DB_Inspector**: Visual database management interface
- **Unified_Response**: Standardized JSON response format for all API endpoints

## Requirements

### Requirement 1: Unified Response Model

**User Story:** As a frontend developer, I want all API responses to follow a consistent structure, so that I can handle responses uniformly in interceptors.

#### Acceptance Criteria

1. THE Backend SHALL return responses in JSON format with fields: code (integer), data (object), and msg (string)
2. WHEN an operation succeeds, THE Backend SHALL set code to 0 and include result data in the data field
3. WHEN an operation fails, THE Backend SHALL set code to a non-zero error code and include an error message in the msg field
4. THE Backend SHALL provide helper functions (Ok, Fail, OkWithDetailed) for generating Unified_Response objects
5. THE Frontend SHALL parse all API responses using the Unified_Response structure

### Requirement 2: JWT Authentication System

**User Story:** As a user, I want to authenticate securely with automatic token refresh, so that I can access the system without frequent re-login.

#### Acceptance Criteria

1. WHEN a user submits valid credentials, THE Backend SHALL generate an Access_Token with 15-minute expiration and a Refresh_Token with 7-day expiration
2. THE Backend SHALL encrypt passwords using Bcrypt before storage
3. THE Backend SHALL validate JWT tokens on protected API endpoints using middleware
4. WHEN an Access_Token expires, THE Frontend SHALL automatically use the Refresh_Token to obtain a new Access_Token without user intervention
5. WHEN a Refresh_Token is used, THE Backend SHALL validate it and issue a new Access_Token
6. WHEN a token refresh fails, THE Frontend SHALL redirect the user to the login page
7. THE Backend SHALL support token blacklisting using Redis for forced logout functionality

### Requirement 3: Multi-Level RBAC Permission System

**User Story:** As an administrator, I want to control access at menu, API, and button levels, so that I can enforce fine-grained security policies.

#### Acceptance Criteria

1. THE System SHALL implement three permission levels: menu-level, API-level, and button-level
2. WHEN a user logs in, THE Backend SHALL return a Menu_Tree containing only menus the user is authorized to access
3. THE Frontend SHALL dynamically generate routes based on the received Menu_Tree
4. THE Backend SHALL use Casbin middleware to enforce API-level permissions based on role, request path, and HTTP method
5. THE Backend SHALL store Casbin policies in the sys_casbin_rules table with format: (role, path, method)
6. THE Frontend SHALL provide an AuthButton component that shows/hides buttons based on user permissions
7. WHEN a user attempts to access an unauthorized API endpoint, THE Backend SHALL return a 403 Forbidden response
8. THE System SHALL support hierarchical role inheritance where child roles inherit parent role permissions

### Requirement 4: User Management Module

**User Story:** As an administrator, I want to manage user accounts, so that I can control who has access to the system.

#### Acceptance Criteria

1. THE Backend SHALL store user data in the sys_users table with fields: id, username, password, header_img, role_id, active
2. THE System SHALL support creating users with username, password, role assignment, and avatar
3. THE System SHALL support updating user information including password changes
4. THE System SHALL support enabling and disabling user accounts via the active field
5. THE System SHALL support deleting users (soft delete preferred)
6. THE Frontend SHALL display a user list with search, pagination, and filtering capabilities
7. WHEN displaying users, THE System SHALL mask password fields
8. THE System SHALL validate username uniqueness before creating new users

### Requirement 5: Role Management Module

**User Story:** As an administrator, I want to define roles with specific permissions, so that I can group users by their access needs.

#### Acceptance Criteria

1. THE Backend SHALL store role data in the sys_roles table with fields: id, role_name, role_key, data_scope
2. THE System SHALL support creating roles with name, key identifier, and data scope configuration
3. THE System SHALL support assigning menu permissions to roles
4. THE System SHALL support assigning API permissions to roles via Casbin policies
5. THE System SHALL support updating role information and permissions
6. THE System SHALL support deleting roles that have no associated users
7. THE Frontend SHALL provide a role management interface with permission tree selection
8. THE System SHALL support data scope permissions (all data, department data, personal data)

### Requirement 6: Menu Management Module

**User Story:** As an administrator, I want to configure the system menu structure, so that I can customize navigation and routing.

#### Acceptance Criteria

1. THE Backend SHALL store menu data in the sys_menus table with fields: id, parent_id, path, name, component, sort, meta, btn_perms
2. THE System SHALL support hierarchical menu structures with unlimited nesting levels
3. THE System SHALL store menu metadata (icon, title, hidden, keepAlive) in JSON format
4. THE System SHALL support defining button-level permissions within menu records via btn_perms field
5. THE Frontend SHALL render menus as a tree structure in the sidebar navigation
6. THE System SHALL support sorting menus via the sort field
7. WHEN a menu has component path defined, THE Frontend SHALL dynamically load the corresponding React component
8. THE System SHALL support hiding menus from navigation while keeping routes accessible via the hidden meta property

### Requirement 7: Visual Database Management Tool

**User Story:** As a developer, I want to manage database tables visually, so that I can inspect and modify data without external tools.

#### Acceptance Criteria

1. THE DB_Inspector SHALL support connecting to multiple database sources
2. THE DB_Inspector SHALL display a list of all tables in the connected database
3. WHEN a table is selected, THE DB_Inspector SHALL display the table schema including column names, types, comments, and keys
4. THE DB_Inspector SHALL support browsing table data with pagination
5. THE DB_Inspector SHALL support basic CRUD operations on table records
6. THE DB_Inspector SHALL provide a SQL console using Monaco Editor for executing custom queries
7. THE DB_Inspector SHALL restrict dangerous SQL operations (DROP, TRUNCATE) to super administrators with confirmation
8. THE DB_Inspector SHALL support read-only mode to prevent accidental data modifications
9. WHEN SQL execution fails, THE DB_Inspector SHALL display detailed error messages

### Requirement 8: Intelligent Code Generator

**User Story:** As a developer, I want to generate CRUD code from database tables, so that I can accelerate development of standard features.

#### Acceptance Criteria

1. THE Code_Generator SHALL allow selecting a database table as input
2. THE Code_Generator SHALL extract table metadata including columns, types, and constraints
3. THE Code_Generator SHALL provide configuration options for struct name, package name, and frontend path
4. THE Code_Generator SHALL generate backend code including: model struct, service layer, API controller, and router registration
5. THE Code_Generator SHALL generate frontend code including: TypeScript types, API definitions, table page component, and form modal component
6. THE Code_Generator SHALL use Go text/template for backend code generation
7. THE Code_Generator SHALL use template files for frontend code generation
8. THE Code_Generator SHALL support previewing generated code before writing to files
9. THE Code_Generator SHALL support automatic table creation from field definitions
10. THE Code_Generator SHALL generate code that follows project conventions and structure

### Requirement 9: Frontend Dynamic Routing

**User Story:** As a user, I want to see only the pages I have permission to access, so that the interface is tailored to my role.

#### Acceptance Criteria

1. WHEN a user logs in, THE Frontend SHALL fetch the user's Menu_Tree from the backend
2. THE Frontend SHALL recursively transform the Menu_Tree into React Router route configurations
3. THE Frontend SHALL dynamically import components based on the component path in menu records
4. THE Frontend SHALL register generated routes using React Router v6 useRoutes or createBrowserRouter
5. THE Frontend SHALL render sidebar navigation based on the Menu_Tree structure
6. WHEN a user navigates to an unauthorized route, THE Frontend SHALL redirect to a 403 error page
7. THE Frontend SHALL support nested routes matching the menu hierarchy

### Requirement 10: Theme Switching and Keep-Alive Tabs

**User Story:** As a user, I want to customize the interface theme and keep my work context when switching pages, so that I have a comfortable and efficient experience.

#### Acceptance Criteria

1. THE Frontend SHALL support light and dark theme modes using Ant Design 5 token system
2. THE Frontend SHALL persist theme preference in localStorage
3. THE Frontend SHALL implement a tab system that displays visited pages
4. THE Frontend SHALL support Keep-Alive functionality to preserve component state when switching tabs
5. WHEN a user switches between tabs, THE Frontend SHALL restore the previous page state including form inputs and scroll position
6. THE Frontend SHALL support closing individual tabs or all tabs except the current one
7. THE Frontend SHALL support refreshing individual tabs to reload component state
8. THE Frontend SHALL use React Router route names for Keep-Alive identification

### Requirement 11: Request Interceptor and Error Handling

**User Story:** As a developer, I want centralized request handling and error management, so that I can maintain consistent behavior across the application.

#### Acceptance Criteria

1. THE Frontend SHALL use Axios with request and response interceptors
2. THE Frontend SHALL automatically attach Access_Token to all API requests via Authorization header
3. WHEN a response has code 0, THE Frontend SHALL extract and return the data field
4. WHEN a response has non-zero code, THE Frontend SHALL display the msg field as an error notification
5. WHEN a response returns 401 status, THE Frontend SHALL attempt token refresh before retrying the original request
6. THE Frontend SHALL implement automatic loading state management for requests
7. THE Backend SHALL implement panic recovery middleware to prevent server crashes
8. THE Backend SHALL log all errors with stack traces using Zap logger
9. THE Frontend SHALL implement Error Boundary components to catch React rendering errors

### Requirement 12: Configuration Management

**User Story:** As a system administrator, I want to configure the application via files and environment variables, so that I can deploy to different environments easily.

#### Acceptance Criteria

1. THE Backend SHALL use Viper for configuration management
2. THE Backend SHALL support loading configuration from YAML, JSON, and environment variables
3. THE Backend SHALL define configuration structure for: server (port, mode), database (host, port, name, credentials), JWT (secret, expiration), Redis (host, port, password), and logging (level, path)
4. THE Backend SHALL validate required configuration fields on startup
5. WHEN configuration is invalid or missing, THE Backend SHALL log detailed error messages and exit gracefully
6. THE Backend SHALL support hot-reloading of non-critical configuration changes

### Requirement 13: Logging System

**User Story:** As a developer, I want comprehensive logging with rotation, so that I can troubleshoot issues and monitor system behavior.

#### Acceptance Criteria

1. THE Backend SHALL use Zap for structured logging
2. THE Backend SHALL use Lumberjack for log file rotation and archival
3. THE Backend SHALL support configurable log levels: debug, info, warn, error, fatal
4. THE Backend SHALL log all HTTP requests with: timestamp, method, path, status code, latency, and client IP
5. THE Backend SHALL log all database queries in debug mode
6. THE Backend SHALL rotate log files based on size (default 100MB) and age (default 7 days)
7. THE Backend SHALL output logs to both console and file in development mode
8. THE Backend SHALL output logs to file only in production mode

### Requirement 14: API Documentation

**User Story:** As a developer, I want automatically generated API documentation, so that I can understand and test endpoints easily.

#### Acceptance Criteria

1. THE Backend SHALL use Swag to generate Swagger documentation from code annotations
2. THE Backend SHALL expose Swagger UI at /swagger/index.html endpoint
3. THE Backend SHALL document all API endpoints with: description, parameters, request body schema, response schema, and status codes
4. THE Backend SHALL group API documentation by modules (system, tools, business)
5. THE Backend SHALL include authentication requirements in API documentation
6. WHEN code annotations change, THE Backend SHALL regenerate Swagger documentation via swag init command

### Requirement 15: Database Connection and ORM

**User Story:** As a developer, I want reliable database connectivity with type-safe queries, so that I can build robust data access layers.

#### Acceptance Criteria

1. THE Backend SHALL use Gorm as the ORM framework
2. THE Backend SHALL support MySQL 8.0+ as the primary database
3. THE Backend SHALL implement database connection pooling with configurable max connections
4. THE Backend SHALL implement automatic reconnection on connection loss
5. THE Backend SHALL support database migrations via Gorm AutoMigrate
6. THE Backend SHALL define common model fields (ID, CreatedAt, UpdatedAt, DeletedAt) in a base struct
7. THE Backend SHALL support soft deletes using Gorm's DeletedAt field
8. THE Backend SHALL log slow queries (threshold configurable, default 200ms)

### Requirement 16: Middleware System

**User Story:** As a developer, I want reusable middleware components, so that I can apply cross-cutting concerns consistently.

#### Acceptance Criteria

1. THE Backend SHALL implement JWT authentication middleware that validates tokens and extracts user information
2. THE Backend SHALL implement Casbin authorization middleware that enforces API-level permissions
3. THE Backend SHALL implement CORS middleware with configurable allowed origins
4. THE Backend SHALL implement rate limiting middleware to prevent abuse
5. THE Backend SHALL implement request logging middleware that logs all HTTP requests
6. THE Backend SHALL implement panic recovery middleware that catches panics and returns 500 errors
7. THE Backend SHALL support middleware chaining in a defined order
8. THE Backend SHALL allow excluding specific routes from middleware application

### Requirement 17: Deployment and Containerization

**User Story:** As a DevOps engineer, I want containerized deployment configurations, so that I can deploy the system consistently across environments.

#### Acceptance Criteria

1. THE System SHALL provide a Dockerfile for the Backend with multi-stage builds
2. THE System SHALL provide a Dockerfile for the Frontend with Nginx serving static files
3. THE System SHALL provide a docker-compose.yml that orchestrates Backend, Frontend, MySQL, and Redis services
4. THE System SHALL support environment-specific configuration via Docker environment variables
5. THE System SHALL expose Backend on port 8080 and Frontend on port 80 by default
6. THE System SHALL include health check endpoints for container orchestration
7. THE System SHALL support volume mounting for persistent data (database, logs, uploads)
