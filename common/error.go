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
	ERROR_USER_NOT_UID_AND_EMAIL        = 100008
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

// 公共库相关错误
const (
	ERROR_COMMON_SNOWFLOWS_GENERATE     = 400001
	ERROR_COMMON_RECAPTCHA_VERIFICATION = 400002
)

// 路由相关错误
const (
	ERROR_ROUTER_PARSEJSON = 500001
)

// 鉴权相关错误
const (
	ERROR_AUTH_GET_TOKEN     = 600001
	ERROR_AUTH_TOKEN_INVALID = 600002
)

var errMap = map[int]string{
	ERROR_USER_NAME_LENGTH_NOT_MATCH:    "用户名称长度不符合规定",
	ERROR_USER_PASSWORD_NOT_MATCH_RULES: "用户密码不符合规定",
	ERROR_USER_EMAIL_NOT_MATHCH_RULES:   "用户邮箱不符合规定",
	ERROR_USER_NOT_UID_AND_EMAIL:        "用户ID或用户邮箱未给出",
	ERROR_USER_MKDIR_FAILED:             "用户注册新建个人文件夹错误",
	ERROR_DB_WRITE_FAILED:               "数据库写入错误",
	ERROR_DB_READ_FAILED:                "数据库读入错误",
	ERROR_USER_TOEKN_INVALIED:           "token无效或非法",
	ERROR_USER_UID_PASSWORD_WRONG:       "uid/邮箱不存在或密码错误",
	ERROR_JWT_GENERATE_TOKEN_FAILED:     "token生成失败",
	ERROR_USER_EMAIL_CONFLICT:           "邮箱已被注册",
	ERROR_COMMON_SNOWFLOWS_GENERATE:     "雪花算法生成错误",
	ERROR_ROUTER_PARSEJSON:              "JSON结构体解析错误",
	ERROR_COMMON_RECAPTCHA_VERIFICATION: "recaptcha验证码错误",
	ERROR_AUTH_GET_TOKEN:                "无法拿到token",
	ERROR_AUTH_TOKEN_INVALID:            "token无效",
}

func NewError(errorNum int) error {
	return errors.New(errMap[errorNum])
}
