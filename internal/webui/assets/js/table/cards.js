const rankNameByCode = {
  A: 'ace',
  K: 'king',
  Q: 'queen',
  J: 'jack',
  T: '10',
  '9': '9',
  '8': '8',
  '7': '7',
  '6': '6',
  '5': '5',
  '4': '4',
  '3': '3',
  '2': '2'
};

const suitNameByCode = {
  S: 'spades',
  H: 'hearts',
  D: 'diamonds',
  C: 'clubs'
};

function cardParts(cardValue) {
  if (!cardValue || cardValue.length < 2) {
    return { rank: '', suit: '' };
  }

  return {
    rank: cardValue.slice(0, -1),
    suit: cardValue.slice(-1)
  };
}

export function cardImageURL(cardValue) {
  const parts = cardParts(cardValue);
  const rank = rankNameByCode[parts.rank];
  const suit = suitNameByCode[parts.suit];
  if (!rank || !suit) {
    return '';
  }

  return `/assets/cards/${rank}_of_${suit}.svg`;
}

export function cardAltText(cardValue) {
  const parts = cardParts(cardValue);
  const rank = rankNameByCode[parts.rank] || parts.rank;
  const suit = suitNameByCode[parts.suit] || parts.suit;
  if (!rank || !suit) {
    return cardValue;
  }

  return `${rank} of ${suit}`;
}
