package env

import (
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"coin-server/common/values"

	"go.uber.org/zap/zapcore"
)

//环境变量定义
const (
	// 应用名
	APP_NAME = "APP_NAME"
	// 服务启动时间，设置该值后，将会将服务启动时间设置为该时间，格式为："2006.01.02 15:04:05 -0700"
	SERVER_START_TIME = "SERVER_START_TIME"
	// 修改器开关，0为关，1为开
	CHEAT_TOGGLE = "CHEAT_TOGGLE"
	// 非核心日志开关，0为关，1为开，缺省值为0
	NONCORE_LOG = "NONCORE_LOG"
	// 开启打印接口请求信息、请求时长以及请求的PO数量，0，1为开，缺省值为0
	FORMAT_LOG_OPEN = "FORMAT_LOG_OPEN"
	// 开启控制台打印接口请求信息、请求时长以及请求的PO数量，0为关，1为开，缺省值为0
	FORMAT_LOG_STDOUT = "FORMAT_LOG_STDOUT"
	// 打印接口请求信息、请求时长以及请求的PO数量的日志文件路径
	FORMAT_LOG_FILE = "FORMAT_LOG_FILE"
	// 所属时区 默认东八区
	TIME_ZONE = "TIME_ZONE"
	// Mode
	APP_MODE = "APP_MODE"
	// 接收请求的区服列表
	SERVER_ID = "SERVER_ID"
	// 竞技场
	ARENA_SERVER_ID = "ARENA_SERVER_ID"
	// 竞技场
	ACTIVITY_RANKING_SERVER_ID = "ACTIVITY_RANKING_SERVER_ID"
	//GameId
	GAME_ID = "GAME_ID"
	//逻辑大区Id
	LogicalRegionId = "LogicalRegionId"
	//IdentityURL
	IdentityGobURL = "IdentityGobURL"
	//IdentityGobLrIdURL
	IdentityGobLrIdURL = "IdentityGobLrIdURL"
	//IdentityToken
	IdentityToken = "IdentityToken"
	//NodeName
	NODE_NAME = "NODE_NAME"

	/* ============================== */
	/* ========== 配置中心 =========== */
	/* ============================== */

	CONF_TYPE   = "CONF_TYPE"   // 配置类型
	CONF_HOSTS  = "CONF_HOSTS"  // 配置集群地址
	CONF_PATH   = "CONF_PATH"   // 配置节点路径
	CONF_AUTH   = "CONF_AUTH"   // 配置认证信息
	CONF_FORMAT = "CONF_FORMAT" // 配置文件格式

	/* ============================== */
	/* ============ 网络 ============= */
	/* ============================== */

	// 监听的HTTP服务地址
	HTTP_ADDR = "HTTP_ADDR"
	// 监听的有状态服务地址
	TCP_ADDR     = "TCP_ADDR"
	Gateway_Addr = "GATEWAY_ADDR"
	// 监听grpc服务地址
	GRPC_ADDR = "GRPC_ADDR"
	// 监听HTTP_ACTION服务地址
	HTTP_ACTION_ADDR = "HTTP_ACTION_ADDR"
	// 服务发现命名空间
	DISCOVERY_NAMESPACE = "DISCOVERY_NAMESPACE"
	// 服务发现服务名
	DISCOVERY_SERVICE = "DISCOVERY_SERVICE"

	/* ============================== */
	/* ============ 日志 ============= */
	/* ============================== */

	// 日志等级
	// 可为以下几种日志等级： INFO, DEBUG, WARNING, ERROR, CRITICAL
	// 缺省值为INFO
	LOG_LEVEL = "LOG_LEVEL"
	//lumberjack 文件到多大时进行rotate
	LOG_MAX_SIZE = "LOGGER_MAX_SIZE"
	//lumberjack 最多保留多少个rotate
	LOG_MAX_BACKUP = "LOGGER_MAX_BACKUP"
	//日志使用utc时间
	LOG_UTC_TIME = "LOG_UTC_TIME"
	// 是否将日志输入到stdout, 值为1时表示true
	LOG_STDOUT = "LOG_STDOUT"
	// 是否将日志发送到远程日志中心, 值为1时表示true
	LOG_REMOTE = "LOG_REMOTE"
	// 远程日志服务地址
	LOG_REMOTE_ADDR = "LOG_REMOTE_ADDR"
	// 日志输出格式 默认"json",另有"console"
	LOG_ENCODING_MODE = "LOG_ENCODING_MODE"
	// 控制台的基本启动日志是否关闭 true-close false-open 默认false
	BASIC_LOG_CLOSE = "BASIC_LOG_CLOSE"
	/* ============================== */
	/* =========== pprof ============ */
	/* ============================== */

	// 是否开启PPROF
	PPROF_OPEN = "PPROF_OPEN"
	// PPROF监听端口号
	PPROF_ADDR = "PPROF_ADDR"
	// METRICS_ADDR prometheus metrics 监听端口号
	METRICS_ADDR = "METRICS_ADDR"

	// 是否开启PPROF
	GRACEFUL_STOP_OPEN = "GRACEFUL_STOP_OPEN"
	// 优雅关闭监听端口号
	GRACEFUL_STOP_ADDR = "GRACEFUL_STOP_ADDR"
	// 规则中心tag
	RULE_TAGE = "RULE_TAG"

	ERROR_CODE_STACK = "ERROR_CODE_STACK"

	//压力测试相关
	LOCUST_MASTER_HOST           = "LOCUST_MASTER_HOST"
	LOCUST_MASTER_PORT           = "LOCUST_MASTER_PORT"
	LOCUST_USE_SAVE_USER         = "LOCUST_USE_SAVE_USER"   // 是否使用数据库内已有的测试账号
	LOCUST_TARGET_NET_TYPE       = "LOCUST_TARGET_NET_TYPE" // 目标服务器网络类型 tcp http
	LOCUST_TARGET_SERVER_ADDR    = "LOCUST_TARGET_SERVER_ADDR"
	LOCUST_TARGET_SERVER_ID      = "LOCUST_TARGET_SERVER_ID"
	LOCUST_TARGET_LESS_SERVER_ID = "LOCUST_TARGET_LESS_SERVER_ID"

	//分布式追踪
	OPEN_TRACEING   = "OPEN_TRACEING"
	JEAGER_ADDR     = "JEAGER_ADDR"
	JEAGER_UDP_ADDR = "JEAGER_UDP_ADDR"

	// 战斗服可视化
	VISUAL_Addr = "VISUAL_Addr"
	OPEN_VISUAL = "OPEN_VISUAL"

	BATTLE_SERVER_PATH = "BATTLE_SERVER_PATH"

	CENTER_SERVER_ID = "CENTER_SERVER_ID" //中心调度服务id，默认为1

	OPEN_GM_HANDLER = "OPEN_GM_HANDLER" // 是否开启GM指令，默认关闭

	ROGUE_LIKE_SERVER_ID = "ROGUE_LIKE_SERVER_ID" // rogue_like 匹配服Id 默认为1

	IGG_SDK_ALARM_TOKEN = "IGG_SDK_ALARM_TOKEN" // IGG 报警SDK token 如果环境值为空则不会报警

	OPEN_MIDDLE_ERROR = "OPEN_MIDDLE_ERROR"

	GUILD_GVG_SERVER_ID = "GUILD_GVG_SERVER_ID" // 工会GVG SERVER ID
)

