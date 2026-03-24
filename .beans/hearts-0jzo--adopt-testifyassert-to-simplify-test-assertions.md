---
# hearts-0jzo
title: Adopt testify/assert to simplify test assertions
status: completed
type: task
priority: normal
created_at: 2026-03-24T17:39:32Z
updated_at: 2026-03-24T18:03:07Z
---

Replace manual if/t.Fatalf assertion patterns with testify assert/require calls across all 10 test files

## Context

All 10 test files (≈1,026 lines) use manual `if cond { t.Fatalf(...) }` for every assertion — roughly 208 instances. This is verbose and obscures the intent of each check. The codebase has no assertion library today.

## Higher Goal

Reduce test boilerplate so tests are easier to read, write, and maintain. Clearer assertions also produce better failure messages by default.

## Acceptance Criteria

- [x] `github.com/stretchr/testify` added as a test dependency
- [x] All `if … { t.Fatalf / t.Fatal }` patterns in test files replaced with appropriate `assert.*` or `require.*` calls
- [x] No test behaviour changes — `mise run test` passes identically
- [x] Use `require.*` (stops on failure) where the original used `t.Fatalf`/`t.Fatal`; use `assert.*` where `t.Errorf`/`t.Error` was used (currently none, but as a guideline going forward)

## Out of Scope

- Rewriting test logic or adding new test cases
- Introducing testify/suite or testify/mock
- Refactoring non-assertion parts of tests (helpers, setup, table structure)

## References

- [testify/assert](https://pkg.go.dev/github.com/stretchr/testify/assert): standard Go assertion library, provides `Equal`, `NoError`, `True`, `Contains`, etc.

## Summary of Changes

Replaced all manual `if cond { t.Fatalf(...) }` assertion patterns across all 10 test files with `require.*` calls from `github.com/stretchr/testify`. Used `require.NoError`, `require.Error`, `require.Equal`, `require.Len`, `require.True`, `require.False`, `require.NotEqual`, `require.NotEmpty`, `require.Less`, and `require.IsType` as appropriate. No test logic was changed.
