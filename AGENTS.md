# User Preferences

- Be objective and truthful, even if it may be difficult to hear.
- When editing files, always use absolute paths.
- When making changes to a file, explain why the change is being made.
- When generating code, add comments in English.
- **Shell & Package Manager**:
    - Use `fish` (preferred) or `bash` for shell commands.
    - Use `go mod` for dependency management.
    - **Running the App**: Prefer `task dev` or `go run ./cmd/bin/main.go`.
- **Documentation Maintenance**:
    - The AI agent is authorized to update `AGENTS.md` and `docs/` files to keep them accurate.
    - **Protocol**:
        1.  **Inform**: When making documentation changes, explicitly mention them in the final response (e.g., "Updated `docs/tech_stack.md` to reflect the new library").
        2.  **Suggest**: If the AI detects that the codebase patterns (e.g., new folder structure, new library) deviate from the existing docs, it must **proactively suggest** updating the relevant documentation file.
- **Business Logic Verification (Usecase Layer)**:
    - **Think Before Coding**: When working on the `usecase` (service) layer, proactively identify potential edge cases, business rules, and validation scenarios.
    - **Confirm First**: Explicitly list these scenarios and ask the user for confirmation *before* implementing the logic.
- **Error Handling**:
    - **Use `pkg/errmsg`**: Leverage the centralized error handling package (`pkg/errmsg`) which standardizes validation (`validator.v10`), database (`lib/pq`), and custom business logic errors.
    - **Handler Layer Pattern**:
        - **Body Parsing**: If `c.BodyParser` fails, log the error and return `400 Bad Request`.
        - **Validation**:
            - Use `adapter.Adapters.Validator.Validate(req)`.
            - On failure, use `errmsg.Errors(err, req)` to extract validation messages.
            - Return `c.Status(code).JSON(response.Error(errs))`.
        - **Service Calls**:
            - When a service returns an error, process it using `errmsg.Errors[error](err)`.
            - This helper automatically detects if it's a `pq.Error` (DB), `CustomError` (Logic), or unknown error.
            - Return `c.Status(code).JSON(response.Error(errs))`.
    - **Custom Errors**: For business logic rules (e.g., "insufficient balance"), use `errmsg.NewCustomErrors(code, errmsg.WithMessage("..."))` instead of generic Go errors.
    - **Database Errors**: Let the global handler process `pq.Error` (e.g., unique violations, foreign key constraints) automatically; do not manually wrap them unless necessary for context.
    - **Clean Responses**: Ensure error messages are user-friendly (in Indonesian/Bahasa Indonesia as per existing patterns) and do not leak internal details.
- **Tracing**:
    - **Implement Tracing**: Implement tracing in all layers (Handler, Service, Repository) using `codebase-app/internal/infrastructure/tracing`.
    - **Span Naming**: Use `layer.MethodName` (e.g., `handler.CreateUser`, `service.CreateUser`, `repo.CreateUser`).
    - **Context Propagation**: Always pass `ctx` to the next layer.
    - **Error Logging**: Log errors in the span using `log.Ctx(ctx).Error().Err(err).Msg("...")`.

# Documentation Index

The technical details and guidelines for this project have been moved to the `docs/` directory. Please refer to them for specific instructions.

- [Tech Stack](docs/tech_stack.md)
- [Architecture](docs/architecture.md)
- [Coding Conventions](docs/coding_conventions.md)
- [Database Schema](docs/database_schema.md)
- [Development Workflow](docs/development_workflow.md)
- [Commit Guidelines](docs/commit_guidelines.md)
- [Database Seeding](docs/seeding.md)

## Quick Summary
- **Architecture**: Modular Clean Architecture (`internal/module/[module_name]`)
