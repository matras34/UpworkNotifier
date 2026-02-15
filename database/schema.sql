-- Web SSH Service Database Schema
-- PostgreSQL 14+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Organizations table (multi-tenant)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    max_concurrent_sessions INTEGER DEFAULT 10,
    max_saved_connections INTEGER DEFAULT 50,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_organizations_created_at ON organizations(created_at);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'user', 'viewer')),
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_role ON users(role);

-- SSH Connections (saved connection configs)
CREATE TABLE ssh_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 22 CHECK (port > 0 AND port < 65536),
    username VARCHAR(255) NOT NULL,
    auth_type VARCHAR(50) NOT NULL CHECK (auth_type IN ('password', 'privatekey')),
    encrypted_password TEXT,
    encrypted_private_key TEXT,
    encryption_salt VARCHAR(255),
    encryption_nonce VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_connection_name UNIQUE (user_id, name)
);

CREATE INDEX idx_ssh_connections_user_id ON ssh_connections(user_id);
CREATE INDEX idx_ssh_connections_host ON ssh_connections(host);

-- Host Keys (for host key verification)
CREATE TABLE host_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    connection_id UUID REFERENCES ssh_connections(id) ON DELETE CASCADE,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint_sha256 VARCHAR(255) NOT NULL,
    fingerprint_md5 VARCHAR(255),
    verified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_by_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_host_port_algorithm UNIQUE (host, port, algorithm)
);

CREATE INDEX idx_host_keys_host_port ON host_keys(host, port);
CREATE INDEX idx_host_keys_fingerprint ON host_keys(fingerprint_sha256);
CREATE INDEX idx_host_keys_connection_id ON host_keys(connection_id);

-- SSH Sessions (active and historical)
CREATE TABLE ssh_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    connection_id UUID REFERENCES ssh_connections(id) ON DELETE SET NULL,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    username VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('active', 'closed', 'failed', 'timeout')),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    disconnect_reason VARCHAR(255),
    bytes_sent BIGINT DEFAULT 0,
    bytes_received BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ssh_sessions_user_id ON ssh_sessions(user_id);
CREATE INDEX idx_ssh_sessions_status ON ssh_sessions(status);
CREATE INDEX idx_ssh_sessions_started_at ON ssh_sessions(started_at DESC);
CREATE INDEX idx_ssh_sessions_connection_id ON ssh_sessions(connection_id);
CREATE INDEX idx_ssh_sessions_session_token ON ssh_sessions(session_token);

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- Partition audit_logs by month for better performance
-- This should be done programmatically or via cron
-- Example for manual partitioning:
-- CREATE TABLE audit_logs_y2024m01 PARTITION OF audit_logs
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Rate Limit Tracking (can also be done in Redis)
CREATE TABLE rate_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint VARCHAR(255) NOT NULL,
    request_count INTEGER DEFAULT 0,
    window_start TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_endpoint_window UNIQUE (user_id, endpoint, window_start)
);

CREATE INDEX idx_rate_limits_user_id ON rate_limits(user_id);
CREATE INDEX idx_rate_limits_window_start ON rate_limits(window_start);

-- API Keys (for programmatic access)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    scopes TEXT[] DEFAULT ARRAY[]::TEXT[],
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ssh_connections_updated_at BEFORE UPDATE ON ssh_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rate_limits_updated_at BEFORE UPDATE ON rate_limits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default organization and admin user (for development)
INSERT INTO organizations (id, name, max_concurrent_sessions, max_saved_connections)
VALUES ('00000000-0000-0000-0000-000000000001', 'Default Organization', 100, 100);

-- Default password is 'admin123' (bcrypt hash)
-- Change this in production!
INSERT INTO users (id, organization_id, email, password_hash, role)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'admin@example.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin'
);

-- Views for analytics
CREATE VIEW active_sessions_summary AS
SELECT
    u.organization_id,
    COUNT(*) as active_session_count,
    COUNT(DISTINCT s.user_id) as active_user_count
FROM ssh_sessions s
JOIN users u ON s.user_id = u.id
WHERE s.status = 'active'
GROUP BY u.organization_id;

CREATE VIEW connection_usage_stats AS
SELECT
    sc.id,
    sc.name,
    sc.host,
    COUNT(ss.id) as total_sessions,
    COUNT(CASE WHEN ss.status = 'active' THEN 1 END) as active_sessions,
    MAX(ss.last_activity_at) as last_used_at
FROM ssh_connections sc
LEFT JOIN ssh_sessions ss ON sc.id = ss.connection_id
GROUP BY sc.id, sc.name, sc.host;
