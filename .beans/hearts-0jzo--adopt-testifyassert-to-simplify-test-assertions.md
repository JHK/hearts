---
# hearts-0jzo
title: Adopt testify/assert to simplify test assertions
status: todo
type: task
priority: normal
created_at: 2026-03-24T17:39:32Z
updated_at: 2026-03-24T17:39:45Z
---

Replace manual if/t.Fatalf assertion patterns with testify assert/require calls across all 10 test files

## Context

All 10 test files (≈1,026 lines) use manual `if cond { t.Fatalf(...) }` for every assertion — roughly 208 instances. This is verbose and obscures the intent of each check. The codebase has no assertion library today.

## Higher Goal

Reduce test boilerplate so tests are easier to read, write, and maintain. Clearer assertions also produce better failure messages by default.

## Acceptance Criteria

- [ ] `github.com/stretchr/testify` added as a test dependency
- [ ] All `if … { t.Fatalf / t.Fatal }` patterns in test files replaced with appropriate `assert.*` or `require.*` calls
- [ ] No test behaviour changes — `mise run test` passes identically
- [ ] Use `require.*` (stops on failure) where the original used `t.Fatalf`/`t.Fatal`; use `assert.*` where `t.Errorf`/`t.Error` was used (currently none, but as a guideline going forward)

## Out of Scope

- Rewriting test logic or adding new test cases
- Introducing testify/suite or testify/mock
- Refactoring non-assertion parts of tests (helpers, setup, table structure)

## References

- [testify/assert](https://pkg.go.dev/github.com/stretchr/testify/assert): standard Go assertion library, provides `Equal`, `NoError`, `True`, `Contains`, etc.
