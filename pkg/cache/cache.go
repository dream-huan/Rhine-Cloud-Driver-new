package cache

import (
	"Rhine-Cloud-Driver/pkg/conf"
)

func Init(cf conf.Config) {
	InitRedis(cf.RedisManager)
}
