package logging

import (
  "context"
  "net/http"
  "time"

  "github.com/gin-gonic/gin"
  "go.uber.org/zap"
  "gorm.io/gorm"
  "gorm.io/gorm/logger"
)

// InitLogger 初始化 Zap 日志
func InitLogger() (*zap.Logger, error) {
  return zap.NewProduction()
}

// GinLogger 是 Gin 的 Zap 日志中间件
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
  return func(c *gin.Context) {
    start := time.Now()
    c.Next()
    latency := time.Since(start)
    statusCode := c.Writer.Status()
    clientIP := c.ClientIP()
    method := c.Request.Method
    path := c.Request.URL.Path

    logger.Info("gin request",
      zap.Int("status", statusCode),
      zap.String("client_ip", clientIP),
      zap.String("method", method),
      zap.String("path", path),
      zap.Duration("latency", latency),
    )
  }
}

// GormZapLogger 实现 GORM 的 logger.Interface 接口
type GormZapLogger struct {
  Logger *zap.Logger
}

func (l *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
  return l
}

func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
  l.Logger.Info(msg, zap.Any("data", data))
}

func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
  l.Logger.Warn(msg, zap.Any("data", data))
}

func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
  l.Logger.Error(msg, zap.Any("data", data))
}

func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
  sql, rows := fc()
  fields := []zap.Field{
    zap.String("sql", sql),
    zap.Int64("rows", rows),
    zap.Duration("duration", time.Since(begin)),
  }
  if err != nil {
    fields = append(fields, zap.Error(err))
  }
  l.Logger.Info("gorm query", fields...)
}
  