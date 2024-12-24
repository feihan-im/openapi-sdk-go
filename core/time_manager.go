package fhcore

import (
	"sync/atomic"
	"time"
)

type TimeManager interface {
	GetSystemTimestamp() int64
	GetServerTimestamp() int64
	SyncServerTimestamp(timestamp int64)
}

func NewDefaultTimeManager() TimeManager {
	return &defaultTimeManager{}
}

type defaultTimeManager struct {
	serverTimeBase int64
	systemTimeBase int64
}

func (m *defaultTimeManager) GetSystemTimestamp() int64 {
	return time.Now().UnixNano() / 1000000
}

func (m *defaultTimeManager) GetServerTimestamp() int64 {
	return m.GetSystemTimestamp() - atomic.LoadInt64(&m.systemTimeBase) + atomic.LoadInt64(&m.serverTimeBase)
}

func (m *defaultTimeManager) SyncServerTimestamp(timestamp int64) {
	ts := timestamp
	if atomic.LoadInt64(&m.serverTimeBase) > ts {
		return
	}
	atomic.StoreInt64(&m.serverTimeBase, ts)
	atomic.StoreInt64(&m.systemTimeBase, m.GetSystemTimestamp())
}
