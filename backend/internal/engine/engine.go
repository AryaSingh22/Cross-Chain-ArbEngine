package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos-arbengine/backend/internal/db"
	"github.com/cosmos-arbengine/backend/internal/events"
	"github.com/cosmos-arbengine/backend/internal/feeds"
	"github.com/cosmos-arbengine/backend/internal/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// ArbPath represents a potential arbitrage route
type ArbPath struct {
	AssetPair   string
	Chains      []types.ChainID
	Description string
}

// ArbEngine detects arbitrage opportunities across chains
type ArbEngine struct {
	cache         *feeds.PriceCache
	eventBus      *events.EventBus
	repo          *db.Repository
	logger        *zap.Logger
	paths         []ArbPath
	minProfitUSD  decimal.Decimal
	inputAmountUSD decimal.Decimal
	opportunityTTL time.Duration
}

// NewArbEngine creates a new arb engine
func NewArbEngine(
	cache *feeds.PriceCache,
	eventBus *events.EventBus,
	repo *db.Repository,
	logger *zap.Logger,
	minProfitUSD float64,
	inputAmountUSD float64,
	opportunityTTL time.Duration,
) *ArbEngine {
	return &ArbEngine{
		cache:          cache,
		eventBus:       eventBus,
		repo:           repo,
		logger:         logger,
		paths:          buildDefaultPaths(),
		minProfitUSD:   decimal.NewFromFloat(minProfitUSD),
		inputAmountUSD: decimal.NewFromFloat(inputAmountUSD),
		opportunityTTL: opportunityTTL,
	}
}

// Start subscribes to price updates and evaluates arb paths
func (ae *ArbEngine) Start(ctx context.Context) {
	priceCh := ae.eventBus.Subscribe(events.EventPriceUpdate)
	ae.logger.Info("arb engine started", zap.Int("pathCount", len(ae.paths)))

	// Opportunity expiry goroutine
	go ae.expiryLoop(ctx)

	for {
		select {
		case event := <-priceCh:
			priceData, ok := event.Payload.(types.PriceData)
			if !ok {
				continue
			}
			ae.evaluatePaths(ctx, priceData)
		case <-ctx.Done():
			ae.eventBus.Unsubscribe(events.EventPriceUpdate, priceCh)
			return
		}
	}
}

func (ae *ArbEngine) evaluatePaths(ctx context.Context, updatedPrice types.PriceData) {
	for _, path := range ae.paths {
		if path.AssetPair != updatedPrice.AssetPair {
			continue
		}

		// Get prices for all chains in this path
		for i := 0; i < len(path.Chains)-1; i++ {
			sourceChain := path.Chains[i]
			destChain := path.Chains[i+1]

			sourcePrice := ae.cache.Get(sourceChain, path.AssetPair)
			destPrice := ae.cache.Get(destChain, path.AssetPair)

			if sourcePrice == nil || destPrice == nil {
				continue
			}

			// Check both directions
			ae.checkSpread(ctx, path, sourceChain, destChain, sourcePrice, destPrice)
			ae.checkSpread(ctx, path, destChain, sourceChain, destPrice, sourcePrice)
		}
	}
}

