package match

import (
	"log"
	"testing"
	"time"

	"coin-server/common/logger"
	"coin-server/common/proto/models"

	"go.uber.org/zap"
)

func Test_OwnerCron(t *testing.T) {
	InitOwnerCron(logger.MustNew(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_RacingRankServer,
			"serverId":   0,
		},
		Development: true,
	}))
	ownerCron.AddCron(&Cron{
		RoleId: "test+5s",
		When:   time.Now().Add(time.Second * 5).UnixMilli(),
		Exec:   executor,
	})
	ownerCron.AddCron(&Cron{
		RoleId: "test-5s",
		When:   time.Now().Add(-time.Second * 5).UnixMilli(),
		Exec:   executor,
	})
	// time.Sleep(time.Second * 20)
	select {}
}

func executor(cron *Cron) {
	// fmt.Printf("cron exec... %#v\n", cron)
	log.Printf("cron exec: %#v\n", cron)
	if cron.Retry() {
		cron.RetryTimes++
		ownerCron.AddCron(cron)
	}
}
