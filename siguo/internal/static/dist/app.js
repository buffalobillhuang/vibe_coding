const seatNames = ["北", "东", "南", "西"];
const seatPieceClasses = ["piece-red", "piece-green", "piece-yellow", "piece-blue"];
const modeNames = {siguo: "四国 2v2", junqi: "军棋 1v1"};
const rankNames = {
  0: "军棋", 1: "军旗", 2: "地雷", 3: "炸弹", 4: "工兵", 5: "排长", 6: "连长",
  7: "营长", 8: "团长", 9: "旅长", 10: "师长", 11: "军长", 12: "司令"
};
const cellClasses = ["off", "normal", "camp", "railroad", "hq", "frontline", "mountain"];
const setupPoemTitle = "临江仙";
const setupPoemAuthor = "明·杨慎";
const setupPoemLines = [
  "滚滚长江东逝水",
  "浪花淘尽英雄",
  "是非成败转头空",
  "青山依旧在",
  "几度夕阳红",
  "白发渔樵江渚上",
  "惯看秋月春风",
  "一壶浊酒喜相逢",
  "古今多少事",
  "都付笑谈中"
];
const quickChatPhrases = [
  "猪头",
  "不怕神一样的对手，就怕猪一样的队友",
  "以卵击石！。。。",
  "固若金汤 哈哈哈",
  "你太牛了",
  "掐死你 ：-/"
];

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
  setupMusic: null,
  setupMusicEnabled: localStorage.getItem("siguo.setupMusic") === "on",
  setupMusicSource: 0,
  setupMusicBlocked: false,
  combat: null,
  log: [],
  chat: [],
  lowTimeWarned: false,
  viewer: false,
  watchOpen: false,
  watchRooms: [],
  joinOffer: null
};

const app = document.querySelector("#app");

function render() {
  app.innerHTML = `
    <div class="shell">
      <section class="table">
        <div class="topbar">
          <div class="brand"><h1>四国军棋</h1><span id="roomStatus">${statusText()}</span></div>
          <div class="row" style="max-width:720px">
            <button id="watchBtn">观战室</button>
            ${inviteLinkButtonHTML()}
            ${viewerLinkButtonHTML()}
            <button class="primary" id="startBtn" ${canStart() ? "" : "disabled"}>开始</button>
            <button id="randBtn" ${!state.viewer && state.room?.phase === "setup" ? "" : "disabled"}>随机</button>
            <button id="submitBtn" ${!state.viewer && state.room?.phase === "setup" ? "" : "disabled"}>提交</button>
            ${actionButtonsHTML()}
            <button id="soundBtn" class="toggle ${state.sound ? "on" : "off"}">声音</button>
          </div>
        </div>
        ${watchRoomPanelHTML()}
        ${joinOfferHTML()}
        ${requestBannerHTML()}
        <div class="board-wrap">${victoryHTML()}${boardHTML()}</div>
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
          <div class="subtle">${state.viewer ? "观战中" : "本机地址可直接分享给同一局域网玩家"}。模式：${modeNames[currentMode()]}。你的座位：${state.viewer ? "观战" : seatNames[state.seat] || "-"}</div>
        </div>
        ${setupCultureHTML()}
        <div class="panel">
          <div class="seats seats-${currentMode()}">${seatsHTML()}</div>
        </div>
        <div class="panel stack">
          <b>战况</b>
          <div class="log">${state.log.slice(-80).map(x => `<div class="line">${esc(x)}</div>`).join("")}</div>
        </div>
        <div class="panel stack">
          <b>聊天</b>
          <div class="chat-log">${state.chat.slice(-80).map(chatLine).join("")}</div>
          ${state.viewer ? "" : `<div class="row">
            <input id="chatText" maxlength="200" placeholder="输入消息" />
            <button id="sendAll">公屏</button>
            ${currentMode() === "junqi" ? "" : `<button id="sendTeam">队伍</button>`}
          </div>
          ${quickChatHTML()}`}
        </div>
      </aside>
    </div>
  `;
  bind();
  syncSetupMusic();
}

function statusText() {
  if (!state.room) return "创建或加入一个 6 位房间码";
  const phase = {lobby: "大厅", setup: "布阵", playing: "对局", ended: "结束"}[state.room.phase] || state.room.phase;
  const setup = state.room.phase === "setup" ? setupCountdownText() : "";
  const turn = state.room.phase === "playing" ? ` · ${seatNames[state.room.turn]}方行动` : "";
  const role = state.viewer ? "观战" : `房间 ${state.code}`;
  return `${role} · ${modeNames[currentMode()]} · ${phase}${setup}${turn}`;
}

