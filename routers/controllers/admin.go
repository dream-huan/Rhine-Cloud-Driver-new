package controllers

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
	model "Rhine-Cloud-Driver/models"
	"github.com/gin-gonic/gin"
	"strconv"
)

func AdminDemo(c *gin.Context) {

}

type GetAllUserRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetAllUserResponse struct {
	Count    int64      `json:"count"`
	UserData []UserInfo `json:"user_data"`
}

type UserInfo struct {
	Uid   string `json:"uid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetAllUser(c *gin.Context) {
	var data GetAllUserRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	if data.Limit > 50 || data.Limit < 0 || data.Offset < 0 {
		makeResult(c, 200, common.NewError(common.ERROR_PARA_INVALID), nil)
	}
	count, userList := model.GetAllUser(data.Offset, data.Limit)
	users := make([]UserInfo, len(userList))
	for i := range userList {
		users[i] = UserInfo{
			Uid:   strconv.FormatUint(userList[i].Uid, 10),
			Name:  userList[i].Name,
			Email: userList[i].Email,
		}
	}
	makeResult(c, 200, nil, GetAllUserResponse{count, users})
}

type GetUserInfoRequest struct {
	Uid uint64 `json:"uid"`
}

func GetUserInfo(c *gin.Context) {
	var data GetUserInfoRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	user := model.User{Uid: data.Uid}
	user.GetUserDetail()
	makeResult(c, 200, nil, UserDetail{
		Uid:          strconv.FormatUint(user.Uid, 10),
		Email:        user.Email,
		Name:         user.Name,
		CreateTime:   user.CreateTime,
		UsedStorage:  user.UsedStorage,
		TotalStorage: user.TotalStorage,
		Group:        GroupDetail{GroupId: user.GroupId, GroupName: user.GroupName},
	})
}

type GetAllShareRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetAllShareResponse struct {
	Count     int64         `json:"count"`
	ShareData []ShareDetail `json:"share_data"`
}

func GetAllShare(c *gin.Context) {
	var data GetAllShareRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	if data.Limit > 50 || data.Limit < 0 || data.Offset < 0 {
		makeResult(c, 200, common.NewError(common.ERROR_PARA_INVALID), nil)
	}
	count, shareList := model.GetAllShare(data.Offset, data.Limit)
	shares := make([]ShareDetail, len(shareList))
	for i := range shareList {
		file, err := model.GetFileInfo(shareList[i].FileID, "all")
		if err != nil {
			makeResult(c, 200, err, nil)
			return
		}
		shareKey, err := common.HashEncode([]int{int(shareList[i].ShareID)}, 4)
		if err != nil {
			makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED), nil)
			return
		}
		shares[i] = ShareDetail{
			FileName:      file.(model.File).FileName,
			ExpireTime:    shareList[i].ExpireTime,
			ShareKey:      shareKey,
			DownloadTimes: shareList[i].DownloadTimes,
			ViewTimes:     shareList[i].ViewTimes,
			Password:      shareList[i].Password,
			IsDir:         file.(model.File).IsDir,
		}
	}
	makeResult(c, 200, nil, GetAllShareResponse{count, shares})
}

type GetAllGroupRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetAllGroupResponse struct {
	Count     int64         `json:"count"`
	GroupData []GroupDetail `json:"group_data"`
}

func GetAllGroup(c *gin.Context) {
	var data GetAllGroupRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	//if data.Limit > 50 || data.Limit < 0 || data.Offset < 0 {
	//	makeResult(c, 200, common.NewError(common.ERROR_PARA_INVALID), nil)
	//}
	count, groupList := model.GetAllGroup(data.Offset, data.Limit)
	groups := make([]GroupDetail, len(groupList))
	for i := range groupList {
		groups[i] = GroupDetail{
			GroupId:         groupList[i].GroupId,
			GroupName:       groupList[i].GroupName,
			GroupPermission: groupList[i].GroupPermission,
			GroupStorage:    groupList[i].GroupStorage,
		}
	}
	makeResult(c, 200, nil, GetAllGroupResponse{count, groups})
}

type GetAllFileRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetAllFileResponse struct {
	Count    int64        `json:"count"`
	FileData []FileSystem `json:"file_data"`
}

func GetAllFile(c *gin.Context) {
	var data GetAllFileRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	if data.Limit > 50 || data.Limit < 0 || data.Offset < 0 {
		makeResult(c, 200, common.NewError(common.ERROR_PARA_INVALID), nil)
	}
	count, fileList := model.GetAllFile(data.Offset, data.Limit)
	files := make([]FileSystem, len(fileList))
	for i := range fileList {
		files[i] = FileSystem{
			FileID:      fileList[i].FileID,
			FileName:    fileList[i].FileName,
			CreateTime:  fileList[i].CreateTime,
			FileStorage: fileList[i].FileStorage,
		}
	}
	makeResult(c, 200, nil, GetAllFileResponse{count, files})
}

type AdminGetUserDetailRequest struct {
	Uid string `json:"uid"`
}

func AdminGetUserDetail(c *gin.Context) {
	var data AdminGetUserDetailRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	uid, err := strconv.ParseUint(data.Uid, 10, 64)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_ROUTER_PARSEJSON), nil)
	}
	err, user := model.AdminGetUserDetail(uid)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	responseData := UserDetail{
		Name:         user.Name,
		Uid:          strconv.FormatUint(user.Uid, 10),
		Email:        user.Email,
		CreateTime:   user.CreateTime,
		UsedStorage:  user.UsedStorage,
		TotalStorage: user.TotalStorage,
		Group: GroupDetail{
			GroupId:   user.GroupId,
			GroupName: user.GroupName,
		},
	}
	makeResult(c, 200, nil, responseData)
}

type AdminEditUserInfoRequest struct {
	ChangedUid  string `json:"changed_uid"`
	NewName     string `json:"new_name,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
	NewStorage  string `json:"new_storage,omitempty"`
	NewGroupId  string `json:"new_group_id,omitempty"`
}

