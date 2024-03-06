package util

import (
	"Rhine-Cloud-Driver/pkg/log"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/blake2b"
	"os"
)

func Mkdir(dirName string) bool {
	err := os.MkdirAll("./uploads/"+dirName, 0777)
	if err != nil {
		log.Logger.Error("新建文件夹错误", zap.String("dir_name:", dirName), zap.Error(err))
		return false
	}
	return true
}

func RemoveFile(path string) bool {
	err := os.RemoveAll(path)
	if err != nil {
		log.Logger.Error("删除文件/文件夹错误", zap.String("path:", path), zap.Error(err))
		return false
	}
	return true
}

func MkdirIfNoExist(dirName string) bool {
	//os.ReadDir()
	// todo:判断文件夹是否存在,不存在则新建
	err := os.MkdirAll("./uploads/"+dirName, 0777)
	if err != nil {
		// 初始化文件夹错误
		return false
	}
	return true
}

func WriteFile(bytes []byte, path string) bool {
	err := os.WriteFile(path, bytes, 0644)
	if err != nil {
		log.Logger.Error("写文件错误", zap.String("path:", path), zap.Error(err))
		return false
	}
	return true
}

func ReadFile(path string) ([]byte, bool) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Logger.Error("读文件错误", zap.String("path:", path), zap.Error(err))
		return nil, false
	}
	return bytes, true
}

func CreateFileIfNot(path string) bool {
	_, err := os.Stat(path)
	//fmt.Println(err)
	if err == nil {
		return true
	}
	_, err = os.Create(path)
	if err != nil {
		log.Logger.Error("创建文件错误", zap.String("path:", path), zap.Error(err))
		return false
	}
	return true
}

func GetDirAllFiles(path string) ([]os.DirEntry, bool) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Logger.Error("读取文件夹错误", zap.String("path:", path), zap.Error(err))
		return nil, false
	}
	return files, true
}

func RemoveMetaData(uploadID string) {
	if uploadID == "" {
		return
	}
	RemoveFile("./uploads/" + uploadID + "_chunks")
	RemoveFile("./uploads/metadata/metadata_" + uploadID)
}

func IsSameFile(aFilePath, bFilePath string) bool {
	// 解析两个文件的分块blake2b码
	// 取文件大小
	aFileInfo, err := os.Stat(aFilePath)
	if err != nil {
		log.Logger.Error("取对比文件失败", zap.Error(err))
		return false
	}
	bFileInfo, err := os.Stat(bFilePath)
	if err != nil {
		log.Logger.Error("取对比文件失败", zap.Error(err))
		return false
	}
	if aFileInfo.Size() != bFileInfo.Size() {
		return false
	}
	// 取大小进行切块，切为5块
	aFileBytes, _ := ReadFile(aFilePath)
	bFileBytes, _ := ReadFile(bFilePath)
	// 小于5字节的直接对比字节
	if len(aFileBytes) < 5 {
		for i := 0; i < len(aFileBytes); i++ {
			if aFileBytes[i] != bFileBytes[i] {
				return false
			}
		}
		return true
	}
	left, right := 0, 0
	chunkSize := len(aFileBytes) / 5
	lastChunkSize := len(aFileBytes) - chunkSize*4
	right += chunkSize
	for i := 0; i < 5; i++ {
		aChunkHash := blake2b.Sum256(aFileBytes[left:right])
		bChunkHash := blake2b.Sum256(bFileBytes[left:right])
		fmt.Println(aChunkHash, bChunkHash)
		if aChunkHash != bChunkHash {
			return false
		}
		left += chunkSize
		if i < 4 {
			right += chunkSize
		} else {
			right += lastChunkSize
		}
	}
	return true
}
