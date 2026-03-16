---
# hearts-lcxq
title: PWA installability
status: todo
type: feature
priority: normal
created_at: 2026-03-16T07:35:48Z
updated_at: 2026-03-16T08:28:12Z
---

Add Progressive Web App support: installability via web manifest and app icons. Service worker is minimal — no offline caching, just enough to satisfy install requirements.

## Tasks

- [ ] Create `manifest.json` (name, short_name, display: standalone, start_url, theme_color)
- [ ] Generate app icons (192×192, 512×512 minimum; 180×180 for iOS)
- [ ] Link manifest in HTML `<head>`
- [ ] Register a minimal service worker (fetch passthrough only)

## Context
Hearts is a web app but feels disconnected from the desktop/mobile experience users expect from a game. There's no way to install it as a standalone app — users must always open a browser, navigate to the URL, and interact with browser chrome around the game.

## Higher Goal
Make Hearts feel like a native desktop or mobile game. Players on Windows, Android, and iOS should be able to install it to their home screen or Start menu and launch it in a standalone window — no browser chrome, no address bar — just the game.

## Acceptance Criteria
- [ ] The app can be installed on Windows (Chrome/Edge) via the browser's install prompt
- [ ] The app can be added to the home screen on Android
- [ ] The app can be added to the home screen on iOS (via Safari share sheet)
- [ ] When launched from the installed shortcut, the game opens in standalone mode (no browser UI)
- [ ] The app icon appears correctly at all required sizes on install

## Out of Scope
- Offline play or caching game assets for offline use
- Push notifications
- Background sync
- App store distribution (Google Play, Microsoft Store, App Store)
