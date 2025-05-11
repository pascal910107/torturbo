package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/pascal910107/torturbo/internal/cache"
	"github.com/pascal910107/torturbo/internal/scheduler"
	"github.com/rs/zerolog"
)

type Config struct {
	ListenAddr string
	Cache      *cache.Store
	Scheduler  *scheduler.Scheduler
	Logger     zerolog.Logger
}

type Server struct {
	cfg Config
	srv *http.Server
}

type proxyLogger struct {
	logger zerolog.Logger
}

func (l proxyLogger) Printf(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

func New(cfg Config) *Server {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.Logger = proxyLogger{logger: cfg.Logger}

	// Replace transport
	proxy.Tr = cfg.Scheduler.Transport()

	// Caching handler (only GET)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if req.Method != http.MethodGet {
			return req, nil
		}
		if localPath, etag, mod, ok := cfg.Cache.Lookup(req); ok {
			cfg.Logger.Debug().Str("url", req.URL.String()).Msg("cache hit")
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       mustOpen(localPath),
				Request:    req,
			}
			if etag != "" {
				resp.Header.Set("Etag", etag)
			}
			if mod != "" {
				resp.Header.Set("Last-Modified", mod)
			}
			return req, resp
		}
		return req, nil
	})

	// Save to cache
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp.Request.Method != http.MethodGet {
			return resp
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp
		}
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))

		ttl := 24 * time.Hour
		if cc := resp.Header.Get("Cache-Control"); cc != "" && cc == "no-cache" {
			ttl = 0
		}
		if ttl > 0 {
			if _, err := cfg.Cache.Save(resp, body, ttl); err != nil {
				cfg.Logger.Warn().Err(err).Msg("cache save failed")
			}
		}
		return resp
	})

	httpSrv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: proxy,
	}
	return &Server{
		cfg: cfg,
		srv: httpSrv,
	}
}

func mustOpen(path string) io.ReadCloser {
	f, _ := os.ReadFile(path)
	return io.NopCloser(bytes.NewReader(f))
}

func (s *Server) Start() {
	s.cfg.Logger.Info().Str("addr", s.srv.Addr).Msg("proxy listening")
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.cfg.Logger.Fatal().Err(err).Msg("proxy server error")
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	_ = s.srv.Shutdown(ctx)
}
