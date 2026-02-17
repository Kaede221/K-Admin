---
inclusion: always
---

# Project Development Rules

## Build and Compilation

- DO NOT create executable output files (e.g., `go build -o k-admin.exe .`)
- DO compile without output to verify correctness (e.g., `go build` without `-o` flag)
- DO NOT create temporary files or test artifacts during development

## Testing Policy

- DO NOT create or generate unit tests for frontend or backend code
- This project does not require automated test coverage

## File Management

- Keep the workspace clean by avoiding unnecessary file generation
- Focus on source code implementation rather than test infrastructure

## Critical

These rules are mandatory and must be followed without exception.