package common

import "errors"

// 所有错误都在此定义

// 用户相关错误
const (
	ERROR_USER_NAME_LENGTH_NOT_MATCH    = 100001
	ERROR_USER_PASSWORD_NOT_MATCH_RULES = 100002
)

var errMap = map[int]string{
	ERROR_USER_NAME_LENGTH_NOT_MATCH:    "用户名称长度不符合规定",
	ERROR_USER_PASSWORD_NOT_MATCH_RULES: "用户密码不符合规定",
}

func NewError(errorNum int64) error {
	return errors.New(errMap[int(errorNum)])
}
