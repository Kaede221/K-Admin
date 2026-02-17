package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k-admin-system/global"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	RoleID   uint   `json:"roleId"`
	jwt.RegisteredClaims
}

var (
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenBlacklisted = errors.New("token is blacklisted")
)

// GenerateToken 生成访问令牌和刷新令牌
func GenerateToken(userID uint, username string, roleID uint) (accessToken, refreshToken string, err error) {
	// 生成访问令牌
	accessExpiration := time.Duration(global.Config.JWT.AccessExpiration) * time.Minute
	accessClaims := JWTClaims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(global.Config.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshExpiration := time.Duration(global.Config.JWT.RefreshExpiration) * 24 * time.Hour
	refreshClaims := JWTClaims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(global.Config.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析令牌
func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(global.Config.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// 检查令牌是否在黑名单中
		if IsTokenBlacklisted(tokenString) {
			return nil, ErrTokenBlacklisted
		}
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新访问令牌
func RefreshToken(refreshTokenString string) (newAccessToken string, err error) {
	// 解析刷新令牌
	claims, err := ParseToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// 生成新的访问令牌
	accessExpiration := time.Duration(global.Config.JWT.AccessExpiration) * time.Minute
	newClaims := JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		RoleID:   claims.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newAccessToken, err = tokenObj.SignedString([]byte(global.Config.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	return newAccessToken, nil
}

// AddTokenToBlacklist 将令牌添加到黑名单
func AddTokenToBlacklist(tokenString string) error {
	if global.RedisClient == nil {
		return errors.New("redis client is not initialized")
	}

	// 解析令牌获取过期时间
	claims, err := ParseToken(tokenString)
	if err != nil && !errors.Is(err, ErrTokenBlacklisted) {
		return err
	}

	// 计算令牌剩余有效时间
	expiration := time.Until(claims.ExpiresAt.Time)
	if expiration <= 0 {
		// 令牌已过期，无需加入黑名单
		return nil
	}

	// 将令牌添加到Redis黑名单，设置过期时间
	ctx := context.Background()
	key := fmt.Sprintf("blacklist:%s", tokenString)
	err = global.RedisClient.Set(ctx, key, "1", expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsTokenBlacklisted 检查令牌是否在黑名单中
func IsTokenBlacklisted(tokenString string) bool {
	if global.RedisClient == nil {
		return false
	}

	ctx := context.Background()
	key := fmt.Sprintf("blacklist:%s", tokenString)
	result, err := global.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 键不存在，令牌不在黑名单中
			return false
		}
		// 其他错误，为安全起见返回true
		return true
	}

	return result == "1"
}
