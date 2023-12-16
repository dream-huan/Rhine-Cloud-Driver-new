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
		_, isExist := redis.GetRedisKey("groups_name_" + groupId)
		if isExist == false {
			return common.NewError(common.ERROR_GROUP_NOT_EXIST)
		}
		// 判定是否有权赋予这个groupId
		groupID, _ := strconv.ParseUint(groupId, 10, 64)
		if !ChangePermissionVerify(groupID, 0, 0, operatorUid) {
			return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
		}
		changedInfo["group_id"] = groupId
	}
	if storage != "" {
		changedInfo["total_storage"] = storage
	}
	DB.Table("users").Where("uid = ?", changedUid).Updates(changedInfo)
	return nil
}

func AdminGroupPermissionVerify(operatorUid uint64, groupPermission int64) error {
	//// 第一，操作人所在的用户组权限必须有修改后台或最高权限
	//if !PermissionVerify(operatorUid, PERMISSION_ADMIN_WRITE) {
	//	return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	//}
	// 第二，操作人所在的用户组权限必须比创建的要更高
	// 取操作人group_id
	var user User
	err := DB.Table("users").Where("uid = ?", operatorUid).Find(&user).Error
	userPermissionStr, isExist := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(user.GroupId, 10))
	if err != nil || user.Uid == 0 || isExist == false {
		return common.NewError(common.ERROR_USER_NOT_EXIST)
	}
	userPermission, _ := strconv.ParseInt(userPermissionStr.(string), 10, 64)
	if (userPermission&groupPermission < groupPermission) || (userPermission <= groupPermission) {
		return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	}
	return nil
}

func AdminCreateGroup(operatorUid uint64, groupName string, groupPermission int64, groupStorageStr string) error {
	err := AdminGroupPermissionVerify(operatorUid, groupPermission)
	if err != nil {
		return err
	}
	// 执行创建操作，在mysql和redis注册
	groupStorage, err := strconv.ParseUint(groupStorageStr, 10, 64)
	var newGroup Group
	newGroup = Group{
		GroupName:       groupName,
		GroupPermission: uint64(groupPermission),
		GroupStorage:    groupStorage,
	}
	DB.Table("groups").Create(&newGroup)
	groupID := strconv.FormatUint(newGroup.GroupId, 10)
	redis.SetRedisKey("groups_name_"+groupID, newGroup.GroupName, 0)
	redis.SetRedisKey("groups_permission_"+groupID, newGroup.GroupPermission, 0)
	redis.SetRedisKey("groups_storage_"+groupID, newGroup.GroupStorage, 0)
	return nil
}

func AdminEditGroupInfo(operatorUid, groupId uint64, groupName string, groupPermission int64, groupStorage string) error {
	err := AdminGroupPermissionVerify(operatorUid, groupPermission)
	if err != nil {
		return err
	}
	if !ChangePermissionVerify(groupId, 0, 0, operatorUid) {
		return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	}
	//  执行更改
	changedInfo := make(map[string]interface{})
	groupID := strconv.FormatUint(groupId, 10)
	if groupName != "" {
		changedInfo["group_name"] = groupName
		redis.SetRedisKey("groups_name_"+groupID, groupName, 0)
	}
	if groupPermission != 0 {
		changedInfo["group_permission"] = groupPermission
		redis.SetRedisKey("groups_permission_"+groupID, groupPermission, 0)
	}
	if groupStorage != "" {
		changedInfo["group_storage"] = groupStorage
		redis.SetRedisKey("groups_storage_"+groupID, groupStorage, 0)
	}
	DB.Table("groups").Where("group_id = ?", groupId).Updates(changedInfo)
	return nil
}

func AdminDeleteGroup(uid, groupID uint64) error {
	// 判断是否有权限删除
	var user User
	err := DB.Table("users").Where("uid = ?", uid).Find(&user).Error
	if err != nil || user.Uid == 0 {
		// 此用户不存在
		return common.NewError(common.ERROR_AUTH_UID_NOT_EXIST)
	}
	// 1和2不能被删除
	if groupID == 1 || groupID == 2 {
		return common.NewError(common.ERROR_GROUP_DEFAULT)
	}
	if !ChangePermissionVerify(groupID, user.GroupId, 0, 0) {
		return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	}
	// 将该用户组的全部用户移动为其他用户组
	// 默认用户组为注册用户，id为2
	// 当用户组变更时，他们的容量也跟着变化
	groupIDStr := strconv.FormatUint(groupID, 10)
	newGroupStorageStr, _ := redis.GetRedisKey("groups_storage_2")
	newGroupStorage, _ := strconv.ParseUint(newGroupStorageStr.(string), 10, 64)
	DB.Table("users").Where("group_id = ?", groupID).Updates(User{GroupId: 2, TotalStorage: newGroupStorage})
	// 数据库删除，redis注销
	DB.Table("groups").Delete(&Group{}, "group_id = ?", groupID)
	redis.DelRedisKey("groups_permission_" + groupIDStr)
	redis.DelRedisKey("groups_name_" + groupIDStr)
	redis.DelRedisKey("groups_storage_" + groupIDStr)
	return nil
}
