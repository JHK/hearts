---
# hearts-e5b4
title: Web Caching
status: completed
type: epic
priority: normal
created_at: 2026-03-24T10:53:41Z
updated_at: 2026-03-25T09:39:12Z
---

## Vision

Utilize browser caching to make the game load faster, save bandwidth, and work better on mobile/flaky connections. All embedded assets should be served with appropriate cache strategies — immutable for static content, fingerprinted URLs for build-varying content, and conditional requests for HTML.

## Context

Currently no caching headers are set on any route. Browsers re-fetch every asset on every page load, including 56 card SVGs that never change between deploys.
