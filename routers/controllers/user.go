package controllers

import (
	"Rhine-Cloud-Driver/logic/jwt"
	model "Rhine-Cloud-Driver/models"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type RegisterRequest struct {
	Name           string `json:"name"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	RecaptchaToken string `json:"recaptcha_token"`
}

type RegisterResponse struct {
	Uid uint64 `json:"uid"`
}

type LoginRequest struct {
	Uid            string `json:"uid"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	RecaptchaToken string `json:"recaptcha_token"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	var data LoginRequest
	// 从json中取数据
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 先验证recaptcha是否通过
	//if isok := common.VerifyToken(data.recaptchaToken); isok != true {
	//	makeResult(c, 200, common.NewError(common.ERROR_COMMON_RECAPTCHA_VERIFICATION), nil)
	//	return
	//}
	newUser := model.User{}
	var uid uint64
	if len(strings.Split(data.Uid, "@")) > 1 {
		data.Email = data.Uid
		uid = 0
	} else {
		uid, _ = strconv.ParseUint(data.Uid, 10, 64)
	}
	token, err := newUser.VerifyAccess("", uid, data.Email, data.Password)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	responseData := LoginResponse{token}
	makeResult(c, 200, nil, responseData)
}

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	var data RegisterRequest
	// 从json中取数据
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 先验证recaptcha是否通过
	//if isok := common.VerifyToken(data.recaptchaToken); isok != true {
	//	makeResult(c, 200, common.NewError(common.ERROR_COMMON_RECAPTCHA_VERIFICATION), nil)
	//	return
	//}

	newUser := model.User{}
	newUser.Name = data.Name
	newUser.Password = data.Password
	newUser.Email = data.Email

	err := newUser.AddUser()
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	responseData := RegisterResponse{newUser.Uid}
	makeResult(c, 200, nil, responseData)
}

type GroupDetail struct {
	GroupId   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
}

type UserDetail struct {
	Name         string      `json:"name"`
	Uid          string      `json:"uid"` // 因为前端不支持uint64，需要后端转为string来传递一下
	Email        string      `json:"email"`
	CreateTime   string      `json:"create_time"`
	UsedStorage  int64       `json:"used_storage"`
	TotalStorage int64       `json:"total_storage"`
	Group        GroupDetail `json:"group"`
}

// GetUserDetail 获取用户个人信息
func GetUserDetail(c *gin.Context) {
	// 这个时候用户登录凭借的肯定就是token了，把token拿过来校验返回数据即可
	// 因为我们用了鉴权的中间件，所以如果能走到这一步代表token是没有问题的，拿过来直接使用即可，不需要做错误判断了
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	user := model.User{}
	user.Uid = uid
	user.GetUserDetail()
	// 回传的数据有：名称、Uid、创建时间、已用容量、总容量以及用户组的信息
	// 用户组的信息暂时不管
	responseData := UserDetail{
		Name:         user.Name,
		Uid:          strconv.FormatUint(user.Uid, 10),
		Email:        user.Email,
		CreateTime:   user.CreateTime,
		UsedStorage:  user.UsedStorage,
		TotalStorage: user.TotalStorage,
	}
	makeResult(c, 200, nil, responseData)
}