function setupCountdownText() {
  const deadline = state.room?.setupDeadlineMs || 0;
  if (!deadline) return "";
  const seconds = Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
  return ` · ${seconds}秒后自动开局`;
}

function canStart() {
  if (state.viewer) return false;
  const seats = activeSeats();
  const occupied = state.room?.seats?.filter(s => seats.includes(s.seat) && s.name) || [];
  return state.room?.phase === "lobby" && state.host && occupied.length === seats.length;
}

function winnerSeats() {
  return state.room?.winner || state.room?.Winner || state.view?.Winner || state.view?.winner || [];
}

function junqiSeatColor(seat) {
  return Number(seat) === 0 ? "red" : "blue";
}

function winnerLabel() {
  const winners = winnerSeats();
  if (!winners.length) return "和棋";
  if (currentMode() === "junqi") return junqiSeatColor(winners[0]) === "red" ? "红方获胜" : "蓝方获胜";
  return sameTeam(winners[0], 0) ? "北南联军获胜" : "东西联军获胜";
}

function victorySubline() {
  if (currentMode() === "junqi") return "一局定鼎";
  return "联军得胜";
}

function victoryToneClass() {
  const winners = winnerSeats();
  if (!winners.length) return "victory-neutral";
  if (currentMode() === "junqi") return `victory-${junqiSeatColor(winners[0])}`;
  return sameTeam(winners[0], 0) ? "victory-ns" : "victory-ew";
}

function victoryHTML() {
  if (state.room?.phase !== "ended") return "";
  const winners = winnerSeats();
  if (!winners.length) {
    return `<div class="victory-layer victory-draw"><div class="victory-panel"><span class="victory-kicker">终局</span><b>和棋</b>${victoryActionsHTML()}</div></div>`;
  }
  const petals = Array.from({length: 18}, (_, i) => `<span class="petal p${i + 1}"></span>`).join("");
  return `<div class="victory-layer ${victoryToneClass()}" aria-live="polite">
    <div class="victory-panel">
      ${petals}
      <div class="beauty beauty-left"></div>
      <div class="beauty beauty-right"></div>
      <span class="victory-kicker">军旗已定</span>
      <b>${winnerLabel()}</b>
      <small>${victorySubline()} · 另一方获胜</small>
      ${victoryActionsHTML()}
    </div>
  </div>`;
}

function victoryActionsHTML() {
  if (state.viewer) {
    return `<div class="victory-actions"><button id="leaveRoomBtn">返回大厅</button></div>`;
  }
  return `<div class="victory-actions">
    <button id="restartRoomBtn" class="primary">再开一局</button>
    <button id="leaveRoomBtn">返回大厅</button>
  </div>`;
}

function setupCultureHTML() {
  if (!state.room || !["setup", "playing"].includes(state.room.phase)) return "";
  const musicText = state.setupMusicEnabled ? "静音" : state.setupMusicBlocked ? "再试" : "启乐";
  const modeLine = currentMode() === "junqi" ? "一枰对坐，风云入局" : "四方列阵，战局正酣";
  const poemColumns = [
    `<span class="poem-title">${setupPoemTitle}</span>`,
    `<span class="poem-author">${setupPoemAuthor}</span>`,
    ...setupPoemLines.map(line => `<span>${line}</span>`)
  ].join("");
  return `<div class="panel culture-panel">
    <div class="culture-head">
      <div><b>临江仙</b><span>杨慎 · ${modeLine}</span></div>
      <button id="setupMusicBtn" class="music-chip">${musicText}</button>
    </div>
    <div class="poem-window">
      <div class="poem-scroll">
        ${poemColumns}${poemColumns}
      </div>
    </div>
  </div>`;
}

function quickChatHTML() {
  return `<div class="quick-chat" aria-label="常用语">
    ${quickChatPhrases.map(text => `<button type="button" class="quick-chat-btn" data-phrase="${esc(text)}">${esc(text)}</button>`).join("")}
  </div>`;
}

function viewerLinkButtonHTML() {
  if (!state.room || !["setup", "playing"].includes(state.room.phase)) return "";
  return `<button id="viewerLinkBtn">观战链接</button>`;
}

