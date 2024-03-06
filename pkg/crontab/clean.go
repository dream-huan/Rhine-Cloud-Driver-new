package crontab

import (
	model "Rhine-Cloud-Driver/models"
	"Rhine-Cloud-Driver/pkg/log"
	"Rhine-Cloud-Driver/pkg/util"
	"encoding/json"
	"strconv"
	"time"
)

// 部分文件上传失败后分块却仍然保留，当分块过长时间没有使用则被视为垃圾
// 超过一定天数的缩略图也被视为垃圾
func CleanChunksTrash() {
	log.Logger.Info("开始执行定时任务")
	// 时间距离今天超过清理的时间，则不能再使用
	expireTime, err := strconv.ParseInt(model.GetSettingByName("chunks_expire_time"), 10, 64)
	if err != nil {
		// 取过期时间失败，直接取默认两天
		expireTime = 60 * 60 * 24 * 2
	}
	dirPath := "./uploads/metadata/"
	files, isSuccess := util.GetDirAllFiles(dirPath)
	if isSuccess == false {
		return
	}
	for _, file := range files {
		bytes, isExist := util.ReadFile(dirPath + file.Name())
		if !isExist {
			continue
		}
		metaData := model.MetaData{}
		json.Unmarshal(bytes, &metaData)
		if int64(time.Now().Sub(metaData.LastUploaded).Seconds()) >= expireTime {
			// 获取metadata_后面的字段
			if len(file.Name()) >= 9 {
				util.RemoveMetaData(file.Name()[9:])
			}
		}
	}
}
