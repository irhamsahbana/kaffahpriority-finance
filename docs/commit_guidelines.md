# AI Agent Commit Guidelines

This document outlines how the AI Agent should handle git commits. The primary goal is to maintain a clean, semantic history that accurately reflects the changes made.

## Process

1.  **Analyze Changes**:
    - Before generating a commit message, **always** check the actual file changes (using `git status`, `git diff`, or by reviewing your own recent edits).
    - Do not assume; verify exactly what was modified, added, or deleted.

2.  **Determine Scope**:
    - Identify which part of the codebase is affected (e.g., `auth`, `user`, `prisma`, `docs`).
    - If changes span multiple distinct areas, consider splitting them into separate commits if possible, or use a broader scope/description.

3.  **Format Message**:
    - Follow the **Conventional Commits** specification.
    - Structure: `<type>(<scope>): <description>`

## Conventional Commits Specification

### Types
- **feat**: A new feature.
- **fix**: A bug fix.
- **docs**: Documentation only changes.
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc).
- **refactor**: A code change that neither fixes a bug nor adds a feature.
- **perf**: A code change that improves performance.
- **test**: Adding missing tests or correcting existing tests.
- **chore**: Changes to the build process or auxiliary tools and libraries (e.g., documentation generation, dependencies).

### Scope
- The scope provides additional context to the commit type.
- It should be a noun describing a section of the codebase.
- Examples: `auth`, `user`, `api`, `db`, `deps`.

### Description
- Use the imperative, present tense: "change" not "changed" nor "changes".
- Don't capitalize the first letter.
- No dot (.) at the end.

## Examples

- **Good**: `feat(auth): add jwt token generation`
- **Good**: `fix(user): resolve null pointer in profile update`
- **Good**: `docs(readme): update installation instructions`
- **Good**: `chore(deps): upgrade prisma to v7`

- **Bad**: `Fixed the login bug` (Missing type/scope, past tense)
- **Bad**: `Update code` (Too vague)
- **Bad**: `feat: Added new API` (Capitalized, past tense)

## Commit Command

You can use the Taskfile command to ensure linting passes before committing:

```bash
task commit msg="feat(auth): add login handler"
```

This command runs:
1.  `task lint-fix` (Format code)
2.  `task lint-ci` (Run linter)
3.  `git add .`
4.  `git commit -m "..."`

## Auto-Commit Behavior
- When the user asks to "commit", apply these rules automatically.
- Ensure the commit message is derived strictly from the **staged changes**.
