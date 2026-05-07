const seatNames = ["北", "东", "南", "西"];
const seatPieceClasses = ["piece-red", "piece-green", "piece-yellow", "piece-blue"];
const modeNames = {siguo: "四国 2v2", junqi: "军棋 1v1"};
const rankNames = {
  0: "军棋", 1: "军旗", 2: "地雷", 3: "炸弹", 4: "工兵", 5: "排长", 6: "连长",
  7: "营长", 8: "团长", 9: "旅长", 10: "师长", 11: "军长", 12: "司令"
};
const cellClasses = ["off", "normal", "camp", "railroad", "hq", "frontline"];

const state = {
  name: localStorage.getItem("siguo.name") || "",
  mode: localStorage.getItem("siguo.mode") || "siguo",
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
  chat: [],
  lowTimeWarned: false
};

const app = document.querySelector("#app");

function render() {
  app.innerHTML = `
    <div class="shell">
      <section class="table">
        <div class="topbar">
          <div class="brand"><h1>四国军棋</h1><span>${statusText()}</span></div>
          <div class="row" style="max-width:520px">
            <button class="primary" id="startBtn" ${canStart() ? "" : "disabled"}>开始</button>
            <button id="randBtn" ${state.room?.phase === "setup" ? "" : "disabled"}>随机</button>
            <button id="submitBtn" ${state.room?.phase === "setup" ? "" : "disabled"}>提交</button>
            ${actionButtonsHTML()}
            <button id="soundBtn" class="toggle ${state.sound ? "on" : "off"}">声音</button>
          </div>
        </div>
        ${requestBannerHTML()}
        <div class="board-wrap">${boardHTML()}</div>
      </section>
      <aside class="side">
        <div class="panel stack">
          <input id="name" placeholder="昵称" value="${esc(state.name)}" />
          ${modeSwitchHTML()}
          <div class="row">
            <button id="create">创建房间</button>
            <input id="code" placeholder="房间码" value="${esc(state.code)}" />
            <button id="join">加入</button>
          </div>
          <div class="subtle">本机地址可直接分享给同一局域网玩家。模式：${modeNames[currentMode()]}。你的座位：${seatNames[state.seat] || "-"}</div>
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
            ${currentMode() === "junqi" ? "" : `<button id="sendTeam">队伍</button>`}
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
  return `房间 ${state.code} · ${modeNames[currentMode()]} · ${phase}${turn}`;
}

function canStart() {
  const seats = activeSeats();
  const occupied = state.room?.seats?.filter(s => seats.includes(s.seat) && s.name) || [];
  return state.room?.phase === "lobby" && state.host && occupied.length === seats.length;
}

function seatsHTML() {
  const seats = state.room?.seats || activeSeats().map(seat => ({seat}));
  return seats.map(s => `
    <div class="seat ${state.room?.turn === s.seat ? "current" : ""}">
      <b>${seatNames[s.seat]}</b>
      <span class="subtle">${s.name ? esc(s.name) : "空位"} ${s.host ? "房主" : ""} ${s.ready ? "已提交" : ""}</span>
    </div>
  `).join("");
}

function modeSwitchHTML() {
  const locked = state.room && state.room.phase !== "lobby";
  const mode = currentMode();
  return `<div class="mode-switch" role="group" aria-label="模式">
    <button id="modeSiguo" class="${mode === "siguo" ? "on" : ""}" ${locked ? "disabled" : ""}>四国 2v2</button>
    <button id="modeJunqi" class="${mode === "junqi" ? "on" : ""}" ${locked ? "disabled" : ""}>军棋 1v1</button>
  </div>`;
}

function currentMode() {
  return state.room?.mode || state.mode || "siguo";
}

function activeSeats() {
  return currentMode() === "junqi" ? [0, 2] : [0, 1, 2, 3];
}

function boardHTML() {
  const cells = new Map((state.view?.Cells || state.view?.cells || []).map(c => [`${c.Pos?.Row ?? c.pos.row},${c.Pos?.Col ?? c.pos.col}`, c]));
  const mode = currentMode();
  const rows = mode === "junqi" ? 13 : 17;
  const cols = mode === "junqi" ? 5 : 17;
  let html = `<div class="board-stage ${mode === "junqi" ? "board-stage-junqi" : ""}">${deadTrayHTML()}<div class="board board-${mode}">${railOverlayHTML()}${playerTickersHTML()}`;
  for (let displayRow = 0; displayRow < rows; displayRow++) {
    for (let displayCol = 0; displayCol < cols; displayCol++) {
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
      const riverLabel = mode === "junqi" ? junqiRiverLabel(displayRow, displayCol) : centralRiverLabel(displayRow, displayCol);
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
  if (currentMode() === "junqi") return junqiRailOverlayHTML();
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

function junqiRailOverlayHTML() {
  const line = (x1, y1, x2, y2) => `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" />`;
  const parts = [];
  const sleepers = [];
  const addVertical = (x, y1, y2) => {
    parts.push(line(x, y1, x, y2));
    for (let y = Math.ceil(y1); y <= Math.floor(y2); y += 2) sleepers.push(line(x - .16, y + .5, x + .16, y + .5));
  };
  const addHorizontal = (y, x1, x2) => {
    parts.push(line(x1, y, x2, y));
    for (let x = Math.ceil(x1); x <= Math.floor(x2); x += 2) sleepers.push(line(x + .5, y - .16, x + .5, y + .16));
  };
  [1.5, 5.5, 7.5, 11.5].forEach(y => addHorizontal(y, .5, 4.5));
  [0.5, 2.5, 4.5].forEach(x => addVertical(x, 1.5, 11.5));
  return `<svg class="rails rails-junqi" viewBox="0 0 5 13" aria-hidden="true">
    <g class="rail-shadow">${parts.join("")}</g>
    <g class="rail-main">${parts.join("")}</g>
    <g class="rail-sleeper">${sleepers.join("")}</g>
  </svg>`;
}

