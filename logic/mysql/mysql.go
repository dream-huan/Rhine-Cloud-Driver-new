// package mysql

// import (
// 	"database/sql"
// 	"os"

// 	"Rhine-Cloud-Driver/common"
// 	"Rhine-Cloud-Driver/config"
// 	log "Rhine-Cloud-Driver/logic/log"
// 	model "Rhine-Cloud-Driver/models"
// 	_ "github.com/go-sql-driver/mysql"
// 	"go.uber.org/zap"
// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"
// )

// var gormDB *gorm.DB
// var db *sql.DB

// var pwd, _ = os.Getwd()
// var baseURL = pwd + "/upload/"

// func Init(cf config.Config) {
// 	// var err error
// 	// // dsn := "root:SUIbianla123@tcp(127.0.0.1:3306)/project"
// 	dsn := cf.MysqlManager.User + ":" + cf.MysqlManager.Password + "@tcp(" + cf.Server.Host + ")/" + cf.MysqlManager.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
// 	var err error
// 	gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	// db, err := sql.Open("mysql", dsn)
// 	if err != nil {
// 		log.Logger.Error("数据库链接错误", zap.Error(err))
// 	}
// 	log.Logger.Info("MySQL数据库链接成功")
// 	db, err = gormDB.DB()
// 	if err != nil {
// 		log.Logger.Error("获取数据库DB错误", zap.Error(err))
// 	}
// 	db.SetMaxOpenConns(100)
// }

// // func init() {
// // 	var err error
// // 	dsn := "root:SUIbianla123@tcp(127.0.0.1:3306)/project"
// // 	db, err = sql.Open("mysql", dsn)
// // 	if err != nil {
// // 		log.Logger.Error("数据库链接错误", err)
// // 		fmt.Printf("%#v", err)
// // 	}
// // 	db.SetMaxOpenConns(100)
// // }

// func AddUser(uid string, password string, email string) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误错误", zap.Error(err))
// 		return false
// 	}
// 	sqlStr := "insert into users(uid,password,email,create_time) values(?,?,?,NOW())"
// 	_, err = tx.Exec(sqlStr, uid, password, email)
// 	if err != nil {
// 		tx.Rollback()
// 		log.Logger.Error("事务执行错误", zap.Error(err))
// 		return false
// 	}
// 	sqlStr = "insert into storage values(?,?,?)"
// 	_, err = tx.Exec(sqlStr, uid, 0, 1073741824) //新用户分配1G内存
// 	if err != nil {
// 		tx.Rollback()
// 		log.Logger.Error("事务执行错误", zap.Error(err))
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func GetInfo(uid string) (info model.User) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return
// 	}
// 	sqlStr := "select * from users where uid=?"
// 	_ = tx.QueryRow(sqlStr, uid).Scan(&info.Uid, &info.Password, &info.Email, &info.CreateTime)
// 	sqlStr = "select * from storage where uid=?"
// 	_ = tx.QueryRow(sqlStr, uid).Scan(&info.Uid, &info.UsedStorage, &info.TotalStorage)
// 	tx.Commit()
// 	return info
// }

// func VerifyPassword(uid int64, password string) bool {
// 	// 密码不用事务
// 	// 先校验此用户是否存在，再提取密码
// 	var count int64
// 	gormDB.Table("users").Where("uid", uid).Count(&count)
// 	if count == 0 {
// 		return false
// 	}
// 	var actualPassword string
// 	gormDB.Table("users").Where("uid", uid).Find(&actualPassword)
// 	return actualPassword == password
// }

// func EditPassword(uid string, password string) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	sqlStr := "update users set password=? where uid=?"
// 	_, err = tx.Exec(sqlStr, password, uid)
// 	if err != nil {
// 		tx.Rollback()
// 		log.Logger.Error("事务执行错误", zap.Error(err))
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func AddLoginRecord(uid, ip string) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return
// 	}
// 	//fmt.Printf("%#v %#v", uid, ip)
// 	sqlStr := "insert into iprecord(uid,ip,time,city) values(?,?,NOW(),?)"
// 	_, err = tx.Exec(sqlStr, uid, ip, Geoip2.IpQueryCity(ip))
// 	if err != nil {
// 		tx.Rollback()
// 		log.Logger.Error("事务执行错误", zap.Error(err))
// 		return
// 	}
// 	tx.Commit()
// }

