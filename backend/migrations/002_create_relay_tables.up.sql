CREATE TABLE relay_channels (
  id              SERIAL PRIMARY KEY,
  source_chain    TEXT NOT NULL,
  dest_chain      TEXT NOT NULL,
  channel_id      TEXT NOT NULL,
  port_id         TEXT NOT NULL DEFAULT 'transfer',
  status          TEXT NOT NULL DEFAULT 'healthy',
  pending_packets INTEGER DEFAULT 0,
  oldest_packet_age_s INTEGER,
  last_checked_at TIMESTAMPTZ DEFAULT NOW(),
  last_relay_at   TIMESTAMPTZ,
  UNIQUE (source_chain, channel_id, port_id)
);

CREATE TABLE relay_events (
  id              UUID DEFAULT gen_random_uuid() NOT NULL,
  event_at        TIMESTAMPTZ NOT NULL,
  channel_id      TEXT NOT NULL,
  source_chain    TEXT NOT NULL,
  dest_chain      TEXT NOT NULL,
  event_type      TEXT NOT NULL,
  packet_sequence BIGINT,
  relay_latency_ms INTEGER,
  PRIMARY KEY (id, event_at)
);

SELECT create_hypertable('relay_events', 'event_at');
