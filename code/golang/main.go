package main

import (
    "github.com/gin-gonic/gin"
    "yourproject/internal/controller"
)

func main() {
    r := gin.Default()

    // 初始化控制器
    userController := controller.NewUserController()

    // 定义路由
    userGroup := r.Group("/users")
    {
        userGroup.GET("/:id", userController.GetUser)
        userGroup.POST("/", userController.CreateUser)
    }

    // 启动服务
    r.Run(":8080")
}
    