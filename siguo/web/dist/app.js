const seatNames = ["北", "东", "南", "西"];
const seatPieceClasses = ["piece-red", "piece-green", "piece-yellow", "piece-blue"];
const rankNames = {
  0: "军棋", 1: "军旗", 2: "地雷", 3: "炸弹", 4: "工兵", 5: "排长", 6: "连长",
  7: "营长", 8: "团长", 9: "旅长", 10: "师长", 11: "军长", 12: "司令"
};
const cellClasses = ["off", "normal", "camp", "railroad", "hq", "frontline"];

const state = {
  name: localStorage.getItem("siguo.name") || "",
  code: localStorage.getItem("siguo.code") || "",
  token: localStorage.getItem("siguo.token") || "",
  seat: Number(localStorage.getItem("siguo.seat") || 0),
  host: false,
  room: null,
  view: null,
  ws: null,
  seq: 1,
  selected: null,
  sound: localStorage.getItem("siguo.sound") !== "off",
  audio: null,
  combat: null,
  log: [],
  chat: []
};

const app = document.querySelector("#app");

function render() {
  app.innerHTML = `
    <div class="shell">
      <section class="table">
        <div class="topbar">
          <div class="brand"><h1>四国军棋</h1><span>${statusText()}</span></div>
          <div class="row" style="max-width:390px">
            <button class="primary" id="startBtn" ${canStart() ? "" : "disabled"}>开始</button>
            <button id="randBtn" ${state.room?.phase === "setup" ? "" : "disabled"}>随机</button>
            <button id="submitBtn" ${state.room?.phase === "setup" ? "" : "disabled"}>提交</button>
            <button id="soundBtn" class="toggle ${state.sound ? "on" : "off"}">声音</button>
          </div>
        </div>
        <div class="board-wrap">${boardHTML()}</div>
      </section>
      <aside class="side">
        <div class="panel stack">
          <input id="name" placeholder="昵称" value="${esc(state.name)}" />
          <div class="row">
            <button id="create">创建房间</button>
            <input id="code" placeholder="房间码" value="${esc(state.code)}" />
            <button id="join">加入</button>
          </div>
          <div class="subtle">本机地址可直接分享给同一局域网玩家。你的座位：${seatNames[state.seat] || "-"}</div>
        </div>
        <div class="panel">
          <div class="seats">${seatsHTML()}</div>
        </div>
        <div class="panel stack">
          <b>战况</b>
          <div class="log">${state.log.slice(-80).map(x => `<div class="line">${esc(x)}</div>`).join("")}</div>
        </div>
        <div class="panel stack">
          <b>聊天</b>
          <div class="chat-log">${state.chat.slice(-80).map(chatLine).join("")}</div>
          <div class="row">
            <input id="chatText" maxlength="200" placeholder="输入消息" />
            <button id="sendAll">公屏</button>
            <button id="sendTeam">队伍</button>
          </div>
        </div>
      </aside>
    </div>
  `;
  bind();
}

function statusText() {
  if (!state.room) return "创建或加入一个 6 位房间码";
  const phase = {lobby: "大厅", setup: "布阵", playing: "对局", ended: "结束"}[state.room.phase] || state.room.phase;
  const turn = state.room.phase === "playing" ? ` · ${seatNames[state.room.turn]}方行动` : "";
  return `房间 ${state.code} · ${phase}${turn}`;
}

function canStart() {
  return state.room?.phase === "lobby" && state.host && state.room.seats?.every(s => s.name);
}

function seatsHTML() {
  const seats = state.room?.seats || [0,1,2,3].map(seat => ({seat}));
  return seats.map(s => `
    <div class="seat ${state.room?.turn === s.seat ? "current" : ""}">
      <b>${seatNames[s.seat]}</b>
      <span class="subtle">${s.name ? esc(s.name) : "空位"} ${s.host ? "房主" : ""} ${s.ready ? "已提交" : ""}</span>
    </div>
  `).join("");
}

