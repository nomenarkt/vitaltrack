# 🧑‍💻 CONTRIBUTING GUIDELINES
Welcome to this project! This document outlines the standards for contributing code, particularly for Codex (our engineering automation collaborator). All contributions must comply with Clean Architecture, TDD, and production-grade Go standards.
---
## 🔰 Project Architecture
This project uses **Clean Architecture**, with the following layers:
```
/delivery     → HTTP handlers (routing, validation only)
/usecase      → Business logic (pure functions, use case orchestration)
/repository   → Database and external integrations (no logic)
```
- Handlers → call Usecases → which call Repositories.
- No logic should cross layers directly or leak abstractions.
- Services must define interfaces for dependency injection and testability.
---
## 🧠 Codex Development Rules
### ✅ When Implementing Any Task
- Follow tasks issued by The Architect exactly (e.g., `Implement the /refill endpoint`).
- Never invent features or alter scope.
- Stick to the specified models, routes, and flow.
### ✅ Git & Commit Standards
- One pull request = one logical task.
- Each commit must be:
  - Atomic (one purpose only)
  - Formatted using `go fmt`, `goimports`
  - Labeled with Conventional Commits:
    - `feat(api): add /refill endpoint`
    - `fix(repo): correct null check for user lookup`
---
## 🧪 Testing Requirements
### Unit Tests
- All business logic and handler changes must be covered by `_test.go` files.
- Use table-driven tests when applicable.
- Include success cases, edge cases, and validation errors.
### Test Placement
- Test files must live next to the code they test.
- Common mocks may live in `/internal/testutils` or `mocks/`.
### Manual Testing
- All endpoints must be verifiable via:
  - `curl` scripts
  - Postman collections (if applicable)
---
## 🧹 Code Formatting & Linting
- Always run:
  ```bash
  go fmt ./...
  goimports -w .
  staticcheck ./...
  ```
- Code must be idiomatic. No `snake_case`. No redundant getters.
- Respect Go Proverbs & Effective Go.
---
## 🪵 Logging
- Use structured logging at the following points:
  - Entry to each handler and usecase
  - On all errors, with stack/context
  - On branches like timeouts, retries, or nil results
Example:
```go
log.WithContext(ctx).WithFields(log.Fields{
    "userID": id,
    "action": "refill",
}).Info("processing refill request")
```
---
## 🚫 You MUST NOT
- Include TODOs or commented code in final PRs.
- Use global/shared mutable state.
- Commit without test coverage (except for markdown/doc-only changes).
- Talk directly to the database from handlers.
---
## ✅ You MUST
- Write tests before merging any logic-affecting code.
- Follow separation of concerns across layers.
- Include detailed error messages and validations.
- Use context propagation and timeouts in all external calls.
---
By contributing, you agree to maintain the standards of this codebase and prioritize software quality, testability, and maintainability.
Let’s build real, production-grade software together.