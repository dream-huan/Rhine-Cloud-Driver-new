package common

import (
	"Rhine-Cloud-Driver/logic/log"
	"github.com/speps/go-hashids/v2"
	"go.uber.org/zap"

	// log "Rhine-Cloud-Driver/logic/log"
	"math/rand"
	"sync"
	"time"
)

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 生成n位随机数
func RandStringRunes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func HashEncode(v []int, minLen int) (string, error) {
	hd := hashids.NewData()
	hd.Salt = "this is my salt"
	hd.MinLength = minLen
	h, err := hashids.NewWithData(hd)
	if err != nil {
		log.Logger.Error("hashID new error:", zap.Error(err))
		return "", NewError(ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED)
	}
	e, err := h.Encode(v)
	if err != nil {
		log.Logger.Error("encode new error:", zap.Error(err))
		return "", NewError(ERROR_COMMON_TOOLS_HASH_ENCODE_FAILED)
	}
	return e, nil
}

func HashDecode(hashValue string, minLen int) (uint64, error) {
	hd := hashids.NewData()
	hd.Salt = "this is my salt"
	hd.MinLength = minLen
	h, err := hashids.NewWithData(hd)
	if err != nil {
		log.Logger.Error("hashID new error:", zap.Error(err))
		return 0, NewError(ERROR_COMMON_TOOLS_HASH_DECODE_FAILED)
	}
	d, _ := h.DecodeWithError(hashValue)
	return uint64(d[0]), nil
}

// 雪花算法生成ID
type Worker struct {
	m         sync.Mutex
	lastStamp int64
	workerID  int64
	sequence  int64
}

var IDBuilder *Worker

func NewWorker(workerID int64) {
	IDBuilder = &Worker{
		workerID:  workerID,
		lastStamp: 0,
		sequence:  0,
	}
}

const (
	workerIDBits = uint64(10)
	sequenceBits = uint64(12)

	maxWorkerID = int64(-1) ^ (int64(-1) << workerIDBits)
	maxSequence = int64(-1) ^ (int64(-1) << sequenceBits)

	timeLeft = uint8(22)
	workLeft = uint8(12)
	// 2012-12-9 11:06:01 +0800 CST
	twepoch = int64(1355022361000)
)

func getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func (w *Worker) NextID() (uint64, error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.nextID()
}

func (w *Worker) nextID() (uint64, error) {
	timeStamp := getMilliSeconds()
	if timeStamp < w.lastStamp {
		return 0, NewError(ERROR_COMMON_SNOWFLOWS_GENERATE)
	}

	if w.lastStamp == timeStamp {

		w.sequence = (w.sequence + 1) & maxSequence

		if w.sequence == 0 {
			for timeStamp <= w.lastStamp {
				timeStamp = getMilliSeconds()
			}
		}
	} else {
		w.sequence = 0
	}

	w.lastStamp = timeStamp
	id := ((timeStamp - twepoch) << timeLeft) |
		(w.workerID << workLeft) |
		w.sequence

	return uint64(id), nil
}
