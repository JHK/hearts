let ctx = null;

// Create and warm up the AudioContext on first user gesture so it's ready
// when sounds are triggered from WebSocket callbacks (outside gesture window).
document.addEventListener('click', () => {
  if (!ctx) {
    ctx = new AudioContext();
  } else if (ctx.state === 'suspended') {
    ctx.resume();
  }
}, { once: true });

function audioContext() {
  if (!ctx) {
    ctx = new AudioContext();
  }
  return ctx;
}

// Descending sparkle — like a glass heart shattering
export function playHeartsBreaking() {
  const ac = audioContext();
  const t = ac.currentTime;
  [1047, 880, 698].forEach((freq, i) => {
    const osc = ac.createOscillator();
    const gain = ac.createGain();
    osc.connect(gain);
    gain.connect(ac.destination);
    osc.type = 'sine';
    const start = t + i * 0.1;
    osc.frequency.setValueAtTime(freq, start);
    osc.frequency.linearRampToValueAtTime(freq * 0.65, start + 0.4);
    gain.gain.setValueAtTime(0, start);
    gain.gain.linearRampToValueAtTime(0.22, start + 0.008);
    gain.gain.exponentialRampToValueAtTime(0.001, start + 0.4);
    osc.start(start);
    osc.stop(start + 0.45);
  });
}

// Bass drum thud — deep punch with quick decay
export function playQueenOfSpades() {
  const ac = audioContext();
  const t = ac.currentTime;

  // Body: sine pitched very low, fast pitch drop (classic kick drum model)
  const osc = ac.createOscillator();
  const gain = ac.createGain();
  osc.connect(gain);
  gain.connect(ac.destination);
  osc.type = 'sine';
  osc.frequency.setValueAtTime(120, t);
  osc.frequency.exponentialRampToValueAtTime(30, t + 0.15);
  gain.gain.setValueAtTime(0.9, t);
  gain.gain.exponentialRampToValueAtTime(0.001, t + 0.9);
  osc.start(t);
  osc.stop(t + 0.95);

  // Click/transient: short noise burst for the attack
  const bufferSize = ac.sampleRate * 0.04;
  const buffer = ac.createBuffer(1, bufferSize, ac.sampleRate);
  const data = buffer.getChannelData(0);
  for (let i = 0; i < bufferSize; i++) {
    data[i] = (Math.random() * 2 - 1);
  }
  const noise = ac.createBufferSource();
  noise.buffer = buffer;
  const noiseGain = ac.createGain();
  noise.connect(noiseGain);
  noiseGain.connect(ac.destination);
  noiseGain.gain.setValueAtTime(0.3, t);
  noiseGain.gain.exponentialRampToValueAtTime(0.001, t + 0.04);
  noise.start(t);
  noise.stop(t + 0.05);
}
