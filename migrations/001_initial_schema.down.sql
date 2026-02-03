-- Drop triggers
DROP TRIGGER IF EXISTS update_agents_updated_at ON agents;
DROP TRIGGER IF EXISTS update_clusters_updated_at ON clusters;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (cascade will handle foreign keys)
DROP TABLE IF EXISTS agents CASCADE;
DROP TABLE IF EXISTS clusters CASCADE;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
