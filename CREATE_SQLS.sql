-- Create  database
CREATE DATABASE admin_db;
\c admin_db;
--------------------------------------------------------------
-- Gorm user creation and privilege granting
-- Note: Run these commands in the psql shell or as a superuser
-- Connect to postgres as the admin user
psql -U postgres

-- (Optional) Connect to the database and grant privileges on tables/sequences
\c admin_db
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gorm;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gorm;

-- Trigger function used by the updated_at triggers below
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- users/tokens (auth) now live in the shared auth_db, not here.
-- See https://github.com/cade-gray/jokedle-api/blob/main/auth-db-setup.sql
-- for that schema; admin-site-api connects to it via AUTH_DB_NAME.

-- Create projects table
CREATE TABLE projects (
    project_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    github_url VARCHAR(2048),
    image_url VARCHAR(2048),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- If the projects table already exists, run these instead of recreating it
ALTER TABLE projects ADD COLUMN IF NOT EXISTS github_url VARCHAR(2048);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS image_url VARCHAR(2048);

--------------------------------------------------------------
-- Re-grant so gorm has access to the tables created above (the earlier
-- GRANT ALL ON ALL TABLES only covers tables that existed at that point),
-- and set default privileges so this doesn't need repeating for future
-- tables.
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gorm;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gorm;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO gorm;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO gorm;