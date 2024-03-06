package controllers

import (
	model "Rhine-Cloud-Driver/models"
	"Rhine-Cloud-Driver/pkg/log"
	"Rhine-Cloud-Driver/pkg/mq"
	"Rhine-Cloud-Driver/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/blake2b"
	"io/ioutil"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//var uploadInfoChannel map[string]chan UploadInfo

//var uploadStatusChannel map[]

type GetFileSystemRequest struct {
	Path        string   `json:"path"`
	Offset      int      `json:"offset"`
	Limit       int      `json:"limit"`
	TargetDirId uint64   `json:"target_dir_id,omitempty"`
	FilterKey   string   `json:"filter_key"`
	FilterType  []string `json:"filter_type"`
}

type GetFileSystemResponse struct {
	Count     int64        `json:"count"`
	DirFileID uint64       `json:"dir_file_id"`
	Path      string       `json:"path,omitempty"`
	Files     []FileSystem `json:"files"`
}

type FileSystem struct {
	FileID      uint64 `json:"file_id"`
	FileName    string `json:"file_name"`
	MD5         string `json:"md5,omitempty"`
	CreateTime  string `json:"create_time"`
	FileStorage uint64 `json:"file_storage,omitempty"`
	IsDir       bool   `json:"is_dir"`
	Path        string `json:"path,omitempty"`
}

type UploadInfo struct {
	ginContext *gin.Context
	chunkIndex int64
	status     string
}

func GetMyFiles(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data GetFileSystemRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	path, _ := url.QueryUnescape(data.Path)
	count, dirFileID, originFiles, dirPath, err := model.BuildFileSystem(uid, path, data.TargetDirId, data.Limit, data.Offset, data.FilterKey, data.FilterType)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	files := make([]FileSystem, len(originFiles))
	for i := range files {
		files[i] = FileSystem{
			FileID:      originFiles[i].FileID,
			FileName:    originFiles[i].FileName,
			MD5:         originFiles[i].MD5,
			FileStorage: originFiles[i].FileStorage,
			IsDir:       originFiles[i].IsDir,
			CreateTime:  originFiles[i].CreateTime,
			Path:        originFiles[i].Path,
		}
	}
	resp := GetFileSystemResponse{
		Count:     count,
		DirFileID: dirFileID,
		Path:      dirPath,
		Files:     files,
	}
	makeResult(c, 200, nil, resp)
}

func UploadGovern(uploadID string) {
	expireTime := time.Now().Add(time.Second * 60 * 10)
	timer := time.NewTimer(time.Until(expireTime))
	ch := mq.GlobalMQ.Subscribe(uploadID, 1)
	isRenew := false
exit:
	for {
		select {
		case info := <-ch:
			log.Logger.Info("消息消费：", zap.Any("info", info))
			if info.Event == "file_status" && info.Content.(string) == "merging" {
				log.Logger.Info("进程已完成，上传守护线程结束：", zap.Any("upload_id", uploadID))
				break exit
			}
			isRenew = true
		case <-timer.C:
			if isRenew {
				isRenew = false
				model.RenewUploadID(uploadID)
			} else {
				mq.GlobalMQ.Unsubscribe(uploadID, ch)
				log.Logger.Info("超过时间，上传守护线程结束：", zap.Any("upload_id", uploadID))
				// 将缓存写回metadata
				go model.WriteCacheToMetadata(uploadID)
				break exit
			}
		}
	}
}

func Upload(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	//tempValue, _ := c.Get("uid")
	//uid := tempValue.(uint64)
	form, _ := c.MultipartForm()
	if form.Value["file_name"] == nil || form.Value["upload_id"] == nil {
		// 参数缺失
		makeResult(c, 400, util.NewError(util.ERROR_PARA_ABSENT), nil)
		return
	}
	//fileName := form.Value["file_name"][0]
	fileIndex := form.Value["chunk_index"][0]
	//md5 := form.Value["md5"][0]
	//extraMD5 := form.Value["extra_md5"][0]
	uploadID := form.Value["upload_id"][0]
	//parentID, err := strconv.ParseUint(form.Value["parent_id"][0], 10, 64)
	chunkIndex, err := strconv.ParseInt(fileIndex, 10, 64)
	if err != nil {
		log.Logger.Error("ParseInt 转换错误", zap.Error(err))
		makeResult(c, 400, util.NewError(util.ERROR_PARA_INVALID), nil)
		return
	}
	// 验证uploadID是否有效
	metaData := &model.MetaData{}
	isExist, isNotExpire := metaData.CheckUploadID(uploadID)
	if !isNotExpire {
		makeResult(c, 200, util.NewError(util.ERROR_UPLOAD_TIME_EXCEED), nil)
		return
	}
	if !isExist {
		go UploadGovern(uploadID)
	}
	file, _ := c.FormFile("file")
	isExist, err = model.DealFileChunk(chunkIndex, uploadID)
	if err != nil {
		log.Logger.Error("DealFileChunk error", zap.Error(err))
		makeResult(c, 200, err, nil)
		mq.GlobalMQ.Publish(uploadID, mq.Message{
			Event:   "chunk_status",
			Content: "failed",
		})
		return
	}
	if !isExist {
		c.SaveUploadedFile(file, "./uploads/"+uploadID+"_chunks/"+fileIndex)
	}
	makeResult(c, 200, nil, nil)
	mq.GlobalMQ.Publish(uploadID, mq.Message{
		Event:   "chunk_status",
		Content: "success",
	})
}

func MergeFileChunks(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	form, _ := c.MultipartForm()
	if form.Value["file_name"] == nil || form.Value["upload_id"] == nil {
		// 参数缺失
		makeResult(c, 200, util.NewError(util.ERROR_PARA_ABSENT), nil)
		return
	}
	fileName := form.Value["file_name"][0]
	uploadID := form.Value["upload_id"][0]
	parentID, err := strconv.ParseUint(form.Value["parent_id"][0], 10, 64)
	mq.GlobalMQ.Publish(uploadID, mq.Message{
		Event:   "file_status",
		Content: "merging",
	})
	// 若不存在，则合并
	chunkNum, err := model.MergeFileChunks(uploadID)
	if err != nil {
		// 请重试
		makeResult(c, 200, err, nil)
		return
	}
	// 合并
	fi, err := os.Stat("./uploads/" + uploadID + "_chunks/" + strconv.FormatInt(chunkNum-1, 10))
	if err != nil {
		makeResult(c, 200, util.NewError(util.ERROR_FILE_NOT_EXISTS), nil)
		return
	}
	fileSize := (chunkNum-1)*(20*1024*1024) + fi.Size()
	fileID, err := model.AddFile(uid, "", fileName, uint64(fileSize), parentID, true)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	hash, _ := blake2b.New256(nil)
	allFile, _ := os.OpenFile("./uploads/"+uploadID, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	for i := int64(0); i < chunkNum; i++ {
		file, _ := os.OpenFile("./uploads/"+uploadID+"_chunks/"+strconv.FormatInt(i, 10), os.O_RDONLY, os.ModePerm)
		// 计算文件BLAKE2值
		bytes, _ := ioutil.ReadAll(file)
		hash.Write(bytes)
		//util.RemoveFile("./uploads/" + md5 + extraMD5 + "_chunks/" + strconv.FormatInt(i, 10))
		allFile.Write(bytes)
	}
	// 更新文件hash值
	hashValue := fmt.Sprintf("%x", hash.Sum(nil))
	isExist, extraID := model.EditFileHash(fileID, hashValue)
	if isExist {
		os.Rename("./uploads/"+uploadID, "./uploads/"+hashValue+"_"+strconv.FormatUint(extraID, 10))
	} else {
		os.Rename("./uploads/"+uploadID, "./uploads/"+hashValue)
	}
	model.CleanMetaData(uploadID)
	makeResult(c, 200, nil, nil)
}

type UploadTaskRequest struct {
	FileName     string `json:"file_name"`
	FileChunkNum int64  `json:"file_chunk_num"`
	FileSize     uint64 `json:"file_size"`
	TargetDirID  uint64 `json:"target_dir_id"`
	UploadID     string `json:"upload_id,omitempty"`
}

type UploadTaskResponse struct {
	IsExist     bool   `json:"is_exist"`
	FinishQueue string `json:"finish_queue,omitempty"`
	UploadID    string `json:"upload_id"`
}

func UploadTaskCreate(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data UploadTaskRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 400, util.NewError(util.ERROR_PARA_ABSENT), nil)
		return
	}
	if data.FileSize < 8 {
		// 不接受小于1B的文件
		makeResult(c, 200, util.NewError(util.ERROR_FILE_TINY), nil)
		return
	}
	if model.CheckUserStorage(uid, data.FileSize) == false {
		makeResult(c, 200, util.NewError(util.ERROR_USER_STORAGE_EXCEED), nil)
		return
	}
	if data.UploadID != "" {
		// 检验ID是否仍有有效，有效则直接回传
		metaData := model.MetaData{}
		if isExist, isNotExpire := metaData.CheckUploadID(data.UploadID); isNotExpire == true {
			// 取redis当前状态
			if isExist == false {
				go UploadGovern(data.UploadID)
			}
			chunks := model.GetRedisAllBitmap(data.UploadID, data.FileSize)
			resp := UploadTaskResponse{
				//IsExist:     isExist,
				FinishQueue: chunks,
				UploadID:    data.UploadID,
			}
			makeResult(c, 200, nil, resp)
			return
		}
	}
	// 从元数据文件和cache里找最新的上传记录
	//isExist, chunks := model.CheckChunkQueue(data.FileMD5, data.FileExtraMD5)
	//if isExist == false {
	//	// 新建元数据文件和redis文件
	//	var err error
	//	isExist, chunks, err = model.NewUploadTask(data.FileMD5, data.FileExtraMD5, data.FileName, uid, data.FileSize, data.TargetDirID)
	//	if err != nil {
	//		makeResult(c, 200, err, nil)
	//		return
	//	}
	//}
	// 生成新的元数据文件，并将元数据文件的部分数据存放缓存
	metaData := &model.MetaData{
		LastUploaded: time.Now(),
		FileName:     data.FileName,
		Uid:          uid,
		FileSize:     data.FileSize,
		TargetDirID:  data.TargetDirID,
	}
	uploadID := metaData.NewMetaData()
	resp := UploadTaskResponse{
		//IsExist:     isExist,
		FinishQueue: strings.Repeat("0", model.GetChunkNum(data.FileSize)),
		UploadID:    uploadID,
	}
	//if !isExist {
	// 创建上传管理协程
	go UploadGovern(uploadID)
	//}
	makeResult(c, 200, nil, resp)
}

