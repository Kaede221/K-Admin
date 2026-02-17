package system

import (
	"net/http"
	"time"

	"k-admin-system/global"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check the health status of the application and its dependencies
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	services := make(map[string]string)
	allHealthy := true

	// Check database connectivity
	sqlDB, err := global.DB.DB()
	if err != nil {
		services["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		if err := sqlDB.Ping(); err != nil {
			services["database"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			services["database"] = "healthy"
		}
	}

	// Check Redis connectivity
	if global.RedisClient != nil {
		if err := global.RedisClient.Ping(c).Err(); err != nil {
			services["redis"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			services["redis"] = "healthy"
		}
	} else {
		services["redis"] = "not configured"
	}

	status := "healthy"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
	}

	c.JSON(statusCode, response)
}
