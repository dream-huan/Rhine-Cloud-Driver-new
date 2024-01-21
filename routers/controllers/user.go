package controllers

import (
	model "Rhine-Cloud-Driver/models"
	"Rhine-Cloud-Driver/pkg/jwt"
	"Rhine-Cloud-Driver/pkg/log"
	"Rhine-Cloud-Driver/pkg/recaptcha"
	"Rhine-Cloud-Driver/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	Uid string `json:"uid"`
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
	log.Logger.Info("用户登录：", zap.Any("data", data))
	// 先验证recaptcha是否通过
	if isok := recaptcha.VerifyToken(data.RecaptchaToken); isok != true {
		makeResult(c, 200, util.NewError(util.ERROR_COMMON_RECAPTCHA_VERIFICATION), nil)
		return
	}
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

type CheckRegisterResponse struct {
	OpenEnrollment bool `json:open_enrollment`
}

// CheckRegister 是否允许用户注册
func CheckRegister(c *gin.Context) {
	responseData := &CheckRegisterResponse{OpenEnrollment: true}
	if model.GetSettingByName("register_open") == "0" {
		responseData.OpenEnrollment = false
	}
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
	if isok := recaptcha.VerifyToken(data.RecaptchaToken); isok != true {
		makeResult(c, 200, util.NewError(util.ERROR_COMMON_RECAPTCHA_VERIFICATION), nil)
		return
	}

	newUser := model.User{}
	newUser.Name = data.Name
	newUser.Password = data.Password
	newUser.Email = data.Email

	err := newUser.AddUser()
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	responseData := RegisterResponse{strconv.FormatUint(newUser.Uid, 10)}
	makeResult(c, 200, nil, responseData)
}

type GroupDetail struct {
	GroupId         uint64 `json:"group_id"`
	GroupName       string `json:"group_name"`
	GroupPermission uint64 `json:"group_permission"`
	GroupStorage    uint64 `json:"group_storage"`
}

type UserDetail struct {
	Name         string `json:"name"`
	Uid          string `json:"uid"` // 因为前端不支持uint64，需要后端转为string来传递一下
	Email        string `json:"email"`
	CreateTime   string `json:"create_time"`
	UsedStorage  uint64 `json:"used_storage"`
	TotalStorage uint64 `json:"total_storage"`
	//AvatarId     string      `json:"avatar_id"`
	Group GroupDetail `json:"group"`
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
	//hash := md5.New()
	//hashValue := hex.EncodeToString(hash.Sum([]byte(user.Email)))
	responseData := UserDetail{
		Name:         user.Name,
		Uid:          strconv.FormatUint(user.Uid, 10),
		Email:        user.Email,
		CreateTime:   user.CreateTime,
		UsedStorage:  user.UsedStorage,
		TotalStorage: user.TotalStorage,
		Group: GroupDetail{
			GroupId:         user.GroupId,
			GroupName:       user.GroupName,
			GroupPermission: user.GroupPermission,
		},
	}
	//panic(common.NewError(common.ERROR_AUTH_NOT_PERMISSION))
	makeResult(c, 200, nil, responseData)
}

type ChangeUserInfoRequest struct {
	NewName     string `json:"new_name"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func ChangeUserInfo(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data ChangeUserInfoRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.ChangeUserInfo(uid, data.NewName, data.OldPassword, data.NewPassword)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	makeResult(c, 200, nil, nil)
}

func UploadAvatar(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	file, _ := c.FormFile("avatar")
	c.SaveUploadedFile(file, "./avatar/"+strconv.FormatUint(uid, 10))
	makeResult(c, 200, nil, nil)
}

func GetUserAvatar(c *gin.Context) {
	id := c.Query("id")
	c.Header("Content-Disposition", "attachment; filename=avatar.png")
	c.File("./avatar/" + id)
}

//
//func VerifyAdmin(c *gin.Context) {
//	token, _ := c.Cookie("token")
//	_, uid := jwt.TokenGetUid(token)
//	isok := model.VerifyAdmin(uid)
//	if !isok {
//		makeResult(c, 200, common.NewError(common.ERROR_GROUP_NOT_ADMIN), nil)
//		return
//	}
//	// 生成adminToken用于访问
//	jwt.GenerateToken()
//}
