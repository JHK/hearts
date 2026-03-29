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

export const localeKey = 'hearts.locale';

export function initLocaleSelect(selectEl) {
  const all = window.__i18n_all || {};
  const locales = Object.keys(all).sort();
  for (const code of locales) {
    const opt = document.createElement('option');
    opt.value = code;
    opt.textContent = (all[code] && all[code]['_meta.displayName']) || code;
    selectEl.appendChild(opt);
  }
  selectEl.value = window.__i18n_locale || 'en';
  selectEl.onchange = () => {
    localStorage.setItem(localeKey, selectEl.value);
    location.reload();
  };
}

export function initSettingsPopover(toggleEl, panelEl) {
  function setOpen(open) {
    panelEl.classList.toggle('hidden', !open);
    toggleEl.setAttribute('aria-expanded', String(open));
  }

  toggleEl.setAttribute('aria-expanded', 'false');

  toggleEl.onclick = () => {
    setOpen(panelEl.classList.contains('hidden'));
  };

  document.addEventListener('pointerdown', (e) => {
    if (!panelEl.classList.contains('hidden') &&
        !panelEl.contains(e.target) &&
        !toggleEl.contains(e.target)) {
      setOpen(false);
    }
  });

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !panelEl.classList.contains('hidden')) {
      setOpen(false);
      toggleEl.focus();
    }
  });
}