func AdminEditUserInfo(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data AdminEditUserInfoRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	changedUid, err := strconv.ParseUint(data.ChangedUid, 10, 64)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_ROUTER_PARSEJSON), nil)
		return
	}
	err = model.AdminEditUserInfo(changedUid, uid, data.NewName, data.NewPassword, data.NewGroupId, data.NewStorage)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	makeResult(c, 200, nil, nil)
}

func AdminUploadAvatar(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	file, _ := c.FormFile("avatar")
	form, _ := c.MultipartForm()
	changedUid, err := strconv.ParseUint(form.Value["uid"][0], 10, 64)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_ROUTER_PARSEJSON), nil)
		return
	}
	if !model.ChangePermissionVerify(0, 0, changedUid, uid) {
		makeResult(c, 200, common.NewError(common.ERROR_AUTH_NOT_PERMISSION), nil)
		return
	}
	c.SaveUploadedFile(file, "./avatar/"+strconv.FormatUint(changedUid, 10))
	makeResult(c, 200, nil, nil)
}

//func AdminGetFile()

type AdminCreateGroupRequest struct {
	NewGroupName       string `json:"new_group_name"`
	NewGroupPermission int64  `json:"new_group_permission"`
	NewGroupStorage    string `json:"new_group_storage"`
}

func AdminCreateGroup(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data AdminCreateGroupRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.AdminCreateGroup(uid, data.NewGroupName, data.NewGroupPermission, data.NewGroupStorage)
	makeResult(c, 200, err, nil)
}

type AdminEditGroupInfoRequest struct {
	GroupId            int64  `json:"group_id"`
	NewGroupName       string `json:"new_group_name"`
	NewGroupPermission int64  `json:"new_group_permission"`
	NewGroupStorage    string `json:"new_group_storage"`
}

func AdminEditGroupInfo(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data AdminEditGroupInfoRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.AdminEditGroupInfo(uid, uint64(data.GroupId), data.NewGroupName, data.NewGroupPermission, data.NewGroupStorage)
	makeResult(c, 200, err, nil)
}

type AdminDeleteGroupRequest struct {
	GroupId int64 `json:"group_id"`
}

func AdminDeleteGroup(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data AdminDeleteGroupRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.AdminDeleteGroup(uid, uint64(data.GroupId))
	makeResult(c, 200, err, nil)

}

type VerifyAdminRequest struct {
	Password       string `json:"password"`
	RecaptchaToken string `json:"recaptcha_token"`
}

func VerifyAdmin(c *gin.Context) {
	// 查询group权限
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data VerifyAdminRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 先验证recaptcha是否通过
	if isok := common.VerifyToken(data.RecaptchaToken); isok != true {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_RECAPTCHA_VERIFICATION), nil)
		return
	}
	// 验证密码
	user := model.User{Uid: uid}
	user.GetUserDetail()
	if !user.VerifyPassword(data.Password) {
		makeResult(c, 200, common.NewError(common.ERROR_USER_UID_PASSWORD_WRONG), nil)
		return
	}
	if model.PermissionVerify(uid, model.PERMISSION_ADMIN_WRITE) {
		makeResult(c, 200, nil, 2)
		return
	}
	makeResult(c, 200, nil, 1)
}

func GetActiveUserData(c *gin.Context) {

}

func GetUploadFileData(c *gin.Context) {

}
