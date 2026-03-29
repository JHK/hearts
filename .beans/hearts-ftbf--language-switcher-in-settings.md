---
# hearts-ftbf
title: Language switcher in settings
status: completed
type: feature
priority: normal
created_at: 2026-03-29T12:32:21Z
updated_at: 2026-03-29T14:41:09Z
parent: hearts-xf2i
blocked_by:
    - hearts-q4bi
---

Add UI control to settings panel for switching between English and German

**Note:** Read the parent epic (`hearts-xf2i`) and preceding tickets for the full approach and implementation details.

## Context
After `hearts-q4bi`, both English and German locales are fully functional. Users can only switch language by manually editing LocalStorage. This ticket adds a proper UI control.

## Higher Goal
Users can choose their preferred language through the settings UI, with the choice persisted and applied immediately.

## Acceptance Criteria
- [x] Language selector added to the settings panel (dropdown or toggle for en/de)
- [x] Selecting a language writes the value to LocalStorage and reloads the page
- [x] The selector reflects the currently active locale on load
- [x] When no explicit choice has been made (LocalStorage empty), the selector shows the auto-detected language
- [x] Visual design is consistent with the existing settings panel
- [x] Documentation updated (design-system.md if a new component pattern, CLAUDE.md if relevant)

## Out of Scope
- Persisting preference to a database/account system
- Adding more locales
- Changing the detection/fallback logic established in earlier tickets

## Summary of Changes

Added a language switcher dropdown to the settings panel. The `<select>` element is populated dynamically from `window.__i18n_all` keys, displays native locale names (English/Deutsch), reflects the current locale on load, and triggers a page reload on change after writing to `hearts.locale` in LocalStorage. Added `settings.language` i18n key to both en.json and de.json. Styled with `.settings-select` class matching existing panel aesthetics.
