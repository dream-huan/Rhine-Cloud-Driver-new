package model

import (
	"Rhine-Cloud-Driver/config"
	log "Rhine-Cloud-Driver/logic/log"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func init() {
	var cf config.Config
	configFile, err := ioutil.ReadFile("../conf/Rhine-Cloud-Driver.yaml")
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}
	err = yaml.Unmarshal(configFile, &cf)
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}
	log.Logger, err = log.NewLogger(cf.Log.LogPath, cf.Log.LogLevel, cf.Log.MaxSize, cf.Log.MaxBackup,
		cf.Log.MaxAge, cf.Log.Compress, cf.Log.LogConsole, cf.Log.ServiceName)
	if err != nil {
		log.Logger.Error("Unmarshal yaml file error", zap.Error(err))
	}
	Init(cf)
}

func Test_AddUser(t *testing.T) {
	user := User{
		Name:       "test",
		Password:   "123456",
		Email:      "1@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	fmt.Printf("%v\n", user.AddUser())
	user2 := User{
		Name:       "test",
		Password:   "123456",
		Email:      "123@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	fmt.Printf("%v\n", user2.AddUser())
	user3 := User{
		Name:       "test",
		Password:   "123456",
		Email:      "1@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	fmt.Printf("%v\n", user3.AddUser())
}

func Test_Login(t *testing.T) {
	user := User{}
	token, err := user.VerifyAccess("", "1@qq.com", "123456")
	fmt.Printf("%v %v\n", err, token)
	DB.Table("users").Where("email", "1@qq.com").Find(&user)
	fmt.Println(user)
	user = User{}
	token, err = user.VerifyAccess("", "1@qq.com", "321456")
	fmt.Printf("%v %v\n", err, token)
	user = User{}
	token, err = user.VerifyAccess("", "2@qq.com", "123456")
	fmt.Printf("%v %v\n", err, token)
	user = User{}
	token, err = user.VerifyAccess("", "7891374@qq.com", "123456")
	fmt.Printf("%v %v\n", err, token)
}
