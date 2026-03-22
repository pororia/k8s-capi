CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    os_auth_url     VARCHAR(255) NOT NULL,
    os_project_id   VARCHAR(64) NOT NULL,
    os_project_name VARCHAR(100) NOT NULL,
    os_domain_name  VARCHAR(100) NOT NULL DEFAULT 'Default',
    os_credentials  TEXT NOT NULL,           -- AES-256-GCM encrypted JSON
    max_clusters    INT NOT NULL DEFAULT 10,
    max_nodes_per_cluster INT NOT NULL DEFAULT 20,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP
);

CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
