CREATE TABLE IF NOT EXISTS report_schema.report_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_type     VARCHAR(20)   NOT NULL CHECK (report_type IN ('daily', 'on_demand')),
    status          VARCHAR(20)   NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    start_date      DATE          NOT NULL,
    end_date        DATE          NOT NULL,
    recipient_email VARCHAR(255)  NOT NULL,
    total_servers   INTEGER,
    servers_on      INTEGER,
    servers_off     INTEGER,
    avg_uptime_pct  DECIMAL(5,2),
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_report_jobs_type ON report_schema.report_jobs(report_type);
CREATE INDEX IF NOT EXISTS idx_report_jobs_status ON report_schema.report_jobs(status);
CREATE INDEX IF NOT EXISTS idx_report_jobs_created ON report_schema.report_jobs(created_at);
