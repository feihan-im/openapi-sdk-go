package fhcore

import (
	"encoding/json"
	"sync/atomic"
	"time"
)

var (
	JsonMarshal   func(v interface{}) ([]byte, error)    = json.Marshal
	JsonUnmarshal func(data []byte, v interface{}) error = json.Unmarshal
)

func Pretty(obj interface{}) string {
	s, _ := json.MarshalIndent(obj, "", "  ")
	return string(s)
}

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
