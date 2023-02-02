package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"coin-server/common/logger"
	"coin-server/common/statistical2/models"
	"coin-server/common/timer"
	"coin-server/common/utils"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var sqlMap = map[string]SQL{}
var sqlMapWithOutDate = map[string]SQLWithOutDate{}

func InitSQL() {
	for topic, model := range models.TopicModelMap {
		sqlMap[topic] = genSQL(model)
		sqlMapWithOutDate[topic] = genSQLWithOutDate(model)
	}
}

func mustLog(err error) {
	if err != nil {
		logger.DefaultLogger.Warn("error.", zap.Error(err))
	}
}

func SaveToMysql(data map[string][]models.Model) error {
	for topic, ms := range data {
		if len(ms) == 0 {
			continue
		}

		last := LastCreate[topic]
		beginOfDay := timer.BeginOfDay(time.Now().UTC()).UnixMilli()
		if last < beginOfDay {
			LastCreate[topic] = time.Now().UTC().UnixMilli()
			//mustLog(mysqlConn.Migrator().CreateTable(ms[0]))
			utils.Must(mysqlConn.Table(ms[0].TableName()).Migrator().AutoMigrate(ms[0]))
		}

		model := ms[0]
		table := model.TableName()
		args := make([]interface{}, 0)
		count := 0
		for _, m := range ms {
			if table != m.TableName() {
				if err := saveToMysql(model, table, count, args); err != nil {
					return err
				}
				model = m
				table = m.TableName()
				count = 0
				args = args[:0]
			}
			args = append(args, m.ToArgs()...)
			count++
		}

		if err := saveToMysql(model, table, count, args); err != nil {
			return err
		}
		data[topic] = data[topic][:0]
	}
	return nil
}

func ExecProcedure(data map[string][]models.Model) error {
	for topic, ms := range data {
		if len(ms) == 0 {
			continue
		}
		switch topic {
		case models.LoginTopic:
			for _, m := range ms {
				args := m.(*models.Login)
				query := fmt.Sprintf(
					"call sp_add_game_login_log_v6('%s','%s', '%s', '%s', %d, '%s', %d, '%s', '%s', '%s');",
					args.GameId,
					strconv.FormatInt(args.ServerId, 10),
					args.IggId,
					args.IP,
					args.Time.Unix(),
					args.Xid,
					0,
					args.DeviceId,
					"",
					args.ClientVersion,
				)
				if err := loginMysqlConn.Exec(query).Error; err != nil {
					logger.DefaultLogger.Warn("call login PROCEDURE fail", zap.Error(err))
				} else {
					logger.DefaultLogger.Debug("call login PROCEDURE succ", zap.String("query", query))
				}
			}
		case models.LogoutTopic:
			for _, m := range ms {
				args := m.(*models.Logout)
				query := fmt.Sprintf(
					"call sp_add_game_logout_log_v6('%s','%s','%s', '%s', %d, %d,'%s', %d, '%s', '%s', '%s');",
					args.GameId,
					strconv.FormatInt(args.ServerId, 10),
					args.IggId,
					args.IP,
					args.Time.Unix(),
					args.OnlineSeconds,
					args.Xid,
					0,
					args.DeviceId,
					"",
					args.ClientVersion,
				)
				if err := loginMysqlConn.Exec(query).Error; err != nil {
					logger.DefaultLogger.Warn("call logout PROCEDURE fail", zap.Error(err))
				} else {
					logger.DefaultLogger.Debug("call login PROCEDURE succ", zap.String("query", query))
				}
			}
		}
		data[topic] = data[topic][:0]
	}
	return nil
}

func saveToMysql(model models.Model, table string, count int, args []interface{}) error {
	query := genStmt(sqlMapWithOutDate[model.Topic()], table, count)
	err := mysqlConn.Exec(query, args...).Error
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == 1146 { // 表不存在
				utils.Must(mysqlConn.Table(table).Migrator().AutoMigrate(model))
				err = mysqlConn.Exec(query, args...).Error
				if err != nil {
					return err
				}
			}
			return err
		}
		return err
	}
	return nil
}

func saveByTopic(stmt *sql.Stmt, list []models.Model) {
	//for _, model := range list {
	//	_, err := stmt.Exec(model.ToArgs()...)
	//	if err != nil {
	//		logger.DefaultLogger.Error("exec sql error",
	//			zap.Error(err),
	//			zap.Any("data", model.ToArgs()),
	//		)
	//		continue
	//	}
	//}
}
