---
# hearts-q4bi
title: German translations
status: todo
type: task
priority: normal
created_at: 2026-03-29T12:28:36Z
updated_at: 2026-03-29T12:30:31Z
parent: hearts-xf2i
blocked_by:
    - hearts-v2e7
---

Create locales/de.json with complete German translations for all i18n keys

**Note:** Read the parent epic (`hearts-xf2i`) and the preceding ticket (`hearts-v2e7`) for the full approach, architectural decisions, and implementation details (key naming convention, interpolation/pluralization format) before starting.

## Context
After `hearts-v2e7`, all English strings go through `t()` and `locales/en.json` is complete. This ticket adds the German locale.

## Higher Goal
German-speaking users see the full UI in their language without any English leaking through.

## Acceptance Criteria
- [ ] `locales/de.json` exists with translations for every key in `en.json`
- [ ] German translations are concise — match the brevity of the English strings to avoid UI/UX issues from longer text
- [ ] Pluralization forms are correct for German (same `one`/`other` split works for both languages)
- [ ] Interpolated strings read naturally in German (word order may differ from English)
- [ ] Manual walkthrough of all pages with locale set to `de` — no missing keys, no layout breakage
- [ ] Documentation updated if anything noteworthy about the German locale emerged

## Out of Scope
- Language switcher UI (next ticket)
- Translations beyond en/de
- Changing layout or design to accommodate translations (flag as follow-up if needed)
