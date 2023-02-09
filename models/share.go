package model

import (
	"Rhine-Cloud-Driver/common"
	"crypto/md5"
	"encoding/hex"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Share struct {
	ShareID       uint64 `json:"share_id" gorm:"primaryKey;auto_increment"`
	Uid           uint64 `json:"uid" gorm:"index:idx_uid"`
	FileID        uint64 `json:"file_id"`
	ExpireTime    string `json:"expire_time"`
	CreateTime    string `json:"create_time"`
	Password      string `json:"password"`
	Valid         bool   `json:"valid"`
	ViewTimes     uint64 `json:"view_times"`
	DownloadTimes uint64 `json:"download_times"`
}

func CheckSharePathValid(path string, uid, fileID uint64) ([]File, error) {
	if path[0] != '/' {
		path = "/" + path
	}
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	shareFile := File{}
	err := DB.Table("files").Where("file_id=?", fileID).Find(&shareFile).Error
	// 分享的文件不存在
	if err != nil {
		return nil, common.NewError(common.ERROR_SHARE_NOT_EXIST)
	}
	if path == "/" {
		return []File{shareFile}, nil
	} else {
		//拿到previousPath
		previousPath := shareFile.Path[:len(shareFile.Path)-1]
		isValid, fileID := CheckPathValid(uid, path, previousPath)
		if isValid == false {
			return nil, common.NewError(common.ERROR_FILE_PATH_INVALID)
		}
		var files []File
		DB.Table("files").Where("parent_id=?", fileID).Find(&files)
		return files, nil
	}
}

func GetShareDetail(shareID uint64, password string, path string) (string, string, []File, error) {
	var ShareDetail Share
	err := DB.Table("shares").Where("share_id=? and valid = true", shareID).Find(&ShareDetail).Error
	if err != nil || (ShareDetail.ExpireTime != "-" && time.Now().Format("2006-01-02 15:04:05") >= ShareDetail.ExpireTime) {
		// 分享无效或已过期
		return "", "", nil, common.NewError(common.ERROR_SHARE_NOT_EXIST)
	}
	// 密码是否正确或无密码
	user := User{}
	DB.Table("users").Where("uid=?", ShareDetail.Uid).Find(&user)
	hash := md5.New()
	hashValue := hex.EncodeToString(hash.Sum([]byte(user.Email)))
	// 访问次数增加一次
	DB.Table("shares").Where("share_id=?", shareID).Update("view_times", gorm.Expr("view_times+1"))
	if ShareDetail.Password != "" && password != ShareDetail.Password {
		// 仅返回用户名称和头像信息，不返回文件系统
		if password != "" {
			return "", "", nil, common.NewError(common.ERROR_SHARE_PASSWORD_WRONG)
		}
		return user.Name, hashValue, nil, nil
	}
	// 返回全部信息
	files, err := CheckSharePathValid(path, ShareDetail.Uid, ShareDetail.FileID)
	if err != nil {
		return "", "", nil, err
	}
	return user.Name, hashValue, files, nil
}

func GetMyShare(uid uint64) (shareList []Share) {
	DB.Table("shares").Where("uid=? and valid=true and (now()<expire_time or expire_time='-')", uid).Find(&shareList)
	return
}

func TransferFiles(uid, shareID uint64, moveFileList []uint64, targetDirID uint64) error {
	tx := DB.Begin()
	// 拿到分享ID的文件ID
	var shareDetail Share
	err := tx.Table("shares").Where("share_id=? and valid=true and (now()<expire_time or expire_time='-')", shareID).Find(&shareDetail).Error
	if err != nil || shareDetail.FileID == 0 {
		// 该分享不存在或已失效
		tx.Rollback()
		return common.NewError(common.ERROR_SHARE_NOT_EXIST)
	}
	parentMap := make(map[uint64]uint64)
	targetDir := File{}
	err = tx.Table("files").Where("file_id=? and valid=true and is_dir=true", targetDirID).Find(&targetDir).Error
	if err != nil || targetDir.Uid != uid {
		// 目标文件夹不存在
		tx.Rollback()
		return common.NewError(common.ERROR_FILE_TARGETDIR_INVALID)
	}
	// 拿到用户信息
	user := User{}
	tx.Table("users").Where("uid=?", uid).Find(&user)
	allFilesStorage := uint64(0)
	for _, v := range moveFileList {
		thisFile := File{}
		err = tx.Table("files").Where("file_id=?", v).Find(&thisFile).Error
		if err != nil {
			// 这个文件不存在或它的祖先不存在
			tx.Rollback()
			return common.NewError(common.ERROR_FILE_NOT_EXISTS)
		}
		// 溯源，直到找到v的父亲为share的file_id或0为止，如果是0的话，证明要转存的文件不属于这个分享的文件，转存标记为失败
		parentID := v
		for parentID != 0 && parentID != shareDetail.FileID {
			file := File{}
			err = tx.Table("files").Where("file_id=?", parentID).Find(&file).Error
			parentID = file.ParentID
			if err != nil {
				// 这个文件不存在或它的祖先不存在
				tx.Rollback()
				return common.NewError(common.ERROR_FILE_NOT_EXISTS)
			}
		}
		if parentID == 0 {
			// 该文件不属于分享的文件
			tx.Rollback()
			return common.NewError(common.ERROR_FILE_INVALID)
		}

		// 目标文件夹是否有同名文件
		count := int64(0)
		tx.Table("files").Where("parent_id=? and file_name=? and valid=true and is_dir=?", targetDirID, thisFile.FileName, thisFile.IsDir).Count(&count)
		if count > 0 {
			tx.Rollback()
			return common.NewError(common.ERROR_FILE_TARGETDIR_SAME_FILES)
		}
		// 先建立自己
		newFile := File{
			Uid:         uid,
			FileName:    thisFile.FileName,
			ParentID:    targetDirID,
			FileStorage: thisFile.FileStorage,
			CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
			Valid:       true,
			IsDir:       thisFile.IsDir,
			MD5:         thisFile.MD5,
			Path:        targetDir.Path + targetDir.FileName + "/",
		}
		tx.Table("files").Create(&newFile)
		if thisFile.IsDir == true {
			parentMap[thisFile.FileID] = newFile.FileID
			// 拿到全部属于该前缀的文件
			var subFiles []File
			tx.Table("files").Where("uid = ? and path like ?", thisFile.Uid, thisFile.Path+thisFile.FileName+"/%").Order("is_dir").Find(&subFiles)
			for i := range subFiles {
				// 替换前缀
				newSubFile := File{
					Uid:         uid,
					FileName:    subFiles[i].FileName,
					FileStorage: subFiles[i].FileStorage,
					ParentID:    parentMap[subFiles[i].ParentID],
					CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
					Valid:       true,
					IsDir:       subFiles[i].IsDir,
					MD5:         subFiles[i].MD5,
					Path:        strings.Replace(subFiles[i].Path, thisFile.Path, newFile.Path, 1),
				}
				tx.Table("files").Create(&newSubFile)
				if subFiles[i].IsDir == true {
					parentMap[subFiles[i].FileID] = newSubFile.FileID
				} else {
					if user.UsedStorage+allFilesStorage > user.TotalStorage {
						tx.Rollback()
						return common.NewError(common.ERROR_USER_STORAGE_EXCEED)
					}
					allFilesStorage += subFiles[i].FileStorage
				}
			}
		} else {
			if user.UsedStorage+allFilesStorage > user.TotalStorage {
				tx.Rollback()
				return common.NewError(common.ERROR_USER_STORAGE_EXCEED)
			}
			allFilesStorage += newFile.FileStorage
		}
	}
	// 容量增加
	err = tx.Table("users").Where("uid=?", uid).Update("used_storage", gorm.Expr("used_storage+?", allFilesStorage)).Error
	if err != nil {
		tx.Rollback()
		return common.NewError(common.ERROR_USER_STORAGE_EXCEED)
	}
	tx.Commit()
	return nil
}

func CreateShare(uid, fileID uint64, ExpireTime string, password string) (uint64, error) {
	// 检查该fileID是否属于该uid
	var file File
	err := DB.Table("files").Where("file_id=?", fileID).Find(&file).Error
	if err != nil || file.Uid != uid {
		// 无权访问这些文件
		return 0, common.NewError(common.ERROR_SHARE_FILE_INVALID)
	}
	// 不允许重复分享同一个文件
	var count int64
	DB.Table("shares").Where("file_id = ? and valid = true and (now()<expire_time or expire_time='-')", fileID).Count(&count)
	if count > 0 {
		return 0, common.NewError(common.ERROR_SHARE_SAME_FILES)
	}
	newShare := Share{
		Uid:        uid,
		FileID:     fileID,
		ExpireTime: ExpireTime,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Password:   password,
		Valid:      true,
	}
	err = DB.Table("shares").Create(&newShare).Error
	if err != nil {
		return 0, common.NewError(common.ERROR_DB_WRITE_FAILED)
	}
	return newShare.ShareID, nil
}

func CancelShare(uid uint64, shareID uint64) error {
	// 检查该shareID是否属于该uid
	var shareDetail Share
	err := DB.Table("shares").Where("share_id=?", shareID).Find(&shareDetail).Error
	if err != nil || shareDetail.Uid != uid {
		// 无权访问这些文件
		return common.NewError(common.ERROR_SHARE_FILE_INVALID)
	}
	err = DB.Table("shares").Where("share_id=?", shareID).Update("valid", 0).Error
	if err != nil {
		return common.NewError(common.ERROR_DB_WRITE_FAILED)
	}
	return nil
}

func GetShareFile(shareID uint64, password string, fileID uint64) (file File, err error) {
	var shareDetail Share
	err = DB.Table("shares").Where("share_id = ? and valid = true and (now()<expire_time or expire_time='-')", shareID).Find(&shareDetail).Error
	if err != nil || password != shareDetail.Password {
		return File{}, common.NewError(common.ERROR_SHARE_PASSWORD_WRONG)
	}
	DB.Table("shares").Where("share_id = ?", shareID).Update("download_times", gorm.Expr("download_times + 1"))
	var parentFile File
	err = DB.Table("files").Where("file_id = ?", shareDetail.FileID).Find(&parentFile).Error
	if err != nil || parentFile.FileID == 0 {
		return File{}, common.NewError(common.ERROR_FILE_NOT_EXISTS)
	}
	if fileID == parentFile.FileID {
		return parentFile, nil
	}
	err = DB.Table("files").Where("file_id = ?", fileID).Find(&file).Error
	if err != nil || file.FileID == 0 {
		return File{}, common.NewError(common.ERROR_FILE_NOT_EXISTS)
	}
	parentFilePath := parentFile.Path + parentFile.FileName + "/"
	if file.Uid == parentFile.Uid && len(file.Path) >= len(parentFilePath) && parentFilePath == file.Path[:len(parentFilePath)] {
		return file, nil
	}
	return File{}, common.NewError(common.ERROR_DOWNLOAD_FILE_INVALID)
}
