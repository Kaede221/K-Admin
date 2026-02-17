package main

// This file ensures all required dependencies are marked as direct dependencies in go.mod
// These imports will be used throughout the project as we build out the features

import (
	_ "github.com/casbin/casbin/v3"
	_ "github.com/casbin/gorm-adapter/v3"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/spf13/viper"
	_ "go.uber.org/zap"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/gorm"
)