var isOpenMiddleError bool
var middleErrorOnce sync.Once

func InitIsOpenMiddleError() {
	if GetString(OPEN_MIDDLE_ERROR) == "1" {
		isOpenMiddleError = true
	}
}

func GetIsOpenMiddleError() bool {
	middleErrorOnce.Do(InitIsOpenMiddleError)
	return isOpenMiddleError
}

func GetServerId() values.ServerId {
	str := os.Getenv(SERVER_ID)
	if str == "" {
		return 0
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return values.ServerId(v)
}

func GetArenaServerId() values.ServerId {
	str := os.Getenv(ARENA_SERVER_ID)
	if str == "" {
		return 0
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return values.ServerId(v)
}

func GetActivityRankingServerId() values.ServerId {
	str := os.Getenv(ACTIVITY_RANKING_SERVER_ID)
	if str == "" {
		return 0
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return values.ServerId(v)
}

func GetHttpAddr() string {
	return GetString(HTTP_ADDR)
}

func GetPPROFAddr() string {
	return GetString(PPROF_ADDR)
}

func GetTCPAddr() string {
	return GetString(TCP_ADDR)
}

func GetString(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func GetInteger(key string) values.Integer {
	str := os.Getenv(key)
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err.Error() + key)
	}
	return val
}

func GetTCPPort() string {
	_, port, err := net.SplitHostPort(GetTCPAddr())
	if err != nil {
		panic(err)
	}
	return port
}

func GetHTTPPort() string {
	_, port, err := net.SplitHostPort(GetHttpAddr())
	if err != nil {
		panic(err)
	}
	return port
}

func GetPPROFPort() string {
	_, port, err := net.SplitHostPort(GetPPROFAddr())
	if err != nil {
		panic(err)
	}
	return port
}

func OpenTracing() bool {
	str := GetString(OPEN_TRACEING)
	if str == "1" {
		return true
	}
	return false
}

func OpenMetrics() bool {
	str := GetString(METRICS_ADDR)
	return str != ""
}

func GetCenterServerId() values.ServerId {
	return GetInteger(CENTER_SERVER_ID)
}

func GetRoguelikeServerId() values.ServerId {
	return GetInteger(ROGUE_LIKE_SERVER_ID)
}

/*
// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
*/

func GetLogLevel() zapcore.Level {
	switch strings.ToUpper(GetString(LOG_LEVEL)) {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	default:
		return zapcore.DebugLevel
	}
}

func GetLogMaxSize() values.Integer {
	return GetInteger(LOG_MAX_SIZE)
}

func GetLogMaxBackUp() values.Integer {
	return GetInteger(LOG_MAX_BACKUP)
}
