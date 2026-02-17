# Configuration Management

This package provides configuration management for the K-Admin system using Viper.

## Features

- **Multiple Format Support**: Load configuration from YAML or JSON files
- **Environment Variable Override**: Environment variables take precedence over file configuration
- **Validation**: Automatic validation of required fields with sensible defaults
- **Type Safety**: Strongly-typed configuration structs

## Configuration Structure

The configuration is organized into five main sections:

### Server Configuration
```yaml
server:
  port: ":8080"           # Server port (required)
  mode: "debug"           # Gin mode: debug, release, or test (default: debug)
```

### Database Configuration
```yaml
database:
  host: "localhost"       # Database host (required)
  port: 3306              # Database port (required)
  name: "k_admin"         # Database name (required)
  username: "root"        # Database username (required)
  password: "password"    # Database password (optional for local dev)
  max_idle_conns: 10      # Maximum idle connections (default: 10)
  max_open_conns: 100     # Maximum open connections (default: 100)
```

### JWT Configuration
```yaml
jwt:
  secret: "your-secret-key"  # JWT signing secret (required)
  access_expiration: 15      # Access token expiration in minutes (default: 15)
  refresh_expiration: 7      # Refresh token expiration in days (default: 7)
```

### Redis Configuration
```yaml
redis:
  host: "localhost"       # Redis host (required)
  port: 6379              # Redis port (required)
  password: ""            # Redis password (optional)
  db: 0                   # Redis database number (default: 0)
```

### Logger Configuration
```yaml
logger:
  level: "info"           # Log level: debug, info, warn, error, fatal (default: info)
  path: "./logs/app.log"  # Log file path (default: ./logs/app.log)
  max_size: 100           # Max size in MB before rotation (default: 100)
  max_age: 7              # Max age in days to retain logs (default: 7)
  max_backups: 3          # Max number of old log files (default: 3)
  compress: true          # Compress rotated logs (default: false)
```

## Usage

### Using Configuration Files

#### YAML Format (config.yaml)
```yaml
server:
  port: ":8080"
  mode: "debug"

database:
  host: "localhost"
  port: 3306
  name: "k_admin"
  username: "root"
  password: "password"

jwt:
  secret: "your-secret-key"
  access_expiration: 15
  refresh_expiration: 7

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

logger:
  level: "info"
  path: "./logs/app.log"
  max_size: 100
  max_age: 7
  max_backups: 3
  compress: true
```

#### JSON Format (config.json)
```json
{
  "server": {
    "port": ":8080",
    "mode": "debug"
  },
  "database": {
    "host": "localhost",
    "port": 3306,
    "name": "k_admin",
    "username": "root",
    "password": "password"
  },
  "jwt": {
    "secret": "your-secret-key",
    "access_expiration": 15,
    "refresh_expiration": 7
  },
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0
  },
  "logger": {
    "level": "info",
    "path": "./logs/app.log",
    "max_size": 100,
    "max_age": 7,
    "max_backups": 3,
    "compress": true
  }
}
```

### Using Environment Variables

Environment variables use the prefix `KADMIN_` and nested keys are separated by underscores.

**Examples:**
```bash
# Server configuration
export KADMIN_SERVER_PORT=":8080"
export KADMIN_SERVER_MODE="release"

# Database configuration
export KADMIN_DATABASE_HOST="localhost"
export KADMIN_DATABASE_PORT=3306
export KADMIN_DATABASE_NAME="k_admin"
export KADMIN_DATABASE_USERNAME="root"
export KADMIN_DATABASE_PASSWORD="password"

# JWT configuration
export KADMIN_JWT_SECRET="your-secret-key"
export KADMIN_JWT_ACCESS_EXPIRATION=15
export KADMIN_JWT_REFRESH_EXPIRATION=7

# Redis configuration
export KADMIN_REDIS_HOST="localhost"
export KADMIN_REDIS_PORT=6379
export KADMIN_REDIS_PASSWORD=""
export KADMIN_REDIS_DB=0

# Logger configuration
export KADMIN_LOGGER_LEVEL="info"
export KADMIN_LOGGER_PATH="./logs/app.log"
```

### Loading Configuration in Code

```go
package main

import (
    "log"
    "k-admin-system/config"
    "k-admin-system/global"
)

func main() {
    // Load configuration from default locations
    cfg, err := config.LoadConfig("")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    global.Config = cfg

    // Or specify a custom config file path
    cfg, err = config.LoadConfig("./custom-config.yaml")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
}
```

### Command Line Usage

```bash
# Use default config file (searches for config.yaml or config.json in current directory)
go run main.go

# Specify a custom config file
go run main.go -config=/path/to/config.yaml

# Use environment variables only (no config file)
export KADMIN_SERVER_PORT=":8080"
export KADMIN_DATABASE_HOST="localhost"
# ... set other required variables
go run main.go
```

## Configuration Priority

Configuration values are loaded in the following order (later sources override earlier ones):

1. **Default values** (set in validation)
2. **Configuration file** (YAML or JSON)
3. **Environment variables** (highest priority)

This allows you to:
- Use a base configuration file for common settings
- Override specific values with environment variables for different environments
- Keep sensitive data (passwords, secrets) in environment variables

## Validation

The configuration is automatically validated on load. Required fields include:

- `server.port`
- `database.host`, `database.port`, `database.name`, `database.username`
- `jwt.secret`
- `redis.host`, `redis.port`

If any required field is missing, the application will fail to start with a detailed error message.

## Default Values

The following default values are applied if not specified:

- `server.mode`: "debug"
- `database.max_idle_conns`: 10
- `database.max_open_conns`: 100
- `jwt.access_expiration`: 15 minutes
- `jwt.refresh_expiration`: 7 days
- `logger.level`: "info"
- `logger.path`: "./logs/app.log"
- `logger.max_size`: 100 MB
- `logger.max_age`: 7 days
- `logger.max_backups`: 3

## Best Practices

1. **Development**: Use `config.yaml` with local settings
2. **Production**: Use environment variables for sensitive data (passwords, secrets)
3. **Docker**: Pass configuration via environment variables in docker-compose or Kubernetes
4. **Security**: Never commit sensitive data to version control
5. **Validation**: Always validate configuration on startup to catch errors early

## Example: Docker Deployment

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
CMD ["./main"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      - KADMIN_SERVER_PORT=:8080
      - KADMIN_SERVER_MODE=release
      - KADMIN_DATABASE_HOST=mysql
      - KADMIN_DATABASE_PORT=3306
      - KADMIN_DATABASE_NAME=k_admin
      - KADMIN_DATABASE_USERNAME=root
      - KADMIN_DATABASE_PASSWORD=${DB_PASSWORD}
      - KADMIN_JWT_SECRET=${JWT_SECRET}
      - KADMIN_REDIS_HOST=redis
      - KADMIN_REDIS_PORT=6379
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_PASSWORD}
      MYSQL_DATABASE: k_admin

  redis:
    image: redis:7-alpine
```

## Testing

Run the configuration tests:

```bash
cd backend
go test -v ./config/...
```

The test suite covers:
- Loading from YAML files
- Loading from JSON files
- Environment variable overrides
- Validation of required fields
- Default value assignment
