package routers

import (
	"Rhine-Cloud-Driver/config"
	"Rhine-Cloud-Driver/routers/controllers"

	"github.com/gin-gonic/gin"
)

func InitRouter(cf config.Config) *gin.Engine {
	router := gin.Default()
	r := router.Group("/api/v1")

	// 文件路由
	fileRouter := r.Group("file")
	fileRouter.GET("list", controllers.FileDemo)

	// 用户路由
	userRouter := r.Group("user")
	userRouter.Use()
	userRouter.POST("login", controllers.UserDemo)

	// 管理员路由
	adminRouter := r.Group("admin")
	adminRouter.GET("web-setting", controllers.AdminDemo)

	router.Run(cf.Server.Host)
	return router
}
