# Architecture

The project follows the **Standard Go Project Layout** combined with **Clean Architecture** principles.

## Directory Structure

- **cmd/**: Main applications.
  - `bin/main.go`: The entry point for the application.
- **internal/**: Private application and library code.
  - **entity/**: Domain entities and data structures.
  - **module/**: Application modules (Clean Architecture layers).
    - `[module_name]/`
      - `handler/`: HTTP handlers (Controller layer).
      - `service/`: Business logic (Usecase layer).
      - `repository/`: Data access layer.
  - **adapter/**: Interface adapters (e.g., PostgreSQL, Redis, Validator).
  - **infrastructure/**: Frameworks and drivers (e.g., Config, Logging, Metrics).
  - **middleware/**: HTTP middlewares (Auth, Logging, etc.).
  - **route/**: API route definitions.
- **pkg/**: Library code that's ok to use by external applications (if any).
  - `config/`, `response/`, `validator/`, etc.
- **db/**: Database related files.
  - `migrations/`: SQL migration files.
  - `seeds/`: Data seeding logic.

## Clean Architecture Layers

1.  **Handler** (`internal/module/*/handler`):
    - Handles HTTP requests.
    - Validates input.
    - Calls Service layer.
    - Returns HTTP response.
2.  **Service** (`internal/module/*/service`):
    - Contains business logic.
    - Calls Repository layer.
    - Defines transaction boundaries.
3.  **Repository** (`internal/module/*/repository`):
    - Handles database interactions.
    - Uses `sqlx` for SQL queries.
    - Maps database rows to Entities.
4.  **Entity** (`internal/entity`):
    - Plain Go structs representing domain objects.
    - May contain basic validation or domain logic.
