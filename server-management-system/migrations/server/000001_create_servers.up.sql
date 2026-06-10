CREATE TABLE IF NOT EXISTS server_schema.servers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       VARCHAR(100)  NOT NULL UNIQUE,
    server_name     VARCHAR(255)  NOT NULL UNIQUE,
    status          VARCHAR(20)   NOT NULL DEFAULT 'off' CHECK (status IN ('on', 'off')),
    ipv4            VARCHAR(15)   NOT NULL,
    os              VARCHAR(100),
    cpu_cores       INTEGER       CHECK (cpu_cores > 0),
    ram_gb          DECIMAL(10,2) CHECK (ram_gb > 0),
    disk_gb         DECIMAL(10,2) CHECK (disk_gb > 0),
    location        VARCHAR(255),
    description     TEXT,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_servers_server_id 
    ON server_schema.servers(server_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_servers_server_name 
    ON server_schema.servers(server_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_servers_status 
    ON server_schema.servers(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_servers_ipv4 
    ON server_schema.servers(ipv4) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_servers_created_at 
    ON server_schema.servers(created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_servers_status_created 
    ON server_schema.servers(status, created_at DESC) WHERE deleted_at IS NULL;
