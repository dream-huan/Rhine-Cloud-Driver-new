package model

import (
	"database/sql"

	_ "Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/config"
	"Rhine-Cloud-Driver/logic/jwt"
	log "Rhine-Cloud-Driver/logic/log"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var db *sql.DB

func initMysql(cf config.MysqlConfig) {
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
	DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&User{})
}

func initJwt(cf config.JwtConfig) {
	jwt.Init(cf.Key)
}

func Init(cf config.Config) {
	initMysql(cf.MysqlManager)
	initJwt(cf.JwtKey)
}