function boardHTML() {
  const cells = new Map((state.view?.Cells || state.view?.cells || []).map(c => [`${c.Pos?.Row ?? c.pos.row},${c.Pos?.Col ?? c.pos.col}`, c]));
  let html = `<div class="board-stage">${deadTrayHTML()}<div class="board">${railOverlayHTML()}`;
  for (let displayRow = 0; displayRow < 17; displayRow++) {
    for (let displayCol = 0; displayCol < 17; displayCol++) {
      const {row, col} = fromDisplay(displayRow, displayCol);
      const c = cells.get(`${row},${col}`);
      const piece = c?.Piece || c?.piece;
      const type = c?.Type ?? c?.type ?? 0;
      const selected = state.selected && state.selected.row === row && state.selected.col === col;
      const combat = state.combat && state.combat.row === row && state.combat.col === col && Date.now() < state.combat.until;
      html += `<div class="cell ${cellClasses[type] || "off"} ${homeClass(row, col)} ${visualTrackClass(row, col)} ${centralNodeClass(row, col)} ${turnFrontClass(row, col)} ${piece ? "occupied" : ""} ${selected ? "selected" : ""} ${combat ? "combat-hit" : ""}" data-row="${row}" data-col="${col}">`;
      if (piece) {
        const owner = piece.Owner ?? piece.owner;
        const rank = piece.Rank ?? piece.rank;
        const exposed = piece.Exposed ?? piece.exposed;
        const colorClass = seatPieceClasses[owner] || "piece-blue";
        const exposedClass = exposed && rank === 1 ? "flag-exposed" : "";
        const hiddenClass = rank === 0 ? "piece-hidden" : "";
        const orientClass = `piece-rel-${relativeSeat(owner)}`;
        const label = rank === 0 ? "" : (rankNames[rank] || "?");
        html += `<div class="piece ${colorClass} ${hiddenClass} ${orientClass} ${exposedClass}" data-piece="${piece.ID ?? piece.id}" data-owner="${owner}">${label}</div>`;
      }
      const riverLabel = centralRiverLabel(displayRow, displayCol);
      if (riverLabel) {
        html += `<div class="river-label ${riverLabel.className}" aria-hidden="true">${riverLabel.text.split("").join("<br>")}</div>`;
      }
      html += `</div>`;
    }
  }
  return html + `</div></div>`;
}

function deadTrayHTML() {
  const dead = state.view?.DeadPieces || state.view?.deadPieces || [];
  const items = dead.map(piece => {
    const owner = piece.Owner ?? piece.owner;
    const rank = piece.Rank ?? piece.rank;
    const colorClass = seatPieceClasses[owner] || "piece-blue";
    return `<div class="dead-piece ${colorClass}">${rankNames[rank] || "?"}</div>`;
  }).join("");
  return `<div class="dead-tray" aria-label="阵亡棋子"><b>阵亡</b><div class="dead-list">${items || `<span class="subtle">无</span>`}</div></div>`;
}

