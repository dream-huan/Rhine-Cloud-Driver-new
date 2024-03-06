package routers

import (
	"Rhine-Cloud-Driver/middleware"
	model "Rhine-Cloud-Driver/models"
	"Rhine-Cloud-Driver/pkg/conf"
	"Rhine-Cloud-Driver/routers/controllers"
	"github.com/gin-gonic/gin"
	"net/http/pprof"
	_ "net/http/pprof"
)

func registerPprofRoutes(router *gin.Engine) {
	// 创建一个pprof的路由组
	pprofGroup := router.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
		pprofGroup.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
		pprofGroup.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		pprofGroup.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		pprofGroup.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
		pprofGroup.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	}
}

func InitRouter(cf conf.Config) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 1024 << 20
	// 跨域设置
	//router.Use(cors.New(cors.Config{
	//	AllowOrigins:     []string{"http://localhost:3000"},
	//	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	//	AllowHeaders:     []string{"Origin"},
	//	AllowCredentials: true,
	//}))
	router.Use(middleware.Recovery())
	// r为总路由
	r := router.Group("/api/v1")

	// 注册和登录路由
	r.POST("login", controllers.UserLogin)
	r.POST("register", controllers.UserRegister)
	r.GET("check_register", controllers.CheckRegister)
	r.POST("share_info", controllers.GetShareDetail)
	r.GET("download/:key", middleware.TaskIDVerify(middleware.USAGE_DOWNLOAD), controllers.DownloadFile)
	r.POST("get_share_file", controllers.GetShareFile)
	r.GET("get_avatar", controllers.GetUserAvatar)
	// 接下来需要鉴权
	// 用户路由
	userRouter := r.Group("")
	userRouter.Use(middleware.TokenVerify())
	{
		userRouter.GET("get_info", controllers.GetUserDetail)
		// 文件路由
		fileRouter := r.Group("")
		fileRouter.Use(middleware.TokenVerify())
		fileRouter.Use(middleware.PermissionVerify(model.PERMISSION_FILE))
		fileRouter.POST("directory", controllers.GetMyFiles)
		fileRouter.POST("mkdir", controllers.Mkdir)
		fileRouter.POST("task_create", controllers.UploadTaskCreate)
		fileRouter.POST("upload", middleware.TaskIDVerify(middleware.USAGE_UPLOAD), middleware.TrafficLimit(30), controllers.Upload)
		fileRouter.POST("merge", middleware.TaskIDVerify(middleware.USAGE_UPLOAD), controllers.MergeFileChunks)
		fileRouter.POST("move_files", controllers.MoveFiles)
		fileRouter.POST("get_download_key", controllers.GetDownloadKey)
		fileRouter.POST("remove_files", controllers.RemoveFiles)
		fileRouter.POST("rename_file", controllers.ReNameFile)
		fileRouter.POST("get_thumbnails", controllers.GetThumbnails)
		fileRouter.GET("get_thumbnail", controllers.GetThumbnail)

		shareRouter := r.Group("")
		shareRouter.Use(middleware.TokenVerify())
		shareRouter.Use(middleware.PermissionVerify(model.PERMISSION_SHARE))
		shareRouter.POST("new_share", controllers.CreateNewShare)
		shareRouter.POST("transfer_files", controllers.TransferFiles)
		shareRouter.POST("cancel_share", controllers.CancelShare)
		shareRouter.POST("get_my_share", controllers.GetMyShare)

		settingRouter := r.Group("")
		settingRouter.Use(middleware.TokenVerify())
		settingRouter.Use(middleware.PermissionVerify(model.PERMISSION_SETTING))
		settingRouter.POST("change_setting", controllers.ChangeUserInfo)
		settingRouter.POST("upload_avatar", controllers.UploadAvatar)

		// 管理员路由
		adminRouter := r.Group("admin")
		adminRouter.Use(middleware.PermissionVerify(model.PERMISSION_ADMIN_READ))
		adminRouter.POST("verify_admin", controllers.VerifyAdmin)
		adminRouter.GET("web_data", controllers.AdminDemo)
		adminRouter.POST("get_all_users", controllers.GetAllUser)
		adminRouter.POST("get_user_info", controllers.GetUserInfo)
		adminRouter.POST("get_all_shares", controllers.GetAllShare)
		adminRouter.POST("get_all_groups", controllers.GetAllGroup)
		adminRouter.POST("get_all_files", controllers.GetAllFile)
		adminRouter.POST("get_user_detail", controllers.AdminGetUserDetail)

		adminWriteRouter := r.Group("admin")
		adminWriteRouter.Use(middleware.PermissionVerify(model.PERMISSION_ADMIN_WRITE))
		adminWriteRouter.POST("edit_user_info", controllers.AdminEditUserInfo)
		adminWriteRouter.POST("upload_avatar", controllers.AdminUploadAvatar)
		adminWriteRouter.POST("edit_group_info", controllers.AdminEditGroupInfo)
		adminWriteRouter.POST("create_group", controllers.AdminCreateGroup)
		adminWriteRouter.POST("delete_group", controllers.AdminDeleteGroup)
	}

	// 注册pprof的路由到Gin
	registerPprofRoutes(router)

	router.Run(cf.Server.Host)
	return router
}
