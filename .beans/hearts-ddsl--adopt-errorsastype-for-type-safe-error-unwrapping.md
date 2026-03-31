---
# hearts-ddsl
title: Adopt errors.AsType for type-safe error unwrapping
status: scrapped
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-31T15:37:20Z
updated_at: 2026-03-31T15:48:19Z
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


## Summary of Changes

No `errors.As` call sites exist in the codebase. The project uses `errors.Is()` and `fmt.Errorf` with `%w` for error handling, but never unwraps to a concrete error type via `errors.As`. There is nothing to convert to `errors.AsType[T]()`.


## Reasons for Scrapping

The codebase has zero `errors.As` call sites and no custom error types. All errors are sentinels compared with `errors.Is`, with human-readable context added via `fmt.Errorf` wrapping. No code path needs to extract structured fields from errors, so `errors.AsType[T]()` has no applicable use case here.