// func GetIpRecord(uid string) (iprecord []Class.IpRecord) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return
// 	}
// 	rows, _ := tx.Query("select uid,time,ip,city from iprecord where uid=? ORDER BY time DESC", uid)
// 	for rows.Next() {
// 		var temp Class.IpRecord
// 		rows.Scan(&temp.Uid, &temp.Time, &temp.Ip, &temp.City)
// 		iprecord = append(iprecord, temp)
// 	}
// 	tx.Commit()
// 	return iprecord
// }

// func Md5Query(md5 string) (fileid, filesize int64) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return
// 	}
// 	sqlStr := "select fileid,filesize from file where md5=?"
// 	_ = db.QueryRow(sqlStr, md5).Scan(&fileid, &filesize)
// 	tx.Commit()
// 	return fileid, filesize
// }

// func CheckIsExistSame(filename string, parentid int64) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	num := 0
// 	sqlStr := "select count(*) from file where oldfilename=? and parentid=?"
// 	_ = tx.QueryRow(sqlStr).Scan(&num)
// 	if num >= 1 {
// 		return true
// 	}
// 	return false
// }

// func AddOldFile(uid, filename, md5, path string, fileid, size, parentid int64) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	sqlStr := "insert into file(create_time,uid,path,isdir,md5,parentid,originid,filename,filesize,oldfilename) values(NOW(),?,?,0,'null',?,?,?,?,?)"
// 	_, err = tx.Exec(sqlStr, uid, path, parentid, fileid, filename, size, filename)
// 	if err != nil {
// 		tx.Rollback()
// 		log.Logger.Error("事务执行错误", zap.Error(err))
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func AddFile(uid, filename, md5, path, oldfilename string, size, parentid int64) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	sqlStr := "insert into file(create_time,uid,path,isdir,md5,parentid,originid,filename,filesize,oldfilename) values(NOW(),?,?,0,?,?,-1,?,?,?)"
// 	_, err = tx.Exec(sqlStr, uid, path, md5, parentid, filename, size, oldfilename)
// 	if err != nil {
// 		tx.Rollback()
// 		log.Logger.Error("事务执行错误", zap.Error(err))
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func JudgeStorage(uid string, size int64) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	sqlStr := "select usedstorage,totalstorage from storage where uid=?"
// 	var usedstorage int64
// 	var totalstorage int64
// 	_ = db.QueryRow(sqlStr, uid).Scan(&usedstorage, &totalstorage)
// 	if usedstorage+size > totalstorage {
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func ChangeStorage(uid string, size int64, symbol string) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	sqlStr := "update storage set usedstorage=usedstorage" + symbol + "? where uid=?"
// 	_, err = tx.Exec(sqlStr, size, uid)
// 	if err != nil {
// 		tx.Rollback()
// 		log.Logger.Error("事务执行错误", zap.Error(err))
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func GetFile(fileid int64) (temp Class.FileSystem) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return
// 	}
// 	sqlStr := "select * from file where fileid=?"
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&temp.FileId, &temp.Create_time, &temp.Uid, &temp.Path, &temp.Isdir, &temp.Md5, &temp.Parentid, &temp.Originid, &temp.FileName, &temp.FileSize, &temp.Valid, &temp.OldFileName)
// 	tx.Commit()
// 	return temp
// }

