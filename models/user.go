package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
	"Rhine-Cloud-Driver/logic/log"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 用户结构体
type User struct {
	Uid          uint64 `json:"uid" gorm:"primarykey"`                        // 用户ID
	Name         string `json:"name" gorm:"size:30"`                          // 用户名称
	Password     string `json:"password" gorm:"size:255"`                     // 用户密码
	Email        string `json:"email" gorm:"size:255;index:idx_email,unique"` // 用户邮箱
	CreateTime   string `json:"create_time"`                                  // 创建时间
	UsedStorage  int64  `json:"used_storage"`                                 // 已用容量
	TotalStorage int64  `json:"total_storage"`                                // 总容量
	GroupId      int64  `json:"group_id"`                                     // 所属用户组
}

func setHaltHash(password string) string {
	halt := common.RandStringRunes(16)
	hash := sha256.New()
	value := hex.EncodeToString(hash.Sum([]byte(password + string(halt))))
	return value + ":" + halt
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
	matched, err := regexp.Match("^[a-zA-Z0-9]+([-_.][A-Za-zd]+)*@([a-zA-Z0-9]+[-.])+[A-Za-z]{2,5}$", []byte(email))
	if err != nil {
		log.Logger.Error("邮箱正则表达式匹配错误", zap.Error(err))
		return false
	}
	return matched
}

func (user *User) verifyPassword(password string) bool {
	stringArray := strings.Split(user.Password, ":")
	hash := sha256.New()
	value := hex.EncodeToString(hash.Sum([]byte(password + stringArray[1])))
	log.Logger.Debug("test", zap.Any("value", value), zap.Any("p", stringArray[0]))
	return value == stringArray[0]
}

// 验证访问权限
func (user *User) VerifyAccess(token string, email string, password string) (string, error) {
	// token校验
	if token != "" {
		isok, _ := jwt.TokenGetUid(token)
		if !(isok && jwt.TokenValid(token)) {
			return "", common.NewError(common.ERROR_USER_TOEKN_INVALIED)
		}
		return "", nil
	}
	// 密码校验
	DB.Table("users").Where("email", email).Find(&user)
	if user.Uid == 0 {
		return "", common.NewError(common.ERROR_USER_UID_PASSWORD_WRONG)
	}
	if user.verifyPassword(password) {
		// 生成新的token下发
		token, err := jwt.GenerateToken(user.Uid)
		if err != nil {
			log.Logger.Error("生成token错误", zap.Error(err))
			return "", common.NewError(common.ERROR_JWT_GENERATE_TOKEN_FAILED)
		}
		return token, nil
	}
	return "", common.NewError(common.ERROR_USER_UID_PASSWORD_WRONG)
}

// 新用户生成
func (user *User) AddUser() error {
	// 判断名称长度
	if !checkNewName(user.Name) {
		return common.NewError(common.ERROR_USER_NAME_LENGTH_NOT_MATCH)
	}
	// 判断密码长度以及字符规定
	if !checkNewPassword(user.Password) {
		return common.NewError(common.ERROR_USER_PASSWORD_NOT_MATCH_RULES)
	}
	// 判断邮箱是否合法 满足xxx@xxx.xxx条件
	if !checkNewEmail(user.Email) {
		return common.NewError(common.ERROR_USER_EMAIL_NOT_MATHCH_RULES)
	}
	// 加盐并加密
	user.Password = setHaltHash(user.Password)
	tx := DB.Session(&gorm.Session{})
	err := tx.Table("users").Create(&user).Error
	if err != nil {
		// 这里的错误有两种可能的问题，第一个是email冲突了，第二个是数据库真的出现错误
		// 我们返回给用户只考虑前者的情况
		log.Logger.Error("事务执行：插入新用户错误", zap.Any("user", &user), zap.Error(err))
		tx.Rollback()
		return common.NewError(common.ERROR_USER_EMAIL_CONFLICT)
	}
	if !common.Mkdir(user.Email) {
		tx.Rollback()
		return common.NewError(common.ERROR_USER_MKDIR_FAILED)
	}
	tx.Commit()
	return nil
}

// 禁止用户在考虑新建一个用户组，直接没有任何功能，即为封禁。
// todo 修改用户所属用户组
func (user *User) EditUserGroup(newGroupId int64) {

}
