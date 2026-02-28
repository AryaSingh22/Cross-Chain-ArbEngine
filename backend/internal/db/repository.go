package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos-arbengine/backend/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Repository provides CRUD operations for all entities
type Repository struct {
	db *Database
}

// NewRepository creates a new repository
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// InsertOpportunity persists a new arbitrage opportunity
func (r *Repository) InsertOpportunity(ctx context.Context, opp *types.Opportunity) error {
	pathJSON, _ := json.Marshal(opp.Path)
	feeJSON, _ := json.Marshal(opp.FeeBreakdown)

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO opportunities (
			id, discovered_at, asset_pair, source_chain, dest_chain,
			spread_pct, gross_profit_usd, net_profit_usd, path_hops,
			path_json, fee_breakdown, calldata_json,
			input_amount_usd, slippage_est_pct, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`,
		opp.ID, opp.DiscoveredAt, opp.AssetPair,
		string(opp.SourceChain), string(opp.DestChain),
		opp.SpreadPct, opp.GrossProfitUSD, opp.NetProfitUSD,
		opp.PathHops, pathJSON, feeJSON, opp.CalldataJSON,
		opp.InputAmountUSD, opp.SlippageEstPct, opp.Status, opp.ExpiresAt,
	)
	return err
}

// GetOpportunities retrieves live opportunities with optional filters
func (r *Repository) GetOpportunities(ctx context.Context, status string, limit int) ([]types.Opportunity, error) {
	query := `
		SELECT id, discovered_at, asset_pair, source_chain, dest_chain,
			spread_pct, gross_profit_usd, net_profit_usd, path_hops,
			path_json, fee_breakdown, calldata_json,
			input_amount_usd, slippage_est_pct, status, expires_at
		FROM opportunities
		WHERE status = $1
		ORDER BY discovered_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanOpportunities(rows)
}

// GetOpportunityHistory retrieves historical opportunities with date range filter
func (r *Repository) GetOpportunityHistory(ctx context.Context, from, to time.Time, assetPair string, limit, offset int) ([]types.Opportunity, error) {
	query := `
		SELECT id, discovered_at, asset_pair, source_chain, dest_chain,
			spread_pct, gross_profit_usd, net_profit_usd, path_hops,
			path_json, fee_breakdown, calldata_json,
			input_amount_usd, slippage_est_pct, status, expires_at
		FROM opportunities
		WHERE discovered_at BETWEEN $1 AND $2
	`
	args := []interface{}{from, to}
	argIdx := 3

	if assetPair != "" {
		query += fmt.Sprintf(" AND asset_pair = $%d", argIdx)
		args = append(args, assetPair)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY discovered_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanOpportunities(rows)
}

// ExpireOldOpportunities marks stale opportunities as expired
func (r *Repository) ExpireOldOpportunities(ctx context.Context) (int64, error) {
	tag, err := r.db.Pool.Exec(ctx, `
		UPDATE opportunities SET status = 'expired'
		WHERE status = 'live' AND expires_at IS NOT NULL AND expires_at < NOW()
	`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// InsertPriceSnapshot persists a price data point
func (r *Repository) InsertPriceSnapshot(ctx context.Context, pd *types.PriceData) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO chain_price_snapshots (snapshotted_at, chain, asset_pair, price_usd, source_dex, pool_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pd.Timestamp, string(pd.Chain), pd.AssetPair, pd.PriceUSD, pd.SourceDEX, pd.PoolID)
	return err
}

// UpsertRelayChannel updates or inserts relay channel state
func (r *Repository) UpsertRelayChannel(ctx context.Context, ch *types.RelayChannel) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO relay_channels (source_chain, dest_chain, channel_id, port_id, status, pending_packets, oldest_packet_age_s, last_checked_at, last_relay_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), $8)
		ON CONFLICT (source_chain, channel_id, port_id) DO UPDATE SET
			status = EXCLUDED.status,
			pending_packets = EXCLUDED.pending_packets,
			oldest_packet_age_s = EXCLUDED.oldest_packet_age_s,
			last_checked_at = NOW(),
			last_relay_at = EXCLUDED.last_relay_at
	`, ch.SourceChain, ch.DestChain, ch.ChannelID, ch.PortID,
		ch.Status, ch.PendingPackets, ch.OldestPacketAgeS, ch.LastRelayAt)
	return err
}

// GetRelayChannels retrieves all relay channels
func (r *Repository) GetRelayChannels(ctx context.Context) ([]types.RelayChannel, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, source_chain, dest_chain, channel_id, port_id, status,
			pending_packets, oldest_packet_age_s, last_checked_at, last_relay_at
		FROM relay_channels ORDER BY source_chain, dest_chain
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []types.RelayChannel
	for rows.Next() {
		var ch types.RelayChannel
		err := rows.Scan(&ch.ID, &ch.SourceChain, &ch.DestChain, &ch.ChannelID,
			&ch.PortID, &ch.Status, &ch.PendingPackets, &ch.OldestPacketAgeS,
			&ch.LastCheckedAt, &ch.LastRelayAt)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

// InsertRelayEvent persists a relay event
func (r *Repository) InsertRelayEvent(ctx context.Context, ev *types.RelayEvent) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO relay_events (event_at, channel_id, source_chain, dest_chain, event_type, packet_sequence, relay_latency_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, ev.EventAt, ev.ChannelID, ev.SourceChain, ev.DestChain,
		ev.EventType, ev.PacketSequence, ev.RelayLatencyMs)
	return err
}

// GetRelayEvents retrieves events for a channel
func (r *Repository) GetRelayEvents(ctx context.Context, channelID string, limit int) ([]types.RelayEvent, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, event_at, channel_id, source_chain, dest_chain, event_type, packet_sequence, relay_latency_ms
		FROM relay_events
		WHERE channel_id = $1
		ORDER BY event_at DESC
		LIMIT $2
	`, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []types.RelayEvent
	for rows.Next() {
		var ev types.RelayEvent
		err := rows.Scan(&ev.ID, &ev.EventAt, &ev.ChannelID, &ev.SourceChain,
			&ev.DestChain, &ev.EventType, &ev.PacketSequence, &ev.RelayLatencyMs)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, nil
}

func scanOpportunities(rows pgx.Rows) ([]types.Opportunity, error) {
	var opps []types.Opportunity
	for rows.Next() {
		var opp types.Opportunity
		var sourceChain, destChain string
		var pathJSON, feeJSON, calldataJSON []byte
		var spreadPct, grossProfit, netProfit, inputAmount, slippage decimal.Decimal

		err := rows.Scan(
			&opp.ID, &opp.DiscoveredAt, &opp.AssetPair,
			&sourceChain, &destChain,
			&spreadPct, &grossProfit, &netProfit,
			&opp.PathHops, &pathJSON, &feeJSON, &calldataJSON,
			&inputAmount, &slippage, &opp.Status, &opp.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		opp.SourceChain = types.ChainID(sourceChain)
		opp.DestChain = types.ChainID(destChain)
		opp.SpreadPct = spreadPct
		opp.GrossProfitUSD = grossProfit
		opp.NetProfitUSD = netProfit
		opp.InputAmountUSD = inputAmount
		opp.SlippageEstPct = slippage
		opp.CalldataJSON = calldataJSON

		json.Unmarshal(pathJSON, &opp.Path)
		json.Unmarshal(feeJSON, &opp.FeeBreakdown)

		opps = append(opps, opp)
	}
	return opps, nil
}
