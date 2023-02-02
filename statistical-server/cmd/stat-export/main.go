package main

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"coin-server/common/statistical2/models"
	"coin-server/common/utils"
	"coin-server/statistical-server/config"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/sqltocsv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string
var myConf = new(config.MysqlConfig)
var db *sqlx.DB
var date string

func GetModelFieldDesc(model models.Model) []string {
	count := reflect.TypeOf(model).Elem().NumField()
	ret := make([]string, 0, count)
	for i := 0; i < count; i++ {
		f := reflect.TypeOf(model).Elem().Field(i)
		ret = append(ret, f.Tag.Get("desc"))
	}
	return ret
}

var rootCmd = &cobra.Command{
	Use:   "stat-export",
	Short: "埋点数据导出",
	Long:  `埋点数据导出`,
	Run: func(cmd *cobra.Command, args []string) {
		d := strings.TrimSpace(date)
		if d == "" {
			d = time.Now().UTC().Format("20060102")
		}
		archive, err := os.Create(fmt.Sprintf("埋点数据_%s.zip", d))
		utils.Must(err)
		defer archive.Close()
		zipWriter := zip.NewWriter(archive)
		defer zipWriter.Close()

		fmt.Println("开始导出...")
		for _, topic := range models.TopicList {
			desc := GetModelFieldDesc(models.TopicModelMap[topic])
			table := topic + "_" + d
			rows, err := db.Query(fmt.Sprintf("SELECT *, convert_tz(time,'+00:00', @@session.time_zone) server_time FROM %s;", table))
			if err != nil {
				if err == sql.ErrNoRows || strings.Contains(err.Error(), "Error 1146") {
					continue
				}
				utils.Must(err)
			}
			fmt.Println("导出：", table)
			w1, err := zipWriter.Create(table + ".csv")
			utils.Must(err)
			_, err = w1.Write([]byte(strings.Join(desc, ",") + ",服务器时间" + "\n"))
			utils.Must(err)
			conv := sqltocsv.New(rows)
			conv.TimeFormat = "2006-01-02T15:04:05"
			utils.Must(conv.Write(w1))
		}

		query := `
SELECT *
FROM (SELECT date(time) date, IF(tutorial_step != 0, tutorial_step, tutorial_skip_step) step, count(*) role_count
      FROM game_%s WHERE tutorial_step BETWEEN 1 AND 37 or tutorial_skip_step BETWEEN 1 AND 37 GROUP BY date, step) ret ORDER BY step;
`
		rows, err := db.Query(fmt.Sprintf(query, d))
		if err == nil {
			fname := fmt.Sprintf("新手引导人数统计_%s.csv", d)
			fmt.Println("导出：", fname)
			w1, err := zipWriter.Create(fname)
			utils.Must(err)
			_, err = w1.Write([]byte("日期,步骤,人数" + "\n"))
			utils.Must(err)
			conv := sqltocsv.New(rows)
			conv.TimeFormat = "2006-01-02"
			utils.Must(conv.Write(w1))
		}

		fmt.Println("导出完成")
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "指定db配置文件")
	rootCmd.Flags().StringVarP(&date, "date", "d", "", "导出哪天的数据，不填为当天的。例：20220823")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	utils.Must(viper.ReadInConfig())

	utils.Must(viper.Unmarshal(myConf))
	utils.Must(initMysql(myConf))
}

func initMysql(conf *config.MysqlConfig) error {
	cnf := &mysql.Config{
		User:                 conf.Username,
		Passwd:               conf.Password,
		Net:                  "tcp",
		Addr:                 conf.Addr,
		DBName:               conf.Database,
		Loc:                  time.UTC,
		Timeout:              time.Second * 3,
		ParseTime:            true,
		AllowNativePasswords: true,
		AllowOldPasswords:    true,
	}
	dsn := cnf.FormatDSN()
	var err error
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	return err
}

func main() {
	Execute()
}
