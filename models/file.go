package model

import (
	"Rhine-Cloud-Driver/pkg/cache"
	"Rhine-Cloud-Driver/pkg/mq"
	"Rhine-Cloud-Driver/pkg/util"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"os"
	"regexp"
	"sort"
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
	ParentID    uint64 `json:"parent_id,omitempty" gorm:"index:idx_parent_id"`
	Path        string `json:"path,omitempty" gorm:"index:idx_path"`
	//QuickAccess bool   `json:"quick_access"`
	Uid     uint64 `json:"uid,omitempty"`
	Valid   bool   `json:"valid,omitempty"`
	Type    string `json:"type,omitempty" gorm:"index:idx_type;size:6"`
	ExtraID uint64 `json:"extra_id"`
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

const RedisPrefixUploadChunks = "file_chunks_"
const RedisPrefixUploadID = "upload_id_"
const RedisPrefixDownloadID = "download_id_"

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

type Files []File

func (a Files) Len() int           { return len(a) }
func (a Files) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Files) Less(i, j int) bool { return a[i].ExtraID < a[j].ExtraID }

// 去重函数
func uniqueFiles(files []File) []File {
	var unique []File
	for i, file := range files {
		if i == 0 || file.ExtraID != files[i-1].ExtraID {
			unique = append(unique, file)
		}
	}
	return unique
}

