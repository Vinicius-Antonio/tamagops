
const FACE_BITMAPS = {
  HAPPY: [
    ".............",
    ".............",
    ".............",
    "...#.....#...",
    "..#.#...#.#..",
    ".............",
    "....#...#....",
    ".....#.#.....",
    "......#......",
    ".............",
    ".............",
  ],
  NEUTRAL: [
    ".............",
    ".............",
    ".............",
    ".............",
    "..###...###..",
    ".............",
    ".............",
    ".....###.....",
    ".............",
    ".............",
    ".............",
  ],
  SICK: [
    ".............",
    ".............",
    ".............",
    "..#.#...#.#..",
    "...#.....#...",
    "..#.#...#.#..",
    ".............",
    "....#.#.#....",
    ".....#.#.....",
    ".............",
    ".............",
  ],
  DIRTY: [
    ".............",
    ".............",
    ".............",
    "..#.......#..",
    "....#...#....",
    "..#.......#..",
    "....#...#....",
    ".....###.....",
    ".............",
    ".............",
    ".............",
  ],
  SLEEPING: [
    ".............",
    ".............",
    ".............",
    ".............",
    "..###...###..",
    ".............",
    ".............",
    ".....#.#.....",
    "......#......",
    ".............",
    ".............",
  ],
};

const FACE_FALLBACK = FACE_BITMAPS.NEUTRAL;

const MAX_LOG_LINES = 60;

const el = {
  liveDot: document.getElementById("live-dot"),
  petFace: document.getElementById("pet-face"),
  petZzz: document.getElementById("pet-zzz"),
  petName: document.getElementById("pet-name"),
  petLevel: document.getElementById("pet-level"),
  petMood: document.getElementById("pet-mood"),
  hpFill: document.getElementById("hp-fill"),
  xpFill: document.getElementById("xp-fill"),
  debuffList: document.getElementById("debuff-list"),

  cpuFill: document.getElementById("cpu-fill"),
  cpuValue: document.getElementById("cpu-value"),
  cpuTempFill: document.getElementById("cputemp-fill"),
  cpuTempValue: document.getElementById("cputemp-value"),
  boardTempFill: document.getElementById("boardtemp-fill"),
  boardTempValue: document.getElementById("boardtemp-value"),
  ramFill: document.getElementById("ram-fill"),
  ramValue: document.getElementById("ram-value"),
  swapFill: document.getElementById("swap-fill"),
  swapValue: document.getElementById("swap-value"),
  diskFill: document.getElementById("disk-fill"),
  diskValue: document.getElementById("disk-value"),

  loadValue: document.getElementById("load-value"),
  uptimeValue: document.getElementById("uptime-value"),

  topProcessName: document.getElementById("top-process-name"),
  topProcessCpu: document.getElementById("top-process-cpu"),

  logList: document.getElementById("log-list"),
};

function buildFaceGrid() {
  const totalRows = FACE_BITMAPS.NEUTRAL.length;
  const totalCols = FACE_BITMAPS.NEUTRAL[0].length;

  for (let row = 0; row < totalRows; row++) {
    for (let col = 0; col < totalCols; col++) {
      const pixel = document.createElement("div");
      pixel.className = "lcd__pixel";
      pixel.dataset.row = String(row);
      pixel.dataset.col = String(col);
      el.petFace.appendChild(pixel);
    }
  }
}

function updateFace(mood) {
  const bitmap = FACE_BITMAPS[mood] || FACE_FALLBACK;
  const pixels = el.petFace.children;

  for (const pixel of pixels) {
    const row = Number(pixel.dataset.row);
    const col = Number(pixel.dataset.col);
    const shouldBeOn = bitmap[row][col] === "#";
    pixel.classList.toggle("lcd__pixel--on", shouldBeOn);
  }

  el.petZzz.classList.toggle("lcd__zzz--visible", mood === "SLEEPING");
}

function formatPercent(value) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "—%";
  }
  return `${value.toFixed(1)}%`;
}

function formatTemp(value, sensorsAvailable) {
  if (!sensorsAvailable) {
    return "n/d";
  }
  return `${value.toFixed(1)}°C`;
}

function formatBytes(bytes) {
  if (!bytes || bytes <= 0) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }
  return `${value.toFixed(1)} ${units[unitIndex]}`;
}

