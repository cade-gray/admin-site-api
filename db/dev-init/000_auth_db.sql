-- Creates and grants the shared auth_db database for local dev.
-- Schema mirrors cade-gray/jokedle-api's auth-db-setup.sql, since
-- admin-site-api and jokedle-api both authenticate against it.
-- See https://github.com/cade-gray/jokedle-api/blob/main/auth-db-setup.sql

CREATE DATABASE auth_db;

\c auth_db

CREATE TABLE tokens (
  idtokens SERIAL PRIMARY KEY,
  createDateTime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  username VARCHAR(45) NOT NULL,
  token VARCHAR(45) NOT NULL,
  CONSTRAINT token_UNIQUE UNIQUE (token)
);

CREATE TABLE users (
  idusers SERIAL PRIMARY KEY,
  username VARCHAR(45) NOT NULL,
  password VARCHAR(100) NOT NULL,
  CONSTRAINT username_unique UNIQUE (username)
);

-- No seed user here on purpose (jokedle-api's plaintext seed insert won't
-- bcrypt-compare against this app's /login). Seed one manually per the
-- readme's "Creating a login user" section.

GRANT ALL ON SCHEMA public TO gorm;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO gorm;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO gorm;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO gorm;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO gorm;
