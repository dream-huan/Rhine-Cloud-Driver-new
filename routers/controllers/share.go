package controllers

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/jwt"
	model "Rhine-Cloud-Driver/models"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type GetShareDetailRequest struct {
	ShareKey      string `json:"share_key"`
	SharePassword string `json:"share_password,omitempty"`
	SharePath     string `json:"share_path"`
}

type GetShareDetailResponse struct {
	Files     []FileSystem `json:"files,omitempty"`
	Name      string       `json:"name"`
	AvatarURL string       `json:"avatar_url"`
}

func GetShareDetail(c *gin.Context) {
	var data GetShareDetailRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 从shareKey还原出shareID
	shareID, err := common.HashDecode(data.ShareKey, 4)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_DECODE_FAILED), nil)
		return
	}
	name, uid, originFiles, err := model.GetShareDetail(shareID, data.SharePassword, data.SharePath)
	// 对邮箱进行MD5加密后返回给用户
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	//hash := md5.New()
	//hashValue := hex.EncodeToString(hash.Sum([]byte(email)))
	files := make([]FileSystem, len(originFiles))
	for i := range originFiles {
		files[i] = FileSystem{
			FileID:      originFiles[i].FileID,
			FileName:    originFiles[i].FileName,
			CreateTime:  originFiles[i].CreateTime,
			FileStorage: originFiles[i].FileStorage,
			IsDir:       originFiles[i].IsDir,
		}
	}
	makeResult(c, 200, nil, GetShareDetailResponse{
		Name:      name,
		AvatarURL: strconv.FormatUint(uid, 10),
		Files:     files,
	})
}

type CreateNewShareRequest struct {
	FileID       uint64 `json:"file_id"`
	Password     string `json:"password"`
	ExpireChoice int32  `json:"expire_choice"`
}

type CreateNewShareResponse struct {
	ShareKey string `json:"share_key"`
}

func CreateNewShare(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data CreateNewShareRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	ExpireTime := "-"
	if data.ExpireChoice == 1 {
		ExpireTime = time.Now().AddDate(0, 0, 7).Format("2006-01-02 15:04:05")
	} else if data.ExpireChoice == 2 {
		ExpireTime = time.Now().AddDate(0, 1, 0).Format("2006-01-02 15:04:05")
	}
	shareID, err := model.CreateShare(uid, data.FileID, ExpireTime, data.Password)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 将shareID转化为shareKey
	shareKey, err := common.HashEncode([]int{int(shareID)}, 4)
	if err != nil {
		// 转化失败
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED), nil)
		return
	}
	makeResult(c, 200, nil, CreateNewShareResponse{shareKey})
}

type TransferFilesRequest struct {
	TargetDirID   uint64   `json:"target_dir_id"`
	FileIDList    []uint64 `json:"file_id_list"`
	ShareKey      string   `json:"share_key"`
	SharePassword string   `json:"share_password"`
}

func TransferFiles(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data TransferFilesRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 从shareKey还原出shareID
	shareID, err := common.HashDecode(data.ShareKey, 4)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_DECODE_FAILED), nil)
		return
	}
	err = model.TransferFiles(uid, shareID, data.FileIDList, data.TargetDirID)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	makeResult(c, 200, nil, nil)
}

type CancelShareRequest struct {
	ShareKey string `json:"share_key"`
}

func CancelShare(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data CancelShareRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	// 从shareKey还原出shareID
	shareID, err := common.HashDecode(data.ShareKey, 4)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_DECODE_FAILED), nil)
		return
	}
	err = model.CancelShare(uid, shareID)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	makeResult(c, 200, nil, nil)
}

type GetMyShareResponse struct {
	Lists []ShareDetail
}

type ShareDetail struct {
	ShareKey      string `json:"share_key"`
	ExpireTime    string `json:"expire_time"`
	FileName      string `json:"file_name"`
	DownloadTimes uint64 `json:"download_times"`
	ViewTimes     uint64 `json:"view_times"`
	Password      string `json:"password"`
	IsDir         bool   `json:"is_dir"`
}

func GetMyShare(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	list := model.GetMyShare(uid)
	shareList := make([]ShareDetail, len(list))
	for i := range list {
		file, err := model.GetFileInfo(list[i].FileID, "all")
		if err != nil {
			makeResult(c, 200, err, nil)
			return
		}
		shareKey, err := common.HashEncode([]int{int(list[i].ShareID)}, 4)
		if err != nil {
			makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED), nil)
			return
		}
		shareList[i] = ShareDetail{
			FileName:      file.(model.File).FileName,
			ExpireTime:    list[i].ExpireTime,
			ShareKey:      shareKey,
			DownloadTimes: list[i].DownloadTimes,
			ViewTimes:     list[i].ViewTimes,
			Password:      list[i].Password,
			IsDir:         file.(model.File).IsDir,
		}
	}
	makeResult(c, 200, nil, GetMyShareResponse{shareList})
}

type GetShareFileRequest struct {
	FileID        uint64 `json:"file_id"`
	ShareKey      string `json:"share_key"`
	SharePassword string `json:"share_password"`
}

func GetShareFile(c *gin.Context) {
	var data GetShareFileRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	shareID, err := common.HashDecode(data.ShareKey, 4)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_DECODE_FAILED), nil)
		return
	}
	file, err := model.GetShareFile(shareID, data.SharePassword, data.FileID)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	fileKey, err := common.HashEncode([]int{int(file.FileID)}, 6)
	if err != nil {
		makeResult(c, 200, common.NewError(common.ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED), nil)
		return
	}
	downloadID, err := model.GetDownloadKey(file.Uid, file.FileID, fileKey)
	if err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	makeResult(c, 200, nil, GetDownloadKeyResponse{downloadID})
}