func CheckFileRepeat(fileID, extraID uint64, md5 string) {
	// 将其和各个同md5的文件对比
	var files Files
	DB.Table("files").Where("md5=?", md5).Find(&files)
	dirPath := "./uploads/"
	// 对files排序，先校验extra_id小的
	sort.Sort(files)
	files = uniqueFiles(files)
	for _, v := range files {
		fmt.Println(v)
		if v.ExtraID == 0 {
			if isSame := util.IsSameFile(dirPath+v.MD5, dirPath+md5+"_"+strconv.FormatUint(extraID, 10)); isSame {
				// 将所有与该文件extra_id相同的文件都划为这个extra_id
				DB.Table("files").Where("extra_id=?", extraID).Update("extra_id", v.ExtraID)
				// 删除旧的文件
				util.RemoveFile(dirPath + md5 + "_" + strconv.FormatUint(extraID, 10))
				break
			}
		} else {
			if isSame := util.IsSameFile(dirPath+v.MD5+"_"+strconv.FormatUint(v.ExtraID, 10), dirPath+md5+"_"+strconv.FormatUint(extraID, 10)); isSame {
				// 与上面同理
				DB.Table("files").Where("extra_id=?", extraID).Update("extra_id", v.ExtraID)
				// 删除旧的文件
				util.RemoveFile(dirPath + md5 + "_" + strconv.FormatUint(extraID, 10))
				break
			}
		}
	}
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

type MetaData struct {
	//FileExtraMD5 string    `json:"file_extra_md5"` // 文件的额外MD5，包括前100个字节的MD5和后100个字节的MD5加在一起，主要是判断是否为同一文件
	UploadedChunks []byte    `json:"uploaded_chunks"`
	LastUploaded   time.Time `json:"last_uploaded"` // 最后一次上传的时间
	FileName       string    `json:"file_name"`
	Uid            uint64    `json:"uid"`
	FileSize       uint64    `json:"file_size"`
	TargetDirID    uint64    `json:"target_dir_id"`
}

func (data *MetaData) GetDataFromFile(uploadID string) bool {
	// 时间距离今天超过清理的时间，则不能再使用
	expireTime, err := strconv.ParseInt(GetSettingByName("chunks_expire_time"), 10, 64)
	if err != nil {
		// 取过期时间失败，直接取默认两天
		expireTime = 60 * 60 * 24 * 2
	}
	metaDataBytes, isExist := util.ReadFile("./uploads/metadata/metadata_" + uploadID)
	if isExist == false {
		return false
	}
	json.Unmarshal(metaDataBytes, data)
	if int64(time.Now().Sub(data.LastUploaded).Seconds()) >= expireTime {
		// 元数据文件过期时，即便redis有数据也不能使用，并且立即将元数据文件清理掉，以便重新开始创建新的元数据文件
		go util.RemoveMetaData(uploadID)
		return false
	}
	return true
}

func (data *MetaData) NewMetaData() (uploadID string) {
	// 从设置中取redis键值过期时间
	expireTime, err := strconv.ParseInt(GetSettingByName("upload_id_expire_time"), 10, 64)
	if err != nil {
		// 取过期时间失败，直接取默认一小时
		expireTime = 60 * 60
	}
	// 随机生成一个32位新的key，并将其作为UploadID
	uploadID = util.RandStringRunes(32)
	_, isRedisExist := cache.GetRedisKey(RedisPrefixUploadID + uploadID)
	_, isMetadataNotExist := os.Stat("./uploads/metadata/metadata_" + uploadID)
	for isRedisExist != false || isMetadataNotExist == nil {
		uploadID = util.RandStringRunes(32)
		_, isRedisExist = cache.GetRedisKey(RedisPrefixUploadID + uploadID)
		_, isMetadataNotExist = os.Stat("./uploads/metadata/metadata_" + uploadID)
	}
	// 创建新的uploadID文件
	util.CreateFileIfNot("./uploads/metadata/metadata_" + uploadID)
	newMetaDataBytes, _ := json.Marshal(data)
	util.WriteFile(newMetaDataBytes, "./uploads/metadata/metadata_"+uploadID)
	// metadata写好了，接下来写到缓存当中
	uploadSession := &UploadSession{
		FileName:    data.FileName,
		Uid:         data.Uid,
		FileSize:    data.FileSize,
		TargetDirID: data.TargetDirID,
	}
	bytes, _ := json.Marshal(&uploadSession)
	cache.SetRedisKey(RedisPrefixUploadID+uploadID, bytes, time.Second*time.Duration(expireTime))
	cache.SetRedisKey(RedisPrefixUploadChunks+uploadID, nil, time.Second*time.Duration(expireTime))
	return
}

func (data *MetaData) CheckUploadID(uploadID string) (isExist, isNotExpire bool) {
	// 从设置中取redis键值过期时间
	expireTime, err := strconv.ParseInt(GetSettingByName("upload_id_expire_time"), 10, 64)
	if err != nil {
		// 取过期时间失败，直接取默认一小时
		expireTime = 60 * 60
	}
	isExist = true
	isNotExpire = true
	if _, isMetadataNotExist := os.Stat("./uploads/metadata/metadata_" + uploadID); isMetadataNotExist != nil {
		isNotExpire = false
		return
	}
	// 保活
	mq.GlobalMQ.Publish(uploadID, mq.Message{
		Event:   "file_status",
		Content: "uploading",
	})
	// 查看上传管理协程是否仍存在
	if !mq.GlobalMQ.CheckStatus(uploadID) {
		// 要重建上传管理协程，model层重新从metadata里取出相应的uploadsession和chunks即可
		// todo要考虑是否是宕机，如果是宕机则redis仍有数据，从redis取数据就行了
		isExist = false
		if isValid := data.GetDataFromFile(uploadID); isValid == false {
			isNotExpire = false
			return
		}
		bytes, _ := json.Marshal(&UploadSession{
			FileName:    data.FileName,
			Uid:         data.Uid,
			FileSize:    data.FileSize,
			TargetDirID: data.TargetDirID,
		})
		cache.SetRedisKey(RedisPrefixUploadID+uploadID, bytes, time.Second*time.Duration(expireTime))
		cache.SetRedisKey(RedisPrefixUploadChunks+uploadID, nil, time.Second*time.Duration(expireTime))
		return
	}
	return
}

func (data *MetaData) UpdateMetaData(uploadID string, updateTime time.Time) {
	var oldMetaData MetaData
	util.CreateFileIfNot("./uploads/metadata/metadata_" + uploadID)
	isExist := oldMetaData.GetDataFromFile(uploadID)
	if isExist != false {
		// 更新的时间如果要晚则不做操作
		if oldMetaData.LastUploaded.Format("2006-01-02 15:04:05") >= updateTime.Format("2006-01-02 15:04:05") {
			return
		}
		//util.RemoveFile("/upload/metadata/metadata_" + md5 + extraMD5)
	}
	newMetaDataBytes, _ := json.Marshal(data)
	util.WriteFile(newMetaDataBytes, "./uploads/metadata/metadata_"+uploadID)
}

func (data *MetaData) WriteCacheToMetadata(uploadID string) {
	// 将数据更新到metadata中
	bytes, _ := cache.GetRedisKeyBytes(RedisPrefixUploadChunks + uploadID)
	data.UploadedChunks = bytes
	data.LastUploaded = time.Now()
	data.UpdateMetaData(uploadID, time.Now())
}

func WriteCacheToMetadata(uploadID string) {
	data := MetaData{}
	data.GetDataFromFile(uploadID)
	bytes, isExist := cache.GetRedisKeyBytes(RedisPrefixUploadChunks + uploadID)
	if isExist == false {
		return
	}
	data.UploadedChunks = bytes
	data.LastUploaded = time.Now()
	data.UpdateMetaData(uploadID, time.Now())
}

func CheckUserStorage(uid, fileSize uint64) bool {
	var nowUser User
	DB.Table("users").Where("uid=?", uid).Find(&nowUser)
	if nowUser.UsedStorage+fileSize > nowUser.TotalStorage {
		return false
	}
	return true
}

type UploadSession struct {
	FileName    string
	Uid         uint64
	FileSize    uint64
	TargetDirID uint64
}

func NewUploadID(session UploadSession) string {
	// 从设置中取过期时间
	expireTime, err := strconv.ParseInt(GetSettingByName("upload_id_expire_time"), 10, 64)
	if err != nil {
		// 取过期时间失败，直接取默认一小时
		expireTime = 60 * 60
	}
	// 随机生成一个32位新的key，并将其作为UploadID
	uploadID := util.RandStringRunes(32)
	_, isExist := cache.GetRedisKey(RedisPrefixUploadID + uploadID)
	for isExist != false {
		uploadID = util.RandStringRunes(32)
		_, isExist = cache.GetRedisKey(RedisPrefixUploadID + uploadID)
	}
	value, _ := json.Marshal(&session)
	cache.SetRedisKey(RedisPrefixUploadID+uploadID, value, time.Second*time.Duration(expireTime))
	return uploadID
}

func GetChunkNum(fileSize uint64) int {
	chunkSize, err := strconv.ParseUint(GetSettingByName("chunk_size"), 10, 64)
	if err != nil {
		chunkSize = 20971520 // 20MB
	}
	chunkNum := int(fileSize / chunkSize)
	if fileSize%chunkSize != 0 {
		chunkNum++
	}
	return chunkNum
}

func NewUploadTask(uploadID, fileName string, uid, fileSize, targetDirID uint64) (isExist bool, chunkQueue string, err error) {
	expireTime, err := strconv.ParseInt(GetSettingByName("upload_id_expire_time"), 10, 64)
	if err != nil {
		// 取过期时间失败，直接取默认一小时
		expireTime = 60 * 60
	}
	chunkNum := GetChunkNum(fileSize)
	chunks := strings.Repeat("0", chunkNum)
	metaData := &MetaData{
		LastUploaded: time.Now(),
		FileName:     fileName,
		Uid:          uid,
		FileSize:     fileSize,
		TargetDirID:  targetDirID,
	}
	// 创建元数据文件和块文件夹
	metaData.UpdateMetaData(uploadID, time.Now())
	util.Mkdir(uploadID + "_chunks")
	cache.SetRedisKey(RedisPrefixUploadChunks+uploadID, nil, time.Second*time.Duration(expireTime))
	return false, chunks, nil
}

//func UploadPrepare(md5, extraMD5, fileName string, chunkNum int64, uid, fileSize, targetDirID uint64) (isExist bool, chunkQueue, uploadID string, err error) {
//	// 从设置中取过期时间
//	expireTime, err := strconv.ParseInt(GetSettingByName("upload_id_expire_time"), 10, 64)
//	if err != nil {
//		// 取过期时间失败，直接取默认一小时
//		expireTime = 60 * 60
//	}
//	var files []File
//	DB.Table("files").Where("md5=?", md5).Find(&files)
//	if len(files) > 0 {
//		for i := 0; i < len(files); i++ {
//			if IsSameFile(files[i].MD5, files[i].ExtraMD5, md5, extraMD5, files[i].FileStorage, fileSize) {
//				// 存在AddFile即可
//				err := AddFile(uid, md5, extraMD5, fileName, fileSize, targetDirID, false)
//				if err != nil {
//					return true, "", "", err
//				}
//				return true, "", "", nil
//			}
//		}
//	}
//	chunks := make([]byte, chunkNum)
//	for i := int64(0); i < chunkNum; i++ {
//		chunks[i] = byte(0)
//	}
//	metaData := &MetaData{
//		UploadedChunks: string(chunks),
//		TotalChunks:    chunkNum,
//		LastUploaded:   time.Now(),
//	}
//	metaData.UpdateMetaData(md5, extraMD5, time.Now())
//	cache.SetRedisKey(RedisPrefixUploadChunks+md5+extraMD5, chunks, time.Second*time.Duration(expireTime))
//	cache.SetRedisKey(RedisPrefixUploadChunkNum+md5+extraMD5, chunkNum, time.Second*time.Duration(expireTime))
//	return false, string(chunks), uploadID, nil
//}

//func VerifyUploadID(uploadID string, md5 string, extraMD5 string) bool {
//	if value, isExist := cache.GetRedisKey("upload_id_" + uploadID); isExist == false || value != md5+extraMD5 {
//		return false
//	}
//	return true
//}

func DealFileChunk(fileIndex int64, uploadID string) (isExist bool, err error) {
	chunksRedisKey := RedisPrefixUploadChunks + uploadID
	fileSessionRedisKey := RedisPrefixUploadID + uploadID
	// 从redis中拿到这块的情况
	uploadSession := &UploadSession{}
	if tempRedisValue, isExist := cache.GetRedisKeyBytes(fileSessionRedisKey); isExist == false {
		// 认为是数据丢失，直接回传上传失败
		return false, util.NewError(util.ERROR_FILE_NOT_EXISTS)
	} else {
		_ = json.Unmarshal(tempRedisValue, uploadSession)
	}
	chunkNum := int64(GetChunkNum(uploadSession.FileSize))
	if fileIndex < 0 || fileIndex >= chunkNum {
		return false, util.NewError(util.ERROR_FILE_INDEX_INVALID)
	}
	if cache.GetRedisKeyBitmap(chunksRedisKey, fileIndex) == 0 {
		cache.SetRedisKeyBitmap(chunksRedisKey, fileIndex, 1, 0)
		return false, nil
	}
	// 该部分已经被其他完成
	return true, nil
}

func MergeFileChunks(uploadID string) (int64, error) {
	//if _, isExist := cache.GetRedisKey("upload_id_" + uploadID); isExist == false {
	//	return 0, util.NewError(util.ERROR_AUTH_UPLOADID_INVALID)
	//}
	chunksRedisKey := RedisPrefixUploadChunks + uploadID
	uploadSessionRedisKey := RedisPrefixUploadID + uploadID
	uploadSession := &UploadSession{}
	if tempRedisValue, isExist := cache.GetRedisKeyBytes(uploadSessionRedisKey); isExist == false {
		return 0, util.NewError(util.ERROR_FILE_NOT_EXISTS)
	} else {
		_ = json.Unmarshal(tempRedisValue, uploadSession)
	}
	chunkNum := int64(GetChunkNum(uploadSession.FileSize))
	hasFinishedChunk := cache.CountRedisKeyBitmap(chunksRedisKey, 0, chunkNum-1)
	if hasFinishedChunk == chunkNum {
		//cache.DelRedisKey(chunksRedisKey)
		//cache.DelRedisKey(uploadSessionRedisKey)
		return chunkNum, nil
	}
	return 0, util.NewError(util.ERROR_FILE_CHUNK_MISSING)
}

func CleanMetaData(uploadID string) {
	chunksRedisKey := RedisPrefixUploadChunks + uploadID
	uploadSessionRedisKey := RedisPrefixUploadID + uploadID
	// 删除缓存
	cache.DelRedisKey(chunksRedisKey)
	cache.DelRedisKey(uploadSessionRedisKey)
	// 删除元数据文件，删除分块文件夹
	os.RemoveAll("./uploads/metadata/metadata_" + uploadID)
	os.RemoveAll("./uploads/" + uploadID + "_chunks")
}

func EditFileHash(fileID uint64, hash string) (isExist bool, maxValue uint64) {
	isExist = false
	var files []File
	DB.Table("files").Where("md5=?", hash).Find(&files)
	maxValue = 0
	for i := 0; i < len(files); i++ {
		if files[i].ExtraID > maxValue {
			maxValue = files[i].ExtraID
		}
	}
	maxValue++
	if len(files) > 0 {
		isExist = true
		DB.Table("files").Where("file_id", fileID).Update("md5", hash)
		DB.Table("files").Where("file_id", fileID).Update("extra_id", maxValue)
		// todo:给处理重复文件的队列发生消息
		//mq.GlobalMQ.
		//mq.GlobalMQ.Publish("")
		//util.IsSameFile()
		go CheckFileRepeat(fileID, maxValue, hash)
	} else {
		DB.Table("files").Where("file_id", fileID).Update("md5", hash)
	}
	return
}

func AddFile(uid uint64, md5 string, fileName string, fileSize, parentID uint64, isOrigin bool) (uint64, error) {
	// 校验容量是否充足
	var nowUser User
	DB.Table("users").Where("uid=?", uid).Find(&nowUser)
	if nowUser.UsedStorage+fileSize > nowUser.TotalStorage {
		return 0, util.NewError(util.ERROR_USER_STORAGE_EXCEED)
	}
	// 校验parentID是否属于该UID本人和是否存在
	var fileDir File
	err := DB.Table("files").Where("file_id=? and is_dir=1", parentID).First(&fileDir).Error
	if err != nil || fileDir.Uid != uid {
		return 0, util.NewError(util.ERROR_FILE_STORE_PATH_INVALID)
	}
	// 开启事务来增加
	tx := DB.Begin()
	// 不允许同一目录有相同文件名的文件
	var count int64
	tx.Table("files").Where("file_name=? and parent_id=? and is_dir=false and valid=true", fileName, parentID).Count(&count)
	if count > 0 {
		tx.Rollback()
		return 0, util.NewError(util.ERROR_FILE_SAME_NAME)
	}
	tempSlice := strings.Split(fileName, ".")
	fileType := ""
	if len(tempSlice) > 0 {
		fileType = tempSlice[len(tempSlice)-1]
		if len(fileType) > 6 {
			fileType = ""
		}
	}
	file := &File{
		Uid:         uid,
		MD5:         md5,
		FileName:    fileName,
		FileStorage: fileSize,
		ParentID:    parentID,
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		Valid:       true,
		IsDir:       false,
		IsOrigin:    isOrigin,
		Path:        fileDir.Path + fileDir.FileName + "/",
		Type:        fileType,
	}
	err = tx.Table("files").Create(file).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	// 增加容量
	err = tx.Table("users").Where("uid=?", uid).Update("used_storage", gorm.Expr("used_storage+?", fileSize)).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()
	return file.FileID, nil
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

type DownloadSession struct {
	Md5      string `json:"md5"`
	FileName string `json:"file_name"`
	FileID   uint64 `json:"file_id"`
	ExtraID  uint64 `json:"extra_id"`
}

func GetDownloadKey(uid, fileID uint64, fileKey string) (downloadID string, err error) {
	// 验证是否是本人的文件
	var file File
	err = DB.Table("files").Where("file_id = ? and valid = true and is_dir = false", fileID).Find(&file).Error
	if !PermissionVerify(uid, PERMISSION_ADMIN_READ) && (err != nil || file.Uid != uid) {
		return "", util.NewError(util.ERROR_DOWNLOAD_FILE_INVALID)
	}
	if file.MD5 == "" {
		return "", util.NewError(util.ERROR_FILE_STATUS_UPLOADING)
	}
	downloadID = util.RandStringRunes(6) + fileKey
	downloadSession := &DownloadSession{
		Md5:      file.MD5,
		FileName: file.FileName,
		FileID:   file.FileID,
		ExtraID:  file.ExtraID,
	}
	bytes, _ := json.Marshal(&downloadSession)
	cache.SetRedisKey("download_id_"+downloadID, bytes, time.Hour/2)
	return downloadID, nil
}

func DownloadFile(fileInfo any, id string) (fileName, fileMD5 string, err error) {
	err = nil
	//fileInfo, isExist := cache.GetRedisKey(key)
	//if isExist == false {
	//	// 链接无效或已过期
	//	return "", "", util.NewError(util.ERROR_DOWNLOAD_KEY_INVALID)
	//}
	// 前6位是文件ID哈希的结果，检测前6位可以得出下载的文件是否有遭到篡改
	downloadSession := &DownloadSession{}
	json.Unmarshal(fileInfo.([]byte), downloadSession)
	fileID, err := util.HashDecode(id[6:], 6)
	if err != nil || downloadSession.FileID != fileID {
		return "", "", err
	}
	fileName = downloadSession.FileName
	fileMD5 = downloadSession.Md5
	if downloadSession.ExtraID != 0 {
		fileMD5 = fileMD5 + "_" + strconv.FormatUint(downloadSession.ExtraID, 10)
	}
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

//func CheckChunkQueue(uploadID string) (isExist bool, chunks string) {
//	isExist = false
//	// 检查元数据文件的队列情况
//	//metaDataBytes, isMetaDataExist := util.ReadFile("/upload/metadata/metadata_" + md5 + extraMD5)
//	var fileMetaData MetaData
//	isMetaDataExist := fileMetaData.GetDataFromFile(uploadID)
//	// 检查redis的队列的情况
//	tempValue, isRedisExist := cache.GetRedisAllBitmap(RedisPrefixUploadChunks + uploadID)
//	if isMetaDataExist == false && isRedisExist == false {
//		return
//	}
//	isExist = true
//	metaDataChunkNum, RedisChunkNum := 0, 0
//	var metaDataUploadedChunks string
//	var redisUploadedChunks []rune
//	// 如果都有，看谁的数据新，即谁统计的上传块数多
//	if isMetaDataExist {
//		metaDataUploadedChunks = fileMetaData.UploadedChunks
//		for i := 0; i < len(metaDataUploadedChunks); i++ {
//			if metaDataUploadedChunks[i] == '1' {
//				metaDataChunkNum++
//			}
//		}
//	}
//	if isRedisExist {
//		//redisUploadedChunks := tempValue.([]byte)
//		redisUploadedChunks = make([]rune, len(tempValue))
//		for i := 0; i < len(tempValue); i++ {
//			if tempValue[i] == 1 {
//				RedisChunkNum++
//				redisUploadedChunks[i] = '1'
//			} else {
//				redisUploadedChunks[i] = '0'
//			}
//		}
//		fmt.Println(string(redisUploadedChunks))
//	}
//	if metaDataChunkNum > RedisChunkNum {
//		chunks = metaDataUploadedChunks
//		// 更新redis的数据
//		// 从设置中取过期时间
//		expireTime, err := strconv.ParseInt(GetSettingByName("upload_id_expire_time"), 10, 64)
//		if err != nil {
//			// 取过期时间失败，直接取默认一小时
//			expireTime = 60 * 60
//		}
//		cache.SetRedisKey(RedisPrefixUploadChunkNum+md5+extraMD5, fileMetaData.TotalChunks, time.Second*time.Duration(expireTime))
//		cache.SetRedisKey(RedisPrefixUploadChunks+uploadID, tempValue, time.Second*time.Duration(expireTime))
//	} else {
//		chunks = string(redisUploadedChunks)
//		// 更新元数据文件的数据
//		fileMetaData.LastUploaded = time.Now()
//		fileMetaData.UploadedChunks = string(redisUploadedChunks)
//		fileMetaData.TotalChunks = int64(len(redisUploadedChunks))
//		go fileMetaData.UpdateMetaData(uploadID, time.Now())
//	}
//	return
//}

func RenewUploadID(uploadID string) bool {
	expireTime, err := strconv.ParseInt(GetSettingByName("upload_id_expire_time"), 10, 64)
	if err != nil {
		// 取过期时间失败，直接取默认一小时
		expireTime = 60 * 60
	}
	var metaData MetaData
	if isSuccess := cache.RenewRedisKey(RedisPrefixUploadChunks+uploadID, time.Second*time.Duration(expireTime)); !isSuccess {
		// 取元数据文件
		isExist := metaData.GetDataFromFile(uploadID)
		if isExist == false {
			return false
		}
		cache.SetRedisKey(RedisPrefixUploadChunks+uploadID, metaData.UploadedChunks, time.Second*time.Duration(expireTime))
	} else {
		// 将数据更新到metadata中
		metaData.WriteCacheToMetadata(uploadID)
	}
	bytes, _ := json.Marshal(&UploadSession{
		FileName:    metaData.FileName,
		Uid:         metaData.Uid,
		FileSize:    metaData.FileSize,
		TargetDirID: metaData.TargetDirID,
	})
	if isSuccess := cache.RenewRedisKey(RedisPrefixUploadID+uploadID, time.Second*time.Duration(expireTime)); !isSuccess {
		cache.SetRedisKey(RedisPrefixUploadID+uploadID, bytes, time.Second*time.Duration(expireTime))
	}
	return true
}

func GetRedisAllBitmap(uploadID string, fileSize uint64) string {
	// 取整个bitmap
	bytes, _ := cache.GetRedisKeyBytes(RedisPrefixUploadChunks + uploadID)
	chunks := make([]byte, GetChunkNum(fileSize))
	for i := 0; i < len(bytes); i++ {
		// 取出来的byte要将其变为bit才能拿到0和1的情况，1byte=8bit
		tempStr := fmt.Sprintf("%08b", bytes[i])
		for j := 0; j < 8; j++ {
			if i*8+j >= len(chunks) {
				break
			}
			chunks[i*8+j] = tempStr[j]
		}
	}
	for i := 0; i < len(chunks)-len(bytes)*8; i++ {
		chunks[len(chunks)-i-1] = '0'
	}
	return string(chunks)
}
