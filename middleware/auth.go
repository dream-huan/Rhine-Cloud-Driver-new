package middleware

import (
	model "Rhine-Cloud-Driver/models"
	"Rhine-Cloud-Driver/pkg/cache"
	"Rhine-Cloud-Driver/pkg/jwt"
	"Rhine-Cloud-Driver/pkg/util"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

func TokenVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil || !jwt.TokenValid(token) {
			if err != nil {
				err = util.NewError(util.ERROR_AUTH_GET_TOKEN)
			} else {
				err = util.NewError(util.ERROR_AUTH_TOKEN_INVALID)
			}
			c.JSON(401, util.ResponseData{Code: 1, Msg: err.Error(), Data: nil})
			c.Abort()
			return
		}
		_, uid := jwt.TokenGetUid(token)
		c.Set("uid", uid)
		if !model.PermissionVerify(uid, model.PERMISSION_ACCESS) {
			c.JSON(401, util.ResponseData{Code: 1, Msg: "您被禁止访问", Data: nil})
			c.Abort()
			return
		}
		c.Next()
	}
}

func PermissionVerify(permissionCode int) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie("token")
		_, uid := jwt.TokenGetUid(token)
		if !model.PermissionVerify(uid, permissionCode) {
			c.JSON(200, util.ResponseData{Code: 1, Msg: "您被限制访问此功能", Data: nil})
			c.Abort()
		}
		c.Next()
	}
}

var USAGE_UPLOAD = 1
var USAGE_DOWNLOAD = 2

func TaskIDVerify(usage int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 不论是什么ID，先检验有效期
		var prefix string
		var id string
		if usage == USAGE_UPLOAD {
			prefix = model.RedisPrefixUploadID
			form, _ := c.MultipartForm()
			if _, isExist := form.Value["upload_id"]; isExist {
				id = form.Value["upload_id"][0]
			} else {
				c.JSON(200, util.ResponseData{Code: 1, Msg: "请求ID非法", Data: nil})
				c.Abort()
			}
		} else if usage == USAGE_DOWNLOAD {
			prefix = model.RedisPrefixDownloadID
			id = c.Param("key")
		}
		value, isExist := cache.GetRedisKeyBytes(prefix + id)
		if isExist == false {
			c.JSON(200, util.ResponseData{Code: 1, Msg: "请求ID非法", Data: nil})
			c.Abort()
		}
		if usage == USAGE_UPLOAD {
			uploadSession := &model.UploadSession{}
			_ = json.Unmarshal(value, uploadSession)
			tempValue, _ := c.Get("uid")
			uid := tempValue.(uint64)
			if uploadSession.Uid != uid {
				c.JSON(200, util.ResponseData{Code: 1, Msg: "上传用户不一致", Data: nil})
				c.Abort()
			}
		} else if usage == USAGE_DOWNLOAD {
			fileName, fileMD5, err := model.DownloadFile(value, id)
			if err != nil {
				c.JSON(200, util.ResponseData{Code: 1, Msg: "请求ID非法", Data: nil})
				c.Abort()
			}
			c.Set("file_name", fileName)
			c.Set("md5", fileMD5)
		}
		c.Next()
	}
}
