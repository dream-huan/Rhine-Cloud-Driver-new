package model

import (
	"Rhine-Cloud-Driver/pkg/cache"
	"Rhine-Cloud-Driver/pkg/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Setting struct {
	gorm.Model
	Name  string `gorm:"unique;not null;index:setting_key"`
	Value string `gorm:"size:10240"`
}

// 获取所有基础设定
//func GetAllSetting() (SettingDetail Setting) {
//	DB.Table("setting").Find(&SettingDetail)
//
//}

func GetSettingByName(name string) string {
	var setting Setting
	// 先查缓存
	tempValue, isExist := cache.GetRedisKey("setting_" + name)
	if isExist == true {
		return tempValue.(string)
	}
	// 查数据库
	err := DB.Where("name = ?", &name).First(&setting).Error
	if err != nil {
		log.Logger.Error("database cannot find data from table setting: "+name, zap.Error(err))
		return ""
	}
	// 放到缓存中
	cache.SetRedisKey("setting_"+name, setting.Value, 0)
	return setting.Value
}
