package model

import (
	"Rhine-Cloud-Driver/pkg/cache"
	"Rhine-Cloud-Driver/pkg/util"
	"gorm.io/gorm"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type File struct {
	CreateTime  string `json:"create_time,omitempty"`
	FileID      uint64 `json:"file_id" gorm:"primaryKey;auto_increment"`
	FileName    string `json:"file_name,omitempty" gorm:"size:255"`
	FileStorage uint64 `json:"file_storage,omitempty"`
	IsDir       bool   `json:"is_dir,omitempty"`
	IsOrigin    bool   `json:"is_origin,omitempty"`
	MD5         string `json:"md5,omitempty" gorm:"index:idx_md5"`
	ExtraMD5    string `json:"extra_md5,omitempty" gorm:"index:idx_extra_md5"`
	ParentID    uint64 `json:"parent_id,omitempty" gorm:"index:idx_parent_id"`
	Path        string `json:"path,omitempty" gorm:"index:idx_path"`
	//QuickAccess bool   `json:"quick_access"`
	Uid   uint64 `json:"uid,omitempty"`
	Valid bool   `json:"valid,omitempty"`
	Type  string `json:"type,omitempty" gorm:"index:idx_type;size:6"`
}

var invalidChar = map[string]bool{
	"?":  true,
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
//func CheckPathValid(uid uint64, path string, parentID uint64) (bool, uint64) {
//	pathName := ""
//	// path类似于/uploads/study/20230102/
//	// path首位和末尾非/自动补/
//	if path[0] != '/' {
//		path = "/" + path
//	}
//	if path[len(path)-1] != '/' {
//		path = path + "/"
//	}
//	// 边检查边在数据库中进行检索
//	// 拿到该用户的根目录的fileID
//	var file File
//	var lastFileID uint64
//	err := DB.Table("files").Where("uid=? and parent_id=?", uid, parentID).First(&file).Error
//	if err != nil || file.FileID == 0 {
//		//此路径不存在
//		return false, 0
//	}
//	lastFileID = file.FileID
//	for i, v := range []rune(path) {
//		// 首位肯定是/无需校验
//		if i == 0 {
//			continue
//		}
//		if v == '/' {
//			if pathName == "" || len(pathName) > 255 {
//				return false, 0
//			}
//			file = File{}
//			err = DB.Table("files").Where("parent_id=? and file_name=? and is_dir=true and valid=true", lastFileID, pathName).First(&file).Error
//			if err != nil || file.FileID == 0 {
//				return false, 0
//			}
//			pathName = ""
//			lastFileID = file.FileID
//		} else {
//			if invalidChar[string(v)] == true {
//				return false, 0
//			}
//			pathName = pathName + string(v)
//		}
//	}
//	return true, file.FileID
//}

func IsSameFile(md5A, extraMD5A, md5B, extraMD5B string, storageA, storageB uint64) bool {
	return md5A == md5B && extraMD5A == extraMD5B && storageA == storageB
}

func CheckPathValid(uid uint64, path, previousPath string) (bool, uint64) {
	if path[0] != '/' {
		path = "/" + path
	}
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	// 找它的上一层
	// 特殊情况特殊处理
	dirName := ""
	if path != "/" {
		splitValue := strings.Split(path, "/")
		for i := range splitValue {
			if i == len(splitValue)-2 {
				dirName = splitValue[i]
				break
			}
			previousPath += splitValue[i] + "/"
		}
	}
	file := File{}
	err := DB.Table("files").Where("uid = ? and path = ? and file_name = ? and valid = true ", uid, previousPath, dirName).First(&file).Error
	if err != nil || file.FileID == 0 {
		return false, 0
	}
	return true, file.FileID
}

func BuildFileSystem(uid uint64, path string, targetDirId uint64, limit, offset int, filterKey string, filterType []string) (count int64, dirFileID uint64, files []File, filePath string, err error) {
	if targetDirId != 0 {
		// 验证此id是否所属操作者本人
		DB.Table("files").Where("parent_id = ? and valid = true", targetDirId).Find(&files)
		if len(files) == 0 {
			// 要额外验证一次
			DB.Table("files").Where("file_id = ? and valid = true", targetDirId).Find(&files)
			if len(files) == 0 {
				// 文件夹不存在，非法操作
				return 0, 0, nil, "", util.NewError(util.ERROR_FILE_NOT_EXISTS)
			}
		}
		if files[0].Uid != uid {
			// 非法操作
			return 0, 0, nil, "", util.NewError(util.ERROR_FILE_INVALID)
		}
		if files[0].FileID == targetDirId {
			filePath = files[0].Path + files[0].FileName + "/"
			files = []File{}
		} else {
			filePath = files[0].Path
		}
		return int64(len(files)), targetDirId, files, filePath, nil
	}
	if filterKey != "" || len(filterType) != 0 {
		if filterKey != "" && len(filterType) != 0 {
			filterKey = "%" + filterKey + "%"
			DB.Table("files").Where("file_name like ? and type in ? and uid = ? and valid = true", filterKey, filterType, uid).Find(&files)
			count = int64(len(files))
		} else if filterKey != "" {
			filterKey = "%" + filterKey + "%"
			DB.Table("files").Where("file_name like ? and uid = ? and valid = true", filterKey, uid).Find(&files)
			count = int64(len(files))
		} else {
			DB.Table("files").Where("type in ? and uid = ? and valid = true", filterType, uid).Find(&files)
			count = int64(len(files))
		}
		return
	}
	err = nil
	// 判断路径结果是否合法
	isValid, fileID := CheckPathValid(uid, path, "")
	if isValid == false {
		return 0, 0, nil, "", util.NewError(util.ERROR_FILE_PATH_INVALID)
	}
	// 分页查询，每次查询最多50条
	// 结果存储到redis中
	if limit > 50 || offset < 0 {
		return 0, 0, nil, "", util.NewError(util.ERROR_FILE_COUNT_EXCEED_LIMIT)
	}
	dirFileID = fileID
	//DB.Table("files").Where("parent_id = ? and valid = true", fileID).Count(&count)
	//DB.Table("files").Where("parent_id = ? and valid = true", fileID).Offset(offset).Limit(limit).Find(&files)
	DB.Table("files").Where("parent_id = ? and valid = true", fileID).Find(&files)
	count = int64(len(files))
	return
}

func UploadPrepare(md5, extraMD5, fileName string, chunkNum int64, uid, fileSize, targetDirID uint64) (bool, string, string, error) {
	// 校验容量是否充足
	var nowUser User
	DB.Table("users").Where("uid=?", uid).Find(&nowUser)
	if nowUser.UsedStorage+fileSize > nowUser.TotalStorage {
		return false, "", "", util.NewError(util.ERROR_USER_STORAGE_EXCEED)
	}
	// 随机生成一个32位新的key，并将其作为UploadID
	uploadID := util.RandStringRunes(32)
	_, isExist := cache.GetRedisKey("upload_id_" + uploadID)
	for isExist != false {
		uploadID = util.RandStringRunes(32)
		_, isExist = cache.GetRedisKey("upload_id_" + uploadID)
	}
	cache.SetRedisKey("upload_id_"+uploadID, md5, time.Second*60*30)
	var files []File
	DB.Table("files").Where("md5=?", md5).Find(&files)
	if len(files) > 0 {
		for i := 0; i < len(files); i++ {
			if IsSameFile(files[i].MD5, files[i].ExtraMD5, md5, extraMD5, files[i].FileStorage, fileSize) {
				// 存在AddFile即可
				err := AddFile(uid, md5, extraMD5, fileName, fileSize, targetDirID, false)
				if err != nil {
					return true, "", "", err
				}
				return true, "", "", nil
			}
		}
	}
	// 先判断是否存在该md5
	chunksRedisKey := "file_md5_chunks_" + md5
	chunkNumRedisKey := "file_md5_chunk_num_" + md5
	if oldChunkNum, isExist := cache.GetRedisKey(chunkNumRedisKey); isExist != false {
		size, _ := strconv.ParseInt(oldChunkNum.(string), 10, 64)
		chunks := make([]byte, size)
		for i := int64(0); i < size; i++ {
			if cache.GetRedisKeyBitmap(chunksRedisKey, i) == 1 {
				chunks[i] = '1'
			} else {
				chunks[i] = '0'
			}
		}
		cache.RenewRedisKey(chunksRedisKey, time.Second*60*60*24)
		cache.RenewRedisKey(chunkNumRedisKey, time.Second*60*60*24)
		return false, string(chunks), uploadID, nil
	}
	// 向内存中注册块数
	// 超过24小时上传未成功作废
	chunks := make([]byte, chunkNum)
	for i := int64(0); i < chunkNum; i++ {
		chunks[i] = '0'
	}
	cache.RenewRedisKey(chunksRedisKey, time.Second*60*60*24)
	cache.SetRedisKey(chunkNumRedisKey, chunkNum, time.Second*60*60*24)
	return false, string(chunks), uploadID, nil
}

func DealFileChunk(md5 string, fileIndex int64, uploadID string) (isExist bool, chunkNum int64, err error) {
	if _, isExist := cache.GetRedisKey("upload_id_" + uploadID); isExist == false {
		return false, 0, util.NewError(util.ERROR_AUTH_UPLOADID_INVALID)
	}
	chunksRedisKey := "file_md5_chunks_" + md5
	chunkNumRedisKey := "file_md5_chunk_num_" + md5
	// 从redis中拿到这块的情况
	if tempRedisValue, isExist := cache.GetRedisKey(chunkNumRedisKey); isExist == false {
		return false, chunkNum, util.NewError(util.ERROR_FILE_NOT_EXISTS)
	} else {
		chunkNum, _ = strconv.ParseInt(tempRedisValue.(string), 10, 64)
	}
	if fileIndex < 0 || fileIndex >= chunkNum {
		return false, chunkNum, util.NewError(util.ERROR_FILE_INDEX_INVALID)
	}
	if cache.GetRedisKeyBitmap(chunksRedisKey, fileIndex) == 0 {
		cache.SetRedisKeyBitmap(chunksRedisKey, fileIndex, 1, time.Second*60*60*24)
		return false, chunkNum, nil
	}
	// 该部分已经被其他完成
	return true, chunkNum, nil
}

func MergeFileChunks(md5, uploadID string) (int64, error) {
	if _, isExist := cache.GetRedisKey("upload_id_" + uploadID); isExist == false {
		return 0, util.NewError(util.ERROR_AUTH_UPLOADID_INVALID)
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
	if tempRedisValue, isExist := cache.GetRedisKey(chunkNumRedisKey); isExist == false {
		return 0, util.NewError(util.ERROR_FILE_NOT_EXISTS)
	} else {
		chunkNum, _ = strconv.ParseInt(tempRedisValue.(string), 10, 64)
	}
	hasFinishedChunk := cache.CountRedisKeyBitmap(chunksRedisKey, 0, chunkNum-1)
	if hasFinishedChunk == chunkNum {
		// 合并并删除redis的记录，包括uploadID...
		cache.DelRedisKey(chunksRedisKey)
		cache.DelRedisKey(uploadID)
		cache.DelRedisKey(chunkNumRedisKey)
		return chunkNum, nil
	}
	return 0, util.NewError(util.ERROR_FILE_CHUNK_MISSING)
}

func AddFile(uid uint64, md5, extraMD5 string, fileName string, fileSize, parentID uint64, isOrigin bool) error {
	// 校验容量是否充足
	var nowUser User
	DB.Table("users").Where("uid=?", uid).Find(&nowUser)
	if nowUser.UsedStorage+fileSize > nowUser.TotalStorage {
		return util.NewError(util.ERROR_USER_STORAGE_EXCEED)
	}
	// 校验parentID是否属于该UID本人和是否存在
	var fileDir File
	err := DB.Table("files").Where("file_id=? and is_dir=1", parentID).First(&fileDir).Error
	if err != nil || fileDir.Uid != uid {
		return util.NewError(util.ERROR_FILE_STORE_PATH_INVALID)
	}
	// 开启事务来增加
	tx := DB.Begin()
	// 不允许同一目录有相同文件名的文件
	var count int64
	tx.Table("files").Where("file_name=? and parent_id=? and is_dir=false and valid=true", fileName, parentID).Count(&count)
	if count > 0 {
		tx.Rollback()
		return util.NewError(util.ERROR_FILE_SAME_NAME)
	}
	tempSlice := strings.Split(fileName, ".")
	fileType := ""
	if len(tempSlice) > 0 {
		fileType = tempSlice[len(tempSlice)-1]
		if len(fileType) > 6 {
			fileType = ""
		}
	}
	err = tx.Table("files").Create(&File{
		Uid:         uid,
		MD5:         md5,
		ExtraMD5:    extraMD5,
		FileName:    fileName,
		FileStorage: fileSize,
		ParentID:    parentID,
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		Valid:       true,
		IsDir:       false,
		IsOrigin:    isOrigin,
		Path:        fileDir.Path + fileDir.FileName + "/",
		Type:        fileType,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	// 增加容量
	err = tx.Table("users").Where("uid=?", uid).Update("used_storage", gorm.Expr("used_storage+?", fileSize)).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func Mkdir(uid uint64, fileName string, parentID uint64) error {
	// 检验文件夹名称是否非法
	if fileName == "" {
		return util.NewError(util.ERROR_FILE_NAME_INVALID)
	}
	matched, err := regexp.MatchString("[\\/+?:*<>!|]", fileName)
	if err != nil || matched == true {
		return util.NewError(util.ERROR_FILE_NAME_INVALID)
	}
	// 简验parentID是否属于该UID，并且该ID的is_dir和valid为true
	var targetDir File
	err = DB.Table("files").Where("file_id=?", parentID).First(&targetDir).Error
	if err != nil || targetDir.Uid != uid || targetDir.IsDir == false || targetDir.Valid == false {
		return util.NewError(util.ERROR_FILE_STORE_PATH_INVALID)
	}
	// 同名不允许在同一目录
	var count int64
	err = DB.Table("files").Where("file_name=? and parent_id=? and valid=true and is_dir=true", fileName, parentID).Count(&count).Error
	if err != nil || count > 0 {
		return util.NewError(util.ERROR_FILE_SAME_NAME)
	}
	err = DB.Table("files").Create(&File{
		Uid:        uid,
		FileName:   fileName,
		ParentID:   parentID,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		IsDir:      true,
		IsOrigin:   true,
		Valid:      true,
		Path:       targetDir.Path + targetDir.FileName + "/",
	}).Error
	if err != nil {
		return util.NewError(util.ERROR_DB_WRITE_FAILED)
	}
	return nil
}

func RemoveFiles(uid uint64, fileID []uint64) error {
	// 验证该文件是否所属该用户
	var targetFile []File
	err := DB.Table("files").Where("file_id in ?", fileID).Find(&targetFile).Error
	if err != nil {
		// 文件不存在
		return util.NewError(util.ERROR_FILE_NOT_EXISTS)
	}
	tx := DB.Begin()
	for _, v := range targetFile {
		if v.FileID == 0 || v.Uid != uid || v.Valid == false {
			tx.Rollback()
			return util.NewError(util.ERROR_FILE_INVALID)
		}
		if v.IsDir == true {
			var subFiles []File
			tx.Table("files").Where("path like ? and valid = true", v.Path+v.FileName+"/%").Find(&subFiles)
			for _, value := range subFiles {
				tx.Table("files").Where("file_id = ?", value.FileID).Update("valid", 0)
				tx.Table("users").Where("uid = ?", uid).Update("used_storage", gorm.Expr("used_storage - ?", value.FileStorage))
				tx.Table("shares").Where("file_id = ? and valid = true and (now()<expire_time or expire_time='-')", value.FileID).Update("valid", 0)
			}
		}
		// 验证文件是否正被分享，如果是，删除该分享
		tx.Table("shares").Where("file_id = ? and valid = true and (now()<expire_time or expire_time='-')", v.FileID).Update("valid", 0)
		// 恢复用户的所属空间
		tx.Table("users").Where("uid = ?", uid).Update("used_storage", gorm.Expr("used_storage - ?", v.FileStorage))
		// 将valid置为0
		tx.Table("files").Where("file_id = ?", v.FileID).Update("valid", 0)
		// 给kafka传递信息，是否可以将该文件删除
	}
	tx.Commit()
	return nil
}

func MoveFiles(uid uint64, moveFiles []uint64, targetDirID uint64) error {
	tx := DB.Begin()
	// 验证移动的文件和目标文件夹ID属于该用户，并且该文件夹下与要移动的文件无重名
	var targetDir File
	err := tx.Table("files").Where("file_id=? and uid=? and is_dir=true and valid=true", targetDirID, uid).Find(&targetDir).Error
	if err != nil || targetDir.FileID == 0 {
		tx.Rollback()
		return util.NewError(util.ERROR_FILE_TARGETDIR_INVALID)
	}
	var oldPathName string
	var newPathName string
	for _, v := range moveFiles {
		count := int64(0)
		var file File
		err := tx.Table("files").Where("file_id=?", v).First(&file).Error
		if err != nil || file.Uid != uid {
			tx.Rollback()
			return util.NewError(util.ERROR_FILE_MOVEFILE_FAILED)
		}
		tx.Table("files").Where("parent_id=? and file_name=? and is_dir=? and valid=true", targetDirID, file.FileName, file.IsDir).Count(&count)
		if count > 0 {
			tx.Rollback()
			return util.NewError(util.ERROR_FILE_TARGETDIR_SAME_FILES)
		}
		if file.IsDir == true {
			// 对子文件和子文件夹更改path
			oldPathName = file.Path + file.FileName + "/"
			var subFiles []File
			tx.Table("files").Where("uid=? and path like ? and valid=true", uid, oldPathName+"%").Find(&subFiles)
			oldPathName = file.Path
			newPathName = targetDir.Path + targetDir.FileName + "/"
			for i := range subFiles {
				subFiles[i].Path = strings.Replace(subFiles[i].Path, oldPathName, newPathName, 1)
				err = tx.Table("files").Where("file_id = ?", subFiles[i].FileID).Update("path", subFiles[i].Path).Error
				if err != nil {
					tx.Rollback()
					return util.NewError(util.ERROR_FILE_MOVEFILE_FAILED)
				}
			}
		}
		// 更改当前文件的path
		file.Path = targetDir.Path + targetDir.FileName + "/"
		err = tx.Table("files").Where("file_id = ?", file.FileID).Update("path", file.Path).Error
		if err != nil {
			tx.Rollback()
			return util.NewError(util.ERROR_FILE_MOVEFILE_FAILED)
		}
		// 更改parent_id
		err = tx.Table("files").Where("file_id = ?", v).Update("parent_id", targetDirID).Error
		if err != nil {
			tx.Rollback()
			return util.NewError(util.ERROR_FILE_MOVEFILE_FAILED)
		}
	}
	tx.Commit()
	return nil
}

func GetDownloadKey(uid, fileID uint64, fileKey string) (downloadID string, err error) {
	// 验证是否是本人的文件
	var file File
	err = DB.Table("files").Where("file_id = ? and valid = true and is_dir = false", fileID).Find(&file).Error
	if !PermissionVerify(uid, PERMISSION_ADMIN_READ) && (err != nil || file.Uid != uid) {
		return "", util.NewError(util.ERROR_DOWNLOAD_FILE_INVALID)
	}
	downloadID = util.RandStringRunes(6) + fileKey
	cache.SetRedisKey("download_key_"+downloadID, strconv.FormatInt(int64(file.FileID), 10)+":"+file.FileName+":"+file.MD5, time.Hour/2)
	return downloadID, nil
}

func DownloadFile(key string, fileID uint64) (fileName, fileMD5 string, err error) {
	err = nil
	fileInfo, isExist := cache.GetRedisKey("download_key_" + key)
	if isExist == false {
		// 链接无效或已过期
		return "", "", util.NewError(util.ERROR_DOWNLOAD_KEY_INVALID)
	}
	tempSlice := strings.Split(fileInfo.(string), ":")
	if tempSlice[0] != strconv.FormatUint(fileID, 10) {
		return "", "", util.NewError(util.ERROR_DOWNLOAD_KEY_INVALID)
	}
	fileName = tempSlice[1]
	fileMD5 = tempSlice[2]
	return
}

func GetFileInfo(fileID uint64, info string) (interface{}, error) {
	file := File{}
	err := DB.Table("files").Where("file_id=?", fileID).Find(&file).Error
	if err != nil {
		return nil, util.NewError(util.ERROR_FILE_NOT_EXISTS)
	}
	switch info {
	case "all":
		return file, nil
	case "create_time":
		return file.CreateTime, nil
	case "file_name":
		return file.FileName, nil
	case "file_storage":
		return file.FileStorage, nil
	case "is_dir":
		return file.IsDir, nil
	case "md5":
		return file.MD5, nil
	case "parent_id":
		return file.ParentID, nil
	case "uid":
		return file.Uid, nil
	case "valid":
		return file.Valid, nil
	}
	return nil, nil
}

func ReNameFile(fileId, uid uint64, newName string) error {
	// 检验文件夹名称是否非法
	if newName == "" {
		return util.NewError(util.ERROR_FILE_NAME_INVALID)
	}
	matched, err := regexp.MatchString("[\\/+?:*<>!|]", newName)
	if err != nil || matched == true {
		return util.NewError(util.ERROR_FILE_NAME_INVALID)
	}
	// 判断此文件是否属于操作者
	var file File
	err = DB.Table("files").Where("file_id = ? and valid = true", fileId).Find(&file).Error
	if err != nil || file.FileID == 0 {
		// 文件不存在
		return util.NewError(util.ERROR_FILE_NOT_EXISTS)
	}
	if file.Uid != uid {
		// 无权操作
		return util.NewError(util.ERROR_AUTH_NOT_PERMISSION)
	}
	// 判断是否有同名文件
	var count int64
	err = DB.Table("files").Where("file_name = ? and parent_id = ?", newName, file.ParentID).Count(&count).Error
	if count > 0 {
		return util.NewError(util.ERROR_FILE_SAME_NAME)
	}
	tempSlice := strings.Split(newName, ".")
	fileType := ""
	if len(tempSlice) > 0 {
		fileType = tempSlice[len(tempSlice)-1]
		if len(fileType) > 6 {
			fileType = ""
		}
	}
	DB.Table("files").Where("file_id = ?", fileId).Updates(&File{FileName: newName, Type: fileType})
	return nil
}

func GetThumbnails(uid uint64, startDate string, endDate string) []File {
	if startDate != "" {

	}
	if endDate != "" {

	}
	// 从uid拿到图片类相关文件
	var files []File
	DB.Table("files").Where("uid = ? and valid = true and ( type = 'png' or type = 'jpg' or type = 'jpeg' )", uid).Find(&files)
	return files
}

func GetThumbnail(uid, fileId uint64) (string, string, error) {
	// 权限校验
	// 此文件是否属于该用户
	var file File
	DB.Table("files").Where("file_id = ? and valid = true", fileId).Find(&file)
	if file.FileID == 0 || file.Uid != uid {
		return "", "", util.NewError(util.ERROR_AUTH_NOT_PERMISSION)
	}
	return file.MD5, file.Type, nil
}
