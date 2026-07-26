-- Schema for the local Docker dev database (see ../../compose.dev.yaml).
-- The postgres image already creates the "admin_db" database and "gorm"
-- role from POSTGRES_DB/POSTGRES_USER/POSTGRES_PASSWORD, so unlike
-- CREATE_SQLS.sql this file only creates tables/triggers.

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- users/tokens (auth) live in auth_db, seeded by 000_auth_db.sql.

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
