package redis

import (
	"Rhine-Cloud-Driver/config"
	log "Rhine-Cloud-Driver/logic/log"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

var rdb *redis.Client

var ctx = context.Background()

func InitRedis(cf config.RedisConfig) {
	ctx := context.Background()
	fmt.Println(cf)
	rdb = redis.NewClient(&redis.Options{
		Addr:     cf.Address[0],
		Password: cf.Password,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Logger.Error("InitRedis ping error", zap.Error(err))
		return
	}
	log.Logger.Info("Redis链接成功")
}

func SetRedisKey(key string, value interface{}, expiration time.Duration) bool {
	err := rdb.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Logger.Error("redis cluster set key error:", zap.Error(err))
		return false
	}
	return true
}

func GetRedisKey(key string) interface{} {
	value, err := rdb.Get(ctx, key).Result()
	if err != nil {
		log.Logger.Error("redis cluster get key error:", zap.Error(err))
		return ""
	}
	return value
}

func DelRedisKey(key string) bool {
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		log.Logger.Error("redis cluster del key error:", zap.Error(err))
		return false
	}
	return true
}