// //func GetShareFileDirectory(parentid int64) (files []Class.FileSystem) {
// //	tx, _ := db.Begin()
// //	rows, err := tx.Query("select * from file where parentid=?", parentid)
// //	if err != nil {
// //		tx.Rollback()
// //		return
// //	}
// //	for rows.Next() {
// //		var temp Class.FileSystem
// //		rows.Scan(&temp.FileId, &temp.Create_time, &temp.Uid, &temp.Path, &temp.Isdir, &temp.Md5, &temp.Parentid, &temp.Originid, &temp.FileName, &temp.FileSize, &temp.Valid, &temp.OldFileName)
// //		if temp.Valid == true {
// //			files = append(files, temp)
// //		}
// //	}
// //	tx.Commit()
// //	return files
// //}

// func GetFileDirectory(uid string, parentid int64) (files []Class.FileSystem) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return
// 	}
// 	//todo:取文件目录是否需要依赖于uid？考虑是否删除此条件,多个函数都应当考虑，这将牵涉uid是否建立索引
// 	rows, _ := tx.Query("select * from file where uid=? and parentid=?", uid, parentid)
// 	for rows.Next() {
// 		var temp Class.FileSystem
// 		rows.Scan(&temp.FileId, &temp.Create_time, &temp.Uid, &temp.Path, &temp.Isdir, &temp.Md5, &temp.Parentid, &temp.Originid, &temp.FileName, &temp.FileSize, &temp.Valid, &temp.OldFileName)
// 		if temp.Valid == true {
// 			files = append(files, temp)
// 		}
// 	}
// 	tx.Commit()
// 	return files
// }

// func SplitPath(uid, filename string, paths []string, parentid int64) (fileid int64) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return
// 	}
// 	sqlStr := "select fileid from file where uid=? and parentid=? and filename=?"
// 	_ = tx.QueryRow(sqlStr, uid, parentid, filename).Scan(&fileid)
// 	if fileid == 0 {
// 		tx.Rollback()
// 		return -1
// 	}
// 	for i := 0; i < len(paths); i = i + 1 {
// 		parentid = fileid
// 		//parentid = Mysql.SplitPath(uid, paths[i], parentid)
// 		sqlStr := "select fileid from file where uid=? and parentid=? and filename=?"
// 		_ = tx.QueryRow(sqlStr, uid, parentid, paths[i]).Scan(&fileid)
// 		if fileid == 0 {
// 			tx.Rollback()
// 			return -1
// 		}
// 	}
// 	tx.Commit()
// 	return fileid
// }

// func VerifyFileOwnership(uid string, fileid int64) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	var fileuid string
// 	sqlStr := "select uid from file where fileid=?"
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&fileuid)
// 	tx.Commit()
// 	return fileuid == uid
// }

// func MovePath(fileid, parentid int64, uid string) bool {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Logger.Error("事务启动错误", zap.Error(err))
// 		return false
// 	}
// 	//var oldpath string
// 	//var originid int64
// 	//var filename string
// 	//var parentid int64
// 	////sqlStr := "select path,originid,filename from file where fileid=?"
// 	////_ = tx.QueryRow(sqlStr, fileid).Scan(&oldpath, &originid, &filename)
// 	////sqlStr = "select fileid from file where isdir=true and path=? and uid=?"
// 	////_ = tx.QueryRow(sqlStr, newpath, uid).Scan(&parentid)
// 	////sqlStr = "update file set path=?,parentid=? where fileid=?"
// 	////_, _ = tx.Exec(sqlStr, newpath, parentid, fileid)
// 	//只需要将文件归并到新目录的parentid下即可,唯一需要确认的就是，是否有重名，如果有重名就不能直接移动
// 	var filename string
// 	var isdir bool
// 	var fileuid string
// 	sqlStr := "select filename,isdir,uid from file where fileid=?"
// 	tx.QueryRow(sqlStr, fileid).Scan(&filename, &isdir, &fileuid)
// 	if uid != fileuid {
// 		//非本人文件，无权操作
// 		tx.Rollback()
// 		return false
// 	}
// 	//根据得出的filename和parentid判断是否有重复
// 	var num int64
// 	sqlStr = "select count(*) from file where parentid=? and filename=? and isdir=?"
// 	tx.QueryRow(sqlStr, parentid, filename, isdir).Scan(&num)
// 	if num > 0 {
// 		//有重复
// 		tx.Rollback()
// 		return false
// 	}
// 	//最后，不允许自己移动自己
// 	if parentid == fileid {
// 		tx.Rollback()
// 		return false
// 	}
// 	//直接更新即可
// 	sqlStr = "update file set parentid=? where fileid=?"
// 	tx.Exec(sqlStr, parentid, fileid)
// 	tx.Commit()
// 	return true
// }

