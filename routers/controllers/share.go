package controllers

import (
	"Rhine-Cloud-Driver/common"
	"github.com/gin-gonic/gin"
)

type GetShareDetailRequest struct {
	ShareKey      string `json:"share_key"`
	SharePassword string `json:"share_password,omitempty"`
}

type GetShareDetailResponse struct {
	FileID    uint64 `json:"file_id,omitempty"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func GetShareDetail(c *gin.Context) {
	var data GetShareDetailRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 从shareKey还原出shareID
	common.HashDecode(data.ShareKey)
	// 数据库查找出ID对应的用户和分享ID

	// 构建文件系统（如果密码正确或分享无密码，其他情况不构建），并拿到用户的名称和邮箱

	// 对邮箱进行MD5加密后返回给用户
}
