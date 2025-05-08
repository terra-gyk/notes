package controller

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "yourproject/internal/entity"
    "yourproject/internal/service"
)

// UserController 定义用户控制器
type UserController struct {
    userService *service.UserService
}

// NewUserController 创建一个新的用户控制器实例
func NewUserController() *UserController {
    // 初始化仓库
    userRepository := repository.NewInMemoryUserRepository()
    // 初始化服务
    userService := service.NewUserService(userRepository)

    return &UserController{
        userService: userService,
    }
}

// GetUser 处理获取用户信息的 HTTP 请求
func (c *UserController) GetUser(ctx *gin.Context) {
    id := ctx.Param("id")
    user, err := c.userService.GetUser(id)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if user == nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    ctx.JSON(http.StatusOK, user)
}

// CreateUser 处理创建用户的 HTTP 请求
func (c *UserController) CreateUser(ctx *gin.Context) {
    var user entity.User
    if err := ctx.ShouldBindJSON(&user); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := c.userService.CreateUser(&user); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusCreated, user)
}
    