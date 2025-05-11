package ui

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/pascal910107/torturbo/internal/cache"
	"github.com/pascal910107/torturbo/internal/scheduler"
	"github.com/pascal910107/torturbo/internal/tunnel"
	"github.com/rs/zerolog"
)

//go:embed static/*
var staticFS embed.FS

func Serve(addr string, ctrl *tunnel.Controller, sched *scheduler.Scheduler, cacheStore *cache.Store, log zerolog.Logger) {
	sub, _ := fs.Sub(staticFS, "static")
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	// Minimal JSON API endpoints
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		type Circ struct {
			RTT string `json:"rtt"`
		}
		var out struct {
			Circuits []Circ `json:"circuits"`
		}
		// 由於 Controller 使用對象池管理電路，我們無法直接獲取所有電路
		// 這裡我們只返回一個空數組，因為無法直接訪問所有電路
		out.Circuits = []Circ{}
		_ = json.NewEncoder(w).Encode(out)
	})
	go func() {
		log.Info().Str("ui", addr).Msg("UI server")
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Warn().Err(err).Msg("UI closed")
		}
	}()
}