function formatUptime(totalSeconds) {
  if (!totalSeconds || totalSeconds <= 0) {
    return "—";
  }
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

function formatLoad(load1, load5, load15, available) {
  if (!available) {
    return "n/d";
  }
  return `${load1.toFixed(2)} / ${load5.toFixed(2)} / ${load15.toFixed(2)}`;
}

const DEBUFF_LABELS = {
  RAM_OVERLOAD: "RAM sobrecarregada",
  DISK_OVERLOAD: "Disco sobrecarregado",
  ZOMBIE_PROCESS_DETECTED: "Processo consumindo demais",
  OVERHEATING: "Superaquecimento",
};

function applyPayload(payload) {
  if (!payload) {
    return;
  }

  const pet = payload.pet || {};
  const hw = payload.hardware || {};

  updateFace(pet.mood);
  el.petName.textContent = pet.name || "—";
  el.petLevel.textContent = `Nv. ${pet.level ?? "—"}`;
  el.petMood.textContent = pet.mood || "—";

  const hpPercent = pet.max_hp ? (pet.hp / pet.max_hp) * 100 : 0;
  el.hpFill.style.width = `${Math.max(0, Math.min(100, hpPercent))}%`;

  const xpPercent = pet.next_level_xp ? (pet.current_xp / pet.next_level_xp) * 100 : 0;
  el.xpFill.style.width = `${Math.max(0, Math.min(100, xpPercent))}%`;

  el.debuffList.innerHTML = "";
  (pet.active_debuffs || []).forEach((code) => {
    const chip = document.createElement("span");
    chip.className = "debuff-chip";
    chip.textContent = DEBUFF_LABELS[code] || code;
    el.debuffList.appendChild(chip);
  });

  el.cpuFill.style.width = `${Math.max(0, Math.min(100, hw.cpu_percent || 0))}%`;
  el.cpuValue.textContent = formatPercent(hw.cpu_percent);

  const cpuTempPercent = hw.sensors_available ? Math.min(100, (hw.cpu_temp_celsius / 100) * 100) : 0;
  el.cpuTempFill.style.width = `${cpuTempPercent}%`;
  el.cpuTempValue.textContent = formatTemp(hw.cpu_temp_celsius, hw.sensors_available);

  const boardTempPercent = hw.sensors_available ? Math.min(100, (hw.board_temp_celsius / 100) * 100) : 0;
  el.boardTempFill.style.width = `${boardTempPercent}%`;
  el.boardTempValue.textContent = formatTemp(hw.board_temp_celsius, hw.sensors_available);

  el.ramFill.style.width = `${Math.max(0, Math.min(100, hw.ram_used_percent || 0))}%`;
  el.ramValue.textContent = formatPercent(hw.ram_used_percent);

  el.swapFill.style.width = `${Math.max(0, Math.min(100, hw.swap_used_percent || 0))}%`;
  el.swapValue.textContent = formatPercent(hw.swap_used_percent);

  el.diskFill.style.width = `${Math.max(0, Math.min(100, hw.disk_used_percent || 0))}%`;
  el.diskValue.textContent = formatPercent(hw.disk_used_percent);

  el.loadValue.textContent = formatLoad(
    hw.load_avg_1min,
    hw.load_avg_5min,
    hw.load_avg_15min,
    hw.load_available,
  );
  el.uptimeValue.textContent = formatUptime(hw.uptime_seconds);

  const top = hw.top_process || {};
  el.topProcessName.textContent = top.name || "—";
  el.topProcessCpu.textContent = formatPercent(top.cpu_percent);

  if (payload.recent_log) {
    appendLogLine(payload.timestamp, payload.recent_log);
  }

  el.liveDot.style.animationDuration = "2s";
}

function appendLogLine(timestampIso, message) {
  const line = document.createElement("div");
  line.className = "logfeed__line";

  const time = document.createElement("span");
  time.className = "logfeed__line-time";
  time.textContent = formatLogTime(timestampIso);

  line.appendChild(time);
  line.appendChild(document.createTextNode(message));

  el.logList.prepend(line);

  while (el.logList.children.length > MAX_LOG_LINES) {
    el.logList.removeChild(el.logList.lastElementChild);
  }
}

function formatLogTime(timestampIso) {
  if (!timestampIso) {
    return "--:--:--";
  }
  const date = new Date(timestampIso);
  return date.toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
}

function init() {
  buildFaceGrid();

  if (window.go && window.go.main && window.go.main.App) {
    window.go.main.App.GetSnapshot()
      .then((payload) => applyPayload(payload))
      .catch((err) => console.error("failed to fetch initial snapshot", err));
  }

  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("pet:update", (payload) => {
      applyPayload(payload);
    });
  }
}

document.addEventListener("DOMContentLoaded", init);
