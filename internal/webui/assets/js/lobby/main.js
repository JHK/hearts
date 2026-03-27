(() => {
  const nameKey = 'hearts.player.name';
  const tokenKey = 'hearts.player.token';
  const speedKey = 'hearts.animation.speed';
  const soundKey = 'hearts.sound.enabled';
  const notifyKey = 'hearts.notifications.enabled';

  const nameEl = document.getElementById('nameInput');
  const tablesEl = document.getElementById('tables');
  const tablesAreaEl = document.getElementById('tablesArea');
  const settingsToggleEl = document.getElementById('settingsToggle');
  const settingsPanelEl = document.getElementById('settingsPanel');
  const speedToggleEl = document.getElementById('speedToggle');
  const soundToggleEl = document.getElementById('soundToggle');
  const notifyToggleEl = document.getElementById('notifyToggle');

  function ensureToken() {
    let token = localStorage.getItem(tokenKey);
    if (!token) {
      token = (self.crypto && self.crypto.randomUUID) ? self.crypto.randomUUID() : String(Date.now()) + Math.random().toString(16).slice(2);
      localStorage.setItem(tokenKey, token);
    }
    return token;
  }

  function getName() {
    const name = nameEl.value.trim();
    return name || 'Player';
  }

  function storeIdentity() {
    localStorage.setItem(nameKey, getName());
    ensureToken();
  }

  function openTable(tableId) {
    storeIdentity();
    window.location.href = '/table/' + encodeURIComponent(tableId);
  }

  // --- Settings popover ---

  settingsToggleEl.onclick = () => {
    settingsPanelEl.classList.toggle('hidden');
  };

  document.addEventListener('pointerdown', (e) => {
    if (!settingsPanelEl.classList.contains('hidden') &&
        !settingsPanelEl.contains(e.target) &&
        !settingsToggleEl.contains(e.target)) {
      settingsPanelEl.classList.add('hidden');
    }
  });

  // --- Settings controls ---

  speedToggleEl.checked = localStorage.getItem(speedKey) === 'fast';
  speedToggleEl.onchange = () => {
    localStorage.setItem(speedKey, speedToggleEl.checked ? 'fast' : 'normal');
  };

  const soundEnabled = localStorage.getItem(soundKey) !== 'false';
  soundToggleEl.checked = soundEnabled;
  soundToggleEl.onchange = () => {
    localStorage.setItem(soundKey, soundToggleEl.checked ? 'true' : 'false');
  };

  notifyToggleEl.checked = localStorage.getItem(notifyKey) === 'true';
  notifyToggleEl.onchange = async () => {
    if (notifyToggleEl.checked) {
      if ('Notification' in window && Notification.permission !== 'granted') {
        const result = await Notification.requestPermission();
        if (result !== 'granted') {
          notifyToggleEl.checked = false;
          return;
        }
      }
    }
    localStorage.setItem(notifyKey, notifyToggleEl.checked ? 'true' : 'false');
  };

  // --- Table cards ---

  const flippedTables = new Set();

  function activateCard(li, handler) {
    li.tabIndex = 0;
    li.setAttribute('role', 'button');
    li.addEventListener('click', handler);
    li.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        handler(e);
      }
    });
  }

  function buildPlusCard() {
    const li = document.createElement('li');
    li.setAttribute('aria-label', 'Create new table');
    const card = document.createElement('div');
    card.className = 'table-card create-card';

    const front = document.createElement('div');
    front.className = 'table-card-front';

    const icon = document.createElement('span');
    icon.className = 'create-icon';
    icon.textContent = '+';

    front.appendChild(icon);
    card.appendChild(front);
    li.appendChild(card);

    activateCard(li, () => createTable());

    return li;
  }

  function renderTables(tables) {
    // Prune flipped state for tables that no longer exist
    const currentIds = new Set(tables.map(t => t.table_id));
    for (const id of flippedTables) {
      if (!currentIds.has(id)) flippedTables.delete(id);
    }

    tablesEl.innerHTML = '';

    for (const table of tables) {
      const li = document.createElement('li');
      const isFlipped = flippedTables.has(table.table_id);

      const card = document.createElement('div');
      card.className = 'table-card' + (isFlipped ? ' flipped' : '');

      li.setAttribute('aria-label', `Table ${table.table_id}`);
      li.setAttribute('aria-expanded', String(isFlipped));

      // --- Front face ---
      const front = document.createElement('div');
      front.className = 'table-card-front';

      const logo = document.createElement('img');
      logo.src = '/icon.svg';
      logo.alt = '';
      logo.className = 'table-logo';

      const meta = document.createElement('span');
      meta.className = 'table-meta';

      const badge = document.createElement('span');
      if (table.paused) {
        badge.className = 'badge badge-paused';
        badge.textContent = 'Paused';
      } else if (table.started) {
        badge.className = 'badge badge-active';
        badge.textContent = 'In progress';
      } else {
        badge.className = 'badge badge-waiting';
        badge.textContent = 'Waiting';
      }

      const players = document.createElement('span');
      players.className = 'table-players';
      players.textContent = `${table.players}/${table.max_players}`;

      meta.appendChild(badge);
      meta.appendChild(players);

      front.appendChild(logo);
      front.appendChild(meta);

      // --- Back face ---
      const back = document.createElement('div');
      back.className = 'table-card-back';

      const backName = document.createElement('span');
      backName.className = 'table-name';
      backName.textContent = table.table_id;

      const btn = document.createElement('button');
      btn.textContent = 'Join Table';
      btn.onclick = (e) => {
        e.stopPropagation();
        openTable(table.table_id);
      };

      back.appendChild(backName);
      back.appendChild(btn);

      card.appendChild(front);
      card.appendChild(back);
      li.appendChild(card);

      activateCard(li, () => {
        const nowFlipped = card.classList.toggle('flipped');
        li.setAttribute('aria-expanded', String(nowFlipped));
        if (nowFlipped) {
          flippedTables.add(table.table_id);
        } else {
          flippedTables.delete(table.table_id);
        }
      });

      tablesEl.appendChild(li);
    }
    tablesEl.appendChild(buildPlusCard());
  }

  function createTable() {
    if (!lobbyWs || lobbyWs.readyState !== WebSocket.OPEN) return;
    lobbyWs.send(JSON.stringify({ type: 'create_table', table_id: '' }));
  }

  // --- Lobby WebSocket (floating names) ---

  let lobbyWs = null;
  let selfId = null;
  let lastPlayers = [];
  const floatingNames = new Map();
  let floatingAnimId = 0;

  function renderPresence(players) {
    if (selfId == null) return;
    const others = players.filter(p => p.id !== selfId);
    const othersById = new Map(others.map(p => [p.id, p]));

    // Remove departed players
    for (const [id, entry] of floatingNames) {
      if (!othersById.has(id)) {
        entry.el.remove();
        floatingNames.delete(id);
      }
    }

    const hadNames = floatingNames.size > 0;

    // Add new or update existing
    for (const p of others) {
      if (!floatingNames.has(p.id)) {
        const el = document.createElement('span');
        el.className = 'floating-name';
        el.textContent = p.name;
        tablesAreaEl.appendChild(el);

        const rect = tablesAreaEl.getBoundingClientRect();
        const w = Math.max(rect.width - 80, 1);
        const h = Math.max(rect.height - 24, 1);
        floatingNames.set(p.id, {
          el,
          x: Math.random() * w,
          y: Math.random() * h,
          dx: (Math.random() - 0.5) * 0.3,
          dy: (Math.random() - 0.5) * 0.3,
        });
      } else {
        floatingNames.get(p.id).el.textContent = p.name;
      }
    }

    // Start animation loop when names appear; it self-stops when empty
    if (!hadNames && floatingNames.size > 0) {
      floatingAnimId = requestAnimationFrame(animateFloatingNames);
    }
  }

  let cachedAreaRect = null;
  window.addEventListener('resize', () => { cachedAreaRect = null; });

  function animateFloatingNames() {
    if (floatingNames.size === 0) {
      floatingAnimId = 0;
      return;
    }

    if (!cachedAreaRect) cachedAreaRect = tablesAreaEl.getBoundingClientRect();
    const w = Math.max(cachedAreaRect.width - 80, 1);
    const h = Math.max(cachedAreaRect.height - 24, 1);

    for (const entry of floatingNames.values()) {
      entry.x += entry.dx;
      entry.y += entry.dy;

      if (entry.x < 0 || entry.x > w) entry.dx *= -1;
      if (entry.y < 0 || entry.y > h) entry.dy *= -1;
      entry.x = Math.max(0, Math.min(w, entry.x));
      entry.y = Math.max(0, Math.min(h, entry.y));

      entry.el.style.transform = `translate(${entry.x}px, ${entry.y}px)`;
    }
    floatingAnimId = requestAnimationFrame(animateFloatingNames);
  }

  function connectLobbyWs() {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
    lobbyWs = new WebSocket(`${protocol}://${location.host}/ws/lobby`);

    lobbyWs.onopen = () => {
      selfId = null;
      announceSelf();
    };

    lobbyWs.onmessage = (event) => {
      let msg;
      try { msg = JSON.parse(event.data); } catch { return; }
      if (msg.type === 'lobby_presence' && msg.data && Array.isArray(msg.data.players)) {
        lastPlayers = msg.data.players;
        renderPresence(lastPlayers);
      } else if (msg.type === 'lobby_self' && msg.data) {
        selfId = msg.data.id;
        renderPresence(lastPlayers);
      } else if (msg.type === 'lobby_tables' && msg.data && Array.isArray(msg.data.tables)) {
        renderTables(msg.data.tables);
      } else if (msg.type === 'create_table_result' && msg.data) {
        if (msg.data.table_id) {
          openTable(msg.data.table_id);
        }
      } else if (msg.type === 'error' && msg.error) {
        console.warn('lobby error:', msg.error);
      }
    };

    lobbyWs.onclose = () => {
      lobbyWs = null;
      selfId = null;
      // Remove all floating names on disconnect
      for (const entry of floatingNames.values()) entry.el.remove();
      floatingNames.clear();
      if (floatingAnimId) { cancelAnimationFrame(floatingAnimId); floatingAnimId = 0; }
      setTimeout(connectLobbyWs, 2000);
    };
  }

  function announceSelf() {
    if (!lobbyWs || lobbyWs.readyState !== WebSocket.OPEN) return;
    const token = ensureToken();
    const name = getName();
    lobbyWs.send(JSON.stringify({ type: 'announce', name, token }));
  }

  // Re-announce when name changes so the presence list stays current.
  let announceTimer = null;
  nameEl.addEventListener('input', () => {
    clearTimeout(announceTimer);
    announceTimer = setTimeout(announceSelf, 300);
  });

  nameEl.value = localStorage.getItem(nameKey) || 'Player';
  ensureToken();
  connectLobbyWs();
})();
