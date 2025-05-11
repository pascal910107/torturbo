package tunnel

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "net"
    "time"

    "github.com/cretz/bine/tor"
    "golang.org/x/net/proxy"
    "github.com/rs/zerolog"
)

var ErrNoCircuit = errors.New("no circuit available")

type Circuit struct {
    dialer *tor.Dialer
    logger zerolog.Logger
    rtt    int64 // 最近一次 RTT (奈秒)
}

// NewCircuit 透過隨機 proxy.Auth 強制隔離出一條新 circuit
func NewCircuit(t *tor.Tor, log zerolog.Logger, ctx context.Context) (*Circuit, error) {
    // 1. 隨機 8byte 作為 SOCKS5 User，以觸發 Tor 的 StreamIsolation
    buf := make([]byte, 8)
    if _, err := rand.Read(buf); err != nil {
        return nil, err
    }
    auth := &proxy.Auth{User: hex.EncodeToString(buf), Password: ""}

    // 2. 建立帶 Auth 的 Dialer → 每次不同 auth 都是新 circuit
    d, err := t.Dialer(ctx, &tor.DialConf{ProxyAuth: auth})
    if err != nil {
        return nil, err
    }

    c := &Circuit{dialer: d, logger: log}
    go c.heartbeat(ctx)
    return c, nil
}

// heartbeat 定期用心跳測 RTT
func (c *Circuit) heartbeat(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-time.After(30 * time.Second):
            start := time.Now()
            conn, err := c.dialer.DialContext(ctx, "tcp", "example.com:80")
            if err != nil {
                // 連線失敗不更新 RTT
                continue
            }
            conn.Close()
            c.rtt = time.Since(start).Nanoseconds()
        }
    }
}

// Dial 由外部使用
func (c *Circuit) Dial(network, addr string) (net.Conn, error) {
    return c.dialer.DialContext(context.Background(), network, addr)
}

// RTT 回傳最後測得的延遲
func (c *Circuit) RTT() time.Duration {
    return time.Duration(c.rtt)
}
