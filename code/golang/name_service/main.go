package main

import (
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
  "github.com/gin-gonic/gin"
  "yourproject/internal/name_service/controller"
  "yourproject/internal/name_service/repository"
  "yourproject/common/logging"
  "log"
)

func main() {
  // 初始化 Zap 日志
  logger, err := logging.InitLogger()
  if err != nil {
    log.Fatalf("failed to initialize zap logger: %v", err)
  }
  defer logger.Sync()

  // 初始化 GORM 数据库连接
  db, err := gorm.Open(sqlite.Open("name_service.db"), &gorm.Config{
    Logger: &logging.GormZapLogger{Logger: logger},
  })
  if err != nil {
    logger.Fatal("failed to connect database", zap.Error(err))
  }

  // 自动迁移 User 表
  db.AutoMigrate(&repository.User{})

  r := gin.New()

  // 添加 Gin 日志中间件
  r.Use(logging.GinLogger(logger))

  // 初始化控制器
  userController := controller.NewUserController(db, logger)

  // 定义路由
  userGroup := r.Group("/users")
  {
    userGroup.GET("/:id/name", userController.GetUserByName)
    userGroup.PUT("/:id/name", userController.UpdateUserName)
  }

  // 启动服务
  if err := r.Run(":9001"); err != nil {
    logger.Fatal("failed to start server", zap.Error(err))
  }
}
  