function inviteLinkButtonHTML() {
  if (state.viewer || !state.room || !state.code || !["lobby", "setup", "playing"].includes(state.room.phase)) return "";
  return `<button id="inviteLinkBtn">邀请链接</button>`;
}

function joinOfferHTML() {
  const offer = state.joinOffer;
  if (!offer) return "";
  const viewerAction = offer.canView ? `<button id="joinOfferView" class="primary">观战</button>` : "";
  const viewerText = offer.canView ? "可作为观众进入。" : "当前没有可用观战席。";
  return `<div class="panel join-offer">
    <div><b>无法加入 ${esc(offer.code)}</b><span>${esc(offer.message)} ${viewerText}</span></div>
    <div><button id="joinOfferClose">关闭</button>${viewerAction}</div>
  </div>`;
}

function watchRoomPanelHTML() {
  if (!state.watchOpen) return "";
  const rooms = state.watchRooms || [];
  const rows = rooms.length ? rooms.map(room => {
    const names = (room.seats || []).map(s => s.name || seatNames[s.seat]).join(" · ");
    return `<div class="watch-row">
      <div><b>${esc(room.code)}</b><span>${modeNames[room.mode] || room.mode} · ${room.phase === "setup" ? "布阵" : "对局"} · ${room.viewers}/${room.maxViewers} 观战</span><small>${esc(names)}</small></div>
      <button class="watch-join" data-code="${esc(room.code)}" ${room.canJoinView ? "" : "disabled"}>观看</button>
      <button class="watch-copy" data-code="${esc(room.code)}">复制链接</button>
    </div>`;
  }).join("") : `<div class="subtle">暂无进行中的对局</div>`;
  return `<div class="panel watch-panel">
    <div class="watch-head"><b>观战室</b><div><button id="watchRefreshBtn">刷新</button><button id="watchCloseBtn">关闭</button></div></div>
    <div class="watch-list">${rows}</div>
  </div>`;
}

function seatsHTML() {
  const seats = state.room?.seats || activeSeats().map(seat => ({seat}));
  return seats.map(s => `
    <div class="seat ${state.room?.turn === s.seat ? "current" : ""}" data-seat="${s.seat}">
      <b>${seatNames[s.seat]}</b>
      <span class="subtle">${s.name ? esc(s.name) : "空位"} ${s.host ? "房主" : ""} ${s.ready ? "已提交" : ""}</span>
    </div>
  `).join("");
}

