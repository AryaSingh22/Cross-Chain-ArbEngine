package feeds

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/cosmos-arbengine/backend/internal/events"
	"github.com/cosmos-arbengine/backend/internal/types"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// mockPairConfig defines base prices for simulated asset pairs per chain
type mockPairConfig struct {
	AssetPair string
	BasePrice float64
	DEX       string
	PoolID    string
}

var mockChainFeeds = map[types.ChainID][]mockPairConfig{
	types.ChainOsmosis: {
		{AssetPair: "ATOM/USDC", BasePrice: 9.45, DEX: "Osmosis DEX", PoolID: "pool-1"},
		{AssetPair: "OSMO/USDC", BasePrice: 0.62, DEX: "Osmosis DEX", PoolID: "pool-678"},
		{AssetPair: "ATOM/OSMO", BasePrice: 15.24, DEX: "Osmosis DEX", PoolID: "pool-1"},
		{AssetPair: "INJ/USDC", BasePrice: 24.80, DEX: "Osmosis DEX", PoolID: "pool-725"},
		{AssetPair: "NTRN/USDC", BasePrice: 0.48, DEX: "Osmosis DEX", PoolID: "pool-812"},
	},
	types.ChainInjective: {
		{AssetPair: "ATOM/USDC", BasePrice: 9.42, DEX: "Helix", PoolID: "market-1"},
		{AssetPair: "INJ/USDC", BasePrice: 24.95, DEX: "Helix", PoolID: "market-2"},
		{AssetPair: "ATOM/OSMO", BasePrice: 15.30, DEX: "Helix", PoolID: "market-5"},
		{AssetPair: "NTRN/USDC", BasePrice: 0.47, DEX: "Helix", PoolID: "market-8"},
	},
	types.ChainNeutron: {
		{AssetPair: "ATOM/USDC", BasePrice: 9.48, DEX: "Astroport", PoolID: "neutron-pool-1"},
		{AssetPair: "NTRN/USDC", BasePrice: 0.49, DEX: "Astroport", PoolID: "neutron-pool-2"},
		{AssetPair: "OSMO/USDC", BasePrice: 0.61, DEX: "Astroport", PoolID: "neutron-pool-3"},
	},
	types.ChainStride: {
		{AssetPair: "ATOM/USDC", BasePrice: 9.44, DEX: "Stride DEX", PoolID: "stride-pool-1"},
		{AssetPair: "stATOM/USDC", BasePrice: 10.15, DEX: "Stride DEX", PoolID: "stride-pool-2"},
		{AssetPair: "OSMO/USDC", BasePrice: 0.63, DEX: "Stride DEX", PoolID: "stride-pool-3"},
	},
	types.ChainJuno: {
		{AssetPair: "ATOM/USDC", BasePrice: 9.50, DEX: "Wynd DEX", PoolID: "juno-pool-1"},
		{AssetPair: "OSMO/USDC", BasePrice: 0.625, DEX: "Wynd DEX", PoolID: "juno-pool-2"},
		{AssetPair: "JUNO/USDC", BasePrice: 0.28, DEX: "Wynd DEX", PoolID: "juno-pool-3"},
	},
	types.ChainCosmosHub: {
		{AssetPair: "ATOM/USDC", BasePrice: 9.46, DEX: "Gravity DEX", PoolID: "hub-pool-1"},
	},
	types.ChainAkash: {
		{AssetPair: "AKT/USDC", BasePrice: 3.15, DEX: "Osmosis (bridged)", PoolID: "akash-pool-1"},
		{AssetPair: "ATOM/USDC", BasePrice: 9.43, DEX: "Osmosis (bridged)", PoolID: "akash-pool-2"},
	},
}

// MockFeed generates simulated price data for local development
type MockFeed struct {
	chain        types.ChainID
	pairs        []mockPairConfig
	cache        *PriceCache
	eventBus     *events.EventBus
	logger       *zap.Logger
	pollInterval time.Duration
	stopCh       chan struct{}
}

// NewMockFeed creates a mock feed for a chain
func NewMockFeed(chain types.ChainID, cache *PriceCache, eventBus *events.EventBus, logger *zap.Logger, pollInterval time.Duration) *MockFeed {
	pairs, ok := mockChainFeeds[chain]
	if !ok {
		pairs = []mockPairConfig{}
	}
	return &MockFeed{
		chain:        chain,
		pairs:        pairs,
		cache:        cache,
		eventBus:     eventBus,
		logger:       logger,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

func (mf *MockFeed) Name() string           { return "mock-" + string(mf.chain) }
func (mf *MockFeed) Chain() types.ChainID    { return mf.chain }

func (mf *MockFeed) Start(ctx context.Context) error {
	ticker := time.NewTicker(mf.pollInterval)
	defer ticker.Stop()

	// Initial tick
	mf.generatePrices()

	for {
		select {
		case <-ticker.C:
			mf.generatePrices()
		case <-mf.stopCh:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (mf *MockFeed) Stop() {
	close(mf.stopCh)
}

func (mf *MockFeed) generatePrices() {
	for _, pair := range mf.pairs {
		// Add realistic price jitter: ±0.1–2.5% fluctuation
		jitter := 1.0 + (rand.Float64()-0.5)*0.05
		// Add occasional larger moves to create arbitrage opportunities
		if rand.Float64() < 0.15 {
			jitter = 1.0 + (rand.Float64()-0.5)*0.08
		}

		price := pair.BasePrice * jitter
		// Add sine wave component for natural-looking oscillation
		sineComponent := math.Sin(float64(time.Now().UnixNano())/1e10) * pair.BasePrice * 0.01
		price += sineComponent

		priceDecimal := decimal.NewFromFloat(price)

		mf.cache.Set(mf.chain, pair.AssetPair, priceDecimal, pair.DEX, pair.PoolID, 30*time.Second)

		mf.eventBus.Publish(events.Event{
			Type: events.EventPriceUpdate,
			Payload: types.PriceData{
				Chain:     mf.chain,
				AssetPair: pair.AssetPair,
				PriceUSD:  priceDecimal,
				SourceDEX: pair.DEX,
				PoolID:    pair.PoolID,
				Timestamp: time.Now(),
			},
		})
	}
}
