package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/redis"
	"strconv"
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
	CreateTime  string `json:"create_time,omitempty"`
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
func CheckPathValid(uid uint64, path string) (bool, uint64) {
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
	var fileID uint64
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

func BuildFileSystem(uid uint64, path string, limit, offset int) (count int64, dirFileID uint64, files []File, err error) {
	err = nil
	// 判断路径结果是否合法
	isValid, fileID := CheckPathValid(uid, path)
	if isValid == false {
		return 0, 0, nil, common.NewError(common.ERROR_FILE_PATH_INVALID)
	}
	// 分页查询，每次查询最多50条
	// 结果存储到redis中
	if limit > 50 || offset < 0 {
		return 0, 0, nil, common.NewError(common.ERROR_FILE_COUNT_EXCEED_LIMIT)
	}
	dirFileID = fileID
	DB.Table("files").Where("parent_id=?", fileID).Count(&count)
	DB.Table("files").Where("parent_id=?", fileID).Offset(offset).Limit(limit).Find(&files)
	return
}

func UploadPrepare(md5 string, chunkNum int64, uid uint64, fileSize uint64) (bool, string, string, error) {
	// 校验容量是否充足
	var nowUser User
	DB.Table("users").Where("uid=?", uid).Find(&nowUser)
	if nowUser.UsedStorage+fileSize > nowUser.TotalStorage {
		return false, "", "", common.NewError(common.ERROR_USER_STORAGE_EXCEED)
	}
	// 随机生成一个32位新的key，并将其作为UploadID
	uploadID := common.RandStringRunes(32)
	for redis.GetRedisKey(uploadID) != nil {
		uploadID = common.RandStringRunes(32)
	}
	redis.SetRedisKey(uploadID, md5, time.Second*60*30)
	var count int64
	DB.Table("files").Where("md5=?", md5).Count(&count)
	if count > 0 {
		return true, "", "", nil
	}
	// 先判断是否存在该md5
	chunksRedisKey := "file_md5_chunks_" + md5
	chunkNumRedisKey := "file_md5_chunk_num_" + md5
	if oldChunkNum := redis.GetRedisKey(chunkNumRedisKey); oldChunkNum != nil {
		size, _ := strconv.ParseInt(oldChunkNum.(string), 10, 64)
		chunks := make([]byte, size)
		for i := int64(0); i < size; i++ {
			if redis.GetRedisKeyBitmap(chunksRedisKey, i) == 1 {
				chunks[i] = '1'
			} else {
				chunks[i] = '0'
			}
		}
		redis.RenewRedisKey(chunksRedisKey, time.Second*60*60*24)
		redis.RenewRedisKey(chunkNumRedisKey, time.Second*60*60*24)
		return false, string(chunks), uploadID, nil
	}
	// 向内存中注册块数
	// 超过24小时上传未成功作废
	chunks := make([]byte, chunkNum)
	for i := int64(0); i < chunkNum; i++ {
		chunks[i] = '0'
	}
	redis.RenewRedisKey(chunksRedisKey, time.Second*60*60*24)
	redis.SetRedisKey(chunkNumRedisKey, chunkNum, time.Second*60*60*24)
	return false, string(chunks), uploadID, nil
}

func DealFileChunk(md5 string, fileIndex int64, uploadID string) (isExist bool, chunkNum int64, err error) {
	if redis.GetRedisKey(uploadID) == nil {
		return false, 0, common.NewError(common.ERROR_AUTH_UPLOADID_INVALID)
	}
	chunksRedisKey := "file_md5_chunks_" + md5
	chunkNumRedisKey := "file_md5_chunk_num_" + md5
	// 从redis中拿到这块的情况
	if tempRedisValue := redis.GetRedisKey(chunkNumRedisKey); tempRedisValue == nil {
		return false, chunkNum, common.NewError(common.ERROR_FILE_NOT_EXISTS)
	} else {
		chunkNum, _ = strconv.ParseInt(tempRedisValue.(string), 10, 64)
	}
	if fileIndex < 0 || fileIndex >= chunkNum {
		return false, chunkNum, common.NewError(common.ERROR_FILE_INDEX_INVALID)
	}
	if redis.GetRedisKeyBitmap(chunksRedisKey, fileIndex) == 0 {
		redis.SetRedisKeyBitmap(chunksRedisKey, fileIndex, 1, time.Second*60*60*24)
		return false, chunkNum, nil
	}
	// 该部分已经被其他完成
	return true, chunkNum, nil
}

func MergeFileChunks(md5, uploadID string) (int64, error) {
	if redis.GetRedisKey(uploadID) == nil {
		return 0, common.NewError(common.ERROR_AUTH_UPLOADID_INVALID)
	}
	// 查看数据库是否有该MD5
	var count int64
	DB.Table("files").Where("md5=?", md5).Count(&count)
	if count > 0 {
		return 0, nil
	}
	chunksRedisKey := "file_md5_chunks_" + md5
	chunkNumRedisKey := "file_md5_chunk_num_" + md5
	var chunkNum int64
	if tempRedisValue := redis.GetRedisKey(chunkNumRedisKey); tempRedisValue == nil {
		return 0, common.NewError(common.ERROR_FILE_NOT_EXISTS)
	} else {
		chunkNum, _ = strconv.ParseInt(tempRedisValue.(string), 10, 64)
	}
	hasFinishedChunk := redis.CountRedisKeyBitmap(chunksRedisKey, 0, chunkNum-1)
	if hasFinishedChunk == chunkNum {
		// 合并并删除redis的记录，包括uploadID...
		redis.DelRedisKey(chunksRedisKey)
		redis.DelRedisKey(uploadID)
		redis.DelRedisKey(chunkNumRedisKey)
		return chunkNum, nil
	}
	return 0, common.NewError(common.ERROR_FILE_CHUNK_MISSING)
}

func AddFile(uid uint64, md5 string, fileName string, fileSize, parentID uint64) error {
	// 校验容量是否充足
	var nowUser User
	DB.Table("users").Where("uid=?", uid).Find(&nowUser)
	if nowUser.UsedStorage+fileSize > nowUser.TotalStorage {
		return common.NewError(common.ERROR_USER_STORAGE_EXCEED)
	}
	// 校验parentID是否属于该UID本人和是否存在
	var fileDir File
	err := DB.Table("files").Select("uid").Where("file_id=? and is_dir=1", parentID).First(&fileDir).Error
	if err != nil || fileDir.Uid != uid {
		return common.NewError(common.ERROR_FILE_STORE_PATH_INVALID)
	}
	err = DB.Table("files").Create(&File{
		Uid:         uid,
		MD5:         md5,
		FileName:    fileName,
		FileStorage: fileSize,
		ParentID:    parentID,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func Mkdir(uid uint64, fileName string, parentID uint64) {

}
