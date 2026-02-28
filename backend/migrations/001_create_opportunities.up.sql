CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE opportunities (
  id              UUID DEFAULT gen_random_uuid() NOT NULL,
  discovered_at   TIMESTAMPTZ NOT NULL,
  asset_pair      TEXT NOT NULL,
  source_chain    TEXT NOT NULL,
  dest_chain      TEXT NOT NULL,
  spread_pct      NUMERIC(10, 6) NOT NULL,
  gross_profit_usd NUMERIC(18, 6) NOT NULL,
  net_profit_usd  NUMERIC(18, 6) NOT NULL,
  path_hops       SMALLINT NOT NULL,
  path_json       JSONB NOT NULL,
  fee_breakdown   JSONB NOT NULL,
  calldata_json   JSONB NOT NULL,
  input_amount_usd NUMERIC(18, 6),
  slippage_est_pct NUMERIC(10, 6),
  status          TEXT DEFAULT 'live',
  expires_at      TIMESTAMPTZ,
  PRIMARY KEY (id, discovered_at)
);

SELECT create_hypertable('opportunities', 'discovered_at');

CREATE INDEX idx_opp_asset_pair ON opportunities (asset_pair, discovered_at DESC);
CREATE INDEX idx_opp_chains ON opportunities (source_chain, dest_chain, discovered_at DESC);
CREATE INDEX idx_opp_net_profit ON opportunities (net_profit_usd DESC, discovered_at DESC);
CREATE INDEX idx_opp_status ON opportunities (status, discovered_at DESC);
