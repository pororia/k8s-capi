CREATE TABLE cluster_events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cluster_id  UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    severity    VARCHAR(20) NOT NULL DEFAULT 'info',
    message     TEXT NOT NULL,
    source      VARCHAR(100),
    metadata    JSONB,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_cluster_events_severity CHECK (severity IN ('info', 'warning', 'error'))
);

CREATE INDEX idx_cluster_events_cluster_created ON cluster_events(cluster_id, created_at DESC);
