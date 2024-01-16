package middleware

import (
	"github.com/gin-gonic/gin"
	"sync"
	"time"
)

type TokenBucket struct {
	Capacity int64
	Rate     float64
	Tokens   int64
	LastTime time.Time
	Mut      sync.Mutex
}

func (tb *TokenBucket) Allow() bool {
	tb.Mut.Lock()
	defer tb.Mut.Unlock()
	now := time.Now()
	// 限制每一秒只能允许capacity个流量通过
	// 上一次限制到现在时间超过rate则清零tokens
	if now.Sub(tb.LastTime).Seconds() >= tb.Rate {
		tb.Tokens = 0
	}
	tb.LastTime = now
	if tb.Tokens+1 > tb.Capacity {
		return false
	} else {
		tb.Tokens = tb.Tokens + 1
		return true
	}
}

func TrafficLimit(maxConn int64) gin.HandlerFunc {
	tb := &TokenBucket{
		Capacity: maxConn,
		Rate:     1.0,
		Tokens:   0,
		LastTime: time.Now(),
	}
	return func(c *gin.Context) {
		var ok bool
		for i := 0; i < 10; i++ {
			if tb.Allow() {
				ok = true
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if ok == false {
			c.String(503, "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
