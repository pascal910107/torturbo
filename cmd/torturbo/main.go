package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/pascal910107/torturbo/internal/cache"
	"github.com/pascal910107/torturbo/internal/config"
	"github.com/pascal910107/torturbo/internal/logger"
	"github.com/pascal910107/torturbo/internal/proxy"
	"github.com/pascal910107/torturbo/internal/scheduler"
	"github.com/pascal910107/torturbo/internal/tunnel"
	"github.com/pascal910107/torturbo/internal/ui"
)

func main() {
	// CLI flags
	listen := flag.String("listen", "127.0.0.1:8118", "HTTP proxy listen address")
	uiPort := flag.String("ui", "127.0.0.1:18000", "Web UI listen address")
	dataDir := flag.String("datadir", "", "TorTurbo data directory (default: $HOME/.torturbo)")
	verbose := flag.Bool("v", false, "enable debug logs")
	flag.Parse()

	// Logger
	log := logger.New(*verbose)

	// Config & dirs
	cfg := config.Default()
	if *dataDir != "" {
		cfg.DataDir = *dataDir
	}
	if err := cfg.EnsureDirs(); err != nil {
		log.Fatal().Err(err).Msg("create data directories failed")
	}

	// Cache
	cacheStore, err := cache.New(cfg.CachePath, cfg.CacheSizeLimitGB, log)
	if err != nil {
		log.Fatal().Err(err).Msg("init cache failed")
	}
	defer cacheStore.Close()

	// Tor controller
	torCtl, err := tunnel.NewController(tunnel.Config{
		TorBinaryPath: cfg.TorBinaryPath(),
		DataDir:       cfg.TorDataDir(),
		CircuitNum:    cfg.CircuitNum,
		Logger:        log,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("launch tor failed")
	}
	defer torCtl.Close()

	// Scheduler
	sched := scheduler.New(torCtl, log)

	// Proxy server
	proxySrv := proxy.New(proxy.Config{
		ListenAddr: *listen,
		Cache:      cacheStore,
		Scheduler:  sched,
		Logger:     log,
	})
	go proxySrv.Start()

	// Web UI
	go ui.Serve(*uiPort, torCtl, sched, cacheStore, log)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	log.Info().Msg("Shutting down...")
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	proxySrv.Shutdown(shutdownCtx)
	torCtl.Close()
	log.Info().Msg("Bye.")
}