function railOverlayHTML() {
  const line = (x1, y1, x2, y2) => `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" />`;
  const parts = [];
  const sleepers = [];

  const addVertical = (x, y1, y2) => {
    parts.push(line(x, y1, x, y2));
    for (let y = Math.ceil(y1); y <= Math.floor(y2); y += 2) {
      sleepers.push(line(x - .22, y + .5, x + .22, y + .5));
    }
  };
  const addHorizontal = (y, x1, x2) => {
    parts.push(line(x1, y, x2, y));
    for (let x = Math.ceil(x1); x <= Math.floor(x2); x += 2) {
      sleepers.push(line(x + .5, y - .22, x + .5, y + .22));
    }
  };

  addHorizontal(1.5, 6.5, 10.5);
  addHorizontal(5.5, 6.5, 10.5);
  addVertical(6.5, 1.5, 5.5);
  addVertical(10.5, 1.5, 5.5);

  addHorizontal(11.5, 6.5, 10.5);
  addHorizontal(15.5, 6.5, 10.5);
  addVertical(6.5, 11.5, 15.5);
  addVertical(10.5, 11.5, 15.5);

  addVertical(1.5, 6.5, 10.5);
  addVertical(5.5, 6.5, 10.5);
  addHorizontal(6.5, 1.5, 5.5);
  addHorizontal(10.5, 1.5, 5.5);

  addVertical(11.5, 6.5, 10.5);
  addVertical(15.5, 6.5, 10.5);
  addHorizontal(6.5, 11.5, 15.5);
  addHorizontal(10.5, 11.5, 15.5);

  addVertical(6.5, 6.5, 10.5);
  addVertical(8.5, 6.5, 10.5);
  addVertical(10.5, 6.5, 10.5);
  addHorizontal(6.5, 6.5, 10.5);
  addHorizontal(8.5, 6.5, 10.5);
  addHorizontal(10.5, 6.5, 10.5);

  addVertical(8.5, 5.5, 6.5);
  addVertical(8.5, 10.5, 11.5);
  addHorizontal(8.5, 5.5, 6.5);
  addHorizontal(8.5, 10.5, 11.5);
  addHorizontal(6.5, 5.5, 6.5);
  addHorizontal(10.5, 5.5, 6.5);
  addHorizontal(6.5, 10.5, 11.5);
  addVertical(6.5, 5.5, 6.5);
  addVertical(10.5, 5.5, 6.5);
  addVertical(6.5, 10.5, 11.5);
  addVertical(10.5, 10.5, 11.5);
  addHorizontal(10.5, 10.5, 11.5);

  parts.push(line(6.5, 5.5, 5.5, 6.5));
  parts.push(line(10.5, 5.5, 11.5, 6.5));
  parts.push(line(6.5, 11.5, 5.5, 10.5));
  parts.push(line(10.5, 11.5, 11.5, 10.5));

  return `<svg class="rails" viewBox="0 0 17 17" aria-hidden="true">
    <g class="rail-shadow">${parts.join("")}</g>
    <g class="rail-main">${parts.join("")}</g>
    <g class="rail-sleeper">${sleepers.join("")}</g>
  </svg>`;
}

function homeClass(row, col) {
  if (row <= 5 && col >= 6 && col <= 10) return "home-north";
  if (row >= 11 && col >= 6 && col <= 10) return "home-south";
  if (col <= 5 && row >= 6 && row <= 10) return "home-west";
  if (col >= 11 && row >= 6 && row <= 10) return "home-east";
  if (row >= 6 && row <= 10 && col >= 6 && col <= 10) return "home-center";
  return "";
}

function turnFrontClass(row, col) {
  if (state.room?.phase !== "playing") return "";
  const turn = state.room.turn;
  if (turn === 0 && row === 6 && col >= 6 && col <= 10) return "turn-front";
  if (turn === 1 && col === 10 && row >= 6 && row <= 10) return "turn-front";
  if (turn === 2 && row === 10 && col >= 6 && col <= 10) return "turn-front";
  if (turn === 3 && col === 6 && row >= 6 && row <= 10) return "turn-front";
  return "";
}

function visualTrackClass(row, col) {
  const inCenter = row >= 6 && row <= 10 && col >= 6 && col <= 10;
  const onTrack = [6, 8, 10].includes(row) || [6, 8, 10].includes(col);
  const isNode = [6, 8, 10].includes(row) && [6, 8, 10].includes(col);
  return inCenter && onTrack && !isNode ? "track-segment" : "";
}

function centralNodeClass(row, col) {
  return [6, 8, 10].includes(row) && [6, 8, 10].includes(col) ? "central-node" : "";
}

function centralRiverLabel(displayRow, displayCol) {
  const labels = {
    "7,7": {text: "山界", className: "river-label-left river-label-top"},
    "7,9": {text: "楚河", className: "river-label-right river-label-top"},
    "9,7": {text: "汉界", className: "river-label-left river-label-bottom"},
    "9,9": {text: "山界", className: "river-label-right river-label-bottom"},
  };
  return labels[`${displayRow},${displayCol}`] || null;
}

function relativeSeat(owner) {
  return (owner - state.seat + 4) % 4;
}

function toDisplay(row, col) {
  switch (state.seat) {
    case 0: return {row: 16 - row, col: 16 - col};
    case 1: return {row: col, col: 16 - row};
    case 2: return {row, col};
    case 3: return {row: 16 - col, col: row};
    default: return {row, col};
  }
}

