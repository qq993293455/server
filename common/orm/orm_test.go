package orm

import (
	"testing"

	"github.com/go-redis/redis/v8"
)

var client = redis.NewClient(&redis.Options{
	Network:  "tcp",
	Addr:     "10.23.20.53:6379",
	Password: "iggcdl5,.",
})

func TestOrm(t *testing.T) {
	//o := GetOrm()
	//var ins []RedisInterface
	//ins = append(ins, &dao.Item{
	//	ItemId: 1,
	//	Count:  1,
	//}, &dao.Item{
	//	ItemId: 2,
	//	Count:  1,
	//})
	//key := "123123"
	//o.HMSetPB(client, key, ins)
	//o.Do()
	//o = GetOrm()
	//o.HGetPB(client, key, &dao.Item{ItemId: 1})
	//o.HSetPB(client, key, &dao.Item{ItemId: 3})
	//for i := 0; i < 100; i++ {
	//	o.HSetPB(client, key, &dao.Item{ItemId: int64(i + 1)})
	//}
	//items := map[string]*dao.Item{}
	//err := o.HGetAll(client, "123123", items)
	//if err != nil {
	//	panic(err)
	//}
	//count := 100000
	//now := time.Now()
	//for i := 0; i < count; i++ {
	//	items1 := []dao.Item{}
	//	err = o.HGetAll(client, "123123", &items1)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//sub := time.Now().Sub(now)
	//fmt.Println(time.Now().Sub(now), count/int(sub/time.Second))
}
