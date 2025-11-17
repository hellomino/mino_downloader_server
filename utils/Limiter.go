package utils

import (
	"minodl/log"
	"sync"
	"sync/atomic"
	"time"
)

var Limiter *LimiterMap

const tiktok = 1

func init() {
	Limiter = &LimiterMap{
		data:  make(map[interface{}]*limitItem),
		tiker: time.NewTicker(time.Second * 30),
	}
	go Limiter.Clean()
}

type LimiterMap struct {
	sync.RWMutex
	data  map[interface{}]*limitItem
	tiker *time.Ticker
}

type limitItem struct {
	t     time.Time
	limit time.Duration
	times int64
}

func (l *LimiterMap) Add(key interface{}, duration time.Duration) {
	l.Lock()
	defer l.Unlock()
	l.data[key] = &limitItem{
		time.Now(),
		duration,
		tiktok,
	}
	//log.Info("Limiter Add key %v\n", key)
}

func (l *LimiterMap) UnsafeAdd(key interface{}, duration time.Duration) {
	l.data[key] = &limitItem{
		time.Now(),
		duration,
		tiktok,
	}
	//log.Info("Limiter Add key %v\n", key)
}

func (l *LimiterMap) Del(key interface{}) {
	l.Lock()
	defer l.Unlock()
	delete(l.data, key)
	//log.Info("Limiter Del key %v\n", key)
}

func (l *LimiterMap) UnSafeDel(key interface{}) {
	//log.Info("Limiter UnSafeDel key %v\n", key)
	delete(l.data, key)
}

// IsLimited 操作是否被限制, 操作key，周期内最大可以执行多少次，查看测试用例
func (l *LimiterMap) IsLimited(key interface{}, duration time.Duration, max int64) (bool, int64) {
	l.Lock()
	defer l.Unlock()
	// read
	v, ok := l.data[key]
	if !ok {
		// safe write
		l.UnsafeAdd(key, duration)
		return false, tiktok
	}
	atomic.AddInt64(&v.times, tiktok)
	if time.Now().Before(v.t.Add(duration)) {
		//log.Info("v.times:%d , the max:%d\n", v.times, max)
		if v.times > max {
			return true, v.times
		}
		return false, v.times
	} else {
		// repeat
		l.UnsafeAdd(key, v.limit)
	}
	return false, tiktok
}

// Clean self clean
func (l *LimiterMap) Clean() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Limiter Clean error:%v\n", err)
		}
	}()
	for {
		<-l.tiker.C
		// read need clean keys
		timeoutKeys := make([]interface{}, 0)
		l.RLock()
		for k, v := range l.data {
			if time.Now().After(v.t.Add(v.limit)) {
				timeoutKeys = append(timeoutKeys, k)
			}
		}
		l.RUnlock()
		// write data
		l.Lock()
		for _, k := range timeoutKeys {
			l.UnSafeDel(k)
		}
		l.Unlock()
	}
}
