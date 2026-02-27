import { createTableDom } from './dom.js';
import { createRenderer } from './render.js';

const nameKey = 'hearts.player.name';
const tokenKey = 'hearts.player.token';
const trickCardInDurationMs = 520;
const trickCardInBufferMs = 80;

const tableId = decodeURIComponent(location.pathname.replace('/table/', ''));
const eventsEnabled = new URLSearchParams(location.search).get('events') === 'true';

const dom = createTableDom({ tableId, eventsEnabled });

const state = {
  ws: undefined,
  myPlayerId: '',
  lastTrickSignature: '',
  lastHandRenderKey: '',
  lastSnapshot: {},
  selectedPassCards: [],
  lastPlayers: [],
  liveTrickPlays: [],
  trickEventQueue: [],
  processingTrickEventQueue: false,
  pendingStateRefresh: 0,
  lastPhase: '',
  lastPassSubmittedCount: -1,
  lastPassReadyCount: -1,
  loggedPassReviewKey: ''
};

const renderer = createRenderer({ dom, state, send });

const token = ensureToken();

dom.addBotDefaultEl.onclick = () => {
  send({ type: 'add_bot', strategy: dom.botStrategySelectEl.value || '' });
  dom.botStrategySelectEl.value = '';
};

dom.startButtonEl.onclick = () => {
  send({ type: 'start' });
};

dom.submitPassEl.onclick = () => {
  if (state.selectedPassCards.length !== 3) {
    return;
  }
  send({ type: 'pass', cards: state.selectedPassCards });
};

dom.readyAfterPassEl.onclick = () => {
  send({ type: 'ready_after_pass' });
};

connect();

function playerName() {
  const stored = (localStorage.getItem(nameKey) || '').trim();
  return stored || 'Player';
}

function ensureToken() {
  let storedToken = localStorage.getItem(tokenKey);
  if (!storedToken) {
    storedToken = (self.crypto && self.crypto.randomUUID)
      ? self.crypto.randomUUID()
      : String(Date.now()) + Math.random().toString(16).slice(2);
    localStorage.setItem(tokenKey, storedToken);
  }
  return storedToken;
}

function waitMs(durationMs) {
  if (!durationMs || durationMs <= 0) {
    return Promise.resolve();
  }

  return new Promise((resolve) => {
    window.setTimeout(resolve, durationMs);
  });
}

function prefersReducedMotion() {
  return !!(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches);
}

function log(line) {
  if (!eventsEnabled) {
    return;
  }

  const ts = new Date().toLocaleTimeString();
  dom.logsEl.textContent += `[${ts}] ${line}\n`;
  dom.logsEl.scrollTop = dom.logsEl.scrollHeight;
}

function send(msg) {
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN) {
    return;
  }

  state.ws.send(JSON.stringify(msg));
}

function scheduleStateRefresh(delayMs) {
  if (state.pendingStateRefresh) {
    clearTimeout(state.pendingStateRefresh);
  }

  if (!delayMs || delayMs <= 0) {
    state.pendingStateRefresh = 0;
    requestState();
    return;
  }

  state.pendingStateRefresh = setTimeout(() => {
    state.pendingStateRefresh = 0;
    requestState();
  }, delayMs);
}

function requestState() {
  send({ type: 'state' });
}

function syncTrickFromLatestSnapshot() {
  if (state.processingTrickEventQueue || state.trickEventQueue.length > 0) {
    return;
  }

  const players = renderer.trickRenderPlayers();
  if (players.length === 0) {
    return;
  }

  const snapshotTrickPlays = Array.isArray(state.lastSnapshot && state.lastSnapshot.trick_plays)
    ? state.lastSnapshot.trick_plays.slice()
    : [];
  state.liveTrickPlays = snapshotTrickPlays;
  renderer.renderTrick(players, state.liveTrickPlays);
}

function enqueueTrickEvent(type, data) {
  state.trickEventQueue.push({ type, data: data || {} });
  void processTrickEventQueue();
}

async function processQueuedCardPlayed(data) {
  if (!data || !data.player_id || !data.card) {
    return;
  }

  const players = renderer.trickRenderPlayers();
  if (players.length === 0) {
    return;
  }

  const nextTrick = state.liveTrickPlays.filter((play) => play.player_id !== data.player_id);
  nextTrick.push({ player_id: data.player_id, card: data.card });
  state.liveTrickPlays = nextTrick;
  renderer.renderTrick(players, state.liveTrickPlays);

  if (!prefersReducedMotion()) {
    await waitMs(trickCardInDurationMs + trickCardInBufferMs);
  }
}

async function processQueuedTrickCompleted(data) {
  if (data && data.winner_player_id) {
    await renderer.animateTrickCardsToWinner(data.winner_player_id);
  }

  state.liveTrickPlays = [];
  state.lastTrickSignature = '';

  const players = renderer.trickRenderPlayers();
  if (players.length > 0) {
    renderer.renderTrick(players, state.liveTrickPlays);
  }
}

async function processTrickEventQueue() {
  if (state.processingTrickEventQueue) {
    return;
  }

  state.processingTrickEventQueue = true;
  try {
    while (state.trickEventQueue.length > 0) {
      const queued = state.trickEventQueue.shift();
      if (!queued) {
        continue;
      }

      if (queued.type === 'card_played') {
        await processQueuedCardPlayed(queued.data);
        continue;
      }

      if (queued.type === 'trick_completed') {
        await processQueuedTrickCompleted(queued.data);
      }
    }
  } finally {
    state.processingTrickEventQueue = false;
    syncTrickFromLatestSnapshot();
  }
}

function connect() {
  const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
  state.ws = new WebSocket(`${protocol}://${location.host}/ws/table/${encodeURIComponent(tableId)}`);

  state.ws.onopen = () => {
    log('connected');
    send({ type: 'join', name: playerName(), token });
    requestState();
  };

  state.ws.onclose = () => {
    log('disconnected, retrying...');
    setTimeout(connect, 1000);
  };

  state.ws.onerror = () => {
    log('websocket error');
  };

  state.ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);

    if (msg.error) {
      log(`error: ${msg.error}`);
      return;
    }

    switch (msg.type) {
      case 'table_state':
        renderer.renderState(msg.data || {}, { log });
        break;
      case 'join_result':
        if (msg.data && msg.data.accepted) {
          state.myPlayerId = msg.data.player_id;
          log(`joined as ${state.myPlayerId} seat ${msg.data.seat}`);
        } else {
          log(`join rejected: ${msg.data && msg.data.reason ? msg.data.reason : 'join rejected'}`);
        }
        requestState();
        break;
      case 'start_result':
      case 'play_result':
      case 'pass_result':
      case 'ready_after_pass_result':
        if (msg.data && msg.data.accepted) {
          log(`${msg.type}: ok`);
          if (msg.type === 'pass_result') {
            state.selectedPassCards = [];
          }
        } else {
          log(`${msg.type}: ${msg.data && msg.data.reason ? msg.data.reason : 'rejected'}`);
        }
        scheduleStateRefresh(0);
        break;
      case 'card_played':
        enqueueTrickEvent('card_played', msg.data || {});
        scheduleStateRefresh(0);
        break;
      case 'add_bot_result':
        log('bot added');
        scheduleStateRefresh(0);
        break;
      case 'trick_completed':
        enqueueTrickEvent('trick_completed', msg.data || {});
        scheduleStateRefresh(0);
        break;
      default:
        log(`${msg.type}: ${JSON.stringify(msg.data || {})}`);
        scheduleStateRefresh(0);
        break;
    }
  };
}
