package redis

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

func setMemKey(key string, value interface{}, expiration time.Duration) {
	mem.Store(key, value)
	if expiration != 0 {
		go checkKeyExpiration(key, expiration)
	}
}

func getMemKey(key string) (bool, interface{}) {
	if value, ok := mem.Load(key); ok != false {
		return true, value
	}
	return false, nil
}

func delMemKey(key string) {
	mem.Delete(key)
}
