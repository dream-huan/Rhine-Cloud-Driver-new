package controllers

import (
	"Rhine-Cloud-Driver/common"
	"github.com/gin-gonic/gin"
)

func makeResult(c *gin.Context, httpCode int, err error, data interface{}) {
	if err == nil {
		c.JSON(httpCode, common.ResponseData{Code: 0, Data: data})
	} else {
		c.JSON(httpCode, common.ResponseData{Code: 1, Msg: err.Error(), Data: data})
	}
}
