package fhcore

import (
	"sync/atomic"
	"time"
)

var (
	serverTimeBase int64 = 0
	systemTimeBase int64 = 0
)

func getSystemTimestamp() int64 {
	return time.Now().UnixNano() / 1000000
}

func setServerTimeBase(timestamp uint64) {
	ts := int64(timestamp)
	if atomic.LoadInt64(&serverTimeBase) > ts {
		return
	}
	atomic.StoreInt64(&serverTimeBase, ts)
	atomic.StoreInt64(&systemTimeBase, getSystemTimestamp())
}

func getCurrentTimestamp() uint64 {
	return uint64(getSystemTimestamp() - atomic.LoadInt64(&systemTimeBase) + atomic.LoadInt64(&serverTimeBase))
}
