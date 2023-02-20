package routers

import (
	"Rhine-Cloud-Driver/config"
	"Rhine-Cloud-Driver/middleware"
	model "Rhine-Cloud-Driver/models"
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
	r.POST("share_info", controllers.GetShareDetail)
	r.GET("download/:key", controllers.DownloadFile)
	r.POST("get_share_file", controllers.GetShareFile)
	r.GET("get_avatar", controllers.GetUserAvatar)
	// 接下来需要鉴权
	// 用户路由
	userRouter := r.Group("")
	userRouter.Use(middleware.TokenVerify())
	{
		userRouter.GET("get_info", controllers.GetUserDetail)
		userRouter.POST("change_setting", controllers.ChangeUserInfo)
		userRouter.POST("upload_avatar", controllers.UploadAvatar)
		// 文件路由
		fileRouter := r.Group("")
		fileRouter.Use(middleware.PermissionVerify(model.PERMISSION_FILE))
		fileRouter.POST("directory", controllers.GetMyFiles)
		fileRouter.POST("mkdir", controllers.Mkdir)
		fileRouter.POST("task_create", controllers.UploadTaskCreate)
		fileRouter.POST("upload", controllers.Upload)
		fileRouter.POST("merge", controllers.MergeFileChunks)
		fileRouter.POST("move_files", controllers.MoveFiles)
		fileRouter.POST("get_download_key", controllers.GetDownloadKey)
		fileRouter.POST("remove_files", controllers.RemoveFiles)

		shareRouter := r.Group("")
		shareRouter.Use(middleware.PermissionVerify(model.PERMISSION_SHARE))
		shareRouter.POST("new_share", controllers.CreateNewShare)
		shareRouter.POST("transfer_files", controllers.TransferFiles)
		shareRouter.POST("cancel_share", controllers.CancelShare)
		shareRouter.POST("get_my_share", controllers.GetMyShare)

		settingRouter := r.Group("")
		settingRouter.Use(middleware.PermissionVerify(model.PERMISSION_SETTING))

		// 管理员路由
		adminRouter := r.Group("admin")
		adminRouter.Use(middleware.PermissionVerify(model.PERMISSION_ADMIN_READ))
		adminRouter.GET("web_data", controllers.AdminDemo)
		adminRouter.POST("get_all_users", controllers.GetAllUser)
		adminRouter.POST("get_user_info", controllers.GetUserInfo)
		adminRouter.POST("get_all_shares", controllers.GetAllShare)
		adminRouter.POST("get_all_groups", controllers.GetAllGroup)
		adminRouter.POST("get_all_files", controllers.GetAllFile)
		adminRouter.POST("get_user_detail", controllers.AdminGetUserDetail)
		adminRouter.POST("edit_user_info", controllers.AdminEditUserInfo)
		adminRouter.POST("upload_avatar", controllers.AdminUploadAvatar)
	}

	router.Run(cf.Server.Host)
	return router
}
