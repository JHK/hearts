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

  function renderTables(tables) {
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

  function createTable() {
    if (!lobbyWs || lobbyWs.readyState !== WebSocket.OPEN) {
      createResultEl.textContent = 'Not connected, please wait...';
      return;
    }
    lobbyWs.send(JSON.stringify({ type: 'create_table', table_id: newTableIdEl.value.trim() }));
  }

  // --- Lobby WebSocket ---

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
      } else if (msg.type === 'lobby_tables' && msg.data && Array.isArray(msg.data.tables)) {
        renderTables(msg.data.tables);
      } else if (msg.type === 'create_table_result' && msg.data) {
        if (msg.data.table_id) {
          createResultEl.textContent = msg.data.created ? `created ${msg.data.table_id}` : `using existing ${msg.data.table_id}`;
          openTable(msg.data.table_id);
        }
      } else if (msg.type === 'error' && msg.error) {
        createResultEl.textContent = msg.error;
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
  connectLobbyWs();
})();
