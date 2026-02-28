package relay

import (
	"context"
	"math/rand"
	"time"

	"github.com/cosmos-arbengine/backend/internal/db"
	"github.com/cosmos-arbengine/backend/internal/events"
	"github.com/cosmos-arbengine/backend/internal/types"
	"go.uber.org/zap"
)

// Monitor tracks IBC packet state across channels
type Monitor struct {
	repo          *db.Repository
	eventBus      *events.EventBus
	logger        *zap.Logger
	pollInterval  time.Duration
	pendingThreshold int
	stuckAgeSec   int
	channels      []types.RelayChannel
}

// NewMonitor creates a new IBC relay monitor
func NewMonitor(
	repo *db.Repository,
	eventBus *events.EventBus,
	logger *zap.Logger,
	pollInterval time.Duration,
	pendingThreshold int,
	stuckAgeSec int,
) *Monitor {
	return &Monitor{
		repo:          repo,
		eventBus:      eventBus,
		logger:        logger,
		pollInterval:  pollInterval,
		pendingThreshold: pendingThreshold,
		stuckAgeSec:   stuckAgeSec,
		channels:      buildDefaultChannels(),
	}
}

// Start begins the relay monitoring loop
func (m *Monitor) Start(ctx context.Context) {
	m.logger.Info("IBC relay monitor started", zap.Int("channels", len(m.channels)))

	// Initial check
	m.checkAllChannels(ctx)

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAllChannels(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (m *Monitor) checkAllChannels(ctx context.Context) {
	for i := range m.channels {
		ch := &m.channels[i]
		m.simulateChannelState(ch)

		// Persist channel state
		if err := m.repo.UpsertRelayChannel(ctx, ch); err != nil {
			m.logger.Error("failed to upsert relay channel", zap.Error(err), zap.String("channel", ch.ChannelID))
		}

		// Check thresholds
		if ch.PendingPackets > m.pendingThreshold {
			m.eventBus.Publish(events.Event{
				Type:    events.EventRelayAlert,
				Payload: ch,
			})
		}

		if ch.OldestPacketAgeS != nil && *ch.OldestPacketAgeS > m.stuckAgeSec {
			m.eventBus.Publish(events.Event{
				Type:    events.EventPacketStuck,
				Payload: ch,
			})
		}

		// Generate relay events
		m.generateMockRelayEvent(ctx, ch)
	}
}

func (m *Monitor) simulateChannelState(ch *types.RelayChannel) {
	// Simulate realistic IBC channel state
	ch.LastCheckedAt = time.Now()

	// Most channels are healthy
	r := rand.Float64()
	if r < 0.70 {
		ch.Status = "healthy"
		ch.PendingPackets = rand.Intn(3)
		age := rand.Intn(30)
		ch.OldestPacketAgeS = &age
	} else if r < 0.90 {
		ch.Status = "backlogged"
		ch.PendingPackets = 5 + rand.Intn(15)
		age := 60 + rand.Intn(240)
		ch.OldestPacketAgeS = &age
	} else if r < 0.97 {
		ch.Status = "stuck"
		ch.PendingPackets = 10 + rand.Intn(30)
		age := 300 + rand.Intn(600)
		ch.OldestPacketAgeS = &age
	} else {
		ch.Status = "closed"
		ch.PendingPackets = 0
		ch.OldestPacketAgeS = nil
	}

	now := time.Now()
	if ch.Status == "healthy" {
		relayAt := now.Add(-time.Duration(rand.Intn(10)) * time.Second)
		ch.LastRelayAt = &relayAt
	}
}

func (m *Monitor) generateMockRelayEvent(ctx context.Context, ch *types.RelayChannel) {
	if rand.Float64() > 0.3 {
		return // Only generate events sometimes
	}

	eventTypes := []string{"packet_sent", "ack", "timeout"}
	if ch.Status == "stuck" {
		eventTypes = append(eventTypes, "stuck")
	}

	seq := int64(rand.Intn(100000))
	latency := rand.Intn(5000)

	ev := &types.RelayEvent{
		EventAt:        time.Now(),
		ChannelID:      ch.ChannelID,
		SourceChain:    ch.SourceChain,
		DestChain:      ch.DestChain,
		EventType:      eventTypes[rand.Intn(len(eventTypes))],
		PacketSequence: &seq,
		RelayLatencyMs: &latency,
	}

	if err := m.repo.InsertRelayEvent(ctx, ev); err != nil {
		m.logger.Error("failed to insert relay event", zap.Error(err))
	}
}

func buildDefaultChannels() []types.RelayChannel {
	return []types.RelayChannel{
		{SourceChain: "osmosis", DestChain: "cosmoshub", ChannelID: "channel-0", PortID: "transfer", Status: "healthy"},
		{SourceChain: "osmosis", DestChain: "injective", ChannelID: "channel-122", PortID: "transfer", Status: "healthy"},
		{SourceChain: "osmosis", DestChain: "neutron", ChannelID: "channel-874", PortID: "transfer", Status: "healthy"},
		{SourceChain: "osmosis", DestChain: "stride", ChannelID: "channel-326", PortID: "transfer", Status: "healthy"},
		{SourceChain: "osmosis", DestChain: "juno", ChannelID: "channel-42", PortID: "transfer", Status: "healthy"},
		{SourceChain: "osmosis", DestChain: "akash", ChannelID: "channel-1", PortID: "transfer", Status: "healthy"},
		{SourceChain: "injective", DestChain: "cosmoshub", ChannelID: "channel-1", PortID: "transfer", Status: "healthy"},
		{SourceChain: "injective", DestChain: "neutron", ChannelID: "channel-60", PortID: "transfer", Status: "healthy"},
		{SourceChain: "neutron", DestChain: "cosmoshub", ChannelID: "channel-1", PortID: "transfer", Status: "healthy"},
		{SourceChain: "neutron", DestChain: "stride", ChannelID: "channel-8", PortID: "transfer", Status: "healthy"},
		{SourceChain: "stride", DestChain: "cosmoshub", ChannelID: "channel-0", PortID: "transfer", Status: "healthy"},
		{SourceChain: "juno", DestChain: "cosmoshub", ChannelID: "channel-1", PortID: "transfer", Status: "healthy"},
	}
}
