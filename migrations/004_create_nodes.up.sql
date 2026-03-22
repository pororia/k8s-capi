CREATE TABLE nodes (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cluster_id          UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    name                VARCHAR(100) NOT NULL,
    role                VARCHAR(20) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'provisioning',
    machine_id          VARCHAR(100),
    openstack_server_id VARCHAR(64),
    private_ip          VARCHAR(45),
    floating_ip         VARCHAR(45),
    flavor              VARCHAR(100) NOT NULL,
    ready               BOOLEAN NOT NULL DEFAULT false,
    kubernetes_version  VARCHAR(20),
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_nodes_role   CHECK (role IN ('control_plane', 'worker')),
    CONSTRAINT chk_nodes_status CHECK (status IN ('provisioning', 'running', 'failed', 'deleting'))
);

CREATE INDEX idx_nodes_cluster_id ON nodes(cluster_id);
CREATE INDEX idx_nodes_status     ON nodes(cluster_id, status);
