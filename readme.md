# admin-site-api

A Go-based REST API for my Admin Site, built with Gin and GORM, containerized for easy deployment.
This project is designed for learning, prototyping, and as a foundation for a production API on my VPS 2.0.

## Features

- Token-based authentication (login/logout) backed by Postgres
- PostgreSQL integration via GORM ORM
- Dockerized for local and remote deployment
- Environment-based configuration (dev/prod)
- Ready for deployment to GitHub Container Registry and DigitalOcean VPS

## Endpoints

| Method | Path        | Auth required | Description                                                     |
| ------ | ----------- | ------------- | --------------------------------------------------------------- |
| POST   | `/login`    | No            | Authenticate with `username`/`password`, returns a bearer token |
| POST   | `/logout`   | Yes           | Invalidates the bearer token used on the request                |
| GET    | `/projects` | No            | List all projects                                               |
| POST   | `/joke`     | Yes           | Pull a random joke from the RapidAPI dad-jokes API              |

Authenticated routes expect an `Authorization: Bearer <token>` header, using a token returned from `/login`.

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL database
- A [RapidAPI](https://rapidapi.com/KegenGuyll/api/dad-jokes) key if you want to use `POST /joke`

### Local Development

1. Clone the repo:

   ```sh
   git clone https://github.com/cade-gray/admin-site-api.git
   cd admin-site-api
   ```

2. Set up the database (see [Database Setup](#database-setup) below).

3. Edit `compose.yaml` to set your database credentials and RapidAPI key:

   ```yaml
   services:
     server:
       build:
         context: .
         target: final
       ports:
         - 4269:4269
       environment:
         - DB_HOST=host
         - DB_USER=user
         - DB_PASSWORD=password
         - DB_NAME=dbname
         - AUTH_DB_NAME=auth_db
         - DB_PORT=5432
         - DB_SSLMODE=disable
         - RAPIDAPI_KEY=your_rapidapi_key
   ```

4. Build and run with Docker Compose:

   ```sh
   docker compose up --build
   ```

   Or run locally without Docker:

   ```sh
   export DB_HOST=localhost DB_USER=gorm DB_PASSWORD=your_secure_password \
          DB_NAME=admin_db AUTH_DB_NAME=auth_db DB_PORT=5432 DB_SSLMODE=disable \
          RAPIDAPI_KEY=your_rapidapi_key
   go run .
   ```

5. Access the API at [http://localhost:4269](http://localhost:4269)

### Production Deployment

#### Manual Deployment

1. Build and push your image to GitHub Container Registry:

   ```sh
   docker build -t ghcr.io/cade-gray/admin-site-api:latest .
   docker push ghcr.io/cade-gray/admin-site-api:latest
   ```

2. On your VPS, create a `compose.yaml` pointing at the published image, with your production database credentials and RapidAPI key:

   ```yaml
   services:
     server:
       image: ghcr.io/cade-gray/admin-site-api:latest
       ports:
         - 4269:4269
       environment:
         - DB_HOST=prod-db-host
         - DB_USER=produser
         - DB_PASSWORD=prodpass
         - DB_NAME=proddb
         - AUTH_DB_NAME=auth_db
         - DB_PORT=5432
         - DB_SSLMODE=disable
         - RAPIDAPI_KEY=your_rapidapi_key
       restart: always
   ```

3. Pull the image and run with Docker Compose:

   ```sh
   docker compose up -d
   ```

#### Automated Deployment (GitHub Actions)

The project includes a GitHub Actions workflow (`.github/workflows/deploy.yml`) that automatically:

1. **Builds and pushes** Docker images to GitHub Container Registry on every push to `main`
2. **Deploys to your VPS** automatically using SSH
3. **Supports multi-platform builds** (ARM64 for Mac, AMD64 for VPS)

**Setup Requirements:**

1. **GitHub Secrets** - Add these in your repository settings:
   - `VPS_HOST`: Your VPS IP address
   - `VPS_USER`: SSH username (e.g., `root`)
   - `VPS_SSH_KEY`: Your private SSH key
   - `GH_TOKEN`: Token with `packages:write` used to push/pull images from GHCR

2. **VPS Setup**:

   ```sh
   # Create deployment directory
   sudo mkdir -p /srv/admin-site-api

   # Copy your compose file
   scp compose.yaml your-user@your-vps:/srv/admin-site-api/
   ```

3. **SSH Key Generation**:

   ```sh
   ssh-keygen -t ed25519 -C "github-actions-deploy"
   ssh-copy-id -i ~/.ssh/id_ed25519.pub your-user@your-vps
   ```

> **Note:** the deploy workflow currently runs `docker compose -f compose.dev.yaml ...` on the VPS. Make sure a `compose.dev.yaml` exists in `/srv/admin-site-api` on the server (or update the workflow to reference `compose.yaml`) before relying on automated deploys.

**Workflow Triggers:**

- Push to `main` branch → Build + Deploy
- Pull requests → Build only (no deployment)

Your API will be available at your VPS IP on port 4269.

## Database Setup

`CREATE_SQLS.sql` contains the schema for this app's own data (plates, trips\*, projects). Create the database and apply it:

```sql
CREATE DATABASE ;
CREATE USER gorm WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE admin_db TO gorm;
```

```sh
psql -U postgres -d admin_db -f CREATE_SQLS.sql
```

### Auth database (auth_db)

Login/token data lives in a separate `auth_db` database, shared with [jokedle-api](https://github.com/cade-gray/jokedle-api). Its schema is defined in [`auth-db-setup.sql`](https://github.com/cade-gray/jokedle-api/blob/main/auth-db-setup.sql) in that repo — apply it there rather than in this repo. Point this app at it with the `AUTH_DB_NAME` environment variable (same `DB_HOST`/`DB_USER`/`DB_PASSWORD`/`DB_PORT`/`DB_SSLMODE` as the main database).

For local Docker dev, `db/dev-init/000_auth_db.sql` creates and grants `auth_db` automatically in the same dev Postgres container.

### Creating a login user

`POST /login` checks `auth_db`'s `users` table for a bcrypt-hashed password. There's no signup endpoint yet, so seed a user manually. From a Go shell (`go run`) or `gorm` console:

```go
hashed, _ := bcrypt.GenerateFromPassword([]byte("your-password"), bcrypt.DefaultCost)
authDb.Exec("INSERT INTO users (username, password) VALUES (?, ?)", "your-username", string(hashed))
```

## Project Structure

```text
main.go                # Entry point
lib/                    # Shared helpers (auth middleware, token generation)
routes/                 # API route handlers
Dockerfile              # Build instructions
compose.yaml             # Docker Compose config
CREATE_SQLS.sql          # Database schema
.github/workflows/       # CI/CD pipeline
```

## Future Plans

- Harden for production (security, logging, error handling)
- Add token expiry and refresh

---
