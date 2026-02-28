package feeds

import (
	"context"
	"sync"
	"time"

	"github.com/cosmos-arbengine/backend/internal/types"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// PriceCache provides thread-safe access to latest prices
type PriceCache struct {
	prices sync.Map // key: "chain:assetPair" -> *PriceCacheEntry
	logger *zap.Logger
}

// PriceCacheEntry holds a cached price with a TTL
type PriceCacheEntry struct {
	Data      types.PriceData
	ExpiresAt time.Time
}

// NewPriceCache creates a new price cache
func NewPriceCache(logger *zap.Logger) *PriceCache {
	return &PriceCache{logger: logger}
}

// Set stores a price in the cache
func (pc *PriceCache) Set(chain types.ChainID, assetPair string, price decimal.Decimal, dex string, poolID string, ttl time.Duration) {
	key := string(chain) + ":" + assetPair
	entry := &PriceCacheEntry{
		Data: types.PriceData{
			Chain:     chain,
			AssetPair: assetPair,
			PriceUSD:  price,
			SourceDEX: dex,
			PoolID:    poolID,
			Timestamp: time.Now(),
		},
		ExpiresAt: time.Now().Add(ttl),
	}
	pc.prices.Store(key, entry)
}

// Get retrieves a price from cache; returns nil if expired or not found
func (pc *PriceCache) Get(chain types.ChainID, assetPair string) *types.PriceData {
	key := string(chain) + ":" + assetPair
	val, ok := pc.prices.Load(key)
	if !ok {
		return nil
	}
	entry := val.(*PriceCacheEntry)
	if time.Now().After(entry.ExpiresAt) {
		pc.prices.Delete(key)
		return nil
	}
	return &entry.Data
}

// GetAllPricesForPair retrieves prices for an asset pair across all chains
func (pc *PriceCache) GetAllPricesForPair(assetPair string) []types.PriceData {
	var results []types.PriceData
	pc.prices.Range(func(key, value interface{}) bool {
		entry := value.(*PriceCacheEntry)
		if entry.Data.AssetPair == assetPair && time.Now().Before(entry.ExpiresAt) {
			results = append(results, entry.Data)
		}
		return true
	})
	return results
}

// GetAllPrices retrieves all non-expired prices
func (pc *PriceCache) GetAllPrices() []types.PriceData {
	var results []types.PriceData
	pc.prices.Range(func(key, value interface{}) bool {
		entry := value.(*PriceCacheEntry)
		if time.Now().Before(entry.ExpiresAt) {
			results = append(results, entry.Data)
		}
		return true
	})
	return results
}

// StartCleanup periodically removes expired entries
func (pc *PriceCache) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				pc.prices.Range(func(key, value interface{}) bool {
					entry := value.(*PriceCacheEntry)
					if time.Now().After(entry.ExpiresAt) {
						pc.prices.Delete(key)
					}
					return true
				})
			case <-ctx.Done():
				return
			}
		}
	}()
}
