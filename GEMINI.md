# Project Overview

This project is a Go-based application named "kaffapriority-finance". It appears to be a web server with database interactions, background workers, and other features.

## User Preferences

- Be objective and truthful, even if it may be difficult to hear.
- When editing files, always use absolute paths.
- Ask for permission before modifying or creating any files.
- When making changes to a file, explain why the change is being made.
- When generating code, add comments in English.
- When creating a commit message, follow the pattern and style of previous commit messages, check on unstaged changes first and then staged changes.
- When asked to commit, first review the changes that have been made, especially in staging.

## How to Run the Application

- **Development:** `go run ./cmd/bin/main.go` or `task dev`
- **Websocket Server:** `go run ./cmd/bin/main.go ws --port=8080` or `task ws`

## How to Build the Application

- **Build:** `go build -o ./kpf-app ./cmd/bin/main.go` or `task build`
- **Build for Development:** `task build-dev`
- **Build for Staging:** `task build-staging`
- **Build for Production:** `task build-production`

## How to Run Tests

The `Taskfile.yml` does not explicitly define a test command. However, based on the dependencies, the project likely uses the standard `go test` command. You can run tests using:

```bash
go test ./...
```

## How to Manage the Database

- **Create a new migration:** `task create-migration name=<migration_name>`
- **Run migrations:** `task migrate cmd="up"`
- **Rollback migrations:** `task migrate cmd="down"`
- **Seed the database:** `task seed table=<table_name>`
- **Restore database from backup:** `task restore file=<backup_file_name>`

## Linting

- **Fix linting issues:** `task lint-fix`
- **Run linter:** `task lint-ci`

## Key Dependencies

- **Web Framework:** Fiber (`github.com/gofiber/fiber/v2`)
- **Database Driver:** `github.com/lib/pq` (PostgreSQL)
- **Database Query Builder:** `github.com/jmoiron/sqlx`
- **Task Runner:** `github.com/go-task/task`
- **Logging:** `github.com/rs/zerolog`
- **Authentication:** `github.com/golang-jwt/jwt/v5`
- **Websockets:** `github.com/gorilla/websocket`
- **Messaging:** `github.com/nats-io/nats.go`

## Other Information

- The application uses `go.mod` for dependency management.
- The main application entry point is likely `cmd/bin/main.go`.
- The application uses a `.env` file for environment variables.
- The `Taskfile.yml` provides a convenient way to run common commands.


