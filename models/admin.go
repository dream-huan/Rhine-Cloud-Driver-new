package model

func GetAllUser(offset, limit int) (count int64, users []User) {
	// todo:searchKey搜索uid和邮箱
	DB.Table("users").Count(&count)
	DB.Table("users").Offset(offset).Limit(limit).Find(&users)
	return
}

func GetAllShare(offset, limit int) (count int64, shares []Share) {
	DB.Table("shares").Where("valid=true and (now()<expire_time or expire_time='-')").Count(&count)
	DB.Table("shares").Where("valid=true and (now()<expire_time or expire_time='-')").Offset(offset).Limit(limit).Find(&shares)
	return
}
