package dcpool

import (
	"context"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// Pool manages connections to multiple Telegram datacenters.
// Files are stored on different DCs; connecting directly to the
// correct DC avoids proxy-forwarding latency and bandwidth limits.
type Pool struct {
	api         *telegram.Client
	size        int64
	mu          sync.Mutex
	clients     map[int]*tg.Client
	closes      map[int]func() error
	middlewares []telegram.Middleware
	logger      *zap.Logger
}

// New creates a connection pool for accessing Telegram DCs.
// size controls how many connections to open per DC (minimum effective: 2).
// middlewares are applied to every client returned by the pool.
func New(api *telegram.Client, size int64, logger *zap.Logger, middlewares ...telegram.Middleware) *Pool {
	return &Pool{
		api:         api,
		size:        size,
		clients:     make(map[int]*tg.Client),
		closes:      make(map[int]func() error),
		middlewares: middlewares,
		logger:      logger,
	}
}

// Client returns a *tg.Client connected directly to the given DC.
// Connections are lazily created and cached; subsequent calls for the
// same DC return the cached client.
func (p *Pool) Client(ctx context.Context, dc int) *tg.Client {
	p.mu.Lock()
	defer p.mu.Unlock()

	if c, ok := p.clients[dc]; ok {
		return c
	}

	var (
		invoker telegram.CloseInvoker
		err     error
	)

	if dc == p.api.Config().ThisDC {
		// Same DC: use pooling (multiple connections to current DC)
		invoker, err = p.api.Pool(p.size)
	} else {
		// Different DC: create new connections
		invoker, err = p.api.DC(ctx, dc, p.size)
	}

	if err != nil {
		p.logger.Warn("failed to create dc client, falling back to default",
			zap.Int("dc", dc),
			zap.Error(err),
		)
		// Degrade: use the default client
		return tg.NewClient(p.api)
	}

	p.closes[dc] = invoker.Close
	p.clients[dc] = tg.NewClient(chainMiddlewares(invoker, p.middlewares...))

	return p.clients[dc]
}

// Close shuts down all pooled connections.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for dc, closeFn := range p.closes {
		if err := closeFn(); err != nil {
			p.logger.Warn("error closing dc pool", zap.Int("dc", dc), zap.Error(err))
		}
	}

	p.clients = make(map[int]*tg.Client)
	p.closes = make(map[int]func() error)
	return nil
}

// chainMiddlewares applies middlewares to an invoker in reverse order
// so that the first middleware in the list wraps the outermost layer.
func chainMiddlewares(invoker tg.Invoker, middlewares ...telegram.Middleware) tg.Invoker {
	for i := len(middlewares) - 1; i >= 0; i-- {
		invoker = middlewares[i].Handle(invoker)
	}
	return invoker
}
