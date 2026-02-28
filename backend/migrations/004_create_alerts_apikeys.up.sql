CREATE TABLE alert_configs (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  api_key_id          UUID,
  min_net_profit_usd  NUMERIC(18, 6) DEFAULT 10.0,
  min_spread_pct      NUMERIC(10, 6) DEFAULT 0.5,
  asset_pairs         TEXT[] DEFAULT '{}',
  source_chains       TEXT[] DEFAULT '{}',
  notification_type   TEXT NOT NULL,
  webhook_url         TEXT,
  telegram_chat_id    TEXT,
  enabled             BOOLEAN DEFAULT TRUE,
  created_at          TIMESTAMPTZ DEFAULT NOW(),
  updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE api_keys (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key_hash        TEXT NOT NULL UNIQUE,
  name            TEXT NOT NULL,
  rate_limit_rpm  INTEGER DEFAULT 60,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  last_used_at    TIMESTAMPTZ,
  revoked         BOOLEAN DEFAULT FALSE
);
