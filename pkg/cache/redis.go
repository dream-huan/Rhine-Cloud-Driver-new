package cache

import (
	"Rhine-Cloud-Driver/pkg/conf"
	"Rhine-Cloud-Driver/pkg/log"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"strconv"
	"time"
)

var rdb *redis.Client

var ctx = context.Background()

func InitRedis(cf conf.RedisConfig) {
	ctx := context.Background()
	//fmt.Println(cf)
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

func GetRedisKey(key string) (interface{}, bool) {
	value, err := rdb.Get(ctx, key).Result()
	if err != nil {
		//log.Logger.Error("redis cluster get key error:", zap.Error(err))
		return nil, false
	}
	return value, true
}

func GetRedisKeyBytes(key string) ([]byte, bool) {
	value, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		//log.Logger.Error("redis cluster get key error:", zap.Error(err))
		return nil, false
	}
	return value, true
}

func DelRedisKey(key string) bool {
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		log.Logger.Error("redis cluster del key error:", zap.Error(err))
		return false
	}
	return true
}

func RenewRedisKey(key string, expiration time.Duration) bool {
	err := rdb.Expire(ctx, key, expiration).Err()
	if err != nil {
		log.Logger.Error("redis cluster del key error:", zap.Error(err))
		return false
	}
	return true
}

func SetRedisKeyBitmap(key string, offset int64, value int64, expiration time.Duration) bool {
	err := rdb.SetBit(ctx, key, offset, int(value)).Err()
	if expiration != 0 {
		rdb.Expire(ctx, key, expiration)
	}
	if err != nil {
		log.Logger.Error("redis cluster set bitmap  error:", zap.Error(err))
		return false
	}
	return true
}

func GetRedisAllBitmap(chunksKey, chunkNumKey string) (value []byte, isExist bool) {
	tempValue, err := rdb.Get(ctx, chunksKey).Bytes()
	if err != nil {
		//log.Logger.Error("redis cluster get key error:", zap.Error(err))
		return nil, false
	}
	chunkNum, _ := strconv.ParseInt(rdb.Get(ctx, chunkNumKey).String(), 10, 64)
	value = make([]byte, chunkNum)
	for i := 0; i < len(tempValue); i++ {
		// 取出来的byte要将其变为bit才能拿到0和1的情况，1byte=8bit
		tempStr := fmt.Sprintf("%08b", tempValue[i])
		for j := 0; j < 8; j++ {
			value[i*8+j] = tempStr[j] - '0'
		}
	}
	return value, true
}

//func SetRedisAllBitmap(key string, chunks string) bool {
//	// 获取字符串长度，如果不足8的整数位则补位为8的整数位
//	if len(chunks)%8 > 0 {
//		chunks += strings.Repeat("0", 8-len(chunks)%8)
//	}
//	// 将字符串的内容(bit)变为byte
//	value := make([]byte, len(chunks)/8)
//	for i := 0; i < len(chunks); i += 8 {
//		for j := 0; j < 8; j++ {
//			if chunks[i+j] == '1' {
//				// 相当于将二进制还原为整数
//				value[i] += 1 << (7 - j)
//			}
//		}
//	}
//	tempValue, err := rdb.Get(ctx, key).Bytes()
//	if err != nil {
//		//log.Logger.Error("redis cluster get key error:", zap.Error(err))
//		return false
//	}
//	return true
//}

func GetRedisKeyBitmap(key string, offset int64) (value int64) {
	value, err := rdb.GetBit(ctx, key, offset).Result()
	if err != nil {
		return -1
	}
	return value
}

func CountRedisKeyBitmap(key string, start, end int64) (value int64) {
	value, err := rdb.BitCount(ctx, key, &redis.BitCount{
		Start: start,
		End:   end,
	}).Result()
	if err != nil {
		log.Logger.Error("CountRedisKeyBitmap error", zap.Error(err))
		return 0
	}
	return value
}
