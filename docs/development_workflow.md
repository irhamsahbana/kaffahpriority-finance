# Development Workflow

## Prerequisites
- **Go**: Version 1.24.0 or later.
- **Docker**: For running PostgreSQL and other services.
- **Task**: [Taskfile](https://taskfile.dev) for running commands (install via `brew install go-task/tap/go-task` or see docs).
- **Goose**: For database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`).

## Setup

1.  **Clone the repository**:
    ```bash
    git clone <repo_url>
    cd kaffahpriority-finance
    ```
2.  **Environment Variables**:
    - Copy `.env.example` to `.env`.
    - Update values as needed.
    ```bash
    cp .env.example .env
    ```
3.  **Start Dependencies**:
    ```bash
    task docker-up
    ```
4.  **Run Migrations**:
    ```bash
    task migrate cmd=up
    ```
5.  **Seed Database**:
    ```bash
    task seed table=all
    ```

## Running the Application

- **Development Mode** (Hot Reload if configured, or just run):
  ```bash
  task dev
  ```
- **Websocket Server**:
  ```bash
  task ws
  ```

## Database Migrations

- **Create a new migration**:
  ```bash
  task create-migration name=create_users_table
  ```
- **Run migrations**:
  ```bash
  task migrate cmd=up
  ```
- **Rollback migration**:
  ```bash
  task migrate cmd=down
  ```

## Testing & Linting

- **Run Linter**:
  ```bash
  task lint-ci
  ```
- **Fix Formatting**:
  ```bash
  task lint-fix
  ```