type MkdirRequest struct {
	FileName string `json:"file_name"`
	ParentID uint64 `json:"parent_id"`
}

func Mkdir(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data MkdirRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.Mkdir(uid, data.FileName, data.ParentID)
	makeResult(c, 200, err, nil)
}

type MoveFilesRequest struct {
	TargetDirID uint64   `json:"target_dir_id"`
	FileIDList  []uint64 `json:"file_id_list"`
}

func MoveFiles(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data MoveFilesRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.MoveFiles(uid, data.FileIDList, data.TargetDirID)
	makeResult(c, 200, err, nil)
}

type GetDownloadKeyRequest struct {
	FileID uint64 `json:"file_id"`
}

type GetDownloadKeyResponse struct {
	DownloadKey string `json:"download_key"`
}

func GetDownloadKey(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data GetDownloadKeyRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	hashValue, err := util.HashEncode([]int{int(data.FileID)}, 6)
	if err != nil {
		makeResult(c, 200, util.NewError(util.ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED), nil)
		return
	}
	downloadKey, err := model.GetDownloadKey(uid, data.FileID, hashValue)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	makeResult(c, 200, nil, GetDownloadKeyResponse{downloadKey})
}

func DownloadFile(c *gin.Context) {
	fileName, _ := c.Get("file_name")
	fileMD5, _ := c.Get("md5")
	c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(fileName.(string)))
	c.File("./uploads/" + fileMD5.(string))
}

