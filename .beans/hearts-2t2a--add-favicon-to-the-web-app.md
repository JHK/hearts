---
# hearts-2t2a
title: Add favicon to the web app
status: todo
type: task
priority: normal
created_at: 2026-03-16T08:04:47Z
updated_at: 2026-03-16T08:08:57Z
---

## Context
The Hearts web app has no favicon. Browsers show a blank icon in tabs and bookmarks.

## Higher Goal
Small polish details make the app feel complete and make the tab identifiable at a glance.

## Acceptance Criteria
- [ ] `favicon.ico` (32×32) is served and appears in browser tabs
- [ ] `icon.svg` (square) is served for modern browsers and scales crisp at all resolutions
- [ ] `apple-touch-icon.png` (180×180, with background color + padding) is served for iOS home screen and Android fallback
- [ ] The icon design resembles the classic Windows 98 Hearts app icon: a black spade with eyes and three hearts (red, dark red, purple)
- [ ] All three files are embedded via `//go:embed` in the webui package
- [ ] Correct `<link>` tags are present in the HTML `<head>`

## Out of Scope
- PWA `manifest.json` icons (192×192, 512×512)
- Safari pinned tab SVG mask icon
- Windows tile meta tags

## References
- [Microsoft Hearts logo history — Logos Fandom](https://logos.fandom.com/wiki/Microsoft_Hearts#1992%E2%80%932001): the 1992–2001 icon shows a black spade with eyes and three hearts in red, dark red, and purple
- [How to Favicon in 2026 — Evil Martians](https://evilmartians.com/chronicles/how-to-favicon-in-2021-six-files-that-fit-most-needs): canonical minimal favicon guide; recommends 3 files: `favicon.ico` (32×32), `icon.svg` (vector, supports dark mode via CSS media query), `apple-touch-icon.png` (180×180 — used by iOS and Android Chrome fallback)
