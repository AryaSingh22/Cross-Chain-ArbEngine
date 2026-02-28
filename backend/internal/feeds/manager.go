package feeds

import (
	"context"
	"time"

	"github.com/cosmos-arbengine/backend/internal/events"
	"github.com/cosmos-arbengine/backend/internal/types"
	"go.uber.org/zap"
)

// Feed represents a price feed source
type Feed interface {
	// Name returns the feed identifier
	Name() string
	// Chain returns the chain this feed sources from
	Chain() types.ChainID
	// Start begins the polling loop
	Start(ctx context.Context) error
	// Stop gracefully stops the feed
	Stop()
}

// Manager manages all price feeds
type Manager struct {
	feeds    []Feed
	cache    *PriceCache
	eventBus *events.EventBus
	logger   *zap.Logger
}

// NewManager creates a new feed manager
func NewManager(cache *PriceCache, eventBus *events.EventBus, logger *zap.Logger) *Manager {
	return &Manager{
		cache:    cache,
		eventBus: eventBus,
		logger:   logger,
	}
}

// AddFeed registers a feed
func (m *Manager) AddFeed(feed Feed) {
	m.feeds = append(m.feeds, feed)
}

// StartAll launches all feeds as goroutines
func (m *Manager) StartAll(ctx context.Context) {
	for _, feed := range m.feeds {
		go func(f Feed) {
			m.logger.Info("starting price feed",
				zap.String("feed", f.Name()),
				zap.String("chain", string(f.Chain())))

			retryDelay := 2 * time.Second
			maxRetryDelay := 60 * time.Second

			for {
				err := f.Start(ctx)
				if err == nil || ctx.Err() != nil {
					return
				}

				m.logger.Error("feed error, will retry",
					zap.String("feed", f.Name()),
					zap.Error(err),
					zap.Duration("retryIn", retryDelay))

				select {
				case <-time.After(retryDelay):
					retryDelay = retryDelay * 2
					if retryDelay > maxRetryDelay {
						retryDelay = maxRetryDelay
					}
				case <-ctx.Done():
					return
				}
			}
		}(feed)
	}
}

// StopAll gracefully stops all feeds
func (m *Manager) StopAll() {
	for _, feed := range m.feeds {
		feed.Stop()
	}
}

// FeedCount returns the number of registered feeds
func (m *Manager) FeedCount() int {
	return len(m.feeds)
}
