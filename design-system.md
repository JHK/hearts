# Design System

Visual language for the Hearts card game. All future design work should reference this document for consistency.

## Design Philosophy

**Casino-inspired, mobile-first.** Evokes the feel of a real card table — rich felt greens, clean typography, understated chrome — without the visual noise of actual casino software.

**Consistency first.** Shared foundations (background, borders, typography, spacing) are the same across all pages.

Two distinct moods:

- **In-game (table)**: Focused, lean, zen-like. The felt board dominates. Controls are minimal and tucked away. Nothing competes with the cards — players should feel the calm concentration of a real card game.
- **Out-of-game (lobby, pause, game over)**: Playful and engaging. Card-flip animations, clear calls to action. These are social/waiting moments — the UI should feel warm and encourage interaction.

**Bots are placeholders for humans.** The UI never treats bots as a distinct class of player. Seat slots, names, and interactions are designed for human multiplayer; bots fill empty chairs.

**Mobile-friendly.** Touch targets are minimum 34px. Content stacks vertically on small screens and spreads into multi-column layouts on desktop.

## Color Palette

Six semantic roles via CSS custom properties. Values are the same on all pages.

| Role | Variable | Value | Purpose |
|---|---|---|---|
| Background | `--bg` | `#eae7df` | Page canvas |
| Surface | `--panel` | `#ffffff` | Cards, panels, overlays |
| Ink | `--ink` | `#1c2b3a` | Primary text, icon buttons |
| Muted | `--muted` | `#5b6f83` | Secondary text, metadata, labels |
| Border | `--line` | `#d5d1c8` | Dividers, input borders, inactive toggles |
| Accent | `--accent` / `--green` | `#116466` / `#1f6f5f` | Buttons, card surfaces, active states |
| Gold | — | `#8b6914` / `#b8860b` / `#9a7209` | Felt buttons only (gradient triple) |

All text and icon colors use `--ink` or `--muted`.

### Felt

Radial gradient from `#2a7d6a` (center) through `#1f5f51` to `#17493f` (edge), evoking a casino card table. The felt is the visual anchor of the in-game experience.

### Card Backs

Three-layer diagonal stripe pattern in deep blues (`#2b5a86`, `#22496d`, `#173653`). Card faces are white with a light border.

### Interactive Highlights

Translucent overlays on existing surfaces:
- **Turn**: warm yellow tint
- **Selection**: cool blue glow
- **Winner**: golden pulse

### Status Badges

Semantic traffic-light colors: green (waiting), amber (active), red (paused). Tinted backgrounds with matching dark text.

### Chart Player Colors

Cyclic: teal (`#116466`), burnt orange (`#b44f26`), slate (`#5b6f83`), purple (`#8b5cf6`). Scoped to the game-over score chart.

## Typography

**`"Noto Sans", "Segoe UI", sans-serif`** everywhere. Single font family.

Three size tiers: small (`0.78-0.88rem`) for metadata and body, medium (`0.9-1rem`) for section headings and scores, large (`1.28-1.6rem`) for page headings and overlay titles.

Three weights: `400` (body), `600` (labels, badges), `700` (buttons, headings).

Text on the felt gets a subtle shadow for legibility.

## Spacing

Base scale: `4-6px` (tight) / `8-10px` (standard) / `12-14px` (sections) / `20-28px` (panels) / `36px` (generous). Standard flex gap is `10px`. Page padding is `16-20px`.

## Border Radius

Five tiers: `4-5px` (inputs, small cards) / `8-10px` (buttons, panels) / `12px` (board elements) / `14-16px` (sections, board) / `999px` (circles).

## Component Patterns

### Buttons

All buttons share: white text, `border-radius: 10px`, `font-weight: 700`.

**Standard** — Teal `130deg` linear gradient (`--accent` → `#1b8d8f`). Used everywhere except on the felt.

**Felt (gold)** — Three-stop `130deg` gradient (`#8b6914` → `#b8860b` → `#9a7209`), white text. Inset top highlight, inset bottom shadow, outer drop shadow for an embossed feel.

**Disabled** — Gray gradient with reduced text opacity. No shadow.

**Icon** — `34px` circles with `--line` border and translucent white background. Hover lifts with box-shadow. Focus ring: `2px solid rgba(26, 76, 104, 0.55)`.

**Back (navigation)** — `28px` borderless ghost circle, `--muted` color. Hover tints background and darkens to `--ink`.

### Page Header

Shared component, identical on every page. Three-column grid: left (optional back link) · center (logo + "Hearts") · right (icon button actions, always includes settings).

Title is always centered. Headline scales to `1.1rem` and logo to `22px` below `480px`.

Page-specific elements belong in the page content area, not the header.

### Surfaces

Panels and cards use `--panel` background, `1px solid --line` border, and a subtle box-shadow. Lobby card surfaces have a 3D flip interaction. In-game sections are more subdued.

### Overlays

Full-screen scrim with dark semi-transparent background and `backdrop-filter: blur(3px)`. Centered surface card with generous padding. Used for game-over only; game-paused is an inline control in the trick center.

### Scoreboard

Minimal table. Player name column uses `--muted`, data uses `--ink`. Current-round row gets a faint green tint, total row a faint blue tint.

## Play Cards

Hand cards for the local player are slightly larger than opponents'. Cards overlap to form a fan. Hover and selection states use vertical lift and glow. Back cards are small and understated — they indicate hidden information without drawing focus.

Trick center cards animate in from each seat direction with scale, rotation, and opacity. The trick center is the focal point during play. Control buttons (Start, Continue, Pass, Game Paused) appear in a centered overlay within the trick center, backed by a solid felt-colored panel with rounded corners that fully covers trick slots behind it. Text on the trick center uses light mint (`#e7fff8`) with a subtle shadow.

## Animation Timing

CSS custom properties with a fast-mode variant at 50% duration:

| Animation | Standard | Fast |
|---|---|---|
| `--anim-card-in` | `520ms` | `260ms` |
| `--anim-trick-capture` | `1400ms` | `700ms` |
| `--anim-winner-pulse` | `1200ms` | `600ms` |
| `--anim-scoreboard-flip` | `400ms` | `200ms` |

`prefers-reduced-motion` disables all animations.
