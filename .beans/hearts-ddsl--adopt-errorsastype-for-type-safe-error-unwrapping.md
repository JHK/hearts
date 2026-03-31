---
# hearts-ddsl
title: Adopt errors.AsType for type-safe error unwrapping
status: todo
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-31T15:37:20Z
updated_at: 2026-03-31T15:37:55Z
parent: hearts-u20m
---

## Context

Go 1.26 added `errors.AsType[T]()`, a generic type-safe replacement for `errors.As` that avoids declaring a target variable. The codebase uses `fmt.Errorf("...: %w", err)` for error wrapping throughout and may have `errors.As` call sites that could be simplified.

## Higher Goal

Adopt idiomatic Go 1.26 error handling patterns for cleaner, less error-prone error unwrapping.

## Acceptance Criteria

- [ ] All `errors.As` call sites identified and converted to `errors.AsType[T]()` where appropriate
- [ ] All tests pass

## Out of Scope

- Changing error wrapping patterns (`fmt.Errorf` with `%w` is fine)
- Introducing new sentinel errors or error types
