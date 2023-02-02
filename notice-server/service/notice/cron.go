package notice

import (
	"fmt"
	"sync"
	"time"

	"coin-server/common/logger"
	"coin-server/common/values"

	"go.uber.org/zap"

	"github.com/panjf2000/ants/v2"

	"github.com/liyue201/gostl/ds/rbtree"
)

var ownerCron *OwnerCron

const (
	defaultWait   = 100 * time.Millisecond // 树为空时按该时间空转
	maxRetryTimes = 3                      // 单个任务最多重试次数
)

type Cron struct {
	NoticeId   string
	OrderKey   int64
	When       values.TimeStamp
	RetryTimes values.Integer
	Exec       func(self *Cron) // 执行器
}

func (c *Cron) Retry() bool {
	return c.RetryTimes < maxRetryTimes
}

type OwnerCron struct {
	tree   *rbtree.RbTree
	owners map[string]*Cron
	mu     sync.Mutex
	logger *logger.Logger
	ctCh   chan struct{} // 新增的cron小于树中最小执行时间时触发
	pool   *ants.Pool

	counter uint16
}

func genKey(ts values.TimeStamp, counter uint16) int64 {
	return ts*1000000 + int64(counter)
}

func InitOwnerCron(logger *logger.Logger) {
	pool, err := ants.NewPool(10000)
	if err != nil {
		panic(fmt.Sprintf("ants.NewPool failed: %s", err.Error()))
	}
	ownerCron = &OwnerCron{
		tree: rbtree.New(rbtree.WithKeyComparator(func(a, b interface{}) int {
			return int(a.(int64) - b.(int64))
		})),
		owners: map[string]*Cron{},
		mu:     sync.Mutex{},
		logger: logger,
		ctCh:   make(chan struct{}, 1),
		pool:   pool,
	}
	go ownerCron.start()
}

func (oc *OwnerCron) AddCron(cron *Cron) {
	if cron == nil || cron.NoticeId == "" {
		return
	}
	if cron.When <= 0 {
		cron.When = time.Now().UnixMilli()
	}
	oc.mu.Lock()
	defer oc.mu.Unlock()

	var fWhen values.TimeStamp
	if oc.tree.Size() > 0 {
		fWhen = oc.firstCron().When
	}
	if oc.OwnerCronExist(cron) {
		oc.RemoveCron(cron)
		delete(oc.owners, cron.NoticeId)
	}
	oc.owners[cron.NoticeId] = cron

	cron.OrderKey = genKey(cron.When, oc.counter)
	oc.counter++
	oc.tree.Insert(cron.OrderKey, cron)

	if oc.tree.Size() > 0 && cron.When < fWhen {
		select {
		case oc.ctCh <- struct{}{}:
		default:
		}
	}
}

func (oc *OwnerCron) OwnerCronExist(cron *Cron) bool {
	item, ok := oc.owners[cron.NoticeId]
	if !ok {
		return false
	}
	return oc.tree.FindNode(item.OrderKey) != nil
}

func (oc *OwnerCron) start() {
	var wait time.Duration
	var cron *Cron
	st := time.NewTimer(defaultWait)
	for {
		wait = defaultWait
		oc.mu.Lock()
		if oc.tree.Size() == 0 {
			goto Sleep
		}

		cron = oc.firstCron()
		if cron.When > time.Now().UnixMilli() {
			wait = time.Duration(cron.When - time.Now().UnixMilli())
			goto Sleep
		}
		oc.run(cron)

		oc.RemoveCron(cron)
		if oc.tree.Size() == 0 {
			goto Sleep
		}
		wait = time.Duration(oc.firstCron().When - time.Now().UnixMilli())
		goto Sleep

	Sleep:
		oc.mu.Unlock()
		if wait <= 0 {
			continue
		}
		if wait < defaultWait { // 防止timer间隔过小
			st.Reset(defaultWait)
		} else {
			st.Reset(wait)
		}
		select {
		case <-st.C:
		case <-oc.ctCh:
			continue
		}
	}
}

func (oc *OwnerCron) firstCron() *Cron {
	return oc.tree.First().Value().(*Cron)
}

func (oc *OwnerCron) RemoveCron(cron *Cron) {
	item, ok := oc.owners[cron.NoticeId]
	if !ok {
		return
	}
	oc.tree.Delete(oc.tree.FindNode(item.OrderKey))
}

func (oc *OwnerCron) run(cron *Cron) {
	if err := oc.pool.Submit(func() {
		cron.Exec(cron)
	}); err != nil {
		oc.logger.Error("cron exec error", zap.Error(err), zap.String("id", cron.NoticeId))
	}
}
