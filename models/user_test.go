package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/config"
	"Rhine-Cloud-Driver/logic/log"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
	"time"
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
	asserts := assert.New(t)
	user := User{
		Name:       "test",
		Password:   "123456",
		Email:      "1@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err := user.AddUser()
	asserts.NoError(err)
	user = User{
		Name:       "test",
		Password:   "123456",
		Email:      "123@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err = user.AddUser()
	asserts.NoError(err)
	user = User{
		Name:       "test",
		Password:   "123456",
		Email:      "1@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err = user.AddUser()
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_EMAIL_CONFLICT).Error())
	user = User{
		Name:       "",
		Password:   "123456",
		Email:      "18@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err = user.AddUser()
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_NAME_LENGTH_NOT_MATCH).Error())
	user = User{
		Name:       "1",
		Password:   "12345",
		Email:      "19@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err = user.AddUser()
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_PASSWORD_NOT_MATCH_RULES).Error())
	user = User{
		Name:       "1",
		Password:   "123456",
		Email:      "144@qq",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err = user.AddUser()
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_EMAIL_NOT_MATHCH_RULES).Error())
}

func Test_Login(t *testing.T) {
	asserts := assert.New(t)
	user := User{}
	// 邮箱登录 账户密码正确
	token, err := user.VerifyAccess("", 0, "1@qq.com", "123456")
	asserts.Nil(err)
	// token正确
	token, err = user.VerifyAccess(token, 0, "", "")
	asserts.Nil(err)
	// uid登录 账户密码正确
	token, err = user.VerifyAccess("", user.Uid, "", "123456")
	asserts.Nil(err)
	// token错误/非法
	token, err = user.VerifyAccess("12gregreg", 0, "", "")
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_TOEKN_INVALIED).Error())
	// uid登录 密码错误
	token, err = user.VerifyAccess("", user.Uid, "", "321456")
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_UID_PASSWORD_WRONG).Error())
	// 邮箱登录 密码错误
	token, err = user.VerifyAccess("", 0, "1@qq.com", "321456")
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_UID_PASSWORD_WRONG).Error())
	// 邮箱不存在
	token, err = user.VerifyAccess("", 0, "2@qq.com", "123456")
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_UID_PASSWORD_WRONG).Error())
	// 邮箱不填
	token, err = user.VerifyAccess("", 0, "", "123456")
	asserts.Equal(err.Error(), common.NewError(common.ERROR_USER_NOT_UID_AND_EMAIL).Error())
}
