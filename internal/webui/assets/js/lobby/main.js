(() => {
  const nameKey = 'hearts.player.name';
  const tokenKey = 'hearts.player.token';

  const nameEl = document.getElementById('name');
  const newTableIdEl = document.getElementById('newTableId');
  const createResultEl = document.getElementById('createResult');
  const tablesEl = document.getElementById('tables');

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
    const res = await fetch('/api/tables');
    const data = await res.json();
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
      badge.className = table.started ? 'badge badge-active' : 'badge badge-waiting';
      badge.textContent = table.started ? 'In progress' : 'Waiting';

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
  }

  document.getElementById('createTable').onclick = createTable;

  nameEl.value = localStorage.getItem(nameKey) || 'Player';
  ensureToken();
  fetchTables();
  setInterval(fetchTables, 1500);
})();
