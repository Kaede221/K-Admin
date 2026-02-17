# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Shanghai

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/config.prod.yaml ./config.yaml
COPY --from=builder /app/config/casbin_model.conf ./config/

# Create logs directory
RUN mkdir -p logs

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
