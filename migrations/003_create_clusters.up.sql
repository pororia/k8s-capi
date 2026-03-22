CREATE TABLE clusters (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id               UUID NOT NULL REFERENCES tenants(id),
    name                    VARCHAR(100) NOT NULL,
    kubernetes_version      VARCHAR(20) NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending',
    cp_count                INT NOT NULL,
    worker_count            INT NOT NULL,
    cp_flavor               VARCHAR(100) NOT NULL,
    worker_flavor           VARCHAR(100) NOT NULL,
    os_image                VARCHAR(100) NOT NULL,
    network_id              VARCHAR(64) NOT NULL,
    subnet_id               VARCHAR(64) NOT NULL,
    external_network_id     VARCHAR(64) NOT NULL,
    ssh_key_name            VARCHAR(100) NOT NULL,
    pod_cidr                VARCHAR(50) NOT NULL DEFAULT '10.244.0.0/16',
    service_cidr            VARCHAR(50) NOT NULL DEFAULT '10.96.0.0/12',
    cni_plugin              VARCHAR(20) NOT NULL DEFAULT 'calico',
    dns_nameservers         JSONB NOT NULL DEFAULT '["8.8.8.8"]',
    api_endpoint            VARCHAR(255),
    capi_namespace          VARCHAR(100) NOT NULL,
    capi_manifest_snapshot  JSONB,
    error_message           TEXT,
    created_by              UUID REFERENCES users(id),
    created_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMP,
    CONSTRAINT chk_clusters_status CHECK (status IN ('pending', 'provisioning', 'running', 'updating', 'failed', 'deleting', 'deleted'))
);

CREATE UNIQUE INDEX idx_clusters_tenant_name ON clusters(tenant_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_clusters_tenant_status   ON clusters(tenant_id, status);
CREATE INDEX idx_clusters_created_at      ON clusters(created_at DESC);
