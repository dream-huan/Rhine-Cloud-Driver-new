package routers

import (
	"Rhine-Cloud-Driver/config"
	"Rhine-Cloud-Driver/middleware"
	"Rhine-Cloud-Driver/routers/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter(cf config.Config) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 1024 << 20
	// 跨域设置
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin"},
		AllowCredentials: true,
	}))
	// r为总路由
	r := router.Group("/api/v1")

	// 注册和登录路由
	r.POST("login", controllers.UserLogin)
	r.POST("register", controllers.UserRegister)
	r.POST("s/:key", controllers.GetShareDetail)

	// 接下来需要鉴权
	// 用户路由
	userRouter := r.Group("")
	userRouter.Use(middleware.TokenVerify())
	{
		userRouter.GET("get_info", controllers.GetUserDetail)
		// 文件路由
		fileRouter := r.Group("")
		fileRouter.POST("directory", controllers.GetMyFiles)
		fileRouter.POST("mkdir", controllers.Mkdir)
		fileRouter.POST("task_create", controllers.UploadTaskCreate)
		fileRouter.POST("upload", controllers.Upload)
		fileRouter.POST("merge", controllers.MergeFileChunks)
	}

	// 管理员路由
	adminRouter := r.Group("admin")
	adminRouter.GET("web-setting", controllers.AdminDemo)

	router.Run(cf.Server.Host)
	return router
}
