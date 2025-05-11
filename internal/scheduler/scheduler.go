package scheduler

import (
	"net/http"

	"github.com/pascal910107/torturbo/internal/tunnel"
	"github.com/rs/zerolog"
)

type Scheduler struct {
	ctrl *tunnel.Controller
	log  zerolog.Logger
}

func New(ctrl *tunnel.Controller, log zerolog.Logger) *Scheduler {
	return &Scheduler{ctrl: ctrl, log: log}
}

// Transport returns an *http.Transport whose DialContext uses fastest circuit.
func (s *Scheduler) Transport() *http.Transport {
	tr := &http.Transport{
		DialContext:       s.ctrl.FastestDial,
		DisableKeepAlives: false,
	}
	return tr
}
