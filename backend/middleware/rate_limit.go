package middleware

import (
	"context"
	"fmt"
	"k-admin-system/config"
	"k-admin-system/global"
	"k-admin-system/model/common"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit 限流中间件
// 使用滑动窗口算法限制请求频率，防止API滥用
//
// 使用示例:
//
//	router.Use(middleware.RateLimit(global.Config.RateLimit))
//
// 配置示例 (config.yaml):
//
//	rate_limit:
//	  enabled: true
//	  requests: 100      # 允许的请求数
//	  window: 60         # 时间窗口（秒）
//	  key_func: "ip"     # 限流键函数: "ip" 或 "user"
func RateLimit(rateLimitConfig config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未启用限流，直接放行
		if !rateLimitConfig.Enabled {
			c.Next()
			return
		}

		// 如果Redis未初始化，记录警告并放行
		if global.RedisClient == nil {
			global.Logger.Warn("Rate limiting disabled: Redis client not initialized")
			c.Next()
			return
		}

		// 获取限流键
		key := getRateLimitKey(c, rateLimitConfig.KeyFunc)
		if key == "" {
			// 无法获取键，放行请求
			c.Next()
			return
		}

		// 检查是否超过限流
		allowed, err := checkRateLimit(key, rateLimitConfig.Requests, rateLimitConfig.Window)
		if err != nil {
			// Redis错误，记录日志但不阻止请求
			global.Logger.Error(fmt.Sprintf("Rate limit check failed: %v", err))
			c.Next()
			return
		}

		if !allowed {
			// 超过限流，返回429
			common.FailWithCode(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRateLimitKey 根据配置获取限流键
func getRateLimitKey(c *gin.Context, keyFunc string) string {
	switch keyFunc {
	case "ip":
		// 基于IP地址限流
		return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
	case "user":
		// 基于用户ID限流（需要先通过JWT认证）
		userID, exists := c.Get("userId")
		if !exists {
			// 未认证用户，回退到IP限流
			return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		}
		return fmt.Sprintf("rate_limit:user:%v", userID)
	default:
		return ""
	}
}

// checkRateLimit 使用滑动窗口算法检查是否超过限流
// 返回 (是否允许, 错误)
func checkRateLimit(key string, maxRequests int, windowSeconds int) (bool, error) {
	ctx := context.Background()
	now := time.Now().Unix()
	windowStart := now - int64(windowSeconds)

	// 使用Redis的有序集合实现滑动窗口
	// 1. 移除窗口外的旧记录
	err := global.RedisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart)).Err()
	if err != nil {
		return false, fmt.Errorf("failed to remove old records: %w", err)
	}

	// 2. 统计当前窗口内的请求数
	count, err := global.RedisClient.ZCard(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to count requests: %w", err)
	}

	// 3. 检查是否超过限制
	if count >= int64(maxRequests) {
		return false, nil
	}

	// 4. 添加当前请求到窗口
	// 使用当前时间戳作为score和member（加上纳秒确保唯一性）
	member := fmt.Sprintf("%d:%d", now, time.Now().UnixNano())
	err = global.RedisClient.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: member,
	}).Err()
	if err != nil {
		return false, fmt.Errorf("failed to add request record: %w", err)
	}

	// 5. 设置键的过期时间（窗口大小的2倍，确保数据清理）
	err = global.RedisClient.Expire(ctx, key, time.Duration(windowSeconds*2)*time.Second).Err()
	if err != nil {
		// 过期时间设置失败不影响限流逻辑
		global.Logger.Warn(fmt.Sprintf("Failed to set expiration for rate limit key: %v", err))
	}

	return true, nil
}
