package startcheck

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/values"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

var (
	UnlockScript = redis.NewScript(`
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
        `)
	UpdateTTlScript = redis.NewScript(`
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("EXPIRE", KEYS[1], 10)
        else
            return 0
        end
        `)
)

var m = sync.Map{}

func genKey(serverType models.ServerType, serverId values.ServerId) string {
	return fmt.Sprintf("start-check-%s-%d", serverType.String(), serverId)
}

func StopCheck(serverType models.ServerType, serverId values.ServerId) {
	key := genKey(serverType, serverId)
	v, ok := m.Load(key)
	if ok && v != nil {
		cc := v.(*cc)
		close(cc.c)
		<-cc.c1
		UnlockScript.Run(context.Background(), redisclient.GetLockerRedis(), []string{key}, cc.ip)
	}
}

type cc struct {
	c  chan struct{}
	c1 chan struct{}
	ip string
}

func StartCheck(serverType models.ServerType, serverId values.ServerId) {
	if serverType != models.ServerType_DungeonMatchServer && serverId == 0 {
		panic("serverId == 0")
	}
	serverName := serverType.String()
	key := genKey(serverType, serverId)
	ip := self_ip.SelfIpLan

	client := redisclient.GetLockerRedis()
	ok, err := client.SetNX(context.Background(), key, ip, 10*time.Second).Result()
	if err != nil {
		panic(err)
	}
	if !ok {
		curIp, err := client.Get(context.Background(), key).Result()
		if err != nil {
			panic(err)
		}
		if curIp != ip {
			fmt.Printf("服务启动失败, %s-%d 已被 %s 占用\n", serverName, serverId, curIp)
			os.Exit(0)
		} else {
			UpdateTTlScript.Run(context.Background(), client, []string{key}, ip)
		}
	}

	c := make(chan struct{})
	c1 := make(chan struct{})
	m.Store(key, &cc{c: c, c1: c1, ip: ip})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println(e)
			}
		}()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c:
				close(c1)
				return
			case <-ticker.C:
				UpdateTTlScript.Run(context.Background(), client, []string{key}, ip)
			}
		}
	}()
	return

}

//func getClientIp() string {
//	addrs, err := net.InterfaceAddrs()
//	if err != nil {
//		panic(err)
//	}
//	var ipList []string
//	for _, address := range addrs {
//		// 检查ip地址判断是否回环地址
//		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
//			if ipNet.IP.To4() == nil {
//				continue
//			}
//			ip := ipNet.IP.String()
//			if strings.Contains(ip, "192.") || strings.Contains(ip, "10.") {
//				ipList = append(ipList, ip)
//			}
//		}
//	}
//	if len(ipList) == 0 {
//		panic("can not get local ip")
//	}
//	sort.Slice(ipList, func(i, j int) bool {
//		return ipList[i] < ipList[j]
//	})
//	return ipList[0]
//}
