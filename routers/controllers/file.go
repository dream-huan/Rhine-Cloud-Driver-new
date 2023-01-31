package controllers

import (
	"Rhine-Cloud-Driver/logic/jwt"
	model "Rhine-Cloud-Driver/models"
	"github.com/gin-gonic/gin"
)

type GetFileSystemRequest struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
	Limit  int64  `json:"limit"`
}

type GetFileSystemResponse struct {
	Count int64        `json:"count"`
	Files []FileSystem `json:"files"`
}

type FileSystem struct {
	FileID      uint64 `json:"file_id"`
	FileName    string `json:"file_name,omitempty"`
	MD5         string `json:"md5,omitempty"`
	FileStorage uint64 `json:"file_storage,omitempty"`
	ParentID    uint64 `json:"parent_id,omitempty"`
	IsDir       bool   `json:"is_dir,omitempty"`
}

func GetMyFiles(c *gin.Context) {
	token, _ := c.Cookie("token")
	_, uid := jwt.TokenGetUid(token)
	var data GetFileSystemRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		makeResult(c, 200, err, nil)
		return
	}
	count, originFiles, err := model.BuildFileSystem(uid, data.Path, 50, 0)
	if err != nil {
		makeResult(c, 200, err, nil)
	}
	files := make([]FileSystem, len(originFiles))
	for i := range files {
		files[i] = FileSystem{
			FileID:      originFiles[i].FileID,
			FileName:    originFiles[i].FileName,
			MD5:         originFiles[i].MD5,
			FileStorage: originFiles[i].FileStorage,
			ParentID:    originFiles[i].ParentID,
			IsDir:       originFiles[i].IsDir,
		}
	}
	resp := GetFileSystemResponse{
		Count: count,
		Files: files,
	}
	makeResult(c, 200, nil, resp)
}

func Upload(c *gin.Context) {
	//token, _ := c.Cookie("token")
	//_, uid := jwt.TokenGetUid(token)
	//
}
