package types

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// ChainID represents a unique Cosmos chain identifier
type ChainID string

const (
	ChainOsmosis   ChainID = "osmosis"
	ChainInjective ChainID = "injective"
	ChainNeutron   ChainID = "neutron"
	ChainStride    ChainID = "stride"
	ChainJuno      ChainID = "juno"
	ChainCosmosHub ChainID = "cosmoshub"
	ChainAkash     ChainID = "akash"
)

// Opportunity represents a detected cross-chain arbitrage opportunity
type Opportunity struct {
	ID             string          `json:"id"`
	DiscoveredAt   time.Time       `json:"discoveredAt"`
	AssetPair      string          `json:"assetPair"`
	SourceChain    ChainID         `json:"sourceChain"`
	DestChain      ChainID         `json:"destChain"`
	SpreadPct      decimal.Decimal `json:"spreadPct"`
	GrossProfitUSD decimal.Decimal `json:"grossProfitUsd"`
	NetProfitUSD   decimal.Decimal `json:"netProfitUsd"`
	PathHops       int             `json:"pathHops"`
	Path           []PathNode      `json:"path"`
	FeeBreakdown   []FeeEntry      `json:"feeBreakdown"`
	CalldataJSON   json.RawMessage `json:"calldataJson"`
	InputAmountUSD decimal.Decimal `json:"inputAmountUsd"`
	SlippageEstPct decimal.Decimal `json:"slippageEstPct"`
	Status         string          `json:"status"` // live | expired | executed
	ExpiresAt      *time.Time      `json:"expiresAt,omitempty"`
}

// PathNode represents one step in an arbitrage path
type PathNode struct {
	Chain    ChainID         `json:"chain"`
	DEX      string          `json:"dex"`
	AssetIn  string          `json:"assetIn"`
	AssetOut string          `json:"assetOut"`
	Price    decimal.Decimal `json:"price"`
	PoolID   string          `json:"poolId,omitempty"`
}

// FeeEntry represents a fee for one hop in a path
type FeeEntry struct {
	Chain     ChainID         `json:"chain"`
	FeeType   string          `json:"feeType"` // ibc_transfer | swap | gas
	AmountUSD decimal.Decimal `json:"amountUsd"`
	Asset     string          `json:"asset"`
}

// PriceData represents a normalized price for an asset pair on a chain
type PriceData struct {
	Chain      ChainID         `json:"chain"`
	AssetPair  string          `json:"assetPair"`
	PriceUSD   decimal.Decimal `json:"priceUsd"`
	SourceDEX  string          `json:"sourceDex"`
	PoolID     string          `json:"poolId,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

// ChainStatus represents the health status of a monitored chain
type ChainStatus struct {
	ChainID     ChainID   `json:"chainId"`
	Name        string    `json:"name"`
	Connected   bool      `json:"connected"`
	LastSeen    time.Time `json:"lastSeen"`
	BlockHeight int64     `json:"blockHeight"`
	FeedCount   int       `json:"feedCount"`
	ErrorCount  int       `json:"errorCount"`
}
