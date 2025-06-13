# SchoolScienceHelper

A modern, open-source visual programming platform inspired by Scratch, built with Go, Templ, and Tailwind CSS.

## Features

- Visual drag-and-drop programming
- User authentication and account management
- Theme switching (light/dark)
- Project and user data stored in MongoDB
- Fully containerized with Docker

## Getting Started

### Prerequisites

- Go 1.21 or newer
- Docker (optional, for containerized deployment)
- MongoDB instance (local or remote)

### Setup

1. **Clone the repository:**

   ```sh
   git clone https://SchoolScienceHelper.git
   cd SchoolScienceHelper
   ```

2. **Copy the example environment file and edit as needed:**

   ```sh
   cp .env.example .env
   # Edit .env with your MongoDB URI, database name, and session keys
   ```

   - `MONGODB_URI`: Your MongoDB connection string. You can use a local instance or a cloud provider like MongoDB Atlas.
   - `MONGODB_DB`: The database name to use (default: `SchoolScienceHelper`).
   - `SESSION_HASH_KEY` and `SESSION_BLOCK_KEY`: Random strings for secure cookie sessions. You can generate them with:

     ```sh
     # Generate a 32-byte random base64 string for each key
     head -c 32 /dev/urandom | base64
     ```

   - `GO_ENV`: Set to `development` or `production` as needed.

3. **Install dependencies and generate Templ files:**

   ```sh
   go mod download
   go install github.com/a-h/templ/cmd/templ@latest
   templ generate
   ```

4. **Run the application:**

   ```sh
   go run main.go
   # or with Docker
   docker build -t SchoolScienceHelper .
   docker run --env-file .env -p 8090:8090 SchoolScienceHelper
   ```

5. **Visit** [http://localhost:8090](http://localhost:8090) in your browser.

## Environment Variables

See `.env.example` for all required variables. You must set your own secure values for production.

## License

MIT
