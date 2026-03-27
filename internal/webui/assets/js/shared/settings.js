export const nameKey = 'hearts.player.name';
export const tokenKey = 'hearts.player.token';
export const speedKey = 'hearts.animation.speed';
export const soundKey = 'hearts.sound.enabled';
export const notifyKey = 'hearts.notifications.enabled';

export function ensureToken() {
  let token = localStorage.getItem(tokenKey);
  if (!token) {
    token = (self.crypto && self.crypto.randomUUID)
      ? self.crypto.randomUUID()
      : String(Date.now()) + Math.random().toString(16).slice(2);
    localStorage.setItem(tokenKey, token);
  }
  return token;
}

export function initSettingsPopover(toggleEl, panelEl) {
  toggleEl.onclick = () => {
    panelEl.classList.toggle('hidden');
  };

  document.addEventListener('pointerdown', (e) => {
    if (!panelEl.classList.contains('hidden') &&
        !panelEl.contains(e.target) &&
        !toggleEl.contains(e.target)) {
      panelEl.classList.add('hidden');
    }
  });
}
