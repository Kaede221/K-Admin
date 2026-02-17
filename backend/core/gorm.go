package core

import (
	"context"
	"fmt"
	"time"

	"k-admin-system/config"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection with Gorm
// Configures connection pooling, reconnection logic, and slow query logging
func InitDB(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	// Configure Gorm logger
	gormLogger := newGormLogger(log, cfg)

	// Open database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database instance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database connected successfully",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name),
		zap.Int("max_idle_conns", cfg.Database.MaxIdleConns),
		zap.Int("max_open_conns", cfg.Database.MaxOpenConns),
	)

	return db, nil
}

// gormLogger is a custom logger that integrates Gorm with Zap
type gormLogger struct {
	zapLogger         *zap.Logger
	logLevel          logger.LogLevel
	slowThreshold     time.Duration
	ignoreNotFoundErr bool
}

// newGormLogger creates a new Gorm logger that uses Zap
func newGormLogger(log *zap.Logger, cfg *config.Config) logger.Interface {
	// Determine log level based on server mode
	var logLevel logger.LogLevel
	switch cfg.Server.Mode {
	case "debug":
		logLevel = logger.Info
	case "test":
		logLevel = logger.Warn
	default:
		logLevel = logger.Error
	}

	return &gormLogger{
		zapLogger:         log,
		logLevel:          logLevel,
		slowThreshold:     200 * time.Millisecond, // Default slow query threshold
		ignoreNotFoundErr: true,
	}
}

// LogMode sets the log level
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info logs info messages
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.zapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn logs warning messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.zapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error logs error messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.zapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace logs SQL queries with execution time
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Log errors
	if err != nil && (!l.ignoreNotFoundErr || err != gorm.ErrRecordNotFound) {
		l.zapLogger.Error("Database query error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
		return
	}

	// Log slow queries
	if elapsed >= l.slowThreshold {
		l.zapLogger.Warn("Slow query detected",
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.slowThreshold),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
		return
	}

	// Log all queries in debug mode
	if l.logLevel >= logger.Info {
		l.zapLogger.Debug("Database query",
			zap.Duration("elapsed", elapsed),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
	}
}