function modeSwitchHTML() {
  const locked = state.room && !["lobby", "ended"].includes(state.room.phase);
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
  const stageClass = mode === "junqi" ? "board-stage board-stage-junqi" : "board-stage";
  const boardOpen = mode === "junqi"
    ? `<div class="board-surface-wrap board-surface-wrap-junqi">${playerTickersHTML()}<div class="board-clip board-clip-junqi"><div class="board board-${mode} board-junqi-rel-${state.seat}">${railOverlayHTML()}`
    : `<div class="board board-${mode} board-siguo-rel-${state.seat}">${railOverlayHTML()}${playerTickersHTML()}`;
  const boardClose = mode === "junqi" ? `</div></div></div>` : `</div>`;
  let html = `<div class="${stageClass}">${boardOpen}`;
  for (let displayRow = 0; displayRow < rows; displayRow++) {
    for (let displayCol = 0; displayCol < cols; displayCol++) {
      const {row, col} = fromDisplay(displayRow, displayCol);
      const c = cells.get(`${row},${col}`);
      const piece = c?.Piece || c?.piece;
      const type = c?.Type ?? c?.type ?? 0;
      const selected = state.selected && state.selected.row === row && state.selected.col === col;
      const combat = state.combat && state.combat.row === row && state.combat.col === col && Date.now() < state.combat.until;
      const cellStyle = mode === "junqi" ? junqiCellStyle(displayRow, displayCol, row) : siguoCellStyle(displayRow, displayCol);
      html += `<div class="cell ${cellClasses[type] || "off"} ${homeClass(row, col)} ${visualTrackClass(row, col)} ${turnFrontClass(row, col)} ${piece ? "occupied" : ""} ${selected ? "selected" : ""} ${combat ? "combat-hit" : ""}" data-row="${row}" data-col="${col}" ${cellStyle}>`;
      if (piece) {
        const owner = piece.Owner ?? piece.owner;
        const rank = piece.Rank ?? piece.rank;
        const exposed = piece.Exposed ?? piece.exposed;
        const colorClass = pieceColorClass(owner);
        const exposedClass = exposed && rank !== 0 ? "flag-exposed" : "";
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
  return html + `${boardClose}${deadTrayHTML()}</div>`;
}

function deadTrayHTML() {
  const dead = state.view?.DeadPieces || state.view?.deadPieces || [];
  const items = dead.map(piece => {
    const owner = piece.Owner ?? piece.owner;
    const rank = piece.Rank ?? piece.rank;
    const colorClass = pieceColorClass(owner);
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
  [0.5, 4.5].forEach(x => addVertical(x, 1.5, 11.5));
  addVertical(2.5, 5.5, 7.5);
  return `<svg class="rails rails-junqi" viewBox="0 0 5 13" preserveAspectRatio="none" aria-hidden="true">
    <g class="rail-shadow">${parts.join("")}</g>
    <g class="rail-main">${parts.join("")}</g>
    <g class="rail-sleeper">${sleepers.join("")}</g>
  </svg>`;
}

function pieceColorClass(owner) {
  if (currentMode() === "junqi") return `piece-${junqiSeatColor(owner)}`;
  return seatPieceClasses[owner] || "piece-blue";
}

function playerTickersHTML() {
  if (!state.room) return "";
  const phase = state.room.phase;
  if (!["lobby", "setup", "playing", "ended"].includes(phase)) return "";
  const seats = state.room.seats || [];
  const skips = state.room.skips || {};
  const maxSkips = state.room.maxSkips || 5;
  const eliminated = state.room.eliminated || {};
  const turn = state.room.turn;
  const isJunqi = currentMode() === "junqi";
  const moveLimit = state.room.moveLimitSec || 45;
  return seats.map(s => {
    const seat = s.seat;
    const used = skips[seat] || 0;
    const isElim = !!eliminated[seat];
    const reqPending = !!state.room.request;
    const active = phase === "playing" && seat === turn && !isElim && !reqPending;
    const canRequest = !state.viewer && isJunqi && phase === "playing" && seat === state.seat && !isElim && !reqPending;
    const canSkip = !state.viewer && isJunqi && phase === "playing" && seat === state.seat && seat === turn && used < maxSkips && !isElim && !reqPending;
    const orient = `ticker-rel-${relativeSeat(seat)}`;
    return `<div class="player-ticker ${orient} ${active ? "active" : ""} ${isElim ? "elim" : ""}" data-seat="${seat}">
      <div class="ticker-head">
        <span class="ticker-seat">${seatNames[seat]}</span>
        <span class="ticker-turn-logo" aria-label="行动方"></span>
        <span class="ticker-name">${esc(s.name || "空位")}</span>
      </div>
      <div class="ticker-bar"><div class="ticker-bar-fill"></div></div>
      <div class="ticker-stats">
        <span class="ticker-time">${isJunqi ? `${moveLimit}s` : ""}</span>
        ${isJunqi ? "" : `<span class="ticker-skips">跳过 ${used}/${maxSkips}</span>`}
        ${isElim ? `<span class="ticker-tag">已淘汰</span>` : ""}
      </div>
      ${isJunqi ? `<div class="ticker-actions">
        <button class="ticker-skip" data-seat="${seat}" ${canSkip ? "" : "disabled"}>跳过 ${used}/${maxSkips}</button>
        <button class="ticker-surrender" data-seat="${seat}" ${canRequest ? "" : "disabled"}>投降</button>
        <button class="ticker-peace" data-seat="${seat}" ${canRequest ? "" : "disabled"}>求和</button>
      </div>` : ""}
    </div>`;
  }).join("");
}

function actionButtonsHTML() {
  if (state.viewer) return "";
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
    if (!state.viewer && state.seat === seatPartner(fromSeat) && !myElim) {
      actorTag = "需要你回应";
      myAction = `<button id="reqAccept" class="primary">支持</button><button id="reqReject">拒绝</button>`;
    }
  } else {
    statusText = `${seatNames[fromSeat]}方请求${label}，等待对方回应`;
    if (!state.viewer && !sameSide(state.seat, fromSeat) && !myElim) {
      if (acks.includes(state.seat)) {
        statusText += `（你已同意）`;
      } else {
        actorTag = "需要你回应";
        myAction = `<button id="reqAccept" class="primary">同意</button><button id="reqReject">拒绝</button>`;
      }
    }
  }
  const cancelBtn = !state.viewer && state.seat === fromSeat ? `<button id="reqCancel">撤回</button>` : "";
  return `<div class="request-banner ${actorTag ? "request-mine" : ""}">
    <div class="request-text">${actorTag ? `<b>${actorTag}：</b>` : ""}${statusText}</div>
    <div class="request-actions">${myAction}${cancelBtn}</div>
  </div>`;
}

function seatPartner(s) {
  return [2, 3, 0, 1][Number(s)];
}

function sameTeam(a, b) {
  const seatA = Number(a);
  const seatB = Number(b);
  return seatA === seatB || seatPartner(seatA) === seatB;
}

function sameSide(a, b) {
  return currentMode() === "junqi" ? a === b : sameTeam(a, b);
}

function tickTimer() {
  if (!state.room) return;
  const status = document.querySelector("#roomStatus");
  if (status) status.textContent = statusText();
  const phase = state.room.phase;
  const turn = state.room.turn;
  const deadline = state.room.moveDeadlineMs || 0;
  const limitSec = state.room.moveLimitSec || 45;
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
      text.textContent = currentMode() === "junqi" ? `${limitSec}s` : "";
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
    const onTrack = [1, 5, 7, 11].includes(localRow) || [0, 4].includes(localCol) || (localRow === 6 && [0, 2, 4].includes(localCol));
    return onTrack ? "track-segment track-junqi" : "";
  }
  const inCenter = row >= 6 && row <= 10 && col >= 6 && col <= 10;
  const onTrack = [6, 8, 10].includes(row) || [6, 8, 10].includes(col);
  const isNode = [6, 8, 10].includes(row) && [6, 8, 10].includes(col);
  return inCenter && onTrack && !isNode ? "track-segment" : "";
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

const siguoMap8Centers = {
  x: [3.50, 8.13, 12.88, 17.75, 22.63, 27.50, 35.25, 42.75, 50.13, 57.38, 64.63, 72.50, 77.13, 81.75, 86.63, 91.38, 96.13],
  y: [4.13, 8.50, 13.13, 17.63, 22.00, 26.75, 34.88, 42.38, 50.00, 57.75, 65.00, 72.75, 77.13, 81.75, 86.38, 90.88, 95.75]
};

function siguoCellStyle(displayRow, displayCol) {
  const w = 5.2;
  const h = 4.6;
  const x = siguoMap8Centers.x[displayCol] - w / 2;
  const y = siguoMap8Centers.y[displayRow] - h / 2;
  return `style="left:${x.toFixed(2)}%;top:${y.toFixed(2)}%"`;
}

const junqiMap3Centers = {
  x: [13.35, 31.48, 50.00, 67.94, 86.57],
  y: [8.45, 15.10, 21.75, 28.38, 35.03, 41.72, 50.08, 58.11, 64.78, 71.45, 78.04, 84.65, 91.55]
};

function junqiCellStyle(displayRow, displayCol, row) {
  const w = 9.1;
  const h = 4.0;
  const nudge = junqiBackRowNudge(row);
  const x = junqiMap3Centers.x[displayCol] + nudge.x - w / 2;
  const y = junqiMap3Centers.y[displayRow] + nudge.y - h / 2;
  return `style="left:${x.toFixed(2)}%;top:${y.toFixed(2)}%"`;
}

function junqiBackRowNudge(row) {
  const rotate = state.seat === 0 ? -1 : 1;
  if (row >= 2 && row <= 4) {
    const amount = [0.70, 0.55, 0.40][row - 2];
    return {x: rotate * amount, y: -rotate * amount};
  }
  if (row >= 13 && row <= 14) {
    const amount = [0.55, 0.70][row - 13];
    return {x: -rotate * amount, y: rotate * amount};
  }
  return {x: 0, y: 0};
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
  document.querySelector("#watchBtn").onclick = openWatchRoom;
  const inviteLinkBtn = document.querySelector("#inviteLinkBtn");
  if (inviteLinkBtn) inviteLinkBtn.onclick = () => copyInviteLink(state.code);
  const viewerLinkBtn = document.querySelector("#viewerLinkBtn");
  if (viewerLinkBtn) viewerLinkBtn.onclick = () => copyViewerLink(state.code);
  const joinOfferView = document.querySelector("#joinOfferView");
  if (joinOfferView) joinOfferView.onclick = () => connectViewer(state.joinOffer?.code);
  const joinOfferClose = document.querySelector("#joinOfferClose");
  if (joinOfferClose) joinOfferClose.onclick = () => { state.joinOffer = null; render(); };
  const watchRefreshBtn = document.querySelector("#watchRefreshBtn");
  if (watchRefreshBtn) watchRefreshBtn.onclick = refreshWatchRooms;
  const watchCloseBtn = document.querySelector("#watchCloseBtn");
  if (watchCloseBtn) watchCloseBtn.onclick = () => { state.watchOpen = false; render(); };
  document.querySelectorAll(".watch-join").forEach(btn => btn.onclick = () => connectViewer(btn.dataset.code));
  document.querySelectorAll(".watch-copy").forEach(btn => btn.onclick = () => copyViewerLink(btn.dataset.code));
  document.querySelector("#startBtn").onclick = () => send({type:"room.start"});
  document.querySelector("#randBtn").onclick = () => send({type:"setup.randomize"});
  document.querySelector("#submitBtn").onclick = () => send({type:"setup.submit"});
  document.querySelector("#soundBtn").onclick = toggleSound;
  const setupMusicBtn = document.querySelector("#setupMusicBtn");
  if (setupMusicBtn) setupMusicBtn.onclick = () => {
    state.setupMusicEnabled = !state.setupMusicEnabled;
    localStorage.setItem("siguo.setupMusic", state.setupMusicEnabled ? "on" : "off");
    if (state.setupMusicEnabled) playSetupMusic(true);
    else syncSetupMusic();
    render();
  };
  const sendAllBtn = document.querySelector("#sendAll");
  if (sendAllBtn) sendAllBtn.onclick = () => sendChat("all");
  const sendTeamBtn = document.querySelector("#sendTeam");
  if (sendTeamBtn) sendTeamBtn.onclick = () => sendChat("team");
  document.querySelectorAll(".quick-chat-btn").forEach(btn => btn.onclick = () => fillQuickChat(btn.dataset.phrase));
  document.querySelector("#modeSiguo").onclick = () => setMode("siguo");
  document.querySelector("#modeJunqi").onclick = () => setMode("junqi");
  const skipBtn = document.querySelector("#skipBtn");
  if (skipBtn) skipBtn.onclick = () => send({type:"move.skip"});
  const tieBtn = document.querySelector("#tieBtn");
  if (tieBtn) tieBtn.onclick = () => confirmAction("发起求和请求？", () => send({type:"request.tie"}));
  const surrBtn = document.querySelector("#surrenderBtn");
  if (surrBtn) surrBtn.onclick = () => confirmAction(currentMode() === "junqi" ? "确认投降并结束本局？" : "发起投降请求？队友支持后立即结束本局。", () => send({type:"request.surrender"}));
  document.querySelectorAll(".ticker-skip").forEach(btn => btn.onclick = e => {
    e.stopPropagation();
    if (btn.disabled || Number(btn.dataset.seat) !== state.seat) return;
    send({type:"move.skip"});
  });
  document.querySelectorAll(".ticker-surrender").forEach(btn => btn.onclick = e => {
    e.stopPropagation();
    if (btn.disabled || Number(btn.dataset.seat) !== state.seat) return;
    confirmAction("确认投降并结束本局？", () => send({type:"request.surrender"}));
  });
  document.querySelectorAll(".ticker-peace").forEach(btn => btn.onclick = e => {
    e.stopPropagation();
    if (btn.disabled || Number(btn.dataset.seat) !== state.seat) return;
    confirmAction("发起求和请求？", () => send({type:"request.tie"}));
  });
  const reqAccept = document.querySelector("#reqAccept");
  if (reqAccept) reqAccept.onclick = () => send({type:"request.respond", kind: state.room?.request?.kind, accept: true});
  const reqReject = document.querySelector("#reqReject");
  if (reqReject) reqReject.onclick = () => send({type:"request.respond", kind: state.room?.request?.kind, accept: false});
  const reqCancel = document.querySelector("#reqCancel");
  if (reqCancel) reqCancel.onclick = () => send({type:"request.cancel"});
  const restartRoomBtn = document.querySelector("#restartRoomBtn");
  if (restartRoomBtn) restartRoomBtn.onclick = createRoom;
  const leaveRoomBtn = document.querySelector("#leaveRoomBtn");
  if (leaveRoomBtn) leaveRoomBtn.onclick = leaveEndedRoom;
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

async function openWatchRoom() {
  state.watchOpen = true;
  render();
  await refreshWatchRooms();
}

async function refreshWatchRooms() {
  const res = await fetch("/api/rooms");
  if (!res.ok) {
    log(`错误：${await responseErrorText(res)}`);
    return;
  }
  const data = await res.json();
  state.watchRooms = data.rooms || [];
  state.watchOpen = true;
  render();
}

function viewerURL(code) {
  return shareURL("watch", code);
}

function inviteURL(code) {
  return shareURL("join", code);
}

function shareURL(param, code) {
  const origin = shareOrigin();
  if (!origin) return "";
  return `${origin}${location.pathname}?${param}=${encodeURIComponent(code)}`;
}

function shareOrigin() {
  if (!["localhost", "127.0.0.1", "::1"].includes(location.hostname)) return location.origin;
  const saved = localStorage.getItem("siguo.shareOrigin") || "";
  const input = window.prompt("请输入可分享的服务器地址：云端80端口用 http://YOUR_VM_PUBLIC_IP，本机/LAN测试8080用 http://YOUR_LAN_IP:8080", saved || "http://");
  if (!input) return "";
  const origin = input.replace(/\/$/, "");
  localStorage.setItem("siguo.shareOrigin", origin);
  return origin;
}

async function copyInviteLink(code) {
  const url = inviteURL(code);
  if (!url) {
    log("未设置可分享的服务器地址");
    return;
  }
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(url);
    log("邀请链接已复制");
    return;
  }
  log(`邀请链接：${url}`);
}

async function copyViewerLink(code) {
  const url = viewerURL(code);
  if (!url) {
    log("未设置可分享的服务器地址");
    return;
  }
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(url);
    log("观战链接已复制");
    return;
  }
  log(`观战链接：${url}`);
}

function setMode(mode) {
  if (state.room?.phase === "ended") leaveEndedRoom(false);
  if (state.room && state.room.phase !== "lobby") return;
  state.mode = mode;
  localStorage.setItem("siguo.mode", mode);
  if (state.room && state.host) {
    send({type:"room.config", mode});
  } else {
    render();
  }
}

function leaveEndedRoom(shouldRender = true) {
  if (state.ws) {
    state.ws.onclose = null;
    state.ws.close();
  }
  state.ws = null;
  state.room = null;
  state.view = null;
  state.code = "";
  state.token = "";
  state.host = false;
  state.viewer = false;
  state.watchOpen = false;
  state.joinOffer = null;
  state.selected = null;
  state.combat = null;
  localStorage.removeItem("siguo.code");
  localStorage.removeItem("siguo.token");
  if (location.search.includes("watch=") || location.search.includes("join=")) history.replaceState(null, "", location.pathname);
  if (shouldRender) render();
}

async function joinRoom() {
  state.name = document.querySelector("#name").value || "玩家";
  state.code = document.querySelector("#code").value.toUpperCase();
  const res = await fetch(`/api/rooms/${state.code}/join`, {method:"POST", body: JSON.stringify({name: state.name, sessionToken: state.token})});
  await acceptJoin(res);
}

async function joinFromInvite(code) {
  code = String(code || "").toUpperCase();
  if (!code) return;
  state.name = state.name || "玩家";
  state.code = code;
  const res = await fetch(`/api/rooms/${code}/join`, {method:"POST", body: JSON.stringify({name: state.name, sessionToken: state.token})});
  await acceptJoin(res, {inviteCode: code});
}

async function acceptJoin(res, opts = {}) {
  if (!res.ok) {
    const message = await responseErrorText(res);
    if (opts.inviteCode) {
      await offerViewerAfterJoinFailure(opts.inviteCode, message);
      return;
    }
    log(`错误：${message}`);
    return;
  }
  const data = await res.json();
  state.code = data.code;
  state.token = data.sessionToken;
  state.seat = data.seat;
  state.host = data.host;
  state.viewer = false;
  state.watchOpen = false;
  state.joinOffer = null;
  localStorage.setItem("siguo.code", state.code);
  localStorage.setItem("siguo.token", state.token);
  localStorage.setItem("siguo.seat", String(state.seat));
  history.replaceState(null, "", location.pathname);
  connect();
}

async function offerViewerAfterJoinFailure(code, message) {
  const status = await viewerRoomStatus(code);
  state.joinOffer = {code, message, canView: status.ok};
  log(status.ok ? "玩家座位已满，可选择观战" : `错误：${message}`);
  render();
}

async function responseErrorText(res) {
  const text = await res.text();
  if (!text) return res.statusText || "请求失败";
  try {
    const data = JSON.parse(text);
    return data.error || text;
  } catch {
    return text;
  }
}

function connect() {
  if (state.ws) state.ws.close();
  const proto = location.protocol === "https:" ? "wss" : "ws";
  state.ws = new WebSocket(`${proto}://${location.host}/ws?room=${state.code}&token=${encodeURIComponent(state.token)}`);
  state.ws.onopen = () => log("已连接");
  state.ws.onclose = () => log("连接已断开");
  state.ws.onmessage = e => onMessage(JSON.parse(e.data));
}

async function connectViewer(code) {
  code = String(code || "").toUpperCase();
  const status = await viewerRoomStatus(code);
  if (!status.ok) {
    log(status.message);
    return;
  }
  if (state.ws) {
    state.ws.onclose = null;
    state.ws.close();
  }
  state.viewer = true;
  state.watchOpen = false;
  state.joinOffer = null;
  state.code = code;
  state.token = "";
  state.host = false;
  state.seat = 2;
  state.selected = null;
  state.combat = null;
  localStorage.removeItem("siguo.code");
  localStorage.removeItem("siguo.token");
  history.replaceState(null, "", `?watch=${encodeURIComponent(state.code)}`);
  const proto = location.protocol === "https:" ? "wss" : "ws";
  state.ws = new WebSocket(`${proto}://${location.host}/ws?room=${state.code}&viewer=1`);
  state.ws.onopen = () => log("已进入观战");
  state.ws.onclose = () => log("观战已断开");
  state.ws.onmessage = e => onMessage(JSON.parse(e.data));
}

async function viewerRoomStatus(code) {
  const res = await fetch("/api/rooms");
  if (!res.ok) return {ok: false, message: `错误：${await responseErrorText(res)}`};
  const data = await res.json();
  const room = (data.rooms || []).find(r => String(r.code).toUpperCase() === code);
  if (!room) return {ok: false, message: "对局不在进行中"};
  if (!room.canJoinView) return {ok: false, message: "观战席已满，请稍后再试"};
  return {ok: true};
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
  if (state.viewer) {
    log("观战中，不能操作棋子");
    return;
  }
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
  if (state.viewer) {
    log("观战中，不能操作对局");
    return;
  }
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN) {
    log("尚未连接");
    return;
  }
  msg.seq = state.seq++;
  state.ws.send(JSON.stringify(msg));
}

function sendChat(channel) {
  if (state.viewer) return;
  const input = document.querySelector("#chatText");
  send({type:"chat.send", channel, text: input.value});
  input.value = "";
}

function fillQuickChat(text) {
  const input = document.querySelector("#chatText");
  if (!input) return;
  input.value = text || "";
  input.focus();
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

function shouldPlaySetupMusic() {
  return state.setupMusicEnabled && ["setup", "playing"].includes(state.room?.phase);
}

function ensureSetupMusic() {
  if (state.setupMusic) return state.setupMusic;
  const audio = new Audio("/song.mp3");
  audio.loop = true;
  audio.volume = 0.32;
  audio.addEventListener("error", () => {
    if (state.setupMusicSource === 0) {
      state.setupMusicSource = 1;
      audio.src = "/setup-music.ogg";
      if (shouldPlaySetupMusic()) playSetupMusic(true);
    }
  });
  state.setupMusic = audio;
  return audio;
}

function syncSetupMusic() {
  if (shouldPlaySetupMusic()) {
    playSetupMusic(false);
    return;
  }
  if (state.setupMusic) {
    state.setupMusic.pause();
    state.setupMusic.currentTime = 0;
  }
  state.setupMusicBlocked = false;
}

function playSetupMusic(force) {
  if (!shouldPlaySetupMusic() && !force) return;
  const audio = ensureSetupMusic();
  audio.play().then(() => {
    state.setupMusicBlocked = false;
  }).catch(() => {
    state.setupMusicBlocked = true;
  });
}

function toggleSound() {
  state.sound = !state.sound;
  localStorage.setItem("siguo.sound", state.sound ? "on" : "off");
  ensureAudio();
  if (state.sound) {
    playMove();
    syncSetupMusic();
  } else {
    syncSetupMusic();
  }
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
const initialParams = new URLSearchParams(location.search);
const initialWatchCode = initialParams.get("watch");
const initialJoinCode = initialParams.get("join");
if (initialWatchCode) {
  connectViewer(initialWatchCode);
} else if (initialJoinCode) {
  joinFromInvite(initialJoinCode);
} else if (state.code && state.token) {
  connect();
}
