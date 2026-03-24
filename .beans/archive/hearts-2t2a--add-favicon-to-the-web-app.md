---
# hearts-2t2a
title: Add favicon to the web app
status: completed
type: task
priority: normal
created_at: 2026-03-16T08:04:47Z
updated_at: 2026-03-16T13:28:10Z
---

## Context
The Hearts web app has no favicon. Browsers show a blank icon in tabs and bookmarks.

## Higher Goal
Small polish details make the app feel complete and make the tab identifiable at a glance.

## Acceptance Criteria
- [x] `favicon.ico` (32×32) is served and appears in browser tabs
- [x] `icon.svg` (square) is served for modern browsers and scales crisp at all resolutions
- [x] `apple-touch-icon.png` (180×180, with background color + padding) is served for iOS home screen and Android fallback
- [x] The icon design resembles the classic Windows 98 Hearts app icon: a black spade with eyes and three hearts (red, dark red, purple)
- [x] All three files are embedded via `//go:embed` in the webui package
- [x] Correct `<link>` tags are present in the HTML `<head>`

## Out of Scope
- PWA `manifest.json` icons (192×192, 512×512)
- Safari pinned tab SVG mask icon
- Windows tile meta tags

## References
- [Microsoft Hearts logo history — Logos Fandom](https://logos.fandom.com/wiki/Microsoft_Hearts#1992%E2%80%932001): the 1992–2001 icon shows a black spade with eyes and three hearts in red, dark red, and purple
- [How to Favicon in 2026 — Evil Martians](https://evilmartians.com/chronicles/how-to-favicon-in-2021-six-files-that-fit-most-needs): canonical minimal favicon guide; recommends 3 files: `favicon.ico` (32×32), `icon.svg` (vector, supports dark mode via CSS media query), `apple-touch-icon.png` (180×180 — used by iOS and Android Chrome fallback)

## Summary of Changes

- Added favicon.ico (32×32), icon.svg, and apple-touch-icon.png (180×180) to the web app
- Designed 2×2 grid icon: black spade with sinister eyes (top-left), red heart (top-right), dark-red heart (bottom-left), purple heart (bottom-right)
- Suit shapes taken from the vector-playing-cards deck used in the app
- Replaced hand-coded rasterizer in genfavicons with SVG-based rendering via oksvg+rasterx
- Iterated on spade eye design for legibility at favicon scale: bold eyebrows, large sclera/pupils