// func CopyPath(fileid, parentid int64, uid string) bool {
// 	//将前者复制到后者
// 	tx, _ := db.Begin()
// 	isok := true
// 	if Isdir(fileid) {
// 		//先将自己复制一份
// 		sqlStr := "select filename from file where fileid=?"
// 		var filename string
// 		_ = tx.QueryRow(sqlStr, fileid).Scan(&filename)
// 		sqlStr = "insert into file(create_time,uid,path,isdir,md5,parentid,originid,filename,filesize,oldfilename) values(NOW(),?,?,1,'null',?,-1,?,0,?)"
// 		res, err := tx.Exec(sqlStr, uid, "/", parentid, filename, filename)
// 		if err != nil {
// 			tx.Rollback()
// 			return false
// 		}
// 		parentid, err = res.LastInsertId()
// 		if err != nil {
// 			tx.Rollback()
// 			return false
// 		}
// 		//将其所属的文件列举出来，然后依次复制到新的用户手中
// 		//先复制文件，再复制文件夹
// 		sqlStr = "select fileid,isdir from file where parentid=?"
// 		rows, err := tx.Query(sqlStr, fileid)
// 		defer rows.Close()
// 		if err != nil {
// 			return false
// 		}
// 		var files []int64
// 		var dirs []int64
// 		for rows.Next() {
// 			var tfileid int64
// 			var tisdir bool
// 			rows.Scan(&tfileid, &tisdir)
// 			if tisdir == true {
// 				dirs = append(dirs, tfileid)
// 			} else {
// 				files = append(files, tfileid)
// 			}
// 		}
// 		for _, v := range files {
// 			CopyPath(v, parentid, uid)
// 		}
// 		for _, v := range dirs {
// 			//获取此文件夹信息，建立一个一模一样的文件夹，然后递归复制
// 			//sqlStr = "select filename from file where fileid=?"
// 			//var filename string
// 			//_ = tx.QueryRow(sqlStr, v).Scan(&filename)
// 			//sqlStr = "insert into file(create_time,uid,path,isdir,md5,parentid,originid,filename,filesize,oldfilename) values(NOW(),?,?,1,'null',?,-1,?,0,?)"
// 			//res, err := tx.Exec(sqlStr, uid, "/", parentid, filename, filename)
// 			//if err != nil {
// 			//	tx.Rollback()
// 			//	return false
// 			//}
// 			//newparentid, _ := res.LastInsertId()
// 			if CopyPath(v, parentid, uid) == false {
// 				isok = false
// 			}
// 		}
// 	} else {
// 		//复制文件就行
// 		//查询来源，如果是本源直接复制，不是查询源头复制
// 		originfileid := GetOriginFileId(fileid)
// 		sqlStr := "select filename,filesize,oldfilename from file where fileid=?"
// 		var filename string
// 		var oldfilename string
// 		var filesize int64
// 		_ = tx.QueryRow(sqlStr, originfileid).Scan(&filename, &filesize, &oldfilename)
// 		sqlStr = "insert into file(create_time,uid,path,isdir,md5,parentid,originid,filename,filesize,oldfilename) values(NOW(),?,?,0,'null',?,?,?,?,?)"
// 		_, err := tx.Exec(sqlStr, uid, "/", parentid, originfileid, filename, filesize, oldfilename)
// 		if err != nil {
// 			tx.Rollback()
// 			return false
// 		}
// 		ChangeStorage(uid, filesize, "+")
// 	}
// 	tx.Commit()
// 	return isok
// }

