CREATE TABLE IF NOT EXISTS monitor_schema.health_check_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       VARCHAR(100)  NOT NULL UNIQUE,
    check_method    VARCHAR(20)   NOT NULL DEFAULT 'tcp' CHECK (check_method IN ('tcp', 'simulator')),
    tcp_port        INTEGER       DEFAULT 80,
    tcp_timeout_ms  INTEGER       DEFAULT 5000,
    uptime_rate     DECIMAL(3,2)  DEFAULT 0.95 CHECK (uptime_rate >= 0 AND uptime_rate <= 1),
    is_enabled      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hc_configs_server_id 
    ON monitor_schema.health_check_configs(server_id);
CREATE INDEX IF NOT EXISTS idx_hc_configs_enabled 
    ON monitor_schema.health_check_configs(is_enabled) WHERE is_enabled = TRUE;
