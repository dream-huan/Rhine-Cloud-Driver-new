package redis

import (
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

func (m *RedisManager) setRedisKey(key string, value interface{}, expiration time.Duration) bool {
	err := m.rdb.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Logger.Error("redis cluster set key error:", zap.Error(err))
		return false
	}
	return true
}

func (m *RedisManager) getRedisKey(key string) interface{} {
	value, err := m.rdb.Get(ctx, key).Result()
	if err != nil {
		log.Logger.Error("redis cluster get key error:", zap.Error(err))
		return ""
	}
	return value
}

func (m *RedisManager) delRedisKey(key string) bool {
	err := m.rdb.Del(ctx, key).Err()
	if err != nil {
		log.Logger.Error("redis cluster del key error:", zap.Error(err))
		return false
	}
	return true
}