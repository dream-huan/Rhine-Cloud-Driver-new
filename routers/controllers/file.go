package controllers

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
	"Rhine-Cloud-Driver/logic/log"
	model "Rhine-Cloud-Driver/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
)

type GetFileSystemRequest struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type GetFileSystemResponse struct {
	Count     int64        `json:"count"`
	DirFileID uint64       `json:"dir_file_id"`
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

func GetMyFiles(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data GetFileSystemRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	path, _ := url.QueryUnescape(data.Path)
	count, dirFileID, originFiles, err := model.BuildFileSystem(uid, path, data.Limit, data.Offset)
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
		Files:     files,
	}
	makeResult(c, 200, nil, resp)
}

func Upload(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	file, _ := c.FormFile("file")
	form, _ := c.MultipartForm()
	//fileName := form.Value["file_name"][0]
	fileIndex := form.Value["chunk_index"][0]
	md5 := form.Value["md5"][0]
	uploadID := form.Value["upload_id"][0]
	//parentID, err := strconv.ParseUint(form.Value["parent_id"][0], 10, 64)
	chunkIndex, err := strconv.ParseInt(fileIndex, 10, 64)
	if err != nil {
		log.Logger.Error("ParseInt 转换错误", zap.Error(err))
		makeResult(c, 200, err, nil)
		return
	}
	isExist, _, err := model.DealFileChunk(md5, chunkIndex, uploadID)
	if err != nil {
		log.Logger.Error("DealFileChunk error", zap.Error(err))
		makeResult(c, 200, err, nil)
		return
	}
	if !isExist {
		c.SaveUploadedFile(file, "./uploads/"+md5+"-"+fileIndex)
	}
	makeResult(c, 200, nil, nil)
}

func MergeFileChunks(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	form, _ := c.MultipartForm()
	fileName := form.Value["file_name"][0]
	md5 := form.Value["md5"][0]
	uploadID := form.Value["upload_id"][0]
	parentID, err := strconv.ParseUint(form.Value["parent_id"][0], 10, 64)
	// 若不存在，则合并
	chunkNum, err := model.MergeFileChunks(md5, uploadID)
	if err != nil {
		// 请重试
		makeResult(c, 200, err, nil)
		return
	}
	// 合并
	fi, err := os.Stat("./uploads/" + md5 + "-" + strconv.FormatInt(chunkNum-1, 10))
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	fileSize := (chunkNum-1)*(20*1024*1024) + fi.Size()
	err = model.AddFile(uid, md5, fileName, uint64(fileSize), parentID, true)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	if chunkNum == 0 {
		// 证明数据库中已包含该MD5值，有人先合并完成了，那就引用它
		makeResult(c, 200, nil, nil)
		return
	}
	allFile, _ := os.OpenFile("./uploads/"+md5, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	for i := int64(0); i < chunkNum; i++ {
		f, _ := os.OpenFile("./uploads/"+md5+"-"+strconv.FormatInt(i, 10), os.O_RDONLY, os.ModePerm)
		b, _ := ioutil.ReadAll(f)
		common.RemoveFile("./uploads/" + md5 + "-" + strconv.FormatInt(i, 10))
		allFile.Write(b)
	}
	makeResult(c, 200, nil, nil)
}

type UploadTaskRequest struct {
	FileName     string `json:"file_name"`
	FileMD5      string `json:"file_md5"`
	FileChunkNum int64  `json:"file_chunk_num"`
	FileSize     uint64 `json:"file_size"`
	TargetDirId  uint64 `json:"target_dir_id"`
}

type UploadTaskResponse struct {
	IsExist     bool   `json:"is_exist"`
	FinishQueue string `json:"finish_queue,omitempty"`
	UploadID    string `json:"upload_id,omitempty"`
}

func UploadTaskCreate(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data UploadTaskRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	isExists, chunks, uploadID, err := model.UploadPrepare(data.FileMD5, data.FileName, data.FileChunkNum, uid, data.FileSize, data.TargetDirId)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	resp := UploadTaskResponse{
		IsExist:     isExists,
		FinishQueue: chunks,
		UploadID:    uploadID,
	}
	makeResult(c, 200, nil, resp)
}

type MkdirRequest struct {
	FileName string `json:"file_name"`
	ParentID uint64 `json:"parent_id"`
}

func Mkdir(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
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
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
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
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data GetDownloadKeyRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	hashValue, err := common.HashEncode([]int{int(data.FileID)}, 6)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED), nil)
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
	key := c.Param("key")
	// 取fileID
	fileID, err := common.HashDecode(key[6:], 6)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_DECODE_FAILED), nil)
		return
	}
	fileName, fileMD5, err := model.DownloadFile(key, fileID)
	if err != nil {
		makeResult(c, 200, err, nil)
	}
	c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(fileName))
	c.File("./uploads/" + fileMD5)
}

type RemoveFilesRequest struct {
	FileIDList []uint64 `json:"file_id_list"`
}

func RemoveFiles(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
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
