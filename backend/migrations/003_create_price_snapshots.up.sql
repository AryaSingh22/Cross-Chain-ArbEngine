CREATE TABLE chain_price_snapshots (
  id              UUID DEFAULT gen_random_uuid() NOT NULL,
  snapshotted_at  TIMESTAMPTZ NOT NULL,
  chain           TEXT NOT NULL,
  asset_pair      TEXT NOT NULL,
  price_usd       NUMERIC(24, 10) NOT NULL,
  source_dex      TEXT NOT NULL,
  pool_id         TEXT,
  PRIMARY KEY (id, snapshotted_at)
);

SELECT create_hypertable('chain_price_snapshots', 'snapshotted_at');

CREATE INDEX idx_price_chain_pair ON chain_price_snapshots (chain, asset_pair, snapshotted_at DESC);
