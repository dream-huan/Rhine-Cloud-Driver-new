package crontab

import (
	model "Rhine-Cloud-Driver/models"
	"github.com/robfig/cron/v3"
)

var Cron = &cron.Cron{}

// 定时清理缩略图或其他临时文件

func Init() {
	Cron = cron.New()
	cleanTime := model.GetSettingByName("trash_chunks_clean_time")
	if cleanTime == "" {
		// 取默认每天0点清理一次
		cleanTime = "0 0 * * *"
	}
	Cron.AddFunc(cleanTime, CleanChunksTrash)
	Cron.Start()
}
