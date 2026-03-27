# Design System

Visual language for the Hearts card game. All future design work should reference this document for consistency.

## Design Philosophy

**Casino-inspired, mobile-first.** The game draws from the feel of a real card table -- rich felt greens, clean typography, understated chrome -- without the visual noise of actual casino software. Every element earns its place; if it doesn't help the player, it goes.

**Consistency first.** Shared foundations (background, borders, typography, spacing) are the same across all pages. Individual pages express mood through content and interaction, not by diverging from the base palette.

Two distinct moods govern the UI:

- **In-game (table)**: Focused, lean, zen-like. The felt board dominates. Controls are minimal and tucked away. Nothing competes with the cards. Players should feel the calm concentration of a real card game.
- **Out-of-game (lobby, pause, game over)**: Playful and engaging. Card-flip animations, clear calls to action. These are social/waiting moments -- the UI should feel warm and encourage interaction.

**Bots are placeholders for humans.** The UI never treats bots as a distinct class of player. Seat slots, names, and interactions are designed for human multiplayer first; bots simply fill empty chairs until a human arrives.

**Mobile-friendly.** All layouts are designed for phones first, then enhanced for larger screens. Touch targets are minimum 34px. Content stacks vertically on small screens and spreads into multi-column layouts on desktop.

## Color Palette

Six semantic roles, set via CSS custom properties on `.lobby-page` and `.table-page`. Values are the same on both pages.

| Role | Variable | Value | Purpose |
|---|---|---|---|
| Background | `--bg` | `#eae7df` | Page canvas (warm casino cream) |
| Surface | `--panel` | `#ffffff` | Cards, panels, overlays |
| Ink | `--ink` | `#1c2b3a` | Primary text, icon buttons |
| Muted | `--muted` | `#5b6f83` | Secondary text, metadata, labels |
| Border | `--line` | `#d5d1c8` | Dividers, input borders, inactive toggles |
| Accent | `--accent` / `--green` | `#116466` / `#1f6f5f` | Primary actions, active states |

All text and icon colors should use `--ink` or `--muted`. Avoid hardcoding one-off hex values for text.

### Felt

Radial gradient from `#2a7d6a` (center) through `#1f5f51` to `#17493f` (edge), evoking a casino card table. The felt is the visual anchor of the in-game experience.

### Card Backs

Three-layer diagonal stripe pattern in deep blues (`#2b5a86`, `#22496d`, `#173653`). Card faces are white with a light border.

### Interactive Highlights

Translucent overlays on existing surfaces -- not new colors:
- **Turn**: warm yellow tint
- **Selection**: cool blue glow
- **Winner**: golden pulse

### Status Badges

Semantic traffic-light colors: green (waiting), amber (active), red (paused). Tinted backgrounds with matching dark text.

### Chart Player Colors (game-over chart only)

Cyclic: teal (`#116466`), burnt orange (`#b44f26`), slate (`#5b6f83`), purple (`#8b5cf6`). Chosen for distinguishability at small sizes. These colors are scoped to the game-over score chart and should not appear elsewhere in the UI.

## Typography

**`"Noto Sans", "Segoe UI", sans-serif`** everywhere. Single font family, no overrides.

Three size tiers: small (`0.78-0.88rem`) for metadata and body, medium (`0.9-1rem`) for section headings and scores, large (`1.28-1.6rem`) for page headings and overlay titles.

Three weights: `400` (body), `600` (labels, badges), `700` (buttons, headings).

Text on the felt gets a subtle shadow for legibility against the gradient.

## Spacing

Base scale: `4-6px` (tight) / `8-10px` (standard) / `12-14px` (sections) / `20-28px` (panels) / `36px` (generous). Standard flex gap is `10px`. Page padding is `16-20px`.

## Border Radius

Five tiers: `4-5px` (inputs, small cards) / `8-10px` (buttons, panels) / `12px` (board elements) / `14-16px` (sections, board) / `999px` (circles).

## Component Patterns

### Buttons

All buttons use `130deg` linear gradients, white text, `border-radius: 10px`, `font-weight: 700`. Primary uses the accent teal, secondary uses green, disabled uses gray with reduced text opacity. Icon buttons are `34px` circles with `var(--line)` border and translucent white background.

Shared traits across icon buttons (back, settings, etc.): same size, border, background, hover shadow, focus ring. These should be a single visual pattern.

### Surfaces

Panels and cards use `var(--panel)` background, `1px solid var(--line)` border, and a subtle box-shadow. The lobby card surfaces have a 3D flip interaction (part of the playful lobby mood). In-game sections are more subdued.

### Overlays

Full-screen scrim with dark semi-transparent background and `backdrop-filter: blur(3px)`. The overlay panel is a centered surface card with generous padding. Game-over and game-paused overlays share the same visual treatment -- they should use the same base pattern.

### Scoreboard

Minimal table. Player name column uses `--muted`, data uses inherited `--ink`. Current-round row gets a faint green tint, total row a faint blue tint. The scoreboard should recede during play and only draw attention between rounds.

## Play Cards

Hand cards for the bottom (local) player are slightly larger than opponents'. Cards overlap to form a fan. Hover and selection states use vertical lift and glow effects. Back cards are small and understated -- they indicate hidden information without drawing focus.

Trick center cards animate in from each seat direction with scale, rotation, and opacity. The trick center is the focal point during play.

## Animation Timing

CSS custom properties with a fast-mode variant at 50% duration:

| Animation | Standard | Fast |
|---|---|---|
| `--anim-card-in` | `520ms` | `260ms` |
| `--anim-trick-capture` | `1400ms` | `700ms` |
| `--anim-winner-pulse` | `1200ms` | `600ms` |
| `--anim-scoreboard-flip` | `400ms` | `200ms` |

Animations should feel responsive but never distracting. In-game animations support the zen-like focus; lobby animations (card flip) support the playful mood. `prefers-reduced-motion` disables all animations.
