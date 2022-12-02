package model

type Group struct {
	GroupId         int64  // 用户组id
	GroupName       string // 用户组名
	GroupPermission        // 用户组权限
}

type GroupPermission struct {
	GroupStorage int64 // 用户组容量

}