function fromDisplay(row, col) {
  switch (state.seat) {
    case 0: return {row: 16 - row, col: 16 - col};
    case 1: return {row: 16 - col, col: row};
    case 2: return {row, col};
    case 3: return {row: col, col: 16 - row};
    default: return {row, col};
  }
}

function bind() {
  document.querySelector("#create").onclick = createRoom;
  document.querySelector("#join").onclick = joinRoom;
  document.querySelector("#startBtn").onclick = () => send({type:"room.start"});
  document.querySelector("#randBtn").onclick = () => send({type:"setup.randomize"});
  document.querySelector("#submitBtn").onclick = () => send({type:"setup.submit"});
  document.querySelector("#soundBtn").onclick = toggleSound;
  document.querySelector("#sendAll").onclick = () => sendChat("all");
  document.querySelector("#sendTeam").onclick = () => sendChat("team");
  document.querySelector("#name").oninput = e => {
    state.name = e.target.value;
    localStorage.setItem("siguo.name", state.name);
  };
  document.querySelector("#code").oninput = e => state.code = e.target.value.toUpperCase();
  document.querySelectorAll(".cell").forEach(cell => cell.onclick = () => clickCell(cell));
}

async function createRoom() {
  state.name = document.querySelector("#name").value || "玩家";
  const res = await fetch("/api/rooms", {method:"POST", body: JSON.stringify({name: state.name})});
  await acceptJoin(res);
}

async function joinRoom() {
  state.name = document.querySelector("#name").value || "玩家";
  state.code = document.querySelector("#code").value.toUpperCase();
  const res = await fetch(`/api/rooms/${state.code}/join`, {method:"POST", body: JSON.stringify({name: state.name, sessionToken: state.token})});
  await acceptJoin(res);
}

async function acceptJoin(res) {
  if (!res.ok) {
    log(`错误：${await res.text()}`);
    return;
  }
  const data = await res.json();
  state.code = data.code;
  state.token = data.sessionToken;
  state.seat = data.seat;
  state.host = data.host;
  localStorage.setItem("siguo.code", state.code);
  localStorage.setItem("siguo.token", state.token);
  localStorage.setItem("siguo.seat", String(state.seat));
  connect();
}

function connect() {
  if (state.ws) state.ws.close();
  const proto = location.protocol === "https:" ? "wss" : "ws";
  state.ws = new WebSocket(`${proto}://${location.host}/ws?room=${state.code}&token=${encodeURIComponent(state.token)}`);
  state.ws.onopen = () => log("已连接");
  state.ws.onclose = () => log("连接已断开");
  state.ws.onmessage = e => onMessage(JSON.parse(e.data));
}

function onMessage(msg) {
  if (msg.type === "room.state") {
    state.room = msg.room;
    state.view = msg.room?.view || state.view;
  } else if (msg.type === "view") {
    state.view = msg.view;
  } else if (msg.type === "chat.msg") {
    state.chat.push(msg.chat);
  } else if (msg.type === "error") {
    log(`错误：${msg.error?.message || msg.notice}`);
  } else if (msg.event) {
    handleEventEffect(msg.event);
    log(eventText(msg.event));
  }
  render();
}

function clickCell(cell) {
  ensureAudio();
  const row = Number(cell.dataset.row);
  const col = Number(cell.dataset.col);
  const pieceEl = cell.querySelector(".piece");
  const pieceId = pieceEl ? Number(pieceEl.dataset.piece) : 0;
  const owner = pieceEl ? Number(pieceEl.dataset.owner) : -1;
  if (!state.room) return;
  if (state.room.phase === "setup" && state.selected && pieceId && owner === state.seat) {
    if (state.selected.row === row && state.selected.col === col) {
      state.selected = null;
      render();
      return;
    }
    send({type:"setup.place", pieceId: state.selected.pieceId, row, col});
    state.selected = null;
    return;
  }
  if (pieceId && owner === state.seat) {
    if (state.room.phase === "playing" && state.room.turn !== state.seat) {
      log(`现在是${seatNames[state.room.turn]}方行动`);
      return;
    }
    state.selected = {pieceId, row, col};
    render();
    return;
  }
  if (!state.selected) return;
  if (state.room.phase === "setup") {
    send({type:"setup.place", pieceId: state.selected.pieceId, row, col});
  } else if (state.room.phase === "playing") {
    send({type:"move", pieceId: state.selected.pieceId, from:{Row:state.selected.row, Col:state.selected.col}, to:{Row:row, Col:col}});
  } else {
    log("当前不能走棋");
  }
  state.selected = null;
}

