export function createTableDom({ tableId, eventsEnabled }) {
  const titleEl = document.getElementById('title');
  const statusEl = document.getElementById('status');
  const observerBadgeEl = document.getElementById('observerBadge');
  const turnIndicatorEl = document.getElementById('turnIndicator');
  const botControlEl = document.getElementById('botControl');
  const addBotDefaultEl = document.getElementById('addBotDefault');
  const botStrategySelectEl = document.getElementById('botStrategySelect');
  const eventsSectionEl = document.getElementById('eventsSection');
  const logsEl = document.getElementById('logs');
  const startButtonEl = document.getElementById('start');
  const centerControlsEl = document.getElementById('centerControls');
  const passSummaryEl = document.getElementById('passSummary');
  const passDetailsEl = document.getElementById('passDetails');
  const passSelectionEl = document.getElementById('passSelection');
  const submitPassEl = document.getElementById('submitPass');
  const readyAfterPassEl = document.getElementById('readyAfterPass');
  const trickSectionEl = document.getElementById('trickSection');
  const scoreboardSectionEl = document.getElementById('scoreboardSection');
  const scoreboardHeadEl = document.getElementById('scoreboardHead');
  const scoreboardBodyEl = document.getElementById('scoreboardBody');
  const trickCenterEl = document.getElementById('trickCenter');
  const trickAnimationLayerEl = document.getElementById('trickAnimationLayer');
  const seatTopEl = document.getElementById('seatTop');
  const seatTopNameEl = document.getElementById('seatTopName');
  const seatTopHandEl = document.getElementById('seatTopHand');
  const seatLeftEl = document.getElementById('seatLeft');
  const seatLeftNameEl = document.getElementById('seatLeftName');
  const seatLeftHandEl = document.getElementById('seatLeftHand');
  const seatRightEl = document.getElementById('seatRight');
  const seatRightNameEl = document.getElementById('seatRightName');
  const seatRightHandEl = document.getElementById('seatRightHand');
  const seatBottomEl = document.getElementById('seatBottom');
  const seatBottomNameEl = document.getElementById('seatBottomName');
  const seatBottomHandEl = document.getElementById('seatBottomHand');
  const trickSlotEls = createTrickSlots(trickCenterEl);
  const gameOverOverlayEl = document.getElementById('gameOverOverlay');
  const gameOverWinnerEl = document.getElementById('gameOverWinner');
  const gameOverScoresEl = document.getElementById('gameOverScores');

  eventsSectionEl.hidden = !eventsEnabled;
  titleEl.textContent = `Hearts Table ${tableId}`;

  return {
    titleEl,
    statusEl,
    observerBadgeEl,
    turnIndicatorEl,
    botControlEl,
    addBotDefaultEl,
    botStrategySelectEl,
    eventsSectionEl,
    logsEl,
    startButtonEl,
    centerControlsEl,
    passSummaryEl,
    passDetailsEl,
    passSelectionEl,
    submitPassEl,
    readyAfterPassEl,
    trickSectionEl,
    scoreboardSectionEl,
    scoreboardHeadEl,
    scoreboardBodyEl,
    trickCenterEl,
    trickAnimationLayerEl,
    seatTopEl,
    seatTopNameEl,
    seatTopHandEl,
    seatLeftEl,
    seatLeftNameEl,
    seatLeftHandEl,
    seatRightEl,
    seatRightNameEl,
    seatRightHandEl,
    seatBottomEl,
    seatBottomNameEl,
    seatBottomHandEl,
    trickSlotEls,
    gameOverOverlayEl,
    gameOverWinnerEl,
    gameOverScoresEl
  };
}

function createTrickSlots(trickCenterEl) {
  const slots = {};
  const slotSpecs = [
    { key: 'top', className: 'trick-slot trick-slot-top' },
    { key: 'left', className: 'trick-slot trick-slot-left' },
    { key: 'right', className: 'trick-slot trick-slot-right' },
    { key: 'bottom', className: 'trick-slot trick-slot-bottom' }
  ];

  for (const spec of slotSpecs) {
    const slot = document.createElement('div');
    slot.className = spec.className;
    slot.dataset.slotKey = spec.key;
    trickCenterEl.appendChild(slot);
    slots[spec.key] = slot;
  }

  return slots;
}