// func GetOriginFileId(fileid int64) (originfileid int64) {
// 	tx, _ := db.Begin()
// 	sqlStr := "select originid from file where fileid=?"
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&originfileid)
// 	if originfileid == -1 {
// 		originfileid = fileid
// 	}
// 	tx.Commit()
// 	return originfileid
// }

// func GetFilePath(fileid int64) (filename string) {
// 	var uid string
// 	tx, _ := db.Begin()
// 	//先判断此文件是否有来源，如果没有直接返回路径，如果有，则返回来源的路径
// 	sqlStr := "select uid,originid,filename from file where fileid=?"
// 	var originid int64
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&uid, &originid, &filename)
// 	//不是-1证明有其他源文件
// 	if originid != -1 {
// 		//找源文件的路径
// 		sqlStr = "select uid,filename from file where fileid=?"
// 		//_ = db.QueryRow(sqlStr, originid).Scan(&uid, &path, &filename)
// 		_ = tx.QueryRow(sqlStr, originid).Scan(&uid, &filename)
// 		tx.Commit()
// 		return uid + "/" + filename
// 	} else {
// 		tx.Commit()
// 		return uid + "/" + filename
// 	}
// }

// //todo:加事务进行处理，全部命令都如此

// func FindPath(uid string, fileid, parentid int64) bool {
// 	tx, _ := db.Begin()
// 	var filename string
// 	sqlStr := "select filename from file where fileid=?"
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&filename)
// 	var num int64
// 	sqlStr = "select count(*) from file where parentid=? and uid=? and isdir=true and filename=?"
// 	_ = tx.QueryRow(sqlStr, parentid, uid, filename).Scan(&num)
// 	tx.Commit()
// 	return num > 0
// }

// func DeleteFile(uid string, fileid int64) bool {
// 	//首先需要验证，这个文件是否所属这个用户，然后我们需要进行处理
// 	//第一种可能：该文件属于该用户，并且该文件为非源文件，直接删除该项记录即可
// 	//第二种可能：该文件不属于该用户，无论是否文件为源文件一律不予处理
// 	//第三种可能，该文件属于该用户，并且该文件为源文件，此时不删除此文件，再细分为以下两种情况
// 	//如果该源文件无人进行引用，则删除此文件
// 	//如果该源文件有人引用，则仅将文件uid置为-1即可
// 	//所有删除操作仅对表做处理，暂不对文件做处理，具体对文件的处理传达给消息队列，由它们做判定后续处理
// 	tx, _ := db.Begin()
// 	sqlStr := "select uid,originid,filesize from file where fileid=? for update"
// 	var originid int64
// 	var fileuid string
// 	var filename string
// 	var filesize int64
// 	var nums int64
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&fileuid, &originid, &filesize)
// 	//非本人文件，无权操作
// 	if fileuid != uid {
// 		tx.Rollback()
// 		return false
// 	}
// 	//如果非本人的文件
// 	if originid != -1 {
// 		//分两种情况，查看如果这个文件解除引用对应的文件是否已失效并且无人引用，如果是删除，如果不是仅删除本条记录
// 		sqlStr = "select uid,filename,valid from file where fileid=? for read"
// 		var valid bool
// 		_ = tx.QueryRow(sqlStr, originid).Scan(&fileuid, &filename, &valid)
// 		if valid == false {
// 			sqlStr = "select count(*) from file where originid=?"
// 			_ = tx.QueryRow(sqlStr, originid).Scan(&nums)
// 			//只剩自己一个引用它了，双删
// 			if nums == 1 {
// 				err := common.DeleteFile(baseURL + fileuid + "/" + filename)
// 				if err != nil {
// 					tx.Rollback()
// 					return false
// 				}
// 				sqlStr = "delete from file where fileid=?"
// 				_, err = tx.Exec(sqlStr, originid)
// 				if err != nil {
// 					tx.Rollback()
// 					return false
// 				}
// 				_, err = tx.Exec(sqlStr, fileid)
// 				if err != nil {
// 					tx.Rollback()
// 					return false
// 				}
// 				ChangeStorage(uid, filesize, "-")
// 				tx.Commit()
// 				return true
// 			} else {
// 				//删除自己就完了
// 				sqlStr = "delete from file where fileid=?"
// 				_, err := tx.Exec(sqlStr, fileid)
// 				if err != nil {
// 					tx.Rollback()
// 					return false
// 				}
// 				ChangeStorage(uid, filesize, "-")
// 				tx.Commit()
// 				return true
// 			}
// 		} else {
// 			//删除自己就完了
// 			sqlStr = "delete from file where fileid=?"
// 			_, err := tx.Exec(sqlStr, fileid)
// 			if err != nil {
// 				tx.Rollback()
// 				return false
// 			}
// 			ChangeStorage(uid, filesize, "-")
// 			tx.Commit()
// 			return true
// 		}
// 	}
// 	//查看是否有引用此文件
// 	sqlStr = "select count(*) from file where originid=? for read"
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&nums)
// 	if nums == 0 {
// 		//无人引用直接删记录+删文件
// 		sqlStr = "select uid,filename from file where fileid=?"
// 		_ = tx.QueryRow(sqlStr, fileid).Scan(&fileuid, &filename)
// 		sqlStr = "delete from file where fileid=?"
// 		_, err := tx.Exec(sqlStr, fileid)
// 		if err != nil {
// 			tx.Rollback()
// 			return false
// 		}
// 		err = common.DeleteFile(baseURL + fileuid + "/" + filename)
// 		if err != nil {
// 			tx.Rollback()
// 			return false
// 		}
// 		ChangeStorage(uid, filesize, "-")
// 		tx.Commit()
// 		return true
// 	} else {
// 		//有人引用就仅将valid置为0
// 		sqlStr = "update file set valid=0 where fileid=?"
// 		_, err := tx.Exec(sqlStr, fileid)
// 		if err != nil {
// 			tx.Rollback()
// 			return false
// 		}
// 		ChangeStorage(uid, filesize, "-")
// 		tx.Commit()
// 		return true
// 	}
// }

