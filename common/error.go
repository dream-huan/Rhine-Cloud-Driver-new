package common

import "errors"

// 所有错误都在此定义

// 用户相关错误
const (
	ERROR_USER_NAME_LENGTH_NOT_MATCH    = 100001
	ERROR_USER_PASSWORD_NOT_MATCH_RULES = 100002
	ERROR_USER_EMAIL_NOT_MATHCH_RULES   = 100003
	ERROR_USER_MKDIR_FAILED             = 100004
	ERROR_USER_TOEKN_INVALIED           = 100005
	ERROR_USER_UID_PASSWORD_WRONG       = 100006
	ERROR_USER_EMAIL_CONFLICT           = 100007
)

// 数据库相关错误
const (
	ERROR_DB_WRITE_FAILED = 200001
	ERROR_DB_READ_FAILED  = 200002
)

// jwt相关错误
const (
	ERROR_JWT_GENERATE_TOKEN_FAILED = 300001
)

var errMap = map[int]string{
	ERROR_USER_NAME_LENGTH_NOT_MATCH:    "用户名称长度不符合规定",
	ERROR_USER_PASSWORD_NOT_MATCH_RULES: "用户密码不符合规定",
	ERROR_USER_EMAIL_NOT_MATHCH_RULES:   "用户邮箱不符合规定",
	ERROR_USER_MKDIR_FAILED:             "用户注册新建个人文件夹错误",
	ERROR_DB_WRITE_FAILED:               "数据库写入错误",
	ERROR_DB_READ_FAILED:                "数据库读入错误",
	ERROR_USER_TOEKN_INVALIED:           "token无效或非法",
	ERROR_USER_UID_PASSWORD_WRONG:       "uid不存在或密码错误",
	ERROR_JWT_GENERATE_TOKEN_FAILED:     "token生成失败",
	ERROR_USER_EMAIL_CONFLICT:           "邮箱已被注册",
}

func NewError(errorNum int64) error {
	return errors.New(errMap[int(errorNum)])
}
