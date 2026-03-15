---
# hearts-lsoi
title: Sound effects for hearts breaking and Queen of Spades
status: todo
type: feature
created_at: 2026-03-15T16:47:44Z
updated_at: 2026-03-15T16:47:44Z
---

Play a short synthetic sound effect when the first hearts card is played (hearts breaking) and when the Queen of Spades is played. Use the Web Audio API to generate sounds synthetically — no audio files to ship.

## Tasks

- [ ] Research suitable Web Audio API synthesis approach (oscillator, envelope, etc.)
- [ ] Design and tune the sound for hearts breaking
- [ ] Design and tune the sound for Queen of Spades played
- [ ] Hook into the existing WebSocket event stream (EventCardPlayed) to trigger sounds client-side

## Out of scope

- Audio files / assets
- Mute toggle (separate bean if needed)
