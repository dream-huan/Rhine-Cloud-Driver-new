package model

import (
	"Rhine-Cloud-Driver/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_BuildFileSystem(t *testing.T) {
	asserts := assert.New(t)
	// 创建一个新用户，以拿到其的UID
	user := User{
		Name:       "test",
		Password:   "123456",
		Email:      "17960@qq.com",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	user.AddUser()
	//拿到其UID
	DB.Table("users").Where("email=?", user.Email).Find(&user)
	//正常路径
	_, _, _, _, err := BuildFileSystem(user.Uid, "/", 0, 50, 0, "", []string{})
	asserts.Nil(err)
	//含有非法字符
	_, _, _, _, err = BuildFileSystem(user.Uid, "/*", 0, 50, 0, "", []string{})
	asserts.Equal(err.Error(), util.NewError(util.ERROR_FILE_PATH_INVALID).Error())
	//前后未加/且不存在该路径
	_, _, _, _, err = BuildFileSystem(user.Uid, "123", 0, 50, 0, "", []string{})
	asserts.Equal(err.Error(), util.NewError(util.ERROR_FILE_PATH_INVALID).Error())
	// 超过50限制
	_, _, _, _, err = BuildFileSystem(user.Uid, "/", 0, 100, 0, "", []string{})
	asserts.Equal(err.Error(), util.NewError(util.ERROR_FILE_COUNT_EXCEED_LIMIT).Error())
}

//func Test_CheckFileRepeat(t *testing.T) {
//	CheckFileRepeat(525, 1, "a82d91ad01abcb36af53cf433b36eacea35ea3b42082386d9b0f353921945962")
//
//}
