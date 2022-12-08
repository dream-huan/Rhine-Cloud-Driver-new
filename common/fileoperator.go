package common

import (
	"Rhine-Cloud-Driver/logic/log"
	"go.uber.org/zap"
	"os"
)

func Mkdir(dirName string) bool {
	err := os.Mkdir("/rhine-cloud-driver/uploads/"+dirName, 0777)
	if err != nil {
		log.Logger.Error("新建文件夹错误", zap.Error(err))
		return false
	}
	return true
}

func RemoveFile() {

}

func RemoveDir() {

}
