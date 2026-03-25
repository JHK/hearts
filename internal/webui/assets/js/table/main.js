import { createTableDom } from './dom.js';
import { createRenderer } from './render.js';
import { playHeartsBreaking, playQueenOfSpades, setMuted } from './audio.js';

const nameKey = 'hearts.player.name';
const tokenKey = 'hearts.player.token';
const speedKey = 'hearts.animation.speed';
const soundKey = 'hearts.sound.enabled';
const notifyKey = 'hearts.notifications.enabled';
const trickCardInBufferMs = 80;

const tableId = decodeURIComponent(location.pathname.replace('/table/', ''));
const eventsEnabled = new URLSearchParams(location.search).get('events') === 'true';

const dom = createTableDom({ tableId, eventsEnabled });

const state = {
  ws: undefined,
  myPlayerId: '',
  isObserver: false,
  lastTrickSignature: '',
  lastHandRenderKey: '',
  lastSnapshot: {},
  selectedPassCards: [],
  lastPlayers: [],
  liveTrickPlays: [],
  liveTurnPlayerId: undefined,
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

function isFastSpeed() {
  return document.body.dataset.speed === 'fast';
}

function applySpeed(fast) {
  if (fast) {
    document.body.dataset.speed = 'fast';
  } else {
    delete document.body.dataset.speed;
  }
  dom.speedToggleEl.checked = fast;
}

applySpeed(localStorage.getItem(speedKey) === 'fast');

const soundEnabled = localStorage.getItem(soundKey) !== 'false';
dom.soundToggleEl.checked = soundEnabled;
setMuted(!soundEnabled);

dom.soundToggleEl.onchange = () => {
  const enabled = dom.soundToggleEl.checked;
  setMuted(!enabled);
  localStorage.setItem(soundKey, enabled ? 'true' : 'false');
};

let notificationsEnabled = localStorage.getItem(notifyKey) === 'true';
dom.notifyToggleEl.checked = notificationsEnabled;

dom.notifyToggleEl.onchange = async () => {
  if (dom.notifyToggleEl.checked) {
    if ('Notification' in window && Notification.permission !== 'granted') {
      const result = await Notification.requestPermission();
      if (result !== 'granted') {
        dom.notifyToggleEl.checked = false;
        return;
      }
    }
    notificationsEnabled = true;
  } else {
    notificationsEnabled = false;
  }
  localStorage.setItem(notifyKey, notificationsEnabled ? 'true' : 'false');
};

function notifyTurn(body) {
  if (!notificationsEnabled || document.hasFocus()) return;
  if (!('Notification' in window) || Notification.permission !== 'granted') return;
  const n = new Notification('Hearts', { body, tag: 'hearts-turn' });
  n.onclick = () => { window.focus(); n.close(); };
}

dom.settingsToggleEl.onclick = () => {
  dom.settingsPanelEl.classList.toggle('hidden');
};

dom.speedToggleEl.onchange = () => {
  const fast = dom.speedToggleEl.checked;
  applySpeed(fast);
  localStorage.setItem(speedKey, fast ? 'fast' : 'normal');
};

document.addEventListener('click', (e) => {
  if (!dom.settingsPanelEl.classList.contains('hidden') &&
      !dom.settingsPanelEl.contains(e.target) &&
      !dom.settingsToggleEl.contains(e.target)) {
    dom.settingsPanelEl.classList.add('hidden');
  }
});

dom.addBotDefaultEl.onclick = () => {
  send({ type: 'add_bot', strategy: dom.botStrategySelectEl.value || '' });
  dom.botStrategySelectEl.value = 'smart';
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

dom.rematchButtonEl.onclick = () => {
  send({ type: 'rematch' });
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
    console.warn('send: WebSocket not connected, dropping message:', msg.type);
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

  state.liveTurnPlayerId = data.player_id;
  renderer.renderState(state.lastSnapshot, { log });

  if (data.card === 'QS') {
    playQueenOfSpades();
  } else if (data.breaks_hearts) {
    playHeartsBreaking();
  }

  if (!prefersReducedMotion()) {
    const cardInMs = isFastSpeed() ? 260 : 520;
    await waitMs(cardInMs + trickCardInBufferMs);
  }
}

async function processQueuedTrickCompleted(data) {
  if (data && data.winner_player_id) {
    await renderer.animateTrickCardsToWinner(data.winner_player_id);
  }

  state.liveTrickPlays = [];
  state.lastTrickSignature = '';

  if (data && data.winner_player_id) {
    state.liveTurnPlayerId = data.winner_player_id;
  }
  renderer.renderState(state.lastSnapshot, { log });
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
    renderer.renderState(state.lastSnapshot, { log });
    scheduleStateRefresh(0);
  }
}

let initialConnectFailures = 0;
const maxInitialConnectRetries = 2;
let reconnectAttempts = 0;

function reconnectDelayMs() {
  const base = 1000;
  const max = 30000;
  const delay = Math.min(base * Math.pow(2, reconnectAttempts), max);
  return delay + Math.random() * 500;
}

function connect() {
  const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
  state.ws = new WebSocket(`${protocol}://${location.host}/ws/table/${encodeURIComponent(tableId)}`);

  let opened = false;

  state.ws.onopen = () => {
    opened = true;
    initialConnectFailures = 0;
    reconnectAttempts = 0;
    log('connected');
    send({ type: 'join', name: playerName(), token });
    requestState();
  };

  state.ws.onclose = () => {
    if (!opened) {
      initialConnectFailures++;
      if (initialConnectFailures <= maxInitialConnectRetries) {
        log(`connect failed, retrying (${initialConnectFailures}/${maxInitialConnectRetries})...`);
        setTimeout(connect, 1000);
        return;
      }
      window.location.href = '/';
      return;
    }
    const delay = reconnectDelayMs();
    reconnectAttempts++;
    log(`disconnected, reconnecting in ${Math.round(delay / 1000)}s...`);
    setTimeout(connect, delay);
  };

  state.ws.onerror = () => {
    log('websocket error');
  };

  state.ws.onmessage = (event) => {
    let msg;
    try {
      msg = JSON.parse(event.data);
    } catch (err) {
      console.warn('failed to parse WebSocket message:', err);
      return;
    }

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
          state.isObserver = false;
          dom.observerBadgeEl.hidden = true;
          log(`joined as ${state.myPlayerId} seat ${msg.data.seat}`);
        } else {
          const reason = msg.data && msg.data.reason ? msg.data.reason : 'join rejected';
          log(`join rejected: ${reason}`);
          if (reason === 'table is full' || reason === 'round already in progress') {
            state.isObserver = true;
            dom.observerBadgeEl.hidden = false;
          }
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
      case 'your_turn':
        notifyTurn("It's your turn to play!");
        scheduleStateRefresh(0);
        break;
      case 'pass_submitted':
        if (msg.data && msg.data.submitted === 0 && msg.data.direction && !state.isObserver) {
          notifyTurn('Time to pass cards!');
        }
        scheduleStateRefresh(0);
        break;
      case 'player_left':
        log(`${msg.data.player.name} left the table`);
        scheduleStateRefresh(0);
        break;
      case 'game_paused':
        log('game paused — player disconnected');
        scheduleStateRefresh(0);
        break;
      case 'game_resumed':
        log('game resumed');
        scheduleStateRefresh(0);
        break;
      case 'resume_game_result':
        if (msg.data && msg.data.accepted) {
          log('game resumed');
        } else {
          log(`resume failed: ${msg.data && msg.data.reason ? msg.data.reason : 'rejected'}`);
        }
        scheduleStateRefresh(0);
        break;
      case 'trick_completed':
        enqueueTrickEvent('trick_completed', msg.data || {});
        scheduleStateRefresh(0);
        break;
      case 'rematch_result':
        if (msg.data && msg.data.accepted) {
          log('rematch vote accepted');
        } else {
          log(`rematch failed: ${msg.data && msg.data.reason ? msg.data.reason : 'rejected'}`);
        }
        scheduleStateRefresh(0);
        break;
      case 'rematch_vote':
        log(`rematch: ${msg.data.votes}/${msg.data.total} players ready`);
        scheduleStateRefresh(0);
        break;
      case 'rematch_starting':
        log('rematch starting!');
        renderer.resetGameOver();
        state.selectedPassCards = [];
        state.lastTrickSignature = '';
        state.lastHandRenderKey = '';
        state.liveTrickPlays = [];
        state.liveTurnPlayerId = undefined;
        state.trickEventQueue = [];
        state.processingTrickEventQueue = false;
        state.lastPhase = '';
        state.lastPassSubmittedCount = -1;
        state.lastPassReadyCount = -1;
        state.loggedPassReviewKey = '';
        scheduleStateRefresh(0);
        break;
      default:
        log(`${msg.type}: ${JSON.stringify(msg.data || {})}`);
        scheduleStateRefresh(0);
        break;
    }
  };
}
