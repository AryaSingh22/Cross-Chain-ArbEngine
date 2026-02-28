package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Feeds    FeedConfig     `yaml:"feeds"`
	Engine   EngineConfig   `yaml:"engine"`
	Relay    RelayConfig    `yaml:"relay"`
}

type ServerConfig struct {
	Port     string `yaml:"port"`
	GRPCPort string `yaml:"grpcPort"`
	CORSOrigin string `yaml:"corsOrigin"`
}

type DatabaseConfig struct {
	URL             string        `yaml:"url"`
	MaxConns        int           `yaml:"maxConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
}

type RedisConfig struct {
	URL string `yaml:"url"`
}

type FeedConfig struct {
	UseMock      bool          `yaml:"useMock"`
	PollInterval time.Duration `yaml:"pollInterval"`
	Chains       []ChainConfig `yaml:"chains"`
}

type ChainConfig struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	GRPCURL  string `yaml:"grpcUrl"`
	RESTURL  string `yaml:"restUrl"`
	Enabled  bool   `yaml:"enabled"`
}

type EngineConfig struct {
	MinNetProfitUSD  float64       `yaml:"minNetProfitUsd"`
	OpportunityTTL   time.Duration `yaml:"opportunityTtl"`
	MaxPathHops      int           `yaml:"maxPathHops"`
	InputAmountUSD   float64       `yaml:"inputAmountUsd"`
}

type RelayConfig struct {
	PollInterval        time.Duration `yaml:"pollInterval"`
	PendingPacketThreshold int       `yaml:"pendingPacketThreshold"`
	StuckPacketAgeSec   int           `yaml:"stuckPacketAgeSec"`
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:       getEnv("SERVER_PORT", "8080"),
			GRPCPort:   getEnv("GRPC_PORT", "9090"),
			CORSOrigin: getEnv("CORS_ORIGIN", "http://localhost:3000"),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://arbengine:arbengine_dev@localhost:5432/arbengine?sslmode=disable"),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 10),
			ConnMaxLifetime: 30 * time.Minute,
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		Feeds: FeedConfig{
			UseMock:      getEnvBool("USE_MOCK_FEEDS", true),
			PollInterval: 5 * time.Second,
			Chains: []ChainConfig{
				{ID: "osmosis", Name: "Osmosis", GRPCURL: "grpc.osmosis.zone:443", Enabled: true},
				{ID: "injective", Name: "Injective", GRPCURL: "grpc.injective.network:443", Enabled: true},
				{ID: "neutron", Name: "Neutron", RESTURL: "https://rest.neutron.org", Enabled: true},
				{ID: "stride", Name: "Stride", GRPCURL: "grpc.stride.zone:443", Enabled: true},
				{ID: "juno", Name: "Juno", GRPCURL: "grpc.juno.strange.love:443", Enabled: true},
				{ID: "cosmoshub", Name: "Cosmos Hub", GRPCURL: "grpc.cosmos.directory:443", Enabled: true},
				{ID: "akash", Name: "Akash", GRPCURL: "grpc.akash.network:443", Enabled: true},
			},
		},
		Engine: EngineConfig{
			MinNetProfitUSD: getEnvFloat("MIN_NET_PROFIT_USD", 5.0),
			OpportunityTTL:  5 * time.Minute,
			MaxPathHops:     3,
			InputAmountUSD:  10000.0,
		},
		Relay: RelayConfig{
			PollInterval:           15 * time.Second,
			PendingPacketThreshold: 10,
			StuckPacketAgeSec:      300,
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