function playerTickersHTML() {
  if (!state.room) return "";
  const phase = state.room.phase;
  if (phase !== "playing" && phase !== "ended") return "";
  const seats = state.room.seats || [];
  const skips = state.room.skips || {};
  const maxSkips = state.room.maxSkips || 5;
  const eliminated = state.room.eliminated || {};
  const turn = state.room.turn;
  return seats.map(s => {
    const seat = s.seat;
    const used = skips[seat] || 0;
    const isElim = !!eliminated[seat];
    const reqPending = !!state.room.request;
    const active = phase === "playing" && seat === turn && !isElim && !reqPending;
    const orient = `ticker-rel-${relativeSeat(seat)}`;
    return `<div class="player-ticker ${orient} ${active ? "active" : ""} ${isElim ? "elim" : ""}" data-seat="${seat}">
      <div class="ticker-head">
        <span class="ticker-seat">${seatNames[seat]}</span>
        <span class="ticker-name">${esc(s.name || "空位")}</span>
      </div>
      <div class="ticker-bar"><div class="ticker-bar-fill"></div></div>
      <div class="ticker-stats">
        <span class="ticker-time"></span>
        <span class="ticker-skips">跳过 ${used}/${maxSkips}</span>
        ${isElim ? `<span class="ticker-tag">已淘汰</span>` : ""}
      </div>
    </div>`;
  }).join("");
}

function actionButtonsHTML() {
  if (!state.room || state.room.phase !== "playing") return "";
  const turn = state.room.turn;
  const skips = state.room.skips || {};
  const maxSkips = state.room.maxSkips || 5;
  const used = skips[state.seat] || 0;
  const remaining = Math.max(0, maxSkips - used);
  const isElim = !!(state.room.eliminated || {})[state.seat];
  const reqPending = !!state.room.request;
  const myTurn = turn === state.seat;
  return `
    <button id="skipBtn" ${(!myTurn || used >= maxSkips || isElim || reqPending) ? "disabled" : ""}>跳过 (${remaining})</button>
    <button id="tieBtn" ${(isElim || reqPending) ? "disabled" : ""}>求和</button>
    <button id="surrenderBtn" ${(isElim || reqPending) ? "disabled" : ""}>投降</button>
  `;
}

