package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
	"Rhine-Cloud-Driver/logic/log"
	"crypto/sha256"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// 用户结构体
type User struct {
	Uid          string // 用户ID
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

func checkNewName(name string) bool {
	if len(name) <= 0 || len(name) > 12 {
		return false
	}
	return true
}

func checkNewPassword(password string) bool {
	if len(password) < 6 || len(password) > 18 {
		return false
	}
	// 除大小写和数字外，特殊字符仅能包含*、+、-、^、# 、@、!
	allowChar := map[string]bool{
		"*": true,
		"+": true,
		"-": true,
		"^": true,
		"#": true,
		"@": true,
		"!": true,
	}
	for _, v := range password {
		if !((v >= 'A' && v <= 'Z') || (v >= 'a' || v <= 'z') || (v >= '0' && v <= '9') || (allowChar[string(v)])) {
			return false
		}
	}
	return true
}

func checkNewEmail(email string) bool {
	matched, err := regexp.Match("/^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(.[a-zA-Z0-9_-])+/", []byte(email))
	if err != nil {
		log.Logger.Error("邮箱正则表达式匹配错误", zap.Error(err))
		return false
	}
	return matched
}

func (user *User) verifyPassword(password string) bool {
	stringArray := strings.Split(user.Password, ":")
	hash := sha256.New()
	value := hash.Sum([]byte(password + stringArray[1]))
	return string(value) == stringArray[0]
}

// 验证访问权限
func (user *User) VerifyAccess(token string, email string, password string) (bool, string) {
	// token校验
	if token != "" {
		isok, _ := jwt.TokenGetUid(token)
		if !isok {
			return false, ""
		}
		return jwt.TokenValid(token), ""
	}
	// 密码校验
	gormDB.Table("users").Where("email", email).Find(&user)
	if user.Uid == "" {
		return false, ""
	}
	if user.verifyPassword(password) {
		// 生成新的token下发
		token, err := jwt.GenerateToken(user.Uid)
		if err != nil {
			log.Logger.Error("生成token错误", zap.Error(err))
			return false, ""
		}
		return true, token
	}
	return false, ""
}

// 新用户生成
func (user *User) AddUser() (bool, int) {
	// 判断名称长度
	if !checkNewName(user.Name) {
		return false, common.ERROR_USER_NAME_LENGTH_NOT_MATCH
	}
	// 判断密码长度以及字符规定
	if !checkNewPassword(user.Password) {
		return false, common.ERROR_USER_PASSWORD_NOT_MATCH_RULES
	}
	// 判断邮箱是否合法 满足xxx@xxx.xxx条件
	if !checkNewEmail(user.Email) {
		return false, common.ERROR_USER_EMAIL_NOT_MATHCH_RULES
	}
	// 创建文件夹，待补充，文件夹的路径为：项目根目录/邮箱/
	// 生成唯一ID，用户到时登录采用email+密码进行登录
	// 初步考虑采用雪花算法来生成唯一ID

	return true, 0
}

// 待后续完善 禁止用户
func (user *User) BanUser() {

}

// 待后续完善 修改用户所属用户组
func (user *User) EditUserGroup(newGroupId int64) {

}