// func Mkdir(uid, path, filename string, parentid int64) bool {
// 	if filename == "" {
// 		return false
// 	}
// 	tx, _ := db.Begin()
// 	//查看是否有同名目录存在
// 	sqlStr := "select count(*) from file where filename=? and isdir=true and uid=? and parentid=?"
// 	var nums int64
// 	_ = tx.QueryRow(sqlStr, filename, uid, parentid).Scan(&nums)
// 	if nums >= 1 {
// 		tx.Rollback()
// 		return false
// 	}
// 	sqlStr = "insert into file(create_time,uid,path,isdir,md5,parentid,originid,filename,filesize,oldfilename) values(NOW(),?,?,1,'null',?,-1,?,0,?)"
// 	_, err := tx.Exec(sqlStr, uid, path, parentid, filename, filename)
// 	if err != nil {
// 		tx.Rollback()
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func ShareFile(uid, shareid, password, deaddate string, fileid int64) {
// 	tx, _ := db.Begin()
// 	sqlStr := "insert into share values(?,?,?,?,?,0,0)"
// 	_, _ = tx.Exec(sqlStr, shareid, uid, fileid, deaddate, password)
// 	tx.Commit()
// }

// func DeleteDir(uid string, parentid int64) bool {
// 	tx, _ := db.Begin()
// 	//分开处理，文件夹归一类，文件归一类
// 	sqlStr := "select fileid,isdir from file where parentid=?"
// 	rows, err := tx.Query(sqlStr, parentid)
// 	defer rows.Close()
// 	if err != nil {
// 		return false
// 	}
// 	isok := true
// 	for rows.Next() {
// 		var isdir bool
// 		var fileid int64
// 		rows.Scan(&fileid, &isdir)
// 		if isdir == true {
// 			if DeleteDir(uid, fileid) == false {
// 				isok = false
// 			}
// 		} else {
// 			if DeleteFile(uid, fileid) == false {
// 				isok = false
// 			}
// 		}
// 	}
// 	if isok == true {
// 		sqlStr := "delete from file where fileid=?"
// 		tx.Exec(sqlStr, parentid)
// 		tx.Commit()
// 		return true
// 	} else {
// 		tx.Commit()
// 		return false
// 	}
// }

// func Isdir(fileid int64) (isdir bool) {
// 	tx, _ := db.Begin()
// 	sqlStr := "select isdir from file where fileid=?"
// 	_ = tx.QueryRow(sqlStr, fileid).Scan(&isdir)
// 	tx.Commit()
// 	return isdir
// }

// func GetMyShare(uid string) (share []Class.MyShare) {
// 	tx, _ := db.Begin()
// 	sqlStr := "select shareid,fileid,dead_date,password,viewtimes,downloadtimes from share where uid=?"
// 	rows, err := tx.Query(sqlStr, uid)
// 	defer rows.Close()
// 	if err != nil {
// 		tx.Rollback()
// 		return nil
// 	}
// 	var fileidlist []int64
// 	for rows.Next() {
// 		var temp Class.MyShare
// 		var fileid int64
// 		rows.Scan(&temp.Shareid, &fileid, &temp.DeadDate, &temp.Password, &temp.ViewTimes, &temp.DownloadTimes)
// 		share = append(share, temp)
// 		fileidlist = append(fileidlist, fileid)
// 	}
// 	for i, v := range fileidlist {
// 		sqlStr = "select oldfilename,isdir from file where fileid=?"
// 		err = tx.QueryRow(sqlStr, v).Scan(&share[i].Filename, &share[i].Isdir)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil
// 		}
// 	}
// 	tx.Commit()
// 	return share
// }

// func DeleteShare(uid, shareid string) bool {
// 	tx, _ := db.Begin()
// 	sqlStr := "select uid from share where shareid=?"
// 	var fileuid string
// 	_ = tx.QueryRow(sqlStr, shareid).Scan(&fileuid)
// 	if fileuid != uid {
// 		tx.Rollback()
// 		return false
// 	}
// 	sqlStr = "delete from share where shareid=?"
// 	_, err := tx.Exec(sqlStr, shareid)
// 	if err != nil {
// 		tx.Rollback()
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func AddTime(shareid string, unique int64) bool {
// 	tx, _ := db.Begin()
// 	sqlStr := ""
// 	if unique == 1 {
// 		sqlStr = "update share set viewtimes=viewtimes+1 where shareid=?"
// 	} else {
// 		sqlStr = "update share set downloadtimes=downloadtimes+1 where shareid=?"
// 	}
// 	_, err := tx.Exec(sqlStr, shareid)
// 	if err != nil {
// 		tx.Rollback()
// 		return false
// 	}
// 	tx.Commit()
// 	return true
// }

// func VerifyFileAccess(fileid int64, sharefileid int64) bool {
// 	//验证所属权限
// 	if fileid == sharefileid {
// 		return true
// 	}
// 	tx, _ := db.Begin()
// 	var parentid int64
// 	for {
// 		sqlStr := "select parentid from file where fileid=?"
// 		_ = tx.QueryRow(sqlStr, fileid).Scan(&parentid)
// 		fileid = parentid
// 		if fileid == sharefileid {
// 			tx.Commit()
// 			return true
// 		}
// 		if parentid == 0 {
// 			break
// 		}
// 	}
// 	tx.Commit()
// 	return false
// }

// //func GetShareFile(shareid int64) {
// //	var uid string
// //	tx, _ := db.Begin()
// //	sqlStr := "select uid from share where shareid=?"
// //	_ = tx.QueryRow(sqlStr,shareid).Scan(&uid)
// //	sqlStr = "select filename from share where fileid=?"
// //	_, _ = tx.Exec(sqlStr, shareid, uid)
// //}
package mysql
