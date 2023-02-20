package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/redis"
	"strconv"
)

type WebData struct {
}

func GetAllUser(offset, limit int) (count int64, users []User) {
	// todo:searchKey搜索uid和邮箱
	DB.Table("users").Count(&count)
	DB.Table("users").Offset(offset).Limit(limit).Find(&users)
	return
}

func AdminGetUserDetail(uid uint64) (err error, user User) {
	user.Uid = uid
	user.GetUserDetail()
	if user.Uid == 0 {
		return common.NewError(common.ERROR_USER_NOT_EXIST), User{}
	}
	return
}

func GetAllShare(offset, limit int) (count int64, shares []Share) {
	DB.Table("shares").Where("valid=true and (now()<expire_time or expire_time='-')").Count(&count)
	DB.Table("shares").Where("valid=true and (now()<expire_time or expire_time='-')").Offset(offset).Limit(limit).Find(&shares)
	return
}

func GetAllGroup(offset, limit int) (count int64, groups []Group) {
	DB.Table("groups").Count(&count)
	if limit == 0 {
		DB.Table("groups").Find(&groups)
	} else {
		DB.Table("groups").Offset(offset).Limit(limit).Find(&groups)
	}
	return
}

func GetAllFile(offset, limit int) (count int64, files []File) {
	DB.Table("files").Where("is_dir = false and valid = true and is_origin = true").Count(&count)
	DB.Table("files").Where("is_dir = false and valid = true and is_origin = true").Offset(offset).Limit(limit).Find(&files)
	return
}

func AdminEditUserInfo(changedUid, operatorUid uint64, name, password, groupId, storage string) error {
	// 操作者权限必须高于被更改者
	// 取两者的group信息
	if !ChangePermissionVerify(0, 0, changedUid, operatorUid) {
		return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	}
	changedInfo := make(map[string]interface{})
	if name != "" {
		if !checkNewName(name) {
			return common.NewError(common.ERROR_USER_NAME_LENGTH_NOT_MATCH)
		}
		changedInfo["name"] = name
	}
	if password != "" {
		if !checkNewPassword(password) {
			return common.NewError(common.ERROR_USER_PASSWORD_NOT_MATCH_RULES)
		}
		changedInfo["password"] = setHaltHash(password)
	}
	if groupId != "0" && groupId != "" {
		// 判断指定的group_id是否存在
		// 取redis
		if redis.GetRedisKey("groups_name_"+groupId) == nil {
			return common.NewError(common.ERROR_GROUP_NOT_EXIST)
		}
		// 判定是否有权赋予这个groupId
		groupID, _ := strconv.ParseUint(groupId, 10, 64)
		if !ChangePermissionVerify(groupID, 0, 0, operatorUid) {
			return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
		}
		changedInfo["group_id"] = groupId
	}
	if storage != "0" && storage != "" {
		changedInfo["total_storage"] = storage
	}
	DB.Table("users").Where("uid = ?", changedUid).Updates(changedInfo)
	return nil
}
