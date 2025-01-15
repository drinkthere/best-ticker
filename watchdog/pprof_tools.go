package watchdog

import (
	"best-ticker/config"
	"best-ticker/utils/logger"
	"net/http"
	_ "net/http/pprof"
)

func StartPprofNet(cfg *config.Config) {
	if cfg.PprofListenAddress == "" {
		logger.Info("[Watchdog] No Need Start Pprof Net")
		return
	}
	go func() {
		http.ListenAndServe(cfg.PprofListenAddress, nil)
	}()
	logger.Info("[Watchdog] Start Pprof Net")
}
