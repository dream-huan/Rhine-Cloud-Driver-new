package model

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/logic/redis"
	"strconv"
)

type Group struct {
	GroupId         uint64 `json:"group_id" gorm:"primaryKey;auto_increment"` // 用户组id
	GroupName       string `json:"group_name" gorm:"size:255"`                // 用户组名
	GroupPermission uint64 `json:"group_permission"`                          // 用户组权限
	GroupStorage    uint64 `json:"group_storage"`                             // 用户组容量
}

// 权限
const (
	PERMISSION_ACCESS      = iota // 访问网站
	PERMISSION_SETTING            // 更改个人设置
	PERMISSION_SHARE              // 分享
	PERMISSION_FILE               // 文件
	PERMISSION_ADMIN_READ         // 访问后台
	PERMISSION_ADMIN_WRITE        // 管理后台
	PERMISSION_MAXIMUM            // 最高权限
)

// 考虑直接将云盘系统功能进行划分
// 用二进制位表示功能
// 访问云盘、上传、下载、分享、删除、进入管理界面、文件总览查看（修改）、分享总览查看（修改）、用户总览查看（修改）、用户组总览查看
// 访问、个人设置、分享、文件、管理（读）权限、管理（写）权限、全局最高权限
// 0000000
// 1 2 4 8 16 32 64

// 将所有数据存到redis中
func InitGroupPermission() {
	var count int64
	DB.Table("groups").Count(&count)
	if count <= 0 {
		// 新建两个用户组，一个为普通用户10G内存、一个为管理员1T内存
		DB.Table("groups").Create(&Group{
			GroupId:         1,
			GroupName:       "管理员",
			GroupPermission: 127,
			GroupStorage:    1024 * 1024 * 1024 * 1024,
		})
		DB.Table("groups").Create(&Group{
			GroupId:         2,
			GroupName:       "注册用户",
			GroupPermission: 15,
			GroupStorage:    1024 * 1024 * 1024 * 10,
		})
		// 新建一个管理员用户
		var adminUser = User{
			Uid:      1,
			Name:     "初始网站拥有者账户",
			Password: "123456",
			Email:    "admin@admin.com",
			GroupId:  1,
		}
		adminUser.AddUser()
	}
	// 储存用户组信息到redis中
	var allGroups []Group
	DB.Table("groups").Find(&allGroups)
	for _, v := range allGroups {
		groupID := strconv.FormatUint(v.GroupId, 10)
		redis.SetRedisKey("groups_name_"+groupID, v.GroupName, 0)
		redis.SetRedisKey("groups_permission_"+groupID, v.GroupPermission, 0)
		redis.SetRedisKey("groups_storage_"+groupID, v.GroupStorage, 0)
	}
}

func ChangeUsersGroup(changedUid, operatorUid, newGroupId uint64) error {
	// 是否有权限，用户是否存在
	var changedUser User
	var operatorUser User
	err := DB.Table("users").Where("uid = ?", changedUid).Find(&changedUser).Error
	if err != nil || changedUser.Uid == 0 {
		// 此用户不存在
		return common.NewError(common.ERROR_AUTH_UID_NOT_EXIST)
	}
	err = DB.Table("users").Where("uid = ?", operatorUid).Find(&operatorUser).Error
	if err != nil || changedUser.Uid == 0 {
		// 此用户不存在
		return common.NewError(common.ERROR_AUTH_UID_NOT_EXIST)
	}
	changeUserPermission := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(changedUid, 10))
	operatorUserPermission := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(operatorUid, 10))
	// 鉴定操作者有网站修改权限
	isok := (operatorUserPermission.(int) & (1 << PERMISSION_ADMIN_WRITE)) | (operatorUserPermission.(int) & (1 << PERMISSION_MAXIMUM))
	if isok <= 0 || changeUserPermission.(int) >= operatorUserPermission.(int) {
		return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	}
	// 数据库执行更改
	DB.Table("users").Where("uid = ?", changedUid).Update("group_id", newGroupId)
	return nil
}

