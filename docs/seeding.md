# Database Seeding

This project uses a custom Go-based seeding mechanism to populate the database with initial data.

## Seeding Structure

- **Location**: `db/seeds/` and `cmd/seed.go`.
- **Command**: `task seed`.

## How to Run Seeds

To seed the database, use the Taskfile command:

```bash
# Seed a specific table
task seed table=users

# Seed all tables (if supported by implementation)
task seed table=all
```

## Adding New Seeds

1.  Create or update seeding logic in `db/seeds/`.
2.  Ensure the seed function is registered in the `seed` command handler (likely in `cmd/seed.go` or `db/seeds/seeds.go`).
3.  The seeder should be idempotent if possible (check if data exists before inserting).

## Data Sources
- Some seeds may load data from Excel files located in `db/seeds/excel/`.
