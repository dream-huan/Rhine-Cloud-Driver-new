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
	ERROR_USER_STORAGE_EXCEED           = 100009
	ERROR_USER_NOT_EXIST                = 100010
)

// 数据库相关错误
const (
	ERROR_DB_WRITE_FAILED = 200001
	ERROR_DB_READ_FAILED  = 200002
	ERROR_DB_NOT_DATA     = 200003
)

// jwt相关错误
const (
	ERROR_JWT_GENERATE_TOKEN_FAILED = 300001
)

// 公共库相关错误
const (
	ERROR_COMMON_SNOWFLOWS_GENERATE       = 400001
	ERROR_COMMON_RECAPTCHA_VERIFICATION   = 400002
	ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED = 400003
	ERROR_COMMON_TOOLS_HASH_DECODE_FAILED = 400004
)

// 路由相关错误
const (
	ERROR_ROUTER_PARSEJSON = 500001
)

// 鉴权相关错误
const (
	ERROR_AUTH_GET_TOKEN        = 600001
	ERROR_AUTH_TOKEN_INVALID    = 600002
	ERROR_AUTH_UPLOADID_INVALID = 600003
	ERROR_AUTH_UID_NOT_EXIST    = 600004
	ERROR_AUTH_NOT_PERMISSION   = 600005
)

// 文件相关错误
const (
	ERROR_FILE_COUNT_EXCEED_LIMIT   = 700001
	ERROR_FILE_PATH_INVALID         = 700002
	ERROR_FILE_NEWUSER_MKDIR        = 700003
	ERROR_FILE_NOT_EXISTS           = 700004
	ERROR_FILE_INDEX_INVALID        = 700005
	ERROR_FILE_CHUNK_MISSING        = 700006
	ERROR_FILE_STORE_PATH_INVALID   = 700007
	ERROR_FILE_NAME_INVALID         = 700008
	ERROR_FILE_SAME_NAME            = 700009
	ERROR_FILE_TARGETDIR_INVALID    = 700010
	ERROR_FILE_MOVEFILE_FAILED      = 700011
	ERROR_FILE_INVALID              = 700012
	ERROR_FILE_TARGETDIR_SAME_FILES = 700013
)

// 分享相关
const (
	ERROR_SHARE_NOT_EXIST      = 800001
	ERROR_SHARE_PASSWORD_WRONG = 800002
	ERROR_SHARE_FILE_INVALID   = 800003
	ERROR_SHARE_SAME_FILES     = 800004
)

// 下载相关
const (
	ERROR_DOWNLOAD_KEY_INVALID  = 900001
	ERROR_DOWNLOAD_FILE_INVALID = 900002
)

// 用户组相关
const (
	ERROR_GROUP_NOT_EXIST = 1000001
	ERROR_GROUP_NOT_ADMIN = 1000002
	ERROR_GROUP_DEFAULT   = 1000003
)

// 传值相关
const (
	ERROR_PARA_INVALID = 1100001
)

var errMap = map[int]string{
	ERROR_USER_NAME_LENGTH_NOT_MATCH:      "用户名称长度不符合规定",
	ERROR_USER_PASSWORD_NOT_MATCH_RULES:   "用户密码不符合规定",
	ERROR_USER_EMAIL_NOT_MATHCH_RULES:     "用户邮箱不符合规定",
	ERROR_USER_NOT_UID_AND_EMAIL:          "用户ID或用户邮箱未给出",
	ERROR_USER_MKDIR_FAILED:               "用户注册新建个人文件夹错误",
	ERROR_DB_WRITE_FAILED:                 "数据库写入错误",
	ERROR_DB_READ_FAILED:                  "数据库读入错误",
	ERROR_USER_TOEKN_INVALIED:             "token无效或非法",
	ERROR_USER_UID_PASSWORD_WRONG:         "uid/邮箱不存在或密码错误",
	ERROR_JWT_GENERATE_TOKEN_FAILED:       "token生成失败",
	ERROR_USER_EMAIL_CONFLICT:             "邮箱已被注册",
	ERROR_COMMON_SNOWFLOWS_GENERATE:       "雪花算法生成错误",
	ERROR_ROUTER_PARSEJSON:                "JSON结构体解析错误",
	ERROR_COMMON_RECAPTCHA_VERIFICATION:   "recaptcha验证码错误",
	ERROR_AUTH_GET_TOKEN:                  "无法拿到token",
	ERROR_AUTH_TOKEN_INVALID:              "token无效",
	ERROR_FILE_COUNT_EXCEED_LIMIT:         "请求文件数量超限",
	ERROR_FILE_PATH_INVALID:               "文件路径无效",
	ERROR_FILE_NEWUSER_MKDIR:              "新用户新建根目录失败",
	ERROR_FILE_NOT_EXISTS:                 "文件未存在",
	ERROR_FILE_INDEX_INVALID:              "文件分片下标非法",
	ERROR_AUTH_UPLOADID_INVALID:           "上传ID非法",
	ERROR_FILE_CHUNK_MISSING:              "文件切片缺失，请重试",
	ERROR_USER_STORAGE_EXCEED:             "用户容量不足以存储新文件",
	ERROR_FILE_STORE_PATH_INVALID:         "用户指定存储目录非法",
	ERROR_FILE_NAME_INVALID:               "文件/文件夹名非法",
	ERROR_FILE_SAME_NAME:                  "该目录下有文件/文件夹与新创建文件/文件夹同名",
	ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED: "哈希值生成失败",
	ERROR_COMMON_TOOLS_HASH_DECODE_FAILED: "哈希值解码失败",
	ERROR_FILE_TARGETDIR_INVALID:          "目标文件夹你无权访问",
	ERROR_FILE_MOVEFILE_FAILED:            "要移动的文件失效",
	ERROR_SHARE_NOT_EXIST:                 "无效的分享链接或分享已过期",
	ERROR_SHARE_PASSWORD_WRONG:            "输入的密码错误",
	ERROR_SHARE_FILE_INVALID:              "要分享的文件你无法访问",
	ERROR_FILE_INVALID:                    "文件不存在或你无法访问",
	ERROR_SHARE_SAME_FILES:                "该文件你已经分享过了",
	ERROR_FILE_TARGETDIR_SAME_FILES:       "目标文件夹有同名文件/文件夹",
	ERROR_DOWNLOAD_KEY_INVALID:            "此下载链接无效或已过期",
	ERROR_DOWNLOAD_FILE_INVALID:           "此文件你无权下载",
	ERROR_AUTH_UID_NOT_EXIST:              "操作者或被操作者的UID不存在",
	ERROR_AUTH_NOT_PERMISSION:             "你无权操作",
	ERROR_GROUP_NOT_EXIST:                 "该用户组不存在",
	ERROR_GROUP_NOT_ADMIN:                 "你没有进入管理界面的权限",
	ERROR_PARA_INVALID:                    "参数非法",
	ERROR_USER_NOT_EXIST:                  "目标用户不存在",
	ERROR_GROUP_DEFAULT:                   "默认用户组不能被删除",
}

func NewError(errorNum int) error {
	return errors.New(errMap[errorNum])
}
