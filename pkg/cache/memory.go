package cache

import (
	"sync"
	"time"
	// "time"
)

var mem sync.Map

func checkKeyExpiration(key string, expiration time.Duration) {
	ticker := time.NewTicker(expiration)
	select {
	case <-ticker.C:
		mem.Delete(key)
		return
	}
}

func SetMemKey(key string, value interface{}, expiration time.Duration) {
	mem.Store(key, value)
	if expiration != 0 {
		go checkKeyExpiration(key, expiration)
	}
}

func GetMemKey(key string) (bool, interface{}) {
	if value, ok := mem.Load(key); ok != false {
		return true, value
	}
	return false, nil
}

func DelMemKey(key string) {
	mem.Delete(key)
}
