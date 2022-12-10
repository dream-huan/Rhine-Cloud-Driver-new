package model

type Group struct {
	GroupId         int64  `json:"group_id" gorm:"primarykey"` // 用户组id
	GroupName       string `json:"group_name" gorm:"size:255"` // 用户组名
	GroupPermission uint64 `json:"group_permission"`           // 用户组权限
	GroupStorage    int64  `json:"group_storage"`              // 用户组容量
}

// 考虑直接将云盘系统功能进行划分
// 用二进制位表示功能
// 访问云盘、上传、下载、分享、预览、进入管理界面、文件总览查看（修改）、分享总览查看（修改）、用户总览查看（修改）、网站设置更改、权限控制

// 将所有数据存到redis中
func InitGroupPermission() {

}
