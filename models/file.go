package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/redis"
	"fmt"
	"time"
)

type File struct {
	FileID      uint64 `json:"file_id" gorm:"primaryKey;auto_increment"`
	Uid         uint64 `json:"uid,omitempty"`
	FileName    string `json:"file_name,omitempty" gorm:"size:255;index:idx_file_name"`
	MD5         string `json:"md5,omitempty" gorm:"index:idx_md5"`
	Path        string `json:"path,omitempty"`
	FileStorage uint64 `json:"file_storage,omitempty"`
	ParentID    uint64 `json:"parent_id,omitempty" gorm:"index:idx_parent_id"`
	OriginID    uint64 `json:"origin_id,omitempty"`
	Valid       bool   `json:"valid,omitempty"`
	IsDir       bool   `json:"is_dir,omitempty"`
}

var invalidChar = map[string]bool{
	"?":  true,
	"/":  true,
	"*":  true,
	"\"": true,
	"|":  true,
	":":  true,
	"<":  true,
	">":  true,
	"\\": true,
}

// 检查规则
// 1.不能包含特殊字符如:/\*?<>:|"
// 2.单个路径名长度不能超过255
// 3./与/之间不能为空
// 若合法，err置为空，返回路径切片。若不合法，返回err和空切片
func CheckPathValid(uid uint64, path string) (bool, int64) {
	pathName := ""
	// path类似于/uploads/study/20230102/
	// path首位和末尾非/自动补/
	if path[0] != '/' {
		path = "/" + path
	}
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	// 边检查边在数据库中进行检索
	// 拿到该用户的根目录的fileID
	var fileID int64
	DB.Table("files").Select("file_id").Where("uid=? and parent_id=?", uid, 0).Find(&fileID)
	if fileID == 0 {
		//此路径不存在
		return false, 0
	}
	for i := range []rune(path) {
		// 首位肯定是/无需校验
		if i == 0 {
			continue
		}
		if path[i] == '/' {
			if pathName == "" || len(pathName) > 255 {
				return false, 0
			}
			fmt.Println(pathName)
			err := DB.Table("files").Select("file_id").Where("parent_id=? and file_name=? and is_dir=true and valid=true", fileID, pathName).First(&fileID).Error
			if err != nil || fileID == 0 {
				return false, 0
			}
			pathName = ""
		} else {
			if invalidChar[string(path[i])] == true {
				return false, 0
			}
			pathName = pathName + string(path[i])
		}
	}
	return true, fileID
}

func BuildFileSystem(uid uint64, path string, limit, offset int) (count int64, files []File, err error) {
	err = nil
	// 判断路径结果是否合法
	isValid, fileID := CheckPathValid(uid, path)
	if isValid == false {
		return 0, nil, common.NewError(common.ERROR_FILE_PATH_INVALID)
	}
	// 分页查询，每次查询最多50条
	// 结果存储到redis中
	if limit > 50 || offset < 0 {
		return 0, nil, common.NewError(common.ERROR_FILE_COUNT_EXCEED_LIMIT)
	}
	DB.Table("files").Where("parent_id=?", fileID).Count(&count)
	DB.Table("files").Where("parent_id=?", fileID).Offset(offset).Limit(limit).Find(&files)
	return
}

func UploadPerpare(md5 string, chunkNum int64) (isExists bool) {
	var count int64
	DB.Table("files").Where("md5=?", md5).Count(&count)
	if count > 0 {
		return true
	}
	// 向内存中注册块数
	// 超过24小时上传未成功作废
	redis.SetRedisKey(md5, chunkNum, time.Second*60*60*24)
	return false
}
