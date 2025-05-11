package tunnel

import (
	"context"
	"net"
	"time"

	"github.com/cretz/bine/tor"
	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/rs/zerolog"
)

type Config struct {
	TorBinaryPath string
	DataDir       string
	CircuitNum    int
	Logger        zerolog.Logger
}

type Controller struct {
	t      *tor.Tor
	ctx    context.Context
	cancel context.CancelFunc
	pool   *pool.ObjectPool
	cfg    Config
	logger zerolog.Logger
}

// makeFactory 實作 pool.PooledObjectFactory
type makeFactory struct {
	t      *tor.Tor
	logger zerolog.Logger
	maxRTT time.Duration
}

func (f *makeFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	c, err := NewCircuit(f.t, f.logger, ctx)
	if err != nil {
		return nil, err
	}
	return pool.NewPooledObject(c), nil
}
func (f *makeFactory) DestroyObject(ctx context.Context, obj *pool.PooledObject) error {
	// Circuit 無需特別釋放
	return nil
}
func (f *makeFactory) ValidateObject(ctx context.Context, obj *pool.PooledObject) bool {
	c := obj.Object.(*Circuit)
	// RTT 超過 maxRTT 就視為不健康，踢出重新建
	return c.RTT() <= f.maxRTT
}
func (f *makeFactory) ActivateObject(ctx context.Context, obj *pool.PooledObject) error {
	return nil
}
func (f *makeFactory) PassivateObject(ctx context.Context, obj *pool.PooledObject) error {
	return nil
}

func NewController(cfg Config) (*Controller, error) {
	ctx, cancel := context.WithCancel(context.Background())
	t, err := tor.Start(ctx, &tor.StartConf{
		ExePath: cfg.TorBinaryPath,
		DataDir: cfg.DataDir,
		ExtraArgs: []string{
			"--Log", "notice stdout",
			"--SocksPort", "auto",
			"--ControlPort", "auto",
		},
	})
	if err != nil {
		cancel()
		return nil, err
	}

	// 建立 factory，設定 RTT 最大容忍 500ms（可由 cfg 決定）
	factory := &makeFactory{t: t, logger: cfg.Logger, maxRTT: 500 * time.Millisecond}

	// 建 pool，並調整參數
	p := pool.NewObjectPoolWithDefaultConfig(ctx, factory)
	p.Config.MaxTotal = cfg.CircuitNum
	p.Config.MaxIdle = cfg.CircuitNum
	p.Config.MinIdle = 1
	p.Config.TestOnBorrow = true  // 借用前呼叫 ValidateObject
	p.Config.TestWhileIdle = true // 空閒時也驗證
	p.Config.TimeBetweenEvictionRuns = time.Minute
	p.Config.MinEvictableIdleTime = 5 * time.Minute
	p.StartEvictor() // 啟動空閒檢查

	return &Controller{
		t:      t,
		ctx:    ctx,
		cancel: cancel,
		pool:   p,
		cfg:    cfg,
		logger: cfg.Logger,
	}, nil
}

// FastestDial 從 pool 取出最健康的 Circuit，並用它 Dial
func (c *Controller) FastestDial(ctx context.Context, network, addr string) (net.Conn, error) {
	obj, err := c.pool.BorrowObject(ctx)
	if err != nil {
		return nil, err
	}
	defer c.pool.ReturnObject(ctx, obj)
	circ := obj.(*Circuit)
	return circ.Dial(network, addr)
}

func (c *Controller) Close() {
	c.cancel()
	_ = c.t.Close()
	c.pool.Close(c.ctx)
}
