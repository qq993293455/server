package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/rdb-to-tendis/env"

	"github.com/hdt3213/rdb/parser"
	"github.com/spf13/cobra"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

func main() {
	_log := logger.MustNew(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": "rdb-to-tendis",
		},
		Development: true,
	})
	logger.SetDefaultLogger(_log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), _log)
	utils.Must(err)
	redisclient.Init(cnf)

	Execute()
}

var filepath string
var gonum int

var rootCmd = &cobra.Command{
	Use:   "rdb-to-tendis",
	Short: "rdb导入tendis",
	Long:  `rdb导入tendis`,
	Run: func(cmd *cobra.Command, args []string) {
		if filepath == "" {
			fmt.Print(cmd.UsageString())
			return
		}
		process(filepath, gonum)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&filepath, "file", "f", "", "指定rdb文件")
	rootCmd.Flags().IntVarP(&gonum, "go", "g", 8, "指定并发数")
}

var writeCount = atomic.NewInt32(0)

func process(path string, gonum int) {
	rdbFile, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("读取文件 [%s] 失败", path))
	}
	defer rdbFile.Close()

	decoder := parser.NewDecoder(rdbFile)

	wg := &sync.WaitGroup{}
	ch := make(chan parser.RedisObject, 128)
	for i := 0; i < gonum; i++ {
		go write(ch, wg)
	}

	count := 0
	err = decoder.Parse(func(o parser.RedisObject) bool {
		ch <- o
		count++
		return true
	})
	if err != nil {
		panic(err)
	}
	close(ch)
	wg.Wait()
	log.Printf("解析%d条 -> 写入%d条数据.", count, writeCount.Load())
}

func write(ch chan parser.RedisObject, wg *sync.WaitGroup) {
	ctx := context.Background()

	wg.Add(1)
	for o := range ch {
		switch o.GetType() {
		case parser.StringType:
			str := o.(*parser.StringObject)
			log.Printf("set %s %v", str.Key, str.Value)
			utils.Must(redisclient.GetDefaultRedis().Set(ctx, str.Key, str.Value, 0).Err())
			writeCount.Inc()
		case parser.HashType:
			hash := o.(*parser.HashObject)
			vals := make(map[string]interface{}, len(hash.Hash))
			for k, v := range hash.Hash {
				vals[k] = v
			}
			log.Printf("hmset %s %v", hash.Key, hash.Hash)
			utils.Must(redisclient.GetDefaultRedis().HMSet(ctx, hash.Key, vals).Err())
			writeCount.Inc()
		default:
			log.Printf("[miss] %s %s", o.GetType(), o.GetKey())
		}
	}
	wg.Add(-1)
}
