package util

import (
	"Rhine-Cloud-Driver/pkg/log"
	"go.uber.org/zap"
	"os"
)

func Mkdir(dirName string) bool {
	err := os.MkdirAll("rhine-cloud-driver/uploads/"+dirName, 0777)
	if err != nil {
		log.Logger.Error("新建文件夹错误", zap.Error(err))
		return false
	}
	return true
}

func RemoveFile(path string) bool {
	err := os.RemoveAll(path)
	if err != nil {
		log.Logger.Error("删除文件错误", zap.Error(err))
		return false
	}
	return true
}

func RemoveDir() {

}
