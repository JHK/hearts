---
# hearts-ftbf
title: Language switcher in settings
status: todo
type: feature
priority: normal
created_at: 2026-03-29T12:32:21Z
updated_at: 2026-03-29T12:32:37Z
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
- [ ] Language selector added to the settings panel (dropdown or toggle for en/de)
- [ ] Selecting a language writes the value to LocalStorage and reloads the page
- [ ] The selector reflects the currently active locale on load
- [ ] When no explicit choice has been made (LocalStorage empty), the selector shows the auto-detected language
- [ ] Visual design is consistent with the existing settings panel
- [ ] Documentation updated (design-system.md if a new component pattern, CLAUDE.md if relevant)

## Out of Scope
- Persisting preference to a database/account system
- Adding more locales
- Changing the detection/fallback logic established in earlier tickets
