package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
	"Rhine-Cloud-Driver/logic/log"
	"crypto/sha256"
	"strings"

	"go.uber.org/zap"
)

// 用户结构体
type User struct {
	Uid          int64  // 用户ID
	Name         string // 用户名称
	Password     string // 用户密码
	Email        string // 用户邮箱
	CreateTime   string // 创建时间
	UsedStorage  int64  // 已用容量
	TotalStorage int64  // 总容量
	GroupId      int64  // 所属用户组
}

func setHaltHash(password string) string {
	halt := common.RandStringRunes(16)
	hash := sha256.New()
	value := hash.Sum([]byte(password + string(halt)))
	return string(value) + ":" + string(halt)
}

func (user *User) veifyPassword(password string) bool {
	stringArray := strings.Split(user.Password, ":")
	hash := sha256.New()
	value := hash.Sum([]byte(password + stringArray[1]))
	return string(value) == stringArray[0]
}

// 验证访问权限
func (user *User) VerifyAccess(token string, uid int64, password string) (bool, int64, string) {
	// token校验
	if token != "" {
		isok, uid := jwt.TokenGetUid(token)
		if isok == false {
			return false, -1, ""
		}
		return jwt.TokenValid(token), uid, ""
	}
	// 密码校验
	var count int64
	gormDB.Table("users").Where("uid", uid).Count(&count)
	if count == 0 {
		return false, -1, ""
	}
	var actualPassword string
	gormDB.Table("users").Where("uid", uid).Find(&actualPassword)
	if actualPassword == password {
		// 生成新的token下发
		token, err := jwt.GenerateToken(uid)
		if err != nil {
			log.Logger.Error("生成token错误", zap.Error(err))
			return false, -1, ""
		}
		return true, uid, token
	}
	return false, uid, ""
}

// 新用户生成
func (user *User) AddUser() bool {
	return false
}

// 待后续完善 禁止用户
func (user *User) BanUser() {

}

// 待后续完善 修改用户所属用户组
func (user *User) EditUserGroup(newGroupId int64) {

}
