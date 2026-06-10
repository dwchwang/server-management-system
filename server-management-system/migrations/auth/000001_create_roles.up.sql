CREATE TABLE IF NOT EXISTS auth_schema.roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(50)   NOT NULL UNIQUE,
    description     TEXT,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

INSERT INTO auth_schema.roles (id, name, description) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'admin',    'Full access to all resources'),
    ('a0000000-0000-0000-0000-000000000002', 'operator', 'Can read and update servers, view reports'),
    ('a0000000-0000-0000-0000-000000000003', 'viewer',   'Read-only access')
ON CONFLICT (name) DO NOTHING;
