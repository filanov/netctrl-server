-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Clusters table
CREATE TABLE clusters (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clusters_name ON clusters(name);

-- Agents table
CREATE TABLE agents (
    id UUID PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    hostname TEXT,
    ip_address TEXT,
    version TEXT,
    status TEXT NOT NULL DEFAULT 'AGENT_STATUS_ACTIVE',
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hardware_collected BOOLEAN NOT NULL DEFAULT false,
    network_interfaces JSONB DEFAULT '[]'::jsonb
);

-- Strategic indexes for performance
CREATE INDEX idx_agents_cluster ON agents(cluster_id);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_last_seen ON agents(last_seen);
CREATE INDEX idx_agents_hardware_pending ON agents(hardware_collected) WHERE hardware_collected = false;
CREATE INDEX idx_agents_active_last_seen ON agents(last_seen) WHERE status = 'AGENT_STATUS_ACTIVE';

-- GIN index for JSONB queries on hardware
CREATE INDEX idx_agents_network_interfaces ON agents USING gin(network_interfaces);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for auto-updating updated_at
CREATE TRIGGER update_clusters_updated_at BEFORE UPDATE ON clusters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
