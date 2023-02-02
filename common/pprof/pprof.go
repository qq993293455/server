package pprof

import (
	"encoding/json"
	"net/http"
	_ "net/http/pprof"

	"coin-server/common/logger"
	system_info "coin-server/common/system-info"

	"go.uber.org/zap"
)

func Start(log *logger.Logger, addr string) {
	if addr == "" {
		return
	}
	go func() {
		http.HandleFunc("/debug/pprof/stats", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			si := system_info.StatsInfo()
			data, err := json.Marshal(si)
			if err != nil {
				w.Write([]byte(err.Error()))
			} else {
				_, _ = w.Write(data)
			}
		})
		log.Info("start pprof", zap.String("pprof addr", addr))
		log.Error("pprof exit", zap.Error(http.ListenAndServe(addr, nil)))
	}()
}
