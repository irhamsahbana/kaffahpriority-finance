# Coding Conventions

## Naming
- **Packages**: Short, lowercase, singular (e.g., `user`, `auth`).
- **Files**: `snake_case` (e.g., `user_handler.go`).
- **Structs/Interfaces**: `PascalCase` (e.g., `UserHandler`, `UserRepository`).
- **Functions/Methods**: `PascalCase` for exported, `camelCase` for private.
- **Variables**: `camelCase`.
- **Constants**: `PascalCase` or `UPPER_CASE` for specific values.
- **Database Tables**: `snake_case` (plural).
- **Database Columns**: `snake_case`.

## Code Style
- **Formatting**: Always use `gofmt` (or `goimports`).
- **Linters**: Use `golangci-lint` as configured in the project.

## Error Handling
- **Check Errors**: Always check errors immediately.
  ```go
  if err != nil {
      return err
  }
  ```
- **Custom Errors**: Use `pkg/errmsg` for domain-specific errors.
- **Wrapping**: Wrap errors with context when appropriate, but ensure the root cause is preserved for logging.

## Logging & Tracing
- **Tracing**: Implement tracing in all layers using `codebase-app/internal/infrastructure/tracing`.
- **Span Naming**: Use `layer.MethodName` (e.g., `service.GetPayrollRuns`).
- **Context Propagation**: Always pass `ctx` to the next layer.
- **Logging**: Use `log.Ctx(ctx)` (zerolog) which integrates with tracing.
  - **Pattern**: `log.Ctx(ctx).Error().Err(err).Msg("message")`
  - **No Function Name**: Do NOT include the function name in the log message manually (e.g., avoid `Msgf("%s - error", fnName)`). The tracing context already captures the span/function info.

## Database & SQL
- **Library**: Use `sqlx`.
- **Placeholders**: Use `$` (Postgres) for parameters (e.g., `$1`, `$2`).
- **Context**: Always pass `context.Context` to database methods.
- **Transactions**: Use transactions for operations that modify multiple tables.

## API Response
- Use standard helpers from `pkg/response`.
- Return standardized JSON structure.

## Authentication
- Use `middleware/authorization.go` for protecting routes.
- Access user context from `ctx.Locals` (Fiber) or context keys.
