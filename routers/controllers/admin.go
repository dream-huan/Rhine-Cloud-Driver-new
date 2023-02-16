package controllers

import (
	"Rhine-Cloud-Driver/common"
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

func EditUserInfo(c *gin.Context) {

}

func GetAllFile(c *gin.Context) {

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
