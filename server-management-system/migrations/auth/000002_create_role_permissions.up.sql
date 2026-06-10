CREATE TABLE IF NOT EXISTS auth_schema.role_permissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id         UUID          NOT NULL REFERENCES auth_schema.roles(id) ON DELETE CASCADE,
    scope           VARCHAR(100)  NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, scope)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id 
    ON auth_schema.role_permissions(role_id);

INSERT INTO auth_schema.role_permissions (role_id, scope) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'server:create'),
    ('a0000000-0000-0000-0000-000000000001', 'server:read'),
    ('a0000000-0000-0000-0000-000000000001', 'server:update'),
    ('a0000000-0000-0000-0000-000000000001', 'server:delete'),
    ('a0000000-0000-0000-0000-000000000001', 'server:import'),
    ('a0000000-0000-0000-0000-000000000001', 'server:export'),
    ('a0000000-0000-0000-0000-000000000001', 'report:view'),
    ('a0000000-0000-0000-0000-000000000001', 'report:send'),
    ('a0000000-0000-0000-0000-000000000001', 'user:manage'),
    ('a0000000-0000-0000-0000-000000000002', 'server:read'),
    ('a0000000-0000-0000-0000-000000000002', 'server:update'),
    ('a0000000-0000-0000-0000-000000000002', 'report:view'),
    ('a0000000-0000-0000-0000-000000000003', 'server:read'),
    ('a0000000-0000-0000-0000-000000000003', 'report:view')
ON CONFLICT (role_id, scope) DO NOTHING;
