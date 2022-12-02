package redis

import (
	"Rhine-Cloud-Driver/common"
	"Rhine-Cloud-Driver/config"
	log "Rhine-Cloud-Driver/logic/log"
	"context"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

type RedisManager struct {
	rdb *redis.ClusterClient
}

var ctx = context.Background()

func InitRedis(cf config.RedisConfig) RedisManager {
	ctx := context.Background()
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    cf.Address,
		Password: cf.Password,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Logger.Error("InitRedis ping error", zap.Error(err))
		return RedisManager{
			rdb: nil,
		}
	}
	return RedisManager{
		rdb: rdb,
	}
}

func (m *RedisManager) SetKey(key string, value interface{}, expiration time.Duration) bool {
	err := m.rdb.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Logger.Error("redis cluster set key error:", zap.Error(err))
		return false
	}
	return true
}

func (m *RedisManager) GetKey(key string) interface{} {
	value, err := m.rdb.Get(ctx, key).Result()
	if err != nil {
		log.Logger.Error("redis cluster get key error:", zap.Error(err))
		return ""
	}
	return value
}

func (m *RedisManager) DelKey(key string) bool {
	err := m.rdb.Del(ctx, key).Err()
	if err != nil {
		log.Logger.Error("redis cluster del key error:", zap.Error(err))
		return false
	}
	return true
}

func (m *RedisManager) GetDownloadKey(key string) (bool, string) {
	val, err := m.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, ""
	}
	return true, val
}

func (m *RedisManager) AddDownloadKey(value string) string {
	var key string
	for {
		key = common.RandStringRunes(64)
		if result, _ := m.rdb.Exists(ctx, key).Result(); result == 0 {
			break
		}
	}
	err := m.rdb.Set(ctx, key, value, 60*time.Second).Err()
	if err != nil {
		return ""
	}
	return key
}

func (m *RedisManager) AddUploadKey() string {
	var key string
	for {
		key = common.RandStringRunes(64)
		if result, _ := m.rdb.Exists(ctx, key).Result(); result == 0 {
			break
		}
	}
	err := m.rdb.Set(ctx, key, "1", 12*60*60*time.Second).Err() //12小时失效
	if err != nil {
		return ""
	}
	return key
}

func (m *RedisManager) DelUploadKey(key string) int64 {
	result, _ := m.rdb.Del(ctx, key).Result()
	return result
}

func (m *RedisManager) GetUploadKey(key string) bool {
	if result, _ := m.rdb.Exists(ctx, key).Result(); result == 0 {
		return false
	}
	return true
}

func (m *RedisManager) NewShare(fileid, deadtime int64, uid, password string) (key string) {
	for {
		key = common.RandStringRunes(16)
		if result, _ := m.rdb.Exists(ctx, key).Result(); result == 0 {
			break
		}
	}
	if password == "" {
		m.rdb.HMSet(ctx, key, "fileid", fileid, "uid", uid)
	} else {
		m.rdb.HMSet(ctx, key, "fileid", fileid, "password", password, "uid", uid)
	}
	if deadtime != 3 {
		if deadtime == 1 {
			m.rdb.Expire(ctx, key, 7*24*60*60*time.Second)
		} else {
			m.rdb.Expire(ctx, key, 30*24*60*60*time.Second)
		}
	}
	return key
}

func (m *RedisManager) GetShare(key string) (isexist bool, fileid int64, password, uid string) {
	if result, _ := m.rdb.Exists(ctx, key).Result(); result == 0 {
		return false, 0, "", ""
	}
	m.rdb.HGet(ctx, key, "fileid").Scan(&fileid)
	m.rdb.HGet(ctx, key, "password").Scan(&password)
	m.rdb.HGet(ctx, key, "uid").Scan(&uid)
	return true, fileid, password, uid
}

func (m *RedisManager) DeleteShare(shareid string) bool {
	num := m.rdb.HDel(ctx, shareid, "fileid", "password", "uid")
	return num.Val() != 0
}
