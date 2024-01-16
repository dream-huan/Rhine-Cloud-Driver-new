package middleware

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Allow(t *testing.T) {
	tb := &TokenBucket{
		Capacity: 10000,
		Rate:     1.0,
		Tokens:   0,
		LastTime: time.Now(),
	}
	for i := int64(0); i <= tb.Capacity*10; i++ {
		ok := tb.Allow()
		//fmt.Println(i, ok, time.Now().Format("2006-01-02 15:04:05"))
		if i == tb.Capacity*5 {
			time.Sleep(time.Second)
		}
		if i < tb.Capacity || (i > tb.Capacity*5 && i <= tb.Capacity*5+tb.Capacity) {
			assert.Equal(t, ok, true)
		} else {
			assert.Equal(t, ok, false)
		}
	}
}
