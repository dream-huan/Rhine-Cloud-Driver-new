package model

type Group struct {
	GroupId         int64  // 用户组id
	GroupName       string // 用户组名
	GroupPermission        // 用户组权限
}

type GroupPermission struct {
	GroupStorage int64 // 用户组容量

}

// 考虑直接将云盘系统功能进行划分
// 用二进制位表示功能
// 访问云盘、上传、下载、分享、预览、进入管理界面、文件总览查看、分享总览查看、网站设置更改、权限控制
