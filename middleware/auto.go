package middleware

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
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
		}
		c.Next()
	}
}
