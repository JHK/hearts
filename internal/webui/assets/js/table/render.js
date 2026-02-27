import { cardAltText, cardImageURL } from './cards.js';

export function createRenderer({ dom, state, send }) {
  function setSeatTurnClass(el, player, turnPlayerId) {
    if (player && turnPlayerId && player.player_id === turnPlayerId) {
      el.classList.add('turn');
    } else {
      el.classList.remove('turn');
    }
  }

  function relativeSeatPlayers(players) {
    const me = players.find((p) => p.player_id === state.myPlayerId);
    if (!me) {
      return { me: null, top: null, left: null, right: null };
    }

    const bySeat = new Map();
    for (const player of players) {
      bySeat.set(player.seat, player);
    }

    function atOffset(offset) {
      const seat = (me.seat + offset + 4) % 4;
      return bySeat.get(seat);
    }

    return {
      me,
      top: atOffset(2),
      left: atOffset(1),
      right: atOffset(3)
    };
  }

  function setSeatLabels(relativePlayers, turnPlayerId) {
    function displayName(player) {
      if (!player) {
        return '-';
      }

      return player.is_bot ? `${player.name} [bot]` : player.name;
    }

    if (!relativePlayers.me) {
      dom.seatTopNameEl.textContent = '-';
      dom.seatLeftNameEl.textContent = '-';
      dom.seatRightNameEl.textContent = '-';
      dom.seatBottomNameEl.textContent = '-';

      dom.seatTopEl.classList.remove('turn');
      dom.seatLeftEl.classList.remove('turn');
      dom.seatRightEl.classList.remove('turn');
      dom.seatBottomEl.classList.remove('turn');
      return;
    }

    dom.seatTopNameEl.textContent = displayName(relativePlayers.top);
    dom.seatLeftNameEl.textContent = displayName(relativePlayers.left);
    dom.seatRightNameEl.textContent = displayName(relativePlayers.right);
    dom.seatBottomNameEl.textContent = displayName(relativePlayers.me);

    setSeatTurnClass(dom.seatTopEl, relativePlayers.top, turnPlayerId);
    setSeatTurnClass(dom.seatLeftEl, relativePlayers.left, turnPlayerId);
    setSeatTurnClass(dom.seatRightEl, relativePlayers.right, turnPlayerId);
    setSeatTurnClass(dom.seatBottomEl, relativePlayers.me, turnPlayerId);
  }

  function slotKeyFor(player, me) {
    if (!me || !player) {
      return '';
    }

    const offset = (player.seat - me.seat + 4) % 4;
    switch (offset) {
      case 0:
        return 'bottom';
      case 1:
        return 'left';
      case 2:
        return 'top';
      case 3:
        return 'right';
      default:
        return '';
    }
  }

  function buildTrickCard(cardValue) {
    const card = document.createElement('div');
    card.className = 'play-card';
    const imageURL = cardImageURL(cardValue);
    if (!imageURL) {
      card.textContent = cardValue;
      return card;
    }

    const image = document.createElement('img');
    image.className = 'play-card-image';
    image.src = imageURL;
    image.alt = cardAltText(cardValue);
    image.loading = 'lazy';
    image.decoding = 'async';
    card.appendChild(image);

    return card;
  }

  function renderTrick(players, trickPlays) {
    const me = players.find((player) => player.player_id === state.myPlayerId);
    const trickSignature = `${me ? me.seat : '-'}|${(trickPlays || []).map((play) => `${play.player_id}:${play.card}`).join('|')}`;
    if (trickSignature === state.lastTrickSignature) {
      return;
    }

    state.lastTrickSignature = trickSignature;
    const byPlayer = new Map();
    for (const player of players) {
      byPlayer.set(player.player_id, player);
    }

    const desiredBySlot = {
      top: { signature: '', card: '' },
      left: { signature: '', card: '' },
      right: { signature: '', card: '' },
      bottom: { signature: '', card: '' }
    };

    for (const play of trickPlays || []) {
      const player = byPlayer.get(play.player_id);
      const slotKey = slotKeyFor(player, me);
      if (!slotKey) {
        continue;
      }

      desiredBySlot[slotKey] = {
        signature: `${play.player_id}:${play.card}`,
        card: play.card
      };
    }

    for (const slotKey of Object.keys(dom.trickSlotEls)) {
      const slotEl = dom.trickSlotEls[slotKey];
      const next = desiredBySlot[slotKey];
      const currentSignature = slotEl.dataset.playSignature || '';
      if (currentSignature === next.signature) {
        continue;
      }

      slotEl.dataset.playSignature = next.signature;
      slotEl.replaceChildren();
      if (!next.signature) {
        continue;
      }

      slotEl.appendChild(buildTrickCard(next.card));
    }
  }

  function renderYourHand(cards, phase, isYourTurn, passSubmitted, passReceived, passReady) {
    const safeCards = Array.isArray(cards) ? cards : [];
    const safePassReceived = Array.isArray(passReceived) ? passReceived : [];
    const receivedSet = new Set(safePassReceived);
    state.selectedPassCards = state.selectedPassCards.filter((card) => safeCards.includes(card));
    const handRenderKey = `${phase}|${isYourTurn ? '1' : '0'}|${passSubmitted ? '1' : '0'}|${passReady ? '1' : '0'}|${safePassReceived.join(',')}|${state.selectedPassCards.join(',')}|${safeCards.join(',')}`;
    if (handRenderKey === state.lastHandRenderKey) {
      return;
    }

    state.lastHandRenderKey = handRenderKey;
    dom.seatBottomHandEl.innerHTML = '';

    if (safeCards.length === 0) {
      return;
    }

    for (const cardValue of safeCards) {
      const card = document.createElement('button');
      card.type = 'button';
      card.className = 'hand-card';
      if (phase === 'playing' && isYourTurn) {
        card.className += ' playable';
      }
      if (phase === 'passing' && !passSubmitted) {
        card.className += ' pass-selectable';
      }
      if (state.selectedPassCards.includes(cardValue)) {
        card.className += ' selected';
      }
      if (phase === 'pass_review' && !passReady && receivedSet.has(cardValue)) {
        card.className += ' selected';
      }

      const imageURL = cardImageURL(cardValue);
      if (imageURL) {
        const image = document.createElement('img');
        image.className = 'hand-card-image';
        image.src = imageURL;
        image.alt = cardAltText(cardValue);
        image.loading = 'lazy';
        image.decoding = 'async';
        card.appendChild(image);
      } else {
        card.textContent = cardValue;
      }

      card.onclick = () => {
        if (phase === 'passing' && !passSubmitted) {
          const selected = state.selectedPassCards.includes(cardValue);
          if (selected) {
            state.selectedPassCards = state.selectedPassCards.filter((card) => card !== cardValue);
          } else if (state.selectedPassCards.length < 3) {
            state.selectedPassCards = [...state.selectedPassCards, cardValue];
          }
          renderYourHand(safeCards, phase, isYourTurn, passSubmitted);
          renderPassPanel(state.lastSnapshot);
          return;
        }

        if (phase === 'playing') {
          send({ type: 'play', card: cardValue });
        }
      };
      dom.seatBottomHandEl.appendChild(card);
    }
  }

  function renderSeatBackHand(container, count) {
    container.innerHTML = '';
    for (let i = 0; i < count; i += 1) {
      const back = document.createElement('div');
      back.className = 'back-card';
      container.appendChild(back);
    }
  }

  function renderOtherHands(relativePlayers, handSizes) {
    const seats = [
      { player: relativePlayers.top, container: dom.seatTopHandEl },
      { player: relativePlayers.left, container: dom.seatLeftHandEl },
      { player: relativePlayers.right, container: dom.seatRightHandEl }
    ];

    for (const seat of seats) {
      if (!seat.player) {
        seat.container.innerHTML = '';
        continue;
      }

      const count = handSizes && handSizes[seat.player.player_id] ? handSizes[seat.player.player_id] : 0;
      renderSeatBackHand(seat.container, count);
    }
  }

  function formatCardList(cards) {
    if (!Array.isArray(cards) || cards.length === 0) {
      return '-';
    }

    return cards.join(', ');
  }

  function pointsFor(pointsByPlayer, playerID) {
    const value = pointsByPlayer && pointsByPlayer[playerID];
    const number = Number(value);
    return Number.isFinite(number) ? number : 0;
  }

  function renderScoreboard(snapshot) {
    const players = Array.isArray(snapshot.players) ? snapshot.players : [];
    const history = Array.isArray(snapshot.round_history) ? snapshot.round_history : [];
    const roundPoints = snapshot && snapshot.round_points ? snapshot.round_points : {};
    const totalPoints = snapshot && snapshot.total_points ? snapshot.total_points : {};

    dom.scoreboardHeadEl.innerHTML = '';
    dom.scoreboardBodyEl.innerHTML = '';

    const headerRow = document.createElement('tr');
    const roundHeaderCell = document.createElement('th');
    roundHeaderCell.scope = 'col';
    roundHeaderCell.textContent = 'Round';
    headerRow.appendChild(roundHeaderCell);

    if (players.length === 0) {
      const playersHeaderCell = document.createElement('th');
      playersHeaderCell.scope = 'col';
      playersHeaderCell.textContent = 'Players';
      headerRow.appendChild(playersHeaderCell);

      dom.scoreboardHeadEl.appendChild(headerRow);

      const emptyRow = document.createElement('tr');
      const emptyLabel = document.createElement('th');
      emptyLabel.scope = 'row';
      emptyLabel.textContent = 'Current';
      const emptyValue = document.createElement('td');
      emptyValue.textContent = '-';
      emptyRow.appendChild(emptyLabel);
      emptyRow.appendChild(emptyValue);
      dom.scoreboardBodyEl.appendChild(emptyRow);
      return;
    }

    for (const player of players) {
      const playerHeaderCell = document.createElement('th');
      playerHeaderCell.scope = 'col';
      playerHeaderCell.textContent = player.is_bot ? `${player.name} [bot]` : player.name;
      headerRow.appendChild(playerHeaderCell);
    }
    dom.scoreboardHeadEl.appendChild(headerRow);

    function appendPointsRow(label, pointsByPlayer, rowClassName) {
      const row = document.createElement('tr');
      if (rowClassName) {
        row.classList.add(rowClassName);
      }

      const labelCell = document.createElement('th');
      labelCell.scope = 'row';
      labelCell.textContent = label;
      row.appendChild(labelCell);

      for (const player of players) {
        const valueCell = document.createElement('td');
        valueCell.textContent = String(pointsFor(pointsByPlayer, player.player_id));
        row.appendChild(valueCell);
      }

      dom.scoreboardBodyEl.appendChild(row);
    }

    history.forEach((entry, index) => {
      appendPointsRow(`Round ${index + 1}`, entry, 'scoreboard-history-row');
    });

    const liveTotalPoints = {};
    for (const player of players) {
      liveTotalPoints[player.player_id] = pointsFor(totalPoints, player.player_id) + pointsFor(roundPoints, player.player_id);
    }

    appendPointsRow('Current', roundPoints, 'scoreboard-current-row');
    appendPointsRow('Sum', liveTotalPoints, 'scoreboard-total-row');
  }

  function renderPassPanel(snapshot) {
    const phase = snapshot && snapshot.phase ? snapshot.phase : '';
    const passSubmitted = !!(snapshot && snapshot.pass_submitted);
    const passReady = !!(snapshot && snapshot.pass_ready);

    if (phase !== 'passing' && phase !== 'pass_review') {
      dom.passSummaryEl.textContent = '';
      dom.passDetailsEl.textContent = '';
      dom.passSelectionEl.textContent = '';
      dom.passSummaryEl.hidden = true;
      dom.passDetailsEl.hidden = true;
      dom.passSelectionEl.hidden = true;
      dom.submitPassEl.hidden = true;
      dom.readyAfterPassEl.hidden = true;
      return;
    }

    dom.passSummaryEl.textContent = '';
    dom.passDetailsEl.textContent = '';
    dom.passSelectionEl.textContent = '';
    dom.passSummaryEl.hidden = true;
    dom.passDetailsEl.hidden = true;
    dom.passSelectionEl.hidden = true;

    if (phase === 'passing') {
      dom.submitPassEl.hidden = false;
      dom.submitPassEl.disabled = passSubmitted || state.selectedPassCards.length !== 3;
      dom.readyAfterPassEl.hidden = true;
      return;
    }

    dom.submitPassEl.hidden = true;
    dom.readyAfterPassEl.hidden = false;
    dom.readyAfterPassEl.disabled = passReady;
  }

  function renderState(snapshot, { log }) {
    state.lastSnapshot = snapshot || {};
    const players = snapshot.players || [];
    const phase = snapshot.phase || (snapshot.started ? 'playing' : '');
    const isPassing = phase === 'passing';
    const isPassReview = phase === 'pass_review';
    const isPlaying = phase === 'playing';
    const passSubmittedCount = (snapshot && snapshot.pass_submitted_count) || 0;
    const passReadyCount = (snapshot && snapshot.pass_ready_count) || 0;
    const totalPoints = (snapshot && snapshot.total_points) || {};

    if (phase !== state.lastPhase) {
      if (phase === 'passing') {
        log(`pass phase started (${(snapshot && snapshot.pass_direction) || 'left'})`);
      } else if (phase === 'pass_review') {
        log('pass complete');
      }
    }
    if (isPassing && passSubmittedCount !== state.lastPassSubmittedCount) {
      log(`pass submissions: ${passSubmittedCount}/4`);
    }
    if (isPassReview && passReadyCount !== state.lastPassReadyCount) {
      log(`ready after pass: ${passReadyCount}/4`);
    }
    if (isPassReview) {
      const reviewKey = `${formatCardList(snapshot.pass_sent || [])}|${formatCardList(snapshot.pass_received || [])}`;
      if (reviewKey !== state.loggedPassReviewKey) {
        log(`you passed: ${formatCardList(snapshot.pass_sent || [])}`);
        log(`you received: ${formatCardList(snapshot.pass_received || [])}`);
        state.loggedPassReviewKey = reviewKey;
      }
    }
    if (!isPassReview) {
      state.loggedPassReviewKey = '';
    }
    state.lastPhase = phase;
    state.lastPassSubmittedCount = passSubmittedCount;
    state.lastPassReadyCount = passReadyCount;

    const waitingForPlayers = !snapshot.started && players.length < 4;
    dom.statusEl.hidden = !waitingForPlayers;
    if (waitingForPlayers) {
      dom.statusEl.textContent = '⌛';
    }

    dom.trickSectionEl.hidden = false;

    const canAddBot = !snapshot.started && players.length < 4;
    dom.botControlEl.classList.toggle('hidden', !canAddBot);
    if (!canAddBot) {
      dom.botStrategySelectEl.value = '';
    }

    const showStartControl = !snapshot.started && players.length === 4;
    const showPassControls = isPassing || isPassReview;
    dom.centerControlsEl.classList.toggle('hidden', !showStartControl && !showPassControls);
    dom.startButtonEl.hidden = !showStartControl;
    dom.startButtonEl.disabled = !showStartControl;
    const hasCompletedRound = Object.values(totalPoints).some((points) => Number(points) > 0);
    dom.startButtonEl.textContent = hasCompletedRound ? 'Continue' : 'Start';

    state.lastPlayers = players;
    const relativePlayers = relativeSeatPlayers(players);
    const turnPlayer = players.find((player) => player.player_id === snapshot.turn_player_id);
    const isYourTurn = !!isPlaying && !!state.myPlayerId && snapshot.turn_player_id === state.myPlayerId;
    if (isPlaying) {
      if (!snapshot.turn_player_id) {
        dom.turnIndicatorEl.textContent = 'collecting trick';
      } else {
        dom.turnIndicatorEl.textContent = isYourTurn
          ? 'waiting for you'
          : `waiting for ${turnPlayer ? turnPlayer.name : snapshot.turn_player_id}`;
      }
    } else if (isPassing) {
      dom.turnIndicatorEl.textContent = `pass 3 cards ${snapshot.pass_direction ? `(${snapshot.pass_direction})` : ''}`.trim();
    } else if (isPassReview) {
      dom.turnIndicatorEl.textContent = 'waiting for players to continue';
    } else {
      dom.turnIndicatorEl.textContent = '';
    }

    const serverTrickPlays = (snapshot.trick_plays || []).slice();
    if (!state.processingTrickEventQueue && state.trickEventQueue.length === 0) {
      state.liveTrickPlays = serverTrickPlays;
    }

    setSeatLabels(relativePlayers, snapshot.turn_player_id);
    renderTrick(players, state.liveTrickPlays);
    renderYourHand(
      snapshot.hand || [],
      phase,
      isYourTurn,
      !!snapshot.pass_submitted,
      snapshot.pass_received || [],
      !!snapshot.pass_ready
    );
    renderOtherHands(relativePlayers, snapshot.hand_sizes || {});
    renderPassPanel(snapshot);
    renderScoreboard(snapshot);
  }

  function seatElementForPlayerId(playerId) {
    if (!playerId || !state.lastPlayers || state.lastPlayers.length === 0) {
      return null;
    }

    const me = state.lastPlayers.find((player) => player.player_id === state.myPlayerId);
    const winner = state.lastPlayers.find((player) => player.player_id === playerId);
    if (!me || !winner) {
      return null;
    }

    const offset = (winner.seat - me.seat + 4) % 4;
    switch (offset) {
      case 0:
        return dom.seatBottomEl;
      case 1:
        return dom.seatLeftEl;
      case 2:
        return dom.seatTopEl;
      case 3:
        return dom.seatRightEl;
      default:
        return null;
    }
  }

  function seatNameElementForPlayerId(playerId) {
    if (!playerId || !state.lastPlayers || state.lastPlayers.length === 0) {
      return null;
    }

    const me = state.lastPlayers.find((player) => player.player_id === state.myPlayerId);
    const winner = state.lastPlayers.find((player) => player.player_id === playerId);
    if (!me || !winner) {
      return null;
    }

    const offset = (winner.seat - me.seat + 4) % 4;
    switch (offset) {
      case 0:
        return dom.seatBottomNameEl;
      case 1:
        return dom.seatLeftNameEl;
      case 2:
        return dom.seatTopNameEl;
      case 3:
        return dom.seatRightNameEl;
      default:
        return null;
    }
  }

  function textVisualRect(el) {
    if (!el) {
      return null;
    }

    const text = (el.textContent || '').trim();
    if (!text) {
      return el.getBoundingClientRect();
    }

    const range = document.createRange();
    range.selectNodeContents(el);
    const rect = range.getBoundingClientRect();
    range.detach?.();

    if (!rect || rect.width === 0 || rect.height === 0) {
      return el.getBoundingClientRect();
    }

    return rect;
  }

  function animateTrickWinner(playerId) {
    const seatEl = seatElementForPlayerId(playerId);
    if (!seatEl) {
      return;
    }

    seatEl.classList.remove('trick-winner');
    void seatEl.offsetWidth;
    seatEl.classList.add('trick-winner');
    seatEl.addEventListener('animationend', () => {
      seatEl.classList.remove('trick-winner');
    }, { once: true });
  }

  function prefersReducedMotion() {
    return !!(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches);
  }

  function animateTrickCardsToWinner(playerId) {
    return new Promise((resolve) => {
      const seatEl = seatElementForPlayerId(playerId);
      const seatNameEl = seatNameElementForPlayerId(playerId);
      if (!seatEl || !seatNameEl || !dom.trickAnimationLayerEl) {
        animateTrickWinner(playerId);
        resolve();
        return;
      }
      const nameRect = seatNameEl.getBoundingClientRect();
      const textRect = textVisualRect(seatNameEl) || nameRect;
      const targetCenterX = textRect.left + (textRect.width / 2);
      const targetCenterY = textRect.top + (textRect.height / 2);
      const reduceMotion = prefersReducedMotion();

      const animatedCards = [];
      const slotEntries = Object.entries(dom.trickSlotEls);
      for (const [slotKey, slotEl] of slotEntries) {
        const cardEl = slotEl.querySelector('.play-card');
        if (!cardEl) {
          continue;
        }

        const cardRect = cardEl.getBoundingClientRect();
        const clone = cardEl.cloneNode(true);
        clone.classList.add('trick-capture-card');
        clone.style.left = `${cardRect.left}px`;
        clone.style.top = `${cardRect.top}px`;
        clone.style.width = `${cardRect.width}px`;
        clone.style.height = `${cardRect.height}px`;
        dom.trickAnimationLayerEl.appendChild(clone);

        animatedCards.push({
          el: clone,
          slotKey,
          centerX: cardRect.left + (cardRect.width / 2),
          centerY: cardRect.top + (cardRect.height / 2)
        });

        slotEl.replaceChildren();
        slotEl.dataset.playSignature = '';
      }

      if (animatedCards.length === 0) {
        animateTrickWinner(playerId);
        resolve();
        return;
      }

      if (reduceMotion) {
        for (const card of animatedCards) {
          card.el.remove();
        }
        animateTrickWinner(playerId);
        resolve();
        return;
      }

      const durationMs = 1400;
      const staggerMs = 90;
      const sortPriorityBySlot = { top: 0, left: 1, right: 2, bottom: 3 };
      animatedCards.sort((a, b) => (sortPriorityBySlot[a.slotKey] || 0) - (sortPriorityBySlot[b.slotKey] || 0));

      requestAnimationFrame(() => {
        animatedCards.forEach((card, index) => {
          const deltaX = targetCenterX - card.centerX;
          const deltaY = targetCenterY - card.centerY;
          const angle = (index - ((animatedCards.length - 1) / 2)) * 7;
          card.el.style.transitionDelay = `${index * staggerMs}ms`;
          card.el.style.transform = `translate(${deltaX}px, ${deltaY}px) scale(0.56) rotate(${angle}deg)`;
          card.el.style.opacity = '0';
        });
      });

      window.setTimeout(() => {
        for (const card of animatedCards) {
          card.el.remove();
        }
        animateTrickWinner(playerId);
        resolve();
      }, durationMs + (animatedCards.length - 1) * staggerMs + 80);
    });
  }

  function trickRenderPlayers() {
    if (Array.isArray(state.lastPlayers) && state.lastPlayers.length > 0) {
      return state.lastPlayers;
    }
    if (state.lastSnapshot && Array.isArray(state.lastSnapshot.players)) {
      return state.lastSnapshot.players;
    }

    return [];
  }

  return {
    renderState,
    renderTrick,
    animateTrickCardsToWinner,
    trickRenderPlayers
  };
}
