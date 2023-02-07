package model

import (
	"Rhine-Cloud-Driver/common"
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
	_, _, _, err := BuildFileSystem(user.Uid, "/", 50, 0)
	asserts.Nil(err)
	//含有非法字符
	_, _, _, err = BuildFileSystem(user.Uid, "/*", 50, 0)
	asserts.Equal(err.Error(), common.NewError(common.ERROR_FILE_PATH_INVALID).Error())
	//前后未加/且不存在该路径
	_, _, _, err = BuildFileSystem(user.Uid, "123", 50, 0)
	asserts.Equal(err.Error(), common.NewError(common.ERROR_FILE_PATH_INVALID).Error())
	// 超过50限制
	_, _, _, err = BuildFileSystem(user.Uid, "/", 100, 0)
	asserts.Equal(err.Error(), common.NewError(common.ERROR_FILE_COUNT_EXCEED_LIMIT).Error())
}

func Test_CheckFileExist(t *testing.T) {

}
