package controller

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"
  "yourproject/internal/name_service/entity"
  "yourproject/internal/name_service/repository"
  "yourproject/internal/name_service/service"
  "go.uber.org/zap"
)

// UserController 定义用户控制器
type UserController struct {
  userService *service.UserService
  logger    *zap.Logger
}

// NewUserController 创建一个新的用户控制器实例
func NewUserController(db *gorm.DB, logger *zap.Logger) *UserController {
  // 初始化仓库
  userRepository := repository.NewGormUserRepository(db, logger)
  // 初始化服务
  userService := service.NewUserService(userRepository)

  return &UserController{
    userService: userService,
    logger:    logger,
  }
}

// GetUserByName 处理获取用户姓名的 HTTP 请求
func (c *UserController) GetUserByName(ctx *gin.Context) {
  id := ctx.Param("id")
  user, err := c.userService.GetUserByName(id)
  if err != nil {
    c.logger.Error("failed to get user name", zap.String("id", id), zap.Error(err))
    ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  if user == nil {
    c.logger.Info("user not found", zap.String("id", id))
    ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    return
  }
  ctx.JSON(http.StatusOK, user)
}

// UpdateUserName 处理更新用户姓名的 HTTP 请求
func (c *UserController) UpdateUserName(ctx *gin.Context) {
  id := ctx.Param("id")
  var req struct {
    Name string `json:"name"`
  }
  if err := ctx.ShouldBindJSON(&req); err != nil {
    c.logger.Error("failed to bind user name json", zap.Error(err))
    ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if err := c.userService.UpdateUserName(id, req.Name); err != nil {
    c.logger.Error("failed to update user name", zap.String("id", id), zap.Error(err))
    ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.logger.Info("user name updated successfully", zap.String("id", id))
  ctx.JSON(http.StatusOK, gin.H{"message": "User name updated successfully"})
}
  