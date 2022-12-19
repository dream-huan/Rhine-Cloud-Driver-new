package redis

import "Rhine-Cloud-Driver/config"

func Init(cf config.Config) {
	InitRedis(cf.RedisManager)
}
