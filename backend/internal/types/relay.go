package types

import "time"

// RelayChannel represents the state of an IBC relay channel
type RelayChannel struct {
	ID               int       `json:"id"`
	SourceChain      string    `json:"sourceChain"`
	DestChain        string    `json:"destChain"`
	ChannelID        string    `json:"channelId"`
	PortID           string    `json:"portId"`
	Status           string    `json:"status"` // healthy | backlogged | stuck | closed
	PendingPackets   int       `json:"pendingPackets"`
	OldestPacketAgeS *int      `json:"oldestPacketAgeS,omitempty"`
	LastCheckedAt    time.Time `json:"lastCheckedAt"`
	LastRelayAt      *time.Time `json:"lastRelayAt,omitempty"`
}

// RelayEvent represents an IBC relay event
type RelayEvent struct {
	ID             string    `json:"id"`
	EventAt        time.Time `json:"eventAt"`
	ChannelID      string    `json:"channelId"`
	SourceChain    string    `json:"sourceChain"`
	DestChain      string    `json:"destChain"`
	EventType      string    `json:"eventType"` // packet_sent | ack | timeout | stuck
	PacketSequence *int64    `json:"packetSequence,omitempty"`
	RelayLatencyMs *int      `json:"relayLatencyMs,omitempty"`
}

// AlertConfig represents user-defined alert thresholds
type AlertConfig struct {
	ID               string    `json:"id"`
	APIKeyID         *string   `json:"apiKeyId,omitempty"`
	MinNetProfitUSD  float64   `json:"minNetProfitUsd"`
	MinSpreadPct     float64   `json:"minSpreadPct"`
	AssetPairs       []string  `json:"assetPairs"`
	SourceChains     []string  `json:"sourceChains"`
	NotificationType string    `json:"notificationType"` // webhook | telegram | inapp
	WebhookURL       *string   `json:"webhookUrl,omitempty"`
	TelegramChatID   *string   `json:"telegramChatId,omitempty"`
	Enabled          bool      `json:"enabled"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// APIKey represents an API key for REST/gRPC access
type APIKey struct {
	ID           string     `json:"id"`
	KeyHash      string     `json:"-"`
	Name         string     `json:"name"`
	RateLimitRPM int        `json:"rateLimitRpm"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastUsedAt   *time.Time `json:"lastUsedAt,omitempty"`
	Revoked      bool       `json:"revoked"`
}
