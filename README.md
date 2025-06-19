# KdnSite

A modern, open-source platform for visual programming, interactive quizzes, and creative projects. Built with Go, Templ, Tailwind CSS, and Neon (PostgreSQL).

## Features

- Visual drag-and-drop programming and project creation
- Interactive quizzes and instant feedback
- User authentication and account management (Auth0)
- Theme switching (light/dark)
- User and quiz data stored in PostgreSQL
- Requires a running PostgreSQL database instance
- Fully containerized with Docker
- Community features and progress tracking

## Getting Started

### Prerequisites

- Go 1.21 or newer
- Docker (optional, for containerized deployment)
- Neon (PostgreSQL) database instance

### Setup

1. **Clone the repository:**

   ```sh
   git clone https://KdnSite.git
   cd KdnSite
   ```

2. **Copy the example environment file and edit as needed:**

   ```sh
   cp .env.example .env
   # Edit .env with your Neon DB URL, Auth0 credentials, and session keys
   ```

   - `POSTGRES_DATABASE_URL`: Your PostgreSQL connection string (e.g. `postgres://user:password@localhost:5432/dbname?sslmode=disable`).
   - `AUTH0_DOMAIN`: Your Auth0 domain (e.g., `dev-xxxxxx.eu.auth0.com`).
   - `AUTH0_CLIENT_ID`: Your Auth0 client ID.cookie sessions.
   - `GO_ENV`: Set to `development` or `production` as needed.
   - `PORT`: (Optional) Port to run the server on (default: 8090).

3. **Install dependencies and generate Templ files:**

   ```sh
   go mod download
   go install github.com/a-h/templ/cmd/templ@latest
   templ generate
   ```

4. **Run the application:**

   ```sh
   go run cmd/server/main.go
   # or with Docker
   docker build -t kdnsite .
   docker run --env-file .env -p 8090:8090 kdnsite
   ```

5. **Visit** [http://localhost:8090](http://localhost:8090) in your browser.

## Environment Variables

Edit your `.env` file with your Postgres DB URL, Auth0 credentials, and session keys:

- `POSTGRES_DATABASE_URL`: Your PostgreSQL connection string (e.g. `postgres://user:password@localhost:5432/dbname?sslmode=disable`)
- `AUTH0_DOMAIN`: Your Auth0 domain (e.g., `dev-xxxxxx.eu.auth0.com`).
- `AUTH0_CLIENT_ID`: Your Auth0 client ID.
- `SESSION_HASH_KEY` and `SESSION_BLOCK_KEY`: Random strings for secure cookie sessions. You can generate them with:

  ```sh
  # Generate a 32-byte random base64 string for each key
  head -c 32 /dev/urandom | base64
  ```

- `GO_ENV`: Set to `development` or `production` as needed.
- `PORT`: (Optional) Port to run the server on (default: 8090).

## License

MIT
