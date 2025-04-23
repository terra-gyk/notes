package controller

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"
  "yourproject/internal/score_service/entity"
  "yourproject/internal/score_service/repository"
  "yourproject/internal/score_service/service"
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

// GetUserScore 处理获取用户分数的 HTTP 请求
func (c *UserController) GetUserScore(ctx *gin.Context) {
  id := ctx.Param("id")
  user, err := c.userService.GetUserScore(id)
  if err != nil {
    c.logger.Error("failed to get user score", zap.String("id", id), zap.Error(err))
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

// UpdateUserScore 处理更新用户分数的 HTTP 请求
func (c *UserController) UpdateUserScore(ctx *gin.Context) {
  id := ctx.Param("id")
  var req struct {
    Score int `json:"score"`
  }
  if err := ctx.ShouldBindJSON(&req); err != nil {
    c.logger.Error("failed to bind user score json", zap.Error(err))
    ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if err := c.userService.UpdateUserScore(id, req.Score); err != nil {
    c.logger.Error("failed to update user score", zap.String("id", id), zap.Error(err))
    ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.logger.Info("user score updated successfully", zap.String("id", id))
  ctx.JSON(http.StatusOK, gin.H{"message": "User score updated successfully"})
}
  