type RemoveFilesRequest struct {
	FileIDList []uint64 `json:"file_id_list"`
}

func RemoveFiles(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data RemoveFilesRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.RemoveFiles(uid, data.FileIDList)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	makeResult(c, 200, nil, nil)
}

type ReNameFileRequest struct {
	FileId  uint64 `json:"file_id"`
	NewName string `json:"new_name"`
}

func ReNameFile(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data ReNameFileRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	err := model.ReNameFile(data.FileId, uid, data.NewName)
	makeResult(c, 200, err, nil)
}

type GetThumbnailsRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type ThumbnailTimeLine struct {
	Date  string       `json:"date"`
	Files []model.File `json:"files"`
}

type GetThumbnailsResponse struct {
	Data []ThumbnailTimeLine `json:"data"`
}

type Files []model.File

func (a Files) Len() int           { return len(a) }
func (a Files) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Files) Less(i, j int) bool { return a[i].CreateTime < a[j].CreateTime }

func GetThumbnails(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	var data GetThumbnailsRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	files := model.GetThumbnails(uid, data.StartDate, data.EndDate)
	// 日期处理
	//timeLine := make(map[string][]model.File)
	lastDate := ""
	resp := GetThumbnailsResponse{}
	sort.Sort(Files(files))
	for i := len(files) - 1; i >= 0; i-- {
		// 获取日期
		thisTime := files[i].CreateTime[0:10]
		if lastDate == "" || thisTime != lastDate {
			// 另起一个日期
			var newTimeLine ThumbnailTimeLine
			newTimeLine.Date = thisTime
			newTimeLine.Files = append(newTimeLine.Files, files[i])
			resp.Data = append(resp.Data, newTimeLine)
		} else {
			resp.Data[len(resp.Data)-1].Files = append(resp.Data[len(resp.Data)-1].Files, files[i])
		}
		lastDate = thisTime
	}
	makeResult(c, 200, nil, resp)
}

func GetThumbnail(c *gin.Context) {
	fileId, _ := strconv.ParseUint(c.Query("file_id"), 10, 64)
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	tempValue, _ := c.Get("uid")
	uid := tempValue.(uint64)
	md5, pngType, err := model.GetThumbnail(uid, fileId)
	if err != nil {
		log.Logger.Error("failed to get thumbnail info from database", zap.Error(err))
		makeResult(c, 200, err, nil)
		return
	}
	// 判断是否已有，已有就不再生成
	_, err = os.Stat("./uploads/thumbnail/" + md5 + "." + pngType)
	if err == nil {
		c.File("./uploads/thumbnail/" + md5 + "." + pngType)
		return
	}
	err = util.ThumbGenerate(md5, pngType)
	if err != nil {
		makeResult(c, 503, err, nil)
		return
	}
	c.File("./uploads/thumbnail/" + md5 + "." + pngType)
}