function send(msg) {
  ensureAudio();
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN) {
    log("尚未连接");
    return;
  }
  msg.seq = state.seq++;
  state.ws.send(JSON.stringify(msg));
}

function sendChat(channel) {
  const input = document.querySelector("#chatText");
  send({type:"chat.send", channel, text: input.value});
  input.value = "";
}

function eventText(ev) {
  const type = ev.Type || ev.type;
  const from = ev.From || ev.from;
  const to = ev.To || ev.to;
  if (type === "move") return `移动：${coord(from)} -> ${coord(to)}`;
  if (type === "combat") return `战斗：${coord(from)} -> ${coord(to)}，${ev.Outcome || ev.outcome}`;
  if (type === "flagCaptured") return `${seatNames[ev.Owner ?? ev.owner]}方军旗被夺`;
  if (type === "gameEnded") return "对局结束";
  return "局面更新";
}

function handleEventEffect(ev) {
  const type = ev.Type || ev.type;
  const to = ev.To || ev.to;
  if (type === "combat") {
    const row = to?.Row ?? to?.row;
    const col = to?.Col ?? to?.col;
    state.combat = {row, col, until: Date.now() + 780};
    playCombat();
    setTimeout(() => {
      state.combat = null;
      render();
    }, 820);
  } else if (type === "move") {
    playMove();
  } else if (type === "flagCaptured") {
    playFlag();
  }
}

function toggleSound() {
  state.sound = !state.sound;
  localStorage.setItem("siguo.sound", state.sound ? "on" : "off");
  ensureAudio();
  if (state.sound) playMove();
  render();
}

function ensureAudio() {
  if (!state.sound || state.audio) return;
  const AudioCtx = window.AudioContext || window.webkitAudioContext;
  if (!AudioCtx) return;
  state.audio = new AudioCtx();
  if (state.audio.state === "suspended") state.audio.resume();
}

function tone(freq, start, duration, type = "sine", gain = 0.08) {
  if (!state.sound) return;
  ensureAudio();
  if (!state.audio) return;
  const t = state.audio.currentTime + start;
  const osc = state.audio.createOscillator();
  const amp = state.audio.createGain();
  osc.type = type;
  osc.frequency.setValueAtTime(freq, t);
  amp.gain.setValueAtTime(0.0001, t);
  amp.gain.exponentialRampToValueAtTime(gain, t + 0.015);
  amp.gain.exponentialRampToValueAtTime(0.0001, t + duration);
  osc.connect(amp).connect(state.audio.destination);
  osc.start(t);
  osc.stop(t + duration + 0.02);
}

function playMove() {
  tone(420, 0, 0.09, "triangle", 0.04);
  tone(260, 0.035, 0.08, "triangle", 0.025);
}

function playCombat() {
  tone(90, 0, 0.22, "sawtooth", 0.09);
  tone(155, 0.035, 0.18, "square", 0.05);
  tone(620, 0.11, 0.08, "triangle", 0.035);
}

function playFlag() {
  tone(220, 0, 0.18, "triangle", 0.06);
  tone(330, 0.12, 0.22, "triangle", 0.06);
  tone(440, 0.26, 0.28, "triangle", 0.06);
}

function coord(p) {
  if (!p) return "";
  return `${p.Row ?? p.row},${p.Col ?? p.col}`;
}

function chatLine(c) {
  return `<div class="line">[${c.channel === "team" ? "队伍" : "公屏"}] ${seatNames[c.from]} ${esc(c.name)}：${esc(c.text || c.emote || "")}</div>`;
}

function log(line) {
  state.log.push(line);
  render();
}

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));
}

render();
if (state.code && state.token) connect();
