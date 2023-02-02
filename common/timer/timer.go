package timer

import (
	"sync"
	"sync/atomic"
	"time"

	"coin-server/common/gopool"

	avl "github.com/emirpasic/gods/trees/avltree"
)

/*
timer 应对高频率取系统时间引起syscall的平凡调用
时间精度暂定 1 毫秒，可根据需求调整
*/

type ticker struct {
	sl *avl.Tree
	sync.Mutex
	low uint32
}

type tickerElem struct {
	execTime int64
	f        func()
	d        time.Duration
	tickerF  func() bool
}

var elemPool = sync.Pool{New: func() interface{} {
	return &tickerElem{}
}}

func (t *tickerElem) ExtractKey() float64 {
	return float64(t.execTime)
}

func (t *tickerElem) String() string {
	return "ticker elem"
}

var t1970 = int64(1633017600000) // 2021-10-01 00:00:00

func (this_ *ticker) add(t time.Time, f func()) {
	if f == nil {
		return
	}
	te := elemPool.Get().(*tickerElem)
	tm := t.UnixMilli() - t1970
	tm <<= offset
	te.execTime = tm | int64(atomic.AddUint32(&this_.low, 1))
	te.f = f
	this_.Lock()
	this_.sl.Put(te.execTime, te)
	this_.Unlock()
}

// 1633754016807
const offset = 21

func (this_ *ticker) addTicker(t time.Time, d time.Duration, f func() bool) {
	if f == nil {
		return
	}
	te := elemPool.Get().(*tickerElem)
	tm := t.UnixMilli() - t1970
	tm <<= offset
	te.execTime = tm | (int64(atomic.AddUint32(&this_.low, 1)) & 65535)
	te.tickerF = f
	te.d = d
	this_.Lock()
	this_.sl.Put(te.execTime, te)
	this_.Unlock()
}

func (this_ *ticker) run() {
	go func() {
		timer := time.NewTimer(time.Millisecond)
		defer timer.Stop()
		tes := make([]*tickerElem, 0, 1024)
		for {
			select {
			case t := <-timer.C:
				for this_.doOnce(t.UnixMilli()-t1970, tes) { // 公用tes内存
				}
				timer.Reset(time.Millisecond)
			}
		}
	}()
}

func (this_ *ticker) doOnce(now int64, tes []*tickerElem) (loop bool) {
	this_.Lock()
	cp := cap(tes)
	for i := 0; i < cp; i++ {
		node := this_.sl.Left()
		if node == nil {
			break
		}
		elem := node.Value
		v := elem.(*tickerElem)
		t := v.execTime >> offset
		if t <= now {
			tes = append(tes, v)
			this_.sl.Remove(node.Key)
		} else {
			break
		}
	}
	this_.Unlock()
	if len(tes) >= cp/2 {
		loop = true
	}
	for i := range tes {
		v := tes[i]
		gopool.Submit(func() {
			if v.f != nil {
				v.f()
			} else if v.tickerF != nil {
				if v.tickerF() && v.d > 0 {
					mills := v.execTime >> offset
					this_.addTicker(time.Unix(0, (mills+t1970)*int64(time.Millisecond)).Add(v.d), v.d, v.tickerF)
				}
			}

			v.f = nil
			v.tickerF = nil
			v.d = 0
			v.execTime = 0
			elemPool.Put(v)
		})
	}
	tes = tes[:0]
	return
}

type timer struct {
	// t      unsafe.Pointer
	t      int64
	offset time.Duration // 偏移量，用来修改时间
}

func (this_ *timer) run() {
	n := time.Now().Add(time.Second * this_.offset)
	atomic.StoreInt64(&this_.t, n.UnixNano())
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				t = t.UTC()
				t = t.Add(time.Second * this_.offset)
				atomic.StoreInt64(&this_.t, t.UnixNano())
			}
		}
	}()
}

// Now 当前时间,误差 1 milli
func (this_ *timer) Now() time.Time {
	return time.Unix(0, atomic.LoadInt64(&this_.t)).UTC()
}

// Unix 当前秒级时间戳,误差 1 milli
func (this_ *timer) Unix() int64 {
	return this_.Now().Unix()
}

// UnixMilli 当前毫秒级时间戳,误差 1 milli
func (this_ *timer) UnixMilli() int64 {
	return this_.Now().UnixMilli()
}

// UnixNano 当前纳秒级时间戳,误差 1 milli
func (this_ *timer) UnixNano() int64 {
	return this_.Now().UnixNano()
}

// ToString 当前时间转换成 "2006-01-02 15:04:05" 格式
func (this_ *timer) ToString() string {
	return this_.Now().Format("2006-01-02 15:04:05")
}

// String 当前时间转换成 "2006-01-02 15:04:05.999999999 -0700 MST" 格式
func (this_ *timer) String() string {
	return this_.Now().String()
}

// 修改时间
func (this_ *timer) SetOffset(t time.Duration) {
	this_.offset = t
}

func (this_ *timer) CurrDayFlushTime() {

}

func (this_ *timer) GetOffset() time.Duration {
	return this_.offset * time.Second
}