func (ae *ArbEngine) checkSpread(
	ctx context.Context,
	path ArbPath,
	buyChain, sellChain types.ChainID,
	buyPrice, sellPrice *types.PriceData,
) {
	if buyPrice.PriceUSD.IsZero() {
		return
	}

	// Gross spread = (sellPrice - buyPrice) / buyPrice
	spread := sellPrice.PriceUSD.Sub(buyPrice.PriceUSD).Div(buyPrice.PriceUSD)
	spreadPct := spread.Mul(decimal.NewFromInt(100))

	if spreadPct.LessThanOrEqual(decimal.Zero) {
		return
	}

	// Estimate fees
	ibcFee := decimal.NewFromFloat(0.50) // ~$0.50 IBC transfer fee per hop
	gasFee := decimal.NewFromFloat(0.10) // ~$0.10 gas per tx
	totalFees := ibcFee.Add(gasFee).Mul(decimal.NewFromInt(int64(len(path.Chains) - 1)))

	// Estimate slippage (0.1-0.3% depending on input amount)
	slippagePct := decimal.NewFromFloat(0.002) // 0.2%
	slippageUSD := ae.inputAmountUSD.Mul(slippagePct)

	// Gross profit
	grossProfit := ae.inputAmountUSD.Mul(spread)

	// Net profit = gross - fees - slippage
	netProfit := grossProfit.Sub(totalFees).Sub(slippageUSD)

	if netProfit.LessThan(ae.minProfitUSD) {
		return
	}

	// Build opportunity
	now := time.Now()
	expiresAt := now.Add(ae.opportunityTTL)
	oppID := uuid.New().String()

	pathNodes := []types.PathNode{
		{
			Chain:    buyChain,
			DEX:      buyPrice.SourceDEX,
			AssetIn:  "USDC",
			AssetOut: path.AssetPair,
			Price:    buyPrice.PriceUSD,
			PoolID:   buyPrice.PoolID,
		},
		{
			Chain:    sellChain,
			DEX:      sellPrice.SourceDEX,
			AssetIn:  path.AssetPair,
			AssetOut: "USDC",
			Price:    sellPrice.PriceUSD,
			PoolID:   sellPrice.PoolID,
		},
	}

	feeBreakdown := []types.FeeEntry{
		{Chain: buyChain, FeeType: "gas", AmountUSD: gasFee, Asset: "ATOM"},
		{Chain: buyChain, FeeType: "ibc_transfer", AmountUSD: ibcFee, Asset: path.AssetPair},
		{Chain: sellChain, FeeType: "gas", AmountUSD: gasFee, Asset: "ATOM"},
	}

	calldata := map[string]interface{}{
		"@type":          "/ibc.applications.transfer.v1.MsgTransfer",
		"source_port":    "transfer",
		"source_channel": fmt.Sprintf("channel-%s-%s", buyChain, sellChain),
		"token": map[string]string{
			"denom":  path.AssetPair,
			"amount": ae.inputAmountUSD.String(),
		},
		"sender":           "cosmos1...",
		"receiver":         "cosmos1...",
		"timeout_height":   map[string]string{"revision_number": "1", "revision_height": "100000"},
		"timeout_timestamp": fmt.Sprintf("%d", now.Add(10*time.Minute).UnixNano()),
	}
	calldataJSON, _ := json.Marshal(calldata)

	opp := &types.Opportunity{
		ID:             oppID,
		DiscoveredAt:   now,
		AssetPair:      path.AssetPair,
		SourceChain:    buyChain,
		DestChain:      sellChain,
		SpreadPct:      spreadPct,
		GrossProfitUSD: grossProfit,
		NetProfitUSD:   netProfit,
		PathHops:       len(path.Chains) - 1,
		Path:           pathNodes,
		FeeBreakdown:   feeBreakdown,
		CalldataJSON:   calldataJSON,
		InputAmountUSD: ae.inputAmountUSD,
		SlippageEstPct: slippagePct.Mul(decimal.NewFromInt(100)),
		Status:         "live",
		ExpiresAt:      &expiresAt,
	}

	// Persist
	if err := ae.repo.InsertOpportunity(ctx, opp); err != nil {
		ae.logger.Error("failed to persist opportunity", zap.Error(err))
	}

	// Broadcast via event bus
	ae.eventBus.Publish(events.Event{
		Type:    events.EventOpportunity,
		Payload: opp,
	})

	ae.logger.Info("opportunity detected",
		zap.String("pair", path.AssetPair),
		zap.String("buy", string(buyChain)),
		zap.String("sell", string(sellChain)),
		zap.String("spread%", spreadPct.StringFixed(4)),
		zap.String("netProfit", netProfit.StringFixed(2)),
	)
}

func (ae *ArbEngine) expiryLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			count, err := ae.repo.ExpireOldOpportunities(ctx)
			if err != nil {
				ae.logger.Error("failed to expire opportunities", zap.Error(err))
			} else if count > 0 {
				ae.logger.Debug("expired opportunities", zap.Int64("count", count))
			}
		case <-ctx.Done():
			return
		}
	}
}

// buildDefaultPaths defines all monitored 2-hop arbitrage routes
func buildDefaultPaths() []ArbPath {
	return []ArbPath{
		// ATOM/USDC across chains
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainInjective}, Description: "ATOM: Osmosis ↔ Injective"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainNeutron}, Description: "ATOM: Osmosis ↔ Neutron"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainStride}, Description: "ATOM: Osmosis ↔ Stride"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainJuno}, Description: "ATOM: Osmosis ↔ Juno"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainCosmosHub}, Description: "ATOM: Osmosis ↔ Cosmos Hub"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainAkash}, Description: "ATOM: Osmosis ↔ Akash"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainInjective, types.ChainNeutron}, Description: "ATOM: Injective ↔ Neutron"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainInjective, types.ChainStride}, Description: "ATOM: Injective ↔ Stride"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainInjective, types.ChainJuno}, Description: "ATOM: Injective ↔ Juno"},
		{AssetPair: "ATOM/USDC", Chains: []types.ChainID{types.ChainNeutron, types.ChainStride}, Description: "ATOM: Neutron ↔ Stride"},
		// OSMO/USDC
		{AssetPair: "OSMO/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainNeutron}, Description: "OSMO: Osmosis ↔ Neutron"},
		{AssetPair: "OSMO/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainStride}, Description: "OSMO: Osmosis ↔ Stride"},
		{AssetPair: "OSMO/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainJuno}, Description: "OSMO: Osmosis ↔ Juno"},
		// INJ/USDC
		{AssetPair: "INJ/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainInjective}, Description: "INJ: Osmosis ↔ Injective"},
		// NTRN/USDC
		{AssetPair: "NTRN/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainInjective}, Description: "NTRN: Osmosis ↔ Injective"},
		{AssetPair: "NTRN/USDC", Chains: []types.ChainID{types.ChainOsmosis, types.ChainNeutron}, Description: "NTRN: Osmosis ↔ Neutron"},
		// ATOM/OSMO
		{AssetPair: "ATOM/OSMO", Chains: []types.ChainID{types.ChainOsmosis, types.ChainInjective}, Description: "ATOM/OSMO: Osmosis ↔ Injective"},
	}
}
