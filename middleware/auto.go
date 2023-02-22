package middleware

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
	model "Rhine-Cloud-Driver/models"
	"github.com/gin-gonic/gin"
)

func TokenVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil || !jwt.TokenValid(token) {
			if err != nil {
				err = common.NewError(common.ERROR_AUTH_GET_TOKEN)
			} else {
				err = common.NewError(common.ERROR_AUTH_TOKEN_INVALID)
			}
			c.JSON(401, common.ResponseData{Code: 1, Msg: err.Error(), Data: nil})
			c.Abort()
			return
		}
		_, uid := jwt.TokenGetUid(token)
		if !model.PermissionVerify(uid, model.PERMISSION_ACCESS) {
			c.JSON(401, common.ResponseData{Code: 1, Msg: "您被禁止访问", Data: nil})
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
			c.JSON(200, common.ResponseData{Code: 1, Msg: "您被限制访问此功能", Data: nil})
			c.Abort()
		}
		c.Next()
	}
}