function requestBannerHTML() {
  if (!state.room || !state.room.request) return "";
  const req = state.room.request;
  const label = req.kind === "tie" ? "和棋" : "投降";
  const fromSeat = req.from;
  const fromInfo = (state.room.seats || []).find(s => s.seat === fromSeat) || {};
  const fromName = fromInfo.name || "";
  const eliminated = state.room.eliminated || {};
  const myElim = !!eliminated[state.seat];
  const acks = req.acks || [];
  let myAction = "";
  let statusText = "";
  let actorTag = "";
  if (req.stage === "teammate") {
    statusText = `${seatNames[fromSeat]}方 ${esc(fromName)} 请求${label}，等待队友支持`;
    if (state.seat === seatPartner(fromSeat) && !myElim) {
      actorTag = "需要你回应";
      myAction = `<button id="reqAccept" class="primary">支持</button><button id="reqReject">拒绝</button>`;
    }
  } else {
    statusText = `${seatNames[fromSeat]}方请求${label}，等待对方回应`;
    if (!sameSide(state.seat, fromSeat) && !myElim) {
      if (acks.includes(state.seat)) {
        statusText += `（你已同意）`;
      } else {
        actorTag = "需要你回应";
        myAction = `<button id="reqAccept" class="primary">同意</button><button id="reqReject">拒绝</button>`;
      }
    }
  }
  const cancelBtn = state.seat === fromSeat ? `<button id="reqCancel">撤回</button>` : "";
  return `<div class="request-banner ${actorTag ? "request-mine" : ""}">
    <div class="request-text">${actorTag ? `<b>${actorTag}：</b>` : ""}${statusText}</div>
    <div class="request-actions">${myAction}${cancelBtn}</div>
  </div>`;
}

function seatPartner(s) {
  return [2, 3, 0, 1][s];
}

function sameTeam(a, b) {
  return a === b || seatPartner(a) === b;
}

function sameSide(a, b) {
  return currentMode() === "junqi" ? a === b : sameTeam(a, b);
}

function tickTimer() {
  if (!state.room) return;
  const phase = state.room.phase;
  const turn = state.room.turn;
  const deadline = state.room.moveDeadlineMs || 0;
  const limitSec = state.room.moveLimitSec || 15;
  const reqPending = !!state.room.request;
  const eliminated = state.room.eliminated || {};
  document.querySelectorAll(".player-ticker").forEach(el => {
    const seat = Number(el.dataset.seat);
    const bar = el.querySelector(".ticker-bar-fill");
    const text = el.querySelector(".ticker-time");
    if (!bar || !text) return;
    const isActive = phase === "playing" && seat === turn && deadline > 0 && !reqPending && !eliminated[seat];
    if (!isActive) {
      bar.style.width = "0%";
      bar.classList.remove("urgent");
      text.textContent = "";
      el.classList.remove("active");
      return;
    }
    el.classList.add("active");
    const remaining = Math.max(0, deadline - Date.now());
    const seconds = Math.ceil(remaining / 1000);
    const pct = Math.max(0, Math.min(100, (remaining / (limitSec * 1000)) * 100));
    bar.style.width = pct + "%";
    text.textContent = seconds + "s";
    if (seconds <= 5) bar.classList.add("urgent");
    else bar.classList.remove("urgent");
  });
}

function homeClass(row, col) {
  if (currentMode() === "junqi") {
    if (row >= 2 && row <= 7 && col >= 6 && col <= 10) return "home-north home-junqi";
    if (row >= 9 && row <= 14 && col >= 6 && col <= 10) return "home-south home-junqi";
    if (row === 8) return "home-gap";
    return "";
  }
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
  if (currentMode() === "junqi") {
    if ((turn === 0 || turn === 2) && row === 8 && [6, 8, 10].includes(col)) return "turn-front";
    return "";
  }
  if (turn === 0 && row === 6 && col >= 6 && col <= 10) return "turn-front";
  if (turn === 1 && col === 10 && row >= 6 && row <= 10) return "turn-front";
  if (turn === 2 && row === 10 && col >= 6 && col <= 10) return "turn-front";
  if (turn === 3 && col === 6 && row >= 6 && row <= 10) return "turn-front";
  return "";
}

