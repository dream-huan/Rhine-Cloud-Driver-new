package model

import (
	"Rhine-Cloud-Driver/pkg/cache"
	"Rhine-Cloud-Driver/pkg/conf"
	"Rhine-Cloud-Driver/pkg/jwt"
	"Rhine-Cloud-Driver/pkg/log"
	"Rhine-Cloud-Driver/pkg/recaptcha"
	"Rhine-Cloud-Driver/pkg/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var db *sql.DB

// todo:设立钩子，进行权限认证

func initMysql(cf conf.MysqlConfig) {
	// var err error
	// // dsn := "root:SUIbianla123@tcp(127.0.0.1:3306)/project"
	dsn := cf.User + ":" + cf.Password + "@tcp(" + cf.Address + ")/" + cf.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Logger.Error("数据库链接错误", zap.Error(err))
		return
	}
	log.Logger.Info("MySQL数据库链接成功")
	db, err = DB.DB()
	if err != nil {
		log.Logger.Error("获取数据库DB错误", zap.Error(err))
	}
	db.SetMaxOpenConns(100)

	// 自动建表+建立索引
	DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&User{}, &Group{}, &File{}, &Share{}, &Setting{})
	DB.Table("settings").Create(&Setting{Name: "register_open", Value: "0"})
}

func initJwt(cf conf.JwtConfig) {
	jwt.Init(cf.Key)
}

func Init(cf conf.Config) {
	initMysql(cf.MysqlManager)
	initJwt(cf.JwtKey)
	util.NewWorker(1)
	cache.InitRedis(cf.RedisManager)
	InitGroupPermission()
	recaptcha.InitRecaptcha(cf.GoogleRecaptchaPrivateKey.Key)
	util.MkdirIfNoExist("metadata")
}
