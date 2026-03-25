(() => {
  const nameKey = 'hearts.player.name';
  const tokenKey = 'hearts.player.token';

  const nameEl = document.getElementById('name');
  const newTableIdEl = document.getElementById('newTableId');
  const createResultEl = document.getElementById('createResult');
  const tablesEl = document.getElementById('tables');
  const presenceSection = document.getElementById('presenceSection');
  const presenceText = document.getElementById('presenceText');

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

  async function fetchTables() {
    let data;
    try {
      const res = await fetch('/api/tables');
      data = await res.json();
    } catch (err) {
      console.warn('fetchTables failed, showing stale data:', err);
      return;
    }
    const tables = data.tables || [];

    if (tables.length === 0) {
      tablesEl.innerHTML = '<li class="muted">No open tables yet.</li>';
      return;
    }

    tablesEl.innerHTML = '';
    for (const table of tables) {
      const li = document.createElement('li');

      const info = document.createElement('div');
      info.className = 'table-info';

      const name = document.createElement('span');
      name.className = 'table-name';
      name.textContent = table.table_id;

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
      players.textContent = `${table.players}/${table.max_players} players`;

      meta.appendChild(badge);
      meta.appendChild(players);
      info.appendChild(name);
      info.appendChild(meta);

      const btn = document.createElement('button');
      btn.textContent = 'Join';
      btn.onclick = () => openTable(table.table_id);

      li.appendChild(info);
      li.appendChild(btn);
      tablesEl.appendChild(li);
    }
  }

  async function createTable() {
    const payload = { table_id: newTableIdEl.value.trim() };
    const maxRetries = 2;
    let lastErr;
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        const res = await fetch('/api/tables', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        const data = await res.json();
        if (data.error) {
          createResultEl.textContent = data.error;
          return;
        }
        createResultEl.textContent = data.created ? `created ${data.table_id}` : `using existing ${data.table_id}`;
        openTable(data.table_id);
        return;
      } catch (err) {
        lastErr = err;
        console.warn(`createTable attempt ${attempt + 1} failed:`, err);
        if (attempt < maxRetries) {
          await new Promise((r) => setTimeout(r, 500 * Math.pow(2, attempt)));
        }
      }
    }
    console.error('createTable failed after retries:', lastErr);
    createResultEl.textContent = 'Network error, please try again.';
  }

  // --- Lobby presence via WebSocket ---

  const MAX_VISIBLE_NAMES = 8;
  let lobbyWs = null;
  let selfId = null;
  let lastPlayers = [];

  function renderPresence(players) {
    // Don't render until we know our own ID so we can filter ourselves out.
    if (selfId == null) {
      presenceSection.style.display = 'none';
      return;
    }

    const others = players.filter(p => p.id !== selfId);

    if (others.length === 0) {
      presenceSection.style.display = 'none';
      return;
    }

    presenceSection.style.display = '';
    const names = others.map(p => p.name);

    let text;
    if (names.length <= MAX_VISIBLE_NAMES) {
      text = names.join(', ') + (names.length === 1 ? ' is' : ' are') + ' also in the lobby';
    } else {
      const visible = names.slice(0, MAX_VISIBLE_NAMES);
      const remaining = names.length - MAX_VISIBLE_NAMES;
      text = visible.join(', ') + ' and ' + remaining + (remaining === 1 ? ' other' : ' others') + ' also in the lobby';
    }
    presenceText.textContent = text;
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
      }
    };

    lobbyWs.onclose = () => {
      lobbyWs = null;
      selfId = null;
      presenceSection.style.display = 'none';
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

  document.getElementById('createTable').onclick = createTable;

  nameEl.value = localStorage.getItem(nameKey) || 'Player';
  ensureToken();
  fetchTables();
  setInterval(fetchTables, 1500);
  connectLobbyWs();
})();
