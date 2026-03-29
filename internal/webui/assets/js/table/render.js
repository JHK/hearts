import { cardAltText, cardImageURL } from './cards.js';

const CHART_COLORS = ['#116466', '#b44f26', '#5b6f83', '#8b5cf6'];
const MEDALS = ['🏆', '🥈', '🥉'];

export function createRenderer({ dom, state, send }) {
  let scoreChart = null;
  let gameOverRendered = false;
  function effectiveMe(players) {
    if (!players || players.length === 0) return null;
    const me = players.find((p) => p.player_id === state.myPlayerId);
    if (me) return me;
    if (state.isObserver) {
      return players.find((p) => p.seat === 0) || players[0];
    }
    return null;
  }

  function setSeatTurnClass(el, player, turnPlayerId) {
    if (player && turnPlayerId && player.player_id === turnPlayerId) {
      el.classList.add('turn');
    } else {
      el.classList.remove('turn');
    }
  }

  function relativeSeatPlayers(players) {
    const me = effectiveMe(players);
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
    function setSeatContent(nameEl, player) {
      nameEl.textContent = player ? player.name : '-';
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

    setSeatContent(dom.seatTopNameEl, relativePlayers.top);
    setSeatContent(dom.seatLeftNameEl, relativePlayers.left);
    setSeatContent(dom.seatRightNameEl, relativePlayers.right);
    setSeatContent(dom.seatBottomNameEl, relativePlayers.me);

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
    const me = effectiveMe(players);
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

  function renderYourHand(cards, phase, isYourTurn, passSubmitted, passReceived, passReady, passSent) {
    const safeCards = Array.isArray(cards) ? cards : [];
    const safePassReceived = Array.isArray(passReceived) ? passReceived : [];
    const safePassSent = Array.isArray(passSent) ? passSent : [];
    const receivedSet = new Set(safePassReceived);
    const sentSet = new Set(safePassSent);
    state.selectedPassCards = state.selectedPassCards.filter((card) => safeCards.includes(card));
    const handRenderKey = `${phase}|${isYourTurn ? '1' : '0'}|${passSubmitted ? '1' : '0'}|${passReady ? '1' : '0'}|${safePassReceived.join(',')}|${safePassSent.join(',')}|${state.selectedPassCards.join(',')}|${safeCards.join(',')}`;
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
      if (phase === 'passing' && passSubmitted && sentSet.has(cardValue)) {
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
          renderYourHand(safeCards, phase, isYourTurn, passSubmitted, passReceived, passReady, passSent);
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

    // Compute live totals first so we can sort by them
    const liveTotalPoints = {};
    for (const player of players) {
      liveTotalPoints[player.player_id] = pointsFor(totalPoints, player.player_id) + pointsFor(roundPoints, player.player_id);
    }

    // Sort ascending by live total (fewest = winning = leftmost); stable for ties
    const sortedPlayers = [...players].sort((a, b) => liveTotalPoints[a.player_id] - liveTotalPoints[b.player_id]);

    // Capture pre-rebuild column positions for FLIP animation
    const prevPositions = {};
    if (dom.scoreboardHeadEl.firstElementChild) {
      for (const cell of dom.scoreboardHeadEl.firstElementChild.querySelectorAll('th[data-player-id]')) {
        prevPositions[cell.dataset.playerId] = cell.getBoundingClientRect().left;
      }
    }

    dom.scoreboardHeadEl.innerHTML = '';
    dom.scoreboardBodyEl.innerHTML = '';

    const headerRow = document.createElement('tr');
    const roundHeaderCell = document.createElement('th');
    roundHeaderCell.scope = 'col';
    roundHeaderCell.textContent = '';
    headerRow.appendChild(roundHeaderCell);

    if (sortedPlayers.length === 0) {
      const playersHeaderCell = document.createElement('th');
      playersHeaderCell.scope = 'col';
      playersHeaderCell.textContent = t('table.scoreboard.players');
      headerRow.appendChild(playersHeaderCell);

      dom.scoreboardHeadEl.appendChild(headerRow);

      const emptyRow = document.createElement('tr');
      const emptyLabel = document.createElement('th');
      emptyLabel.scope = 'row';
      emptyLabel.textContent = '►';
      const emptyValue = document.createElement('td');
      emptyValue.textContent = '-';
      emptyRow.appendChild(emptyLabel);
      emptyRow.appendChild(emptyValue);
      dom.scoreboardBodyEl.appendChild(emptyRow);
      return;
    }

    for (const player of sortedPlayers) {
      const playerHeaderCell = document.createElement('th');
      playerHeaderCell.scope = 'col';
      playerHeaderCell.dataset.playerId = player.player_id;
      const nameSpan = document.createElement('span');
      nameSpan.className = 'scoreboard-player-name';
      nameSpan.textContent = player.name;
      playerHeaderCell.appendChild(nameSpan);
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

      for (const player of sortedPlayers) {
        const valueCell = document.createElement('td');
        valueCell.dataset.playerId = player.player_id;
        valueCell.textContent = String(pointsFor(pointsByPlayer, player.player_id));
        row.appendChild(valueCell);
      }

      dom.scoreboardBodyEl.appendChild(row);
    }

    history.forEach((entry, index) => {
      appendPointsRow(`${index + 1}`, entry, 'scoreboard-history-row');
    });

    appendPointsRow('►', roundPoints, 'scoreboard-current-row');
    appendPointsRow('Σ', liveTotalPoints, 'scoreboard-total-row');

    // FLIP animation: slide columns from their old positions to new ones
    if (Object.keys(prevPositions).length > 0) {
      for (const cell of dom.scoreboardHeadEl.firstElementChild.querySelectorAll('th[data-player-id]')) {
        const pid = cell.dataset.playerId;
        if (prevPositions[pid] === undefined) continue;
        const delta = prevPositions[pid] - cell.getBoundingClientRect().left;
        if (Math.abs(delta) < 1) continue;

        const cells = document.querySelectorAll(`.scoreboard-table [data-player-id="${CSS.escape(pid)}"]`);
        for (const c of cells) {
          c.style.transition = 'none';
          c.style.transform = `translateX(${delta}px)`;
        }
        // Double rAF ensures the browser has painted the initial offset before transitioning
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            for (const c of cells) {
              const flipDuration = document.body.dataset.speed === 'fast' ? '0.2s' : '0.4s';
              c.style.transition = `transform ${flipDuration} ease`;
              c.style.transform = 'translateX(0)';
            }
          });
        });
      }
    }
  }

  function renderGameOverPanel(snapshot) {
    const players = Array.isArray(snapshot.players) ? snapshot.players : [];
    const totalPoints = (snapshot && snapshot.total_points) ? snapshot.total_points : {};
    const winners = Array.isArray(snapshot.winners) ? snapshot.winners : [];

    const sorted = [...players].sort((a, b) => (totalPoints[a.player_id] ?? 0) - (totalPoints[b.player_id] ?? 0));

    dom.gameOverScoresEl.innerHTML = '';
    const table = document.createElement('table');
    table.className = 'game-over-scores-table';
    let rank = 0;
    let prevScore = null;
    for (let i = 0; i < sorted.length; i++) {
      const p = sorted[i];
      const score = totalPoints[p.player_id] ?? 0;
      if (score !== prevScore) {
        rank = i;
        prevScore = score;
      }
      const tr = document.createElement('tr');
      if (winners.includes(p.player_id)) {
        tr.classList.add('game-over-winner-row');
      }
      const medalTd = document.createElement('td');
      medalTd.textContent = rank < MEDALS.length ? MEDALS[rank] : '';
      const nameTd = document.createElement('td');
      const swatch = document.createElement('span');
      swatch.className = 'game-over-swatch';
      swatch.style.backgroundColor = CHART_COLORS[i % CHART_COLORS.length];
      nameTd.appendChild(swatch);
      nameTd.appendChild(document.createTextNode(p.name));
      const scoreTd = document.createElement('td');
      scoreTd.textContent = String(score);
      tr.appendChild(medalTd);
      tr.appendChild(nameTd);
      tr.appendChild(scoreTd);
      table.appendChild(tr);
    }
    dom.gameOverScoresEl.appendChild(table);

    renderScoreChart(snapshot, players, winners, sorted);
  }

  function renderRematchStatus(snapshot) {
    const isPlayer = !!state.myPlayerId && !state.isObserver;
    const voted = !!snapshot.rematch_voted;
    const votes = snapshot.rematch_votes || 0;
    const total = snapshot.rematch_total || 0;

    dom.rematchButtonEl.hidden = !isPlayer;
    if (voted) {
      dom.rematchButtonEl.disabled = true;
      dom.rematchButtonEl.textContent = t('game.rematch_waiting', { votes, total });
    } else {
      dom.rematchButtonEl.disabled = false;
      dom.rematchButtonEl.textContent = t('game.play_again');
    }
  }

  function renderScoreChart(snapshot, players, winners, sortedPlayers) {
    const history = Array.isArray(snapshot.round_history) ? snapshot.round_history : [];
    if (history.length === 0 || typeof Chart === 'undefined') {
      dom.gameOverChartEl.hidden = true;
      return;
    }
    dom.gameOverChartEl.hidden = false;

    // Build cumulative scores per player
    const playerIds = sortedPlayers.map(p => p.player_id);
    const datasets = playerIds.map((pid, i) => {
      let cumulative = 0;
      const data = history.map(entry => {
        cumulative += (entry[pid] ?? 0);
        return cumulative;
      });
      const p = players.find(pl => pl.player_id === pid);
      const label = p ? p.name : pid;
      const isWinner = winners.includes(pid);
      return {
        label,
        data,
        borderColor: CHART_COLORS[i % CHART_COLORS.length],
        backgroundColor: CHART_COLORS[i % CHART_COLORS.length],
        borderWidth: isWinner ? 3 : 1.5,
        pointRadius: isWinner ? 3 : 1.5,
        tension: 0.15,
      };
    });

    const labels = history.map((_, i) => String(i + 1));

    if (scoreChart) {
      scoreChart.destroy();
    }

    const ctx = dom.gameOverCanvasEl.getContext('2d');
    const style = getComputedStyle(document.documentElement);
    const lineColor = style.getPropertyValue('--line').trim() || '#d4e1ec';
    const mutedColor = style.getPropertyValue('--muted').trim() || '#5b6f83';

    scoreChart = new Chart(ctx, {
      type: 'line',
      data: { labels, datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              title: (items) => items.length ? t('game.chart_round', { label: items[0].label }) : '',
            },
          },
        },
        scales: {
          x: {
            ticks: { color: mutedColor, font: { size: 10 } },
            grid: { color: lineColor },
          },
          y: {
            ticks: { color: mutedColor, font: { size: 10 } },
            grid: { color: lineColor },
            beginAtZero: true,
          },
        },
      },
    });
  }

  function renderPassPanel(snapshot) {
    if (state.isObserver || !!snapshot.paused) {
      dom.passSummaryEl.hidden = true;
      dom.passDetailsEl.hidden = true;
      dom.passSelectionEl.hidden = true;
      dom.submitPassEl.hidden = true;
      dom.readyAfterPassEl.hidden = true;
      return;
    }

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
      if (passSubmitted) {
        dom.passDetailsEl.textContent = t('game.pass_waiting');
        dom.passDetailsEl.hidden = false;
        dom.submitPassEl.hidden = true;
        dom.readyAfterPassEl.hidden = true;
      } else {
        const dir = snapshot.pass_direction;
        const label = dir ? t('game.pass_direction', { direction: dir.charAt(0).toUpperCase() + dir.slice(1) }) : t('game.pass_cards');
        dom.submitPassEl.textContent = label;
        dom.submitPassEl.hidden = false;
        dom.submitPassEl.disabled = state.selectedPassCards.length !== 3;
        dom.readyAfterPassEl.hidden = true;
      }
      return;
    }

    dom.submitPassEl.hidden = true;
    dom.readyAfterPassEl.hidden = false;
    dom.readyAfterPassEl.disabled = passReady;
  }

  function renderPausedPanel(snapshot) {
    const players = Array.isArray(snapshot.players) ? snapshot.players : [];
    const pid = snapshot.paused_for_player_id || '';
    const match = players.find((p) => p.player_id === pid);
    const who = match ? match.name : t('game.unknown_player');
    dom.gamePausedMessageEl.textContent = t('game.player_disconnected', { name: who });

    dom.gamePausedActionsEl.innerHTML = '';
    if (state.isObserver) {
      return;
    }

    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'felt-btn';
    btn.textContent = t('game.continue_with_bot');
    btn.onclick = () => {
      send({ type: 'resume_game' });
    };
    dom.gamePausedActionsEl.appendChild(btn);
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

    const isGameOver = !!snapshot.game_over;
    dom.gameOverOverlayEl.classList.toggle('hidden', !isGameOver);
    if (isGameOver && !gameOverRendered) {
      gameOverRendered = true;
      renderGameOverPanel(snapshot);
    }
    if (isGameOver) {
      renderRematchStatus(snapshot);
    } else if (gameOverRendered) {
      resetGameOver();
    }

    const isPaused = !!snapshot.paused;
    dom.gamePausedControlsEl.classList.toggle('hidden', !isPaused);
    if (isPaused) {
      renderPausedPanel(snapshot);
    }

    dom.trickSectionEl.hidden = false;

    const botSeats = state.isObserver ? players.filter((p) => p.is_bot) : [];
    const canClaimSeat = botSeats.length > 0 && !isGameOver;
    dom.claimSeatContainerEl.classList.toggle('hidden', !canClaimSeat);
    if (canClaimSeat) {
      dom.claimSeatPanelEl.replaceChildren();
      for (const bot of botSeats) {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'popover-menu-option';
        btn.dataset.seat = bot.seat;
        btn.textContent = bot.name;
        dom.claimSeatPanelEl.appendChild(btn);
      }
    } else {
      dom.claimSeatPanelEl.classList.add('hidden');
      dom.claimSeatToggleEl.setAttribute('aria-expanded', 'false');
    }

    const canAddBot = !state.isObserver && !snapshot.started && players.length < 4 && !isGameOver;
    dom.addBotContainerEl.classList.toggle('hidden', !canAddBot);
    if (!canAddBot) {
      dom.addBotPanelEl.classList.add('hidden');
      dom.addBotToggleEl.setAttribute('aria-expanded', 'false');
    }

    const queueIdle = !state.processingTrickEventQueue && state.trickEventQueue.length === 0;
    const showStartControl = !isPaused && !state.isObserver && !snapshot.started && players.length >= 1 && queueIdle && !isGameOver;
    const showPassControls = !isPaused && !state.isObserver && (isPassing || isPassReview);
    dom.centerControlsEl.classList.toggle('hidden', !showStartControl && !showPassControls && !isPaused);
    const showButtonRow = showStartControl || showPassControls;
    dom.startButtonEl.parentElement.hidden = !showButtonRow;
    dom.startButtonEl.hidden = !showStartControl;
    dom.startButtonEl.disabled = !showStartControl;
    const hasCompletedRound = Object.values(totalPoints).some((points) => Number(points) > 0);
    const seatsOpen = players.length < 4;
    if (hasCompletedRound) {
      dom.startButtonEl.textContent = seatsOpen ? t('game.continue_with_bots') : t('game.continue');
    } else {
      dom.startButtonEl.textContent = seatsOpen ? t('game.start_with_bots') : t('game.start');
    }

    const serverTrickPlays = (snapshot.trick_plays || []).slice();
    if (queueIdle) {
      state.liveTrickPlays = serverTrickPlays;
      state.liveTurnPlayerId = snapshot.turn_player_id;
    }

    state.lastPlayers = players;
    const relativePlayers = relativeSeatPlayers(players);
    const isYourTurn = !!isPlaying && !!state.myPlayerId && state.liveTurnPlayerId === state.myPlayerId;

    setSeatLabels(relativePlayers, state.liveTurnPlayerId);
    renderTrick(players, state.liveTrickPlays);
    if (state.isObserver) {
      const bottomPlayer = relativePlayers.me;
      const bottomCount = bottomPlayer ? ((snapshot.hand_sizes || {})[bottomPlayer.player_id] || 0) : 0;
      renderSeatBackHand(dom.seatBottomHandEl, bottomCount);
    } else {
      renderYourHand(
        snapshot.hand || [],
        phase,
        isYourTurn,
        !!snapshot.pass_submitted,
        snapshot.pass_received || [],
        !!snapshot.pass_ready,
        snapshot.pass_sent || []
      );
    }
    renderOtherHands(relativePlayers, snapshot.hand_sizes || {});
    renderPassPanel(snapshot);
    renderScoreboard(snapshot);
  }

  function seatElementForPlayerId(playerId) {
    if (!playerId || !state.lastPlayers || state.lastPlayers.length === 0) {
      return null;
    }

    const me = effectiveMe(state.lastPlayers);
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

    const me = effectiveMe(state.lastPlayers);
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
        clone.style.position = 'absolute';
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

      const fast = document.body.dataset.speed === 'fast';
      const durationMs = fast ? 700 : 1400;
      const staggerMs = fast ? 45 : 90;
      const sortPriorityBySlot = { top: 0, left: 1, right: 2, bottom: 3 };
      animatedCards.sort((a, b) => (sortPriorityBySlot[a.slotKey] || 0) - (sortPriorityBySlot[b.slotKey] || 0));

      // Force a synchronous reflow so all clones have an established initial
      // computed style before the RAF applies the transition targets. Without
      // this, the last-inserted clone (bottom) has no "from" state and jumps
      // instantly to opacity: 0 instead of transitioning.
      void dom.trickAnimationLayerEl.offsetHeight;

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

  function resetGameOver() {
    gameOverRendered = false;
    if (scoreChart) {
      scoreChart.destroy();
      scoreChart = null;
    }
  }

  return {
    renderState,
    renderTrick,
    animateTrickCardsToWinner,
    trickRenderPlayers,
    resetGameOver
  };
}
