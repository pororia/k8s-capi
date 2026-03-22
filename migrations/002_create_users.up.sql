CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID REFERENCES tenants(id) ON DELETE SET NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'viewer',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    last_login_at   TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT chk_users_role   CHECK (role IN ('super_admin', 'tenant_admin', 'operator', 'viewer')),
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'inactive', 'suspended'))
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email     ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role      ON users(role) WHERE deleted_at IS NULL;
