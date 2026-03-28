import { createTableDom } from './dom.js';
import { createRenderer } from './render.js';
import { playHeartsBreaking, playQueenOfSpades, setMuted } from './audio.js';
import { nameKey, tokenKey, speedKey, soundKey, notifyKey, ensureToken, initSettingsPopover } from '../shared/settings.js';
const trickCardInBufferMs = 80;
let reconnectTimer = null;

const tableId = decodeURIComponent(location.pathname.replace('/table/', ''));

const dom = createTableDom();

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

const renderer = createRenderer({ dom, state, send, claimSeat });

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

dom.nameInputEl.value = playerName();

let renameTimer = null;
dom.nameInputEl.addEventListener('input', () => {
  const name = dom.nameInputEl.value.trim() || 'Player';
  localStorage.setItem(nameKey, name);
  clearTimeout(renameTimer);
  renameTimer = setTimeout(() => {
    if (state.myPlayerId) {
      send({ type: 'rename', name });
    }
  }, 300);
});

initSettingsPopover(dom.settingsToggleEl, dom.settingsPanelEl);
initSettingsPopover(dom.addBotToggleEl, dom.addBotPanelEl);

dom.speedToggleEl.onchange = () => {
  const fast = dom.speedToggleEl.checked;
  applySpeed(fast);
  localStorage.setItem(speedKey, fast ? 'fast' : 'normal');
};

dom.addBotPanelEl.onclick = (e) => {
  const btn = e.target.closest('.bot-strength-option');
  if (!btn) return;
  send({ type: 'add_bot', strategy: btn.dataset.strategy });
  dom.addBotPanelEl.classList.add('hidden');
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

function claimSeat(seat) {
  send({ type: 'claim_seat', seat, name: playerName(), token });
}

connect();

// When the browser restores this page from bfcache, the JS state is
// preserved but the WebSocket is dead. Force a fresh reconnection so
// the hand re-renders with working click handlers.
window.addEventListener('pageshow', (event) => {
  if (event.persisted && (!state.ws || state.ws.readyState !== WebSocket.OPEN)) {
    state.lastHandRenderKey = '';
    state.lastTrickSignature = '';
    clearTimeout(reconnectTimer);
    connect();
  }
});

function playerName() {
  const stored = (localStorage.getItem(nameKey) || '').trim();
  return stored || 'Player';
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
  if (!window.__HEARTS_DEV__) {
    return;
  }

  const ts = new Date().toLocaleTimeString();
  console.log(`[${ts}] ${line}`);
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

function scheduleReconnect(delayMs) {
  clearTimeout(reconnectTimer);
  reconnectTimer = setTimeout(connect, delayMs);
}

function connect() {
  reconnectTimer = null;
  const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
  state.ws = new WebSocket(`${protocol}://${location.host}/ws/table/${encodeURIComponent(tableId)}`);

  let opened = false;

  state.ws.onopen = () => {
    opened = true;
    initialConnectFailures = 0;
    reconnectAttempts = 0;
    state.lastHandRenderKey = '';
    state.lastTrickSignature = '';
    log('connected');
    send({ type: 'join', name: playerName(), token });
    requestState();
  };

  state.ws.onclose = () => {
    if (!opened) {
      initialConnectFailures++;
      if (initialConnectFailures <= maxInitialConnectRetries) {
        log(`connect failed, retrying (${initialConnectFailures}/${maxInitialConnectRetries})...`);
        scheduleReconnect(1000);
        return;
      }
      window.location.href = '/';
      return;
    }
    const delay = reconnectDelayMs();
    reconnectAttempts++;
    log(`disconnected, reconnecting in ${Math.round(delay / 1000)}s...`);
    scheduleReconnect(delay);
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
          log(`joined as ${state.myPlayerId} seat ${msg.data.seat}`);
        } else {
          const reason = msg.data && msg.data.reason ? msg.data.reason : 'join rejected';
          log(`join rejected: ${reason}`);
          if (reason === 'table is full' || reason === 'round already in progress') {
            state.isObserver = true;
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
      case 'claim_seat_result':
        if (msg.data && msg.data.accepted) {
          state.myPlayerId = msg.data.player_id;
          state.isObserver = false;
          log(`claimed seat ${msg.data.seat} as ${state.myPlayerId}`);
        } else {
          log(`claim seat failed: ${msg.data && msg.data.reason ? msg.data.reason : 'rejected'}`);
        }
        scheduleStateRefresh(0);
        break;
      case 'rename_result':
        if (msg.data && !msg.data.accepted) {
          log(`rename failed: ${msg.data.reason || 'rejected'}`);
          // Revert input and localStorage to the server-authoritative name.
          const snap = state.lastSnapshot;
          if (snap && snap.players) {
            const me = snap.players.find(p => p.player_id === state.myPlayerId);
            if (me) {
              dom.nameInputEl.value = me.name;
              localStorage.setItem(nameKey, me.name);
            }
          }
        }
        break;
      case 'player_renamed':
        if (msg.data && msg.data.player) {
          log(`${msg.data.old_name} is now ${msg.data.player.name}`);
        }
        scheduleStateRefresh(0);
        break;
      case 'seat_claimed':
        if (msg.data && msg.data.player) {
          log(`${msg.data.player.name} took ${msg.data.old_name}'s seat`);
        }
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