function visualTrackClass(row, col) {
  if (currentMode() === "junqi") {
    const localRow = row - 2;
    const localCol = col - 6;
    const onTrack = [1, 5, 7, 11].includes(localRow) || [0, 2, 4].includes(localCol);
    return onTrack ? "track-segment track-junqi" : "";
  }
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

function junqiRiverLabel(displayRow, displayCol) {
  if (displayRow === 6 && displayCol === 1) return {text: "山界", className: "river-label-left river-label-bottom"};
  if (displayRow === 6 && displayCol === 3) return {text: "山界", className: "river-label-right river-label-bottom"};
  return null;
}

function relativeSeat(owner) {
  if (currentMode() === "junqi") return owner === state.seat ? 0 : 2;
  return (owner - state.seat + 4) % 4;
}

function toDisplay(row, col) {
  if (currentMode() === "junqi") {
    if (state.seat === 0) return {row: 14 - row, col: 10 - col};
    return {row: row - 2, col: col - 6};
  }
  switch (state.seat) {
    case 0: return {row: 16 - row, col: 16 - col};
    case 1: return {row: col, col: 16 - row};
    case 2: return {row, col};
    case 3: return {row: 16 - col, col: row};
    default: return {row, col};
  }
}

function fromDisplay(row, col) {
  if (currentMode() === "junqi") {
    if (state.seat === 0) return {row: 14 - row, col: 10 - col};
    return {row: row + 2, col: col + 6};
  }
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
  const sendTeamBtn = document.querySelector("#sendTeam");
  if (sendTeamBtn) sendTeamBtn.onclick = () => sendChat("team");
  document.querySelector("#modeSiguo").onclick = () => setMode("siguo");
  document.querySelector("#modeJunqi").onclick = () => setMode("junqi");
  const skipBtn = document.querySelector("#skipBtn");
  if (skipBtn) skipBtn.onclick = () => send({type:"move.skip"});
  const tieBtn = document.querySelector("#tieBtn");
  if (tieBtn) tieBtn.onclick = () => confirmAction("发起求和请求？", () => send({type:"request.tie"}));
  const surrBtn = document.querySelector("#surrenderBtn");
  if (surrBtn) surrBtn.onclick = () => confirmAction(currentMode() === "junqi" ? "确认投降并结束本局？" : "发起投降请求？队友支持后立即结束本局。", () => send({type:"request.surrender"}));
  const reqAccept = document.querySelector("#reqAccept");
  if (reqAccept) reqAccept.onclick = () => send({type:"request.respond", kind: state.room?.request?.kind, accept: true});
  const reqReject = document.querySelector("#reqReject");
  if (reqReject) reqReject.onclick = () => send({type:"request.respond", kind: state.room?.request?.kind, accept: false});
  const reqCancel = document.querySelector("#reqCancel");
  if (reqCancel) reqCancel.onclick = () => send({type:"request.cancel"});
  document.querySelector("#name").oninput = e => {
    state.name = e.target.value;
    localStorage.setItem("siguo.name", state.name);
  };
  document.querySelector("#code").oninput = e => state.code = e.target.value.toUpperCase();
  document.querySelectorAll(".cell").forEach(cell => cell.onclick = () => clickCell(cell));
  tickTimer();
}

function confirmAction(message, fn) {
  if (typeof window !== "undefined" && window.confirm) {
    if (!window.confirm(message)) return;
  }
  fn();
}

async function createRoom() {
  state.name = document.querySelector("#name").value || "玩家";
  const res = await fetch("/api/rooms", {method:"POST", body: JSON.stringify({name: state.name, mode: currentMode()})});
  await acceptJoin(res);
}

function setMode(mode) {
  if (state.room && state.room.phase !== "lobby") return;
  state.mode = mode;
  localStorage.setItem("siguo.mode", mode);
  if (state.room && state.host) {
    send({type:"room.config", mode});
  } else {
    render();
  }
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
    if (msg.room?.mode) {
      state.mode = msg.room.mode;
      localStorage.setItem("siguo.mode", state.mode);
    }
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
setInterval(tickTimer, 200);
if (state.code && state.token) connect();
