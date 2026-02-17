---
inclusion: always
---

# Project Development Rules

## ⛔ ABSOLUTELY FORBIDDEN - Build and Compilation

**CRITICAL: NO BUILD COMMANDS ALLOWED UNDER ANY CIRCUMSTANCES**

- ❌ **NEVER** run `go build` (even without `-o` flag)
- ❌ **NEVER** run `go build -o <filename>`
- ❌ **NEVER** run any compilation commands to "verify correctness"
- ❌ **NEVER** run `go run` to test the code
- ❌ **NEVER** create executable files (.exe, binary files, etc.)
- ❌ **NEVER** use build tools to check syntax or compilation errors

**WHY THIS RULE EXISTS:**
- Build commands waste time and resources
- Code correctness should be verified through code review, not compilation
- The user will handle building and running when needed

**WHAT TO DO INSTEAD:**
- ✅ Read and analyze code carefully
- ✅ Use static analysis tools if absolutely necessary (linters, formatters)
- ✅ Trust your code modifications are correct
- ✅ Let the user build and test when they're ready

## ⛔ ABSOLUTELY FORBIDDEN - Testing

**CRITICAL: NO TEST GENERATION OR EXECUTION**

- ❌ **NEVER** create unit tests
- ❌ **NEVER** create integration tests
- ❌ **NEVER** create test files (*_test.go, *.test.ts, *.spec.ts, etc.)
- ❌ **NEVER** run test commands (go test, npm test, pnpm test, etc.)
- ❌ **NEVER** generate test coverage reports

**WHY THIS RULE EXISTS:**
- This project does not require automated test coverage
- Tests create unnecessary files and clutter
- Focus is on implementation, not testing infrastructure

## File Management

- Keep the workspace clean by avoiding unnecessary file generation
- DO NOT create temporary files or artifacts during development
- DO NOT create documentation files unless explicitly requested
- Focus on source code implementation only

## Summary - What You CAN Do

✅ Modify source code files
✅ Read and analyze existing code
✅ Provide code suggestions and explanations
✅ Use code analysis tools (linters, formatters) if needed
✅ Answer questions about the code

## Summary - What You CANNOT Do

❌ Run ANY build commands (go build, npm build, etc.)
❌ Run ANY compilation commands
❌ Create or run ANY tests
❌ Create executable files
❌ Create test files
❌ Create temporary files

## Critical Warning

**IF YOU VIOLATE THESE RULES, YOU ARE WASTING THE USER'S TIME AND RESOURCES.**

These rules are **MANDATORY** and must be followed **WITHOUT EXCEPTION**.

**NO EXCUSES. NO "JUST TO VERIFY". NO "QUICK CHECK".**

**JUST DON'T DO IT.**