func ChangeGroupInfo(uid uint64, groupID uint64, changedInfo Group) error {
	// 是否有权限，用户组是否存在
	var user User
	err := DB.Table("users").Where("uid = ?", uid).Find(&user).Error
	if err != nil || user.Uid == 0 {
		// 此用户不存在
		return common.NewError(common.ERROR_AUTH_UID_NOT_EXIST)
	}
	userPermission := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(user.GroupId, 10))
	groupPermission := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(groupID, 10))
	// 鉴定操作者有网站修改权限
	isok := (userPermission.(int) & (1 << PERMISSION_ADMIN_WRITE)) | (userPermission.(int) & (1 << PERMISSION_MAXIMUM))
	if isok <= 0 || groupPermission.(int) >= userPermission.(int) {
		return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	}
	// 数据库更改
	// redis更改
	if changedInfo.GroupPermission > 0 {
		DB.Table("groups").Where("group_id = ?", groupID).Update("group_permission", changedInfo.GroupPermission)
		redis.SetRedisKey("groups_permission_"+strconv.FormatUint(groupID, 10), changedInfo.GroupPermission, 0)
	}
	if changedInfo.GroupName != "" {
		DB.Table("groups").Where("group_id = ?", groupID).Update("group_name", changedInfo.GroupName)
		redis.SetRedisKey("groups_name_"+strconv.FormatUint(groupID, 10), changedInfo.GroupName, 0)
	}
	if changedInfo.GroupStorage > 0 {
		DB.Table("groups").Where("group_id = ?", groupID).Update("group_storage", changedInfo.GroupStorage)
		redis.SetRedisKey("groups_storage_"+strconv.FormatUint(groupID, 10), changedInfo.GroupStorage, 0)
	}
	return nil
}

func GetGroupInfo(groupID uint64) (groupDetail Group, err error) {
	err = DB.Table("groups").Where("group_id = ?", groupID).Find(&groupDetail).Error
	if err != nil {
		return groupDetail, common.NewError(common.ERROR_GROUP_NOT_EXIST)
	}
	return
}

func GetAllGroup() (groups []Group) {
	DB.Table("groups").Find(&groups)
	return
}

func DelGroup(uid, groupID uint64) error {
	// 判断是否有权限删除
	var user User
	err := DB.Table("users").Where("uid = ?", uid).Find(&user).Error
	if err != nil || user.Uid == 0 {
		// 此用户不存在
		return common.NewError(common.ERROR_AUTH_UID_NOT_EXIST)
	}
	userPermission := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(user.GroupId, 10))
	groupPermission := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(groupID, 10))
	// 鉴定操作者有网站修改权限
	isok := (userPermission.(int) & (1 << PERMISSION_ADMIN_WRITE)) | (userPermission.(int) & (1 << PERMISSION_MAXIMUM))
	if isok <= 0 || groupPermission.(int) >= userPermission.(int) {
		return common.NewError(common.ERROR_AUTH_NOT_PERMISSION)
	}
	// 将该用户组的全部用户移动为其他用户组
	// 默认用户组为注册用户，id为2
	// 当用户组变更时，他们的容量也跟着变化
	groupIDStr := strconv.FormatUint(groupID, 10)
	newGroupStorage := redis.GetRedisKey("groups_storage_" + groupIDStr)
	DB.Table("users").Where("group_id = ?", groupID).Updates(User{GroupId: groupID, TotalStorage: newGroupStorage.(uint64)})
	// 数据库删除，redis注销
	DB.Table("groups").Delete("group_id = ?", groupID)
	redis.DelRedisKey("groups_permission_" + groupIDStr)
	redis.DelRedisKey("groups_name_" + groupIDStr)
	redis.DelRedisKey("groups_storage_" + groupIDStr)
	return nil
}

func PermissionVerify(uid uint64, permissionCode int) bool {
	var user User
	// 拿到用户的用户组ID
	DB.Table("users").Where("uid = ?", uid).Find(&user)
	// 取得用户的权限
	userPermission := redis.GetRedisKey("groups_permission_" + strconv.FormatUint(user.GroupId, 10))
	if userPermission == nil {
		return false
	}
	return userPermission.(int)&(1<<permissionCode) >= (1<<permissionCode) || userPermission.(int)&(1<<PERMISSION_MAXIMUM) >= (1<<PERMISSION_MAXIMUM)
}
