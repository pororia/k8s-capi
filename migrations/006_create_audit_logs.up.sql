CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID REFERENCES tenants(id),
    user_id         UUID REFERENCES users(id),
    action          VARCHAR(50) NOT NULL,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     VARCHAR(100),
    resource_name   VARCHAR(200),
    request_body    JSONB,
    response_status INT,
    result          VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message   TEXT,
    ip_address      VARCHAR(45),
    user_agent      VARCHAR(500),
    duration_ms     INT,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_audit_logs_result CHECK (result IN ('success', 'failure'))
);

CREATE INDEX idx_audit_logs_tenant_created  ON audit_logs(tenant_id, created_at DESC);
CREATE INDEX idx_audit_logs_user_id         ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action          ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource        ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at      ON audit_logs(created_at DESC);
