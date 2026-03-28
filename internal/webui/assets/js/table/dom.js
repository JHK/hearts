export function createTableDom() {
  const claimSeatContainerEl = document.getElementById('claimSeatContainer');
  const claimSeatToggleEl = document.getElementById('claimSeatToggle');
  const claimSeatPanelEl = document.getElementById('claimSeatPanel');
  const addBotContainerEl = document.getElementById('addBotContainer');
  const addBotToggleEl = document.getElementById('addBotToggle');
  const addBotPanelEl = document.getElementById('addBotPanel');
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
  const gamePausedControlsEl = document.getElementById('gamePausedControls');
  const gamePausedMessageEl = document.getElementById('gamePausedMessage');
  const gamePausedActionsEl = document.getElementById('gamePausedActions');
  const gameOverOverlayEl = document.getElementById('gameOverOverlay');
  const gameOverChartEl = document.getElementById('gameOverChart');
  const gameOverCanvasEl = document.getElementById('gameOverCanvas');
  const gameOverScoresEl = document.getElementById('gameOverScores');
  const rematchButtonEl = document.getElementById('rematchButton');
  const settingsToggleEl = document.getElementById('settingsToggle');
  const settingsPanelEl = document.getElementById('settingsPanel');
  const nameInputEl = document.getElementById('nameInput');
  const speedToggleEl = document.getElementById('speedToggle');
  const soundToggleEl = document.getElementById('soundToggle');
  const notifyToggleEl = document.getElementById('notifyToggle');

  return {
    claimSeatContainerEl,
    claimSeatToggleEl,
    claimSeatPanelEl,
    addBotContainerEl,
    addBotToggleEl,
    addBotPanelEl,
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
    gamePausedControlsEl,
    gamePausedMessageEl,
    gamePausedActionsEl,
    gameOverOverlayEl,
    gameOverChartEl,
    gameOverCanvasEl,
    gameOverScoresEl,
    rematchButtonEl,
    settingsToggleEl,
    settingsPanelEl,
    nameInputEl,
    speedToggleEl,
    soundToggleEl,
    notifyToggleEl
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
