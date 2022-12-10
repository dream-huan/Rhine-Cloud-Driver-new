package redis

import (
	"sync"
	// "time"
)

var mem sync.Map

// func checkKeyExpiration(key string,expiration time.Duration){
// 	select{
// 		case
// 	}
// }

// func setMemKey(key string, value interface{}, expiration time.Duration) {
// 	mem.Set(key,value)
// 	if expiration !=0 {
// 		checkcheckKeyExpiration(key,expiexpiration)
// 	}
// }

func getMemKey(key string) (bool, interface{}) {
	if value, ok := mem.Load(key); ok != false {
		return true, value
	}
	return false, nil
}

func delMemKey(key string) {
	mem.Delete(key)
}
