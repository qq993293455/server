package metrics

import (
	"net/http"
	"strconv"
	"time"

	"coin-server/common/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var isOpen, isInit bool
var startTime = time.Now()

func IsOpen() bool {
	return isOpen
}

func Start(log *logger.Logger, addr string) {
	isOpen = addr != "" && isInit
	if !isOpen {
		return
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Info("start metrics", zap.String("metrics addr", addr))
		log.Error("metrics exit", zap.Error(http.ListenAndServe(addr, nil)))
	}()
}

//var UpTime prometheus.Gauge

var RequestsTotal *prometheus.CounterVec
var RequestsErrorTotal *prometheus.CounterVec
var RequestsLatencyHistogram *prometheus.HistogramVec
var DBProcessTotal *prometheus.CounterVec
var DBDataSizeSummary *prometheus.SummaryVec
var OnlineUserTotal prometheus.Gauge

func Init(server string, serverId int64) {
	//UpTime = prometheus.NewGauge(
	//	prometheus.GaugeOpts{
	//		//Namespace:   server,
	//		Name:        "up_time_seconds",
	//		Help:        "存活时间",
	//		ConstLabels: prometheus.Labels{"server_id": strconv.Itoa(int(serverId))},
	//	},
	//)
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			//Namespace:   server,
			Name:        "requests_total",
			Help:        "总请求次数",
			ConstLabels: prometheus.Labels{"server_id": strconv.Itoa(int(serverId))},
		},
		[]string{"request"},
	)
	RequestsErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			//Namespace:   server,
			Name:        "requests_error_total",
			Help:        "总错误请求次数",
			ConstLabels: prometheus.Labels{"server_id": strconv.Itoa(int(serverId))},
		},
		[]string{"err_code", "err_msg"},
	)
	RequestsLatencyHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			//Namespace:   server,
			Name:        "requests_latency_seconds_histogram",
			Help:        "请求延迟分布",
			ConstLabels: prometheus.Labels{"server_id": strconv.Itoa(int(serverId))},
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"request"},
	)
	DBProcessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			//Namespace:   server,
			Name:        "db_process_total",
			Help:        "数据库操作处理次数",
			ConstLabels: prometheus.Labels{"server_id": strconv.Itoa(int(serverId))},
		},
		[]string{"cmd", "dao"},
	)
	DBDataSizeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			//Namespace:   server,
			Name:        "db_data_size_bytes_summary",
			Help:        "数据库单key数据大小统计",
			ConstLabels: prometheus.Labels{"server_id": strconv.Itoa(int(serverId))},
		},
		[]string{"cmd", "dao"},
	)

	OnlineUserTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			//Namespace:   server,
			Name:        "online_user_total",
			Help:        "在线人数统计",
			ConstLabels: prometheus.Labels{"server_id": strconv.Itoa(int(serverId))},
		},
	)

	prometheus.MustRegister(RequestsTotal, RequestsErrorTotal, RequestsLatencyHistogram,
		DBProcessTotal, DBDataSizeSummary, OnlineUserTotal)
	isInit = true
}
