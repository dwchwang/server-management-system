CREATE TABLE IF NOT EXISTS fileio_schema.import_job_details (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    import_job_id   UUID          NOT NULL REFERENCES fileio_schema.import_jobs(id) ON DELETE CASCADE,
    row_number      INTEGER       NOT NULL,
    server_id       VARCHAR(100),
    server_name     VARCHAR(255),
    status          VARCHAR(20)   NOT NULL CHECK (status IN ('success', 'failed')),
    error_reason    TEXT,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_import_details_job_id ON fileio_schema.import_job_details(import_job_id);
CREATE INDEX IF NOT EXISTS idx_import_details_status ON fileio_schema.import_job_details(status);
