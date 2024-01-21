package controllers

import (
	"Rhine-Cloud-Driver/pkg/util"
	"github.com/gin-gonic/gin"
)

func makeResult(c *gin.Context, httpCode int, err error, data interface{}) {
	if err == nil {
		c.JSON(httpCode, util.ResponseData{Code: 0, Data: data})
	} else {
		c.JSON(httpCode, util.ResponseData{Code: 1, Msg: err.Error(), Data: data})
	}
}
