CREATE TABLE IF NOT EXISTS report_schema.daily_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_date   DATE          NOT NULL UNIQUE,
    total_servers   INTEGER       NOT NULL,
    servers_on      INTEGER       NOT NULL,
    servers_off     INTEGER       NOT NULL,
    avg_uptime_pct  DECIMAL(5,2)  NOT NULL,
    low_uptime_servers JSONB,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_snapshots_date 
    ON report_schema.daily_snapshots(snapshot_date);
