package model

type Share struct {
	ShareID    uint64 `json:"share_id" gorm:"primary key"`
	Uid        uint64 `json:"uid" gorm:"index:idx_uid"`
	FileID     uint64 `json:"file_id"`
	ExpireTime string `json:"expire_time"`
	CreateTime string `json:"create_time"`
	Password   string `json:"password"`
}

func GetShareDetail() {

}

func GetMyShare() {

}

func TransferFiles() {

}

func CancelShare() {

}
