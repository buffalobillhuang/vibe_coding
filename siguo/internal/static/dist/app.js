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
const pieceMarkerValues = ["?", "+", "++", "+++", "!", "!!", "!!!"];
const pieceMarkerActions = [...pieceMarkerValues, "unmark"];
const rulesCopy = {
  hans: {
    dialogTitle: "游戏规则",
    dialogIntro: "适用于四国 2v2 与军棋 1v1。",
    currentModeLabel: "当前模式",
    closeLabel: "关闭",
    modeCompareTitle: "模式区别",
    sections: [
      {
        title: "通用基础",
        items: [
          "双方先完成布阵，全部提交后开始对局。",
          "棋子可以放在兵站和大本营；行营是安全格，驻入后不能被直接攻击。",
          "军旗和地雷不能移动；进入大本营的棋子也不能再移动。"
        ]
      },
      {
        title: "走法与吃子",
        items: [
          "公路线每次只能走一步。",
          "铁路线无阻挡时，工兵可沿铁路连续转弯行走；其他棋子只能沿铁路直线走任意步。",
          "山界为特殊通道，只有工兵可以进入；每个山界位置最多容纳一名工兵。工兵进入山界后，可以在相连山界间穿行，并可接入相邻可达的铁路；若山界内已有敌方工兵，则进入后立即交战。",
          "军衔大小依次为：司令 > 军长 > 师长 > 旅长 > 团长 > 营长 > 连长 > 排长 > 工兵。",
          "工兵可以排雷；炸弹与任意棋子相遇都会同归于尽；同级相遇也会同归于尽。",
          "除工兵和炸弹外，其他棋子撞上地雷会被消灭，地雷保留。"
        ]
      },
      {
        title: "布阵限制",
        items: [
          "炸弹不能放在第一行。",
          "地雷只能放在最后两行。",
          "军旗只能放在大本营。"
        ]
      },
      {
        title: "四国 2v2",
        items: [
          "北南为一队，东西为一队。",
          "按照轮次由当前座位行动，队友共享胜负。",
          "夺取敌方联军军旗，或令对方联军无棋可走，即为联军获胜。"
        ]
      },
      {
        title: "军棋 1v1",
        items: [
          "红蓝双方直接对抗，没有队友协作。",
          "夺取对方军旗，或令对方无棋可走，即可获胜。",
          "对局中可使用跳过、求和、投降等操作。"
        ]
      },
      {
        title: "观战说明",
        items: [
          "观战者只能查看局面与聊天，不能移动棋子或发起对局操作。"
        ]
      }
    ]
  },
  hant: {
    dialogTitle: "遊戲規則",
    dialogIntro: "適用於四國 2v2 與軍棋 1v1。",
    currentModeLabel: "目前模式",
    closeLabel: "關閉",
    modeCompareTitle: "模式差異",
    sections: [
      {
        title: "通用基礎",
        items: [
          "雙方先完成佈陣，全部提交後開始對局。",
          "棋子可以放在兵站和大本營；行營是安全格，駐入後不能被直接攻擊。",
          "軍旗和地雷不能移動；進入大本營的棋子也不能再移動。"
        ]
      },
      {
        title: "走法與吃子",
        items: [
          "公路線每次只能走一步。",
          "鐵路線無阻擋時，工兵可沿鐵路連續轉彎行走；其他棋子只能沿鐵路直線走任意步。",
          "山界為特殊通道，只有工兵可以進入；每個山界位置最多容納一名工兵。工兵進入山界後，可以在相連山界間穿行，並可接入相鄰可達的鐵路；若山界內已有敵方工兵，則進入後立即交戰。",
          "軍銜大小依次為：司令 > 軍長 > 師長 > 旅長 > 團長 > 營長 > 連長 > 排長 > 工兵。",
          "工兵可以排雷；炸彈與任意棋子相遇都會同歸於盡；同級相遇也會同歸於盡。",
          "除工兵和炸彈外，其他棋子撞上地雷會被消滅，地雷保留。"
        ]
      },
      {
        title: "佈陣限制",
        items: [
          "炸彈不能放在第一行。",
          "地雷只能放在最後兩行。",
          "軍旗只能放在大本營。"
        ]
      },
      {
        title: "四國 2v2",
        items: [
          "北南為一隊，東西為一隊。",
          "按照輪次由當前座位行棋，隊友共享勝負。",
          "奪取敵方聯軍軍旗，或令對方聯軍無棋可走，即為聯軍獲勝。"
        ]
      },
      {
        title: "軍棋 1v1",
        items: [
          "紅藍雙方直接對抗，沒有隊友協作。",
          "奪取對方軍旗，或令對方無棋可走，即可獲勝。",
          "對局中可使用跳過、求和、投降等操作。"
        ]
      },
      {
        title: "觀戰說明",
        items: [
          "觀戰者只能查看局面與聊天，不能移動棋子或發起對局操作。"
        ]
      }
    ]
  }
};
const socketWatchdogIntervalMs = 1000;
const socketHeartbeatTimeoutMs = 15000;
const socketStalenessMs = 13000;

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
  lastTrail: null,
  selectedMarker: null,
  pieceMarks: {},
  rulesOpen: false,
  rulesLang: localStorage.getItem("siguo.rulesLang") || "hans",
  boardChats: [],
  boardChatLane: 0,
  boardChatTimer: null,
  log: [],
  chat: [],
  lowTimeWarned: false,
  viewer: false,
  watchOpen: false,
  watchRooms: [],
  joinOffer: null,
  roomGoneOffer: null,
  cultureImageLoaded: false,
  cultureImageRequested: false,
  cultureImageLoader: null,
  connectionStatus: "idle",
  connectionMessage: "",
  reconnectAttempts: 0,
  reconnectTimer: null,
  lastSocketMessageAt: 0,
  socketGeneration: 0
};

const reconnectDelaysMs = [1000, 2000, 5000, 10000];
const imageWarmTimeoutMs = 30000;

const app = document.querySelector("#app");

function formatLoadDuration(ms) {
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`;
  return `${Math.round(ms)}ms`;
}

function render() {
  app.innerHTML = `
    <div class="shell">
      <section class="table">
        <div class="topbar">
          <div class="brand"><h1>四国军棋</h1><button id="rulesBtn" class="rules-link" type="button">游戏规则</button><span id="roomStatus">${statusText()}</span></div>
          <div class="row" style="max-width:720px">
            <button id="watchBtn">观战室</button>
            ${inviteLinkButtonHTML()}
            ${viewerLinkButtonHTML()}
            <button class="primary" id="startBtn" ${canStart() ? "" : "disabled"}>开始</button>
            <button id="randBtn" ${!state.viewer && state.room?.phase === "setup" ? "" : "disabled"}>随机</button>
            <button id="submitBtn" ${!state.viewer && state.room?.phase === "setup" ? "" : "disabled"}>提交</button>
            ${actionButtonsHTML()}
            ${reconnectButtonHTML()}
            <button id="soundBtn" class="toggle ${state.sound ? "on" : "off"}">声音</button>
          </div>
        </div>
        ${rulesDialogHTML()}
        ${watchRoomPanelHTML()}
        ${joinOfferHTML()}
        ${roomGoneOfferHTML()}
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
          <b>聊天</b>
          <div class="chat-log">${state.chat.slice(-80).reverse().map(chatLine).join("")}</div>
          ${state.room ? `<div class="row">
            <input id="chatText" maxlength="200" placeholder="输入消息" />
            <button id="sendAll">公屏</button>
            ${state.viewer || currentMode() === "junqi" ? "" : `<button id="sendTeam">队伍（仅队友可见）</button>`}
          </div>
          ${quickChatHTML()}` : ""}
        </div>
        <div class="panel stack">
          <b>战况</b>
          <div class="log">${state.log.slice(-80).reverse().map(x => `<div class="line">${esc(x)}</div>`).join("")}</div>
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
  const connection = connectionStatusText();
  return `${role} · ${modeNames[currentMode()]} · ${phase}${setup}${turn}${connection ? ` · ${connection}` : ""}`;
}

function connectionStatusText() {
  if (!state.room) return "";
  if (state.connectionMessage) return state.connectionMessage;
  if (state.connectionStatus === "connecting") return "连接中";
  if (state.connectionStatus === "reconnecting") return "重连中";
  if (state.connectionStatus === "offline") return "已断线";
  if (socketLooksStale()) return "连接可能卡住";
  return "";
}

function socketLooksStale() {
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN) return false;
  if (!state.lastSocketMessageAt) return false;
  return Date.now() - state.lastSocketMessageAt > socketStalenessMs;
}

function forceReconnect(viewer) {
  if (state.roomGoneOffer) return;
  clearReconnectTimer();
  if (state.ws) {
    state.ws.onclose = null;
    state.ws.close();
  }
  state.ws = null;
  state.lastSocketMessageAt = 0;
  state.selected = null;
  state.reconnectAttempts = 0;
  state.connectionStatus = "reconnecting";
  state.connectionMessage = "连接卡住，正在重连";
  openSocket(viewer, true);
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

function absoluteWinnerLabel() {
  const winners = winnerSeats();
  if (!winners.length) return "和棋";
  if (currentMode() === "junqi") return junqiSeatColor(winners[0]) === "red" ? "红方获胜" : "蓝方获胜";
  return sameTeam(winners[0], 0) ? "北南联军获胜" : "东西联军获胜";
}

function currentSideWon() {
  const winners = winnerSeats();
  if (!winners.length || state.viewer) return false;
  return sameSide(winners[0], state.seat);
}

function winnerLabel() {
  const winners = winnerSeats();
  if (!winners.length) return "和棋";
  if (state.viewer) return absoluteWinnerLabel();
  return currentSideWon() ? "我方获胜" : "对方获胜";
}

function victoryKicker() {
  if (state.viewer) return "军旗已定";
  return currentSideWon() ? "正义之师 所向披靡" : "胜负乃兵家常事";
}

function victorySubline() {
  if (state.viewer) {
    if (currentMode() === "junqi") return "一局定鼎";
    return "联军得胜";
  }
  if (currentMode() === "junqi") return currentSideWon() ? "我方夺旗得胜" : "对方夺旗得胜";
  return currentSideWon() ? "我方联军得胜" : "对方联军得胜";
}

function victoryToneClass() {
  const winners = winnerSeats();
  if (!winners.length) return "victory-neutral";
  if (state.viewer) {
    if (currentMode() === "junqi") return `victory-${junqiSeatColor(winners[0])}`;
    return sameTeam(winners[0], 0) ? "victory-ns" : "victory-ew";
  }
  if (currentMode() === "junqi") return `victory-${junqiSeatColor(state.seat)}`;
  return sameTeam(state.seat, 0) ? "victory-ns" : "victory-ew";
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
      <span class="victory-kicker">${victoryKicker()}</span>
      <b>${winnerLabel()}</b>
      <small>${victorySubline()}</small>
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
  warmCultureImage();
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
    <div class="poem-window ${state.cultureImageLoaded ? "is-loaded" : "is-loading"}">
      <div class="poem-window-art" aria-hidden="true"></div>
      <div class="poem-scroll">
        ${poemColumns}${poemColumns}
      </div>
    </div>
  </div>`;
}

function warmCultureImage() {
  if (state.cultureImageLoaded || state.cultureImageRequested) return;
  state.cultureImageRequested = true;
  const startedAt = performance.now();
  const img = new Image();
  state.cultureImageLoader = img;
  const timeoutId = setTimeout(() => {
    if (state.cultureImageLoader !== img) return;
    state.cultureImageLoader = null;
    state.cultureImageRequested = false;
    img.src = "";
    log(`诗图加载超时（${formatLoadDuration(imageWarmTimeoutMs)}），重试中`);
  }, imageWarmTimeoutMs);
  img.decoding = "async";
  img.onload = () => {
    if (state.cultureImageLoader !== img) return;
    clearTimeout(timeoutId);
    state.cultureImageLoader = null;
    state.cultureImageRequested = false;
    state.cultureImageLoaded = true;
    log(`诗图加载完成（${formatLoadDuration(performance.now() - startedAt)}）`);
  };
  img.onerror = () => {
    if (state.cultureImageLoader !== img) return;
    clearTimeout(timeoutId);
    state.cultureImageLoader = null;
    state.cultureImageRequested = false;
    log("诗图加载失败，重试中");
  };
  img.src = "/picture02.png";
}

function quickChatHTML() {
  return `<div class="quick-chat" aria-label="常用语">
    ${quickChatPhrases.map(text => `<button type="button" class="quick-chat-btn" data-phrase="${esc(text)}">${esc(text)}</button>`).join("")}
  </div>`;
}

function viewerLinkButtonHTML() {
  if (!state.room || !["lobby", "setup", "playing"].includes(state.room.phase)) return "";
  return `<button id="viewerLinkBtn">观战链接</button>`;
}

function inviteLinkButtonHTML() {
  if (state.viewer || !state.room || !state.code || !["lobby", "setup", "playing"].includes(state.room.phase)) return "";
  return `<button id="inviteLinkBtn">邀请链接</button>`;
}

function reconnectButtonHTML() {
  if (!state.code) return "";
  if (state.roomGoneOffer) return "";
  const urgent = socketLooksStale() || state.connectionStatus === "reconnecting" || state.connectionStatus === "offline" || (state.ws && state.ws.readyState !== WebSocket.OPEN);
  return `<button id="reconnectBtn" class="${urgent ? "urgent" : ""}" title="立即重连">重连</button>`;
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

function currentRulesCopy() {
  return rulesCopy[state.rulesLang] || rulesCopy.hans;
}

function rulesDialogHTML() {
  if (!state.rulesOpen) return "";
  const copy = currentRulesCopy();
  const modeSections = [
    {key: "siguo", section: copy.sections[3]},
    {key: "junqi", section: copy.sections[4]}
  ];
  const sharedSections = [copy.sections[0], copy.sections[1], copy.sections[2], copy.sections[5]];
  return `<div class="rules-overlay" id="rulesOverlay">
    <div class="rules-dialog panel" role="dialog" aria-modal="true" aria-labelledby="rulesTitle">
      <div class="rules-head">
        <div class="rules-title-block">
          <b id="rulesTitle">${esc(copy.dialogTitle)}</b>
          <span>${esc(copy.dialogIntro)}</span>
        </div>
        <button id="rulesClose" type="button">${esc(copy.closeLabel)}</button>
      </div>
      <div class="rules-toolbar">
        <div class="rules-tabs" aria-label="语言切换">
          <button id="rulesLangHans" type="button" class="toggle ${state.rulesLang === "hans" ? "on" : "off"}">简体</button>
          <button id="rulesLangHant" type="button" class="toggle ${state.rulesLang === "hant" ? "on" : "off"}">繁體</button>
        </div>
        <div class="rules-mode">${esc(copy.currentModeLabel)}：${esc(modeNames[currentMode()])}</div>
      </div>
      <div class="rules-body">
        <section class="rules-compare">
          <h3>${esc(copy.modeCompareTitle)}</h3>
          <div class="rules-mode-grid">
            ${modeSections.map(({key, section}) => `<section class="rules-mode-card ${currentMode() === key ? "current" : ""}"><div class="rules-mode-card-head"><h4>${esc(section.title)}</h4>${currentMode() === key ? `<span class="rules-mode-badge">${esc(copy.currentModeLabel)}</span>` : ""}</div><ul>${section.items.map(item => `<li>${esc(item)}</li>`).join("")}</ul></section>`).join("")}
          </div>
        </section>
        ${sharedSections.map(section => `<section class="rules-section"><h3>${esc(section.title)}</h3><ul>${section.items.map(item => `<li>${esc(item)}</li>`).join("")}</ul></section>`).join("")}
      </div>
    </div>
  </div>`;
}

function roomGoneOfferHTML() {
  const offer = state.roomGoneOffer;
  if (!offer) return "";
  const note = offer.lastError ? `<span class="subtle">${esc(offer.lastError)}</span>` : "";
  return `<div class="panel room-gone-offer">
    <div><b>对局 ${esc(offer.code)} 已结束或不存在</b><span>可以尝试重新进入；如果对局已结束，请返回大厅。</span>${note}</div>
    <div><button id="roomGoneLeave">返回大厅</button><button id="roomGoneRejoin" class="primary">重新进入</button></div>
  </div>`;
}

function watchRoomPanelHTML() {
  if (!state.watchOpen) return "";
  const rooms = state.watchRooms || [];
  const rows = rooms.length ? rooms.map(room => {
    const names = (room.seats || []).map(s => s.name || seatNames[s.seat]).join(" · ");
    return `<div class="watch-row">
      <div><b>${esc(room.code)}</b><span>${modeNames[room.mode] || room.mode} · ${watchPhaseLabel(room.phase)} · ${room.viewers}/${room.maxViewers} 观战</span><small>${esc(names)}</small></div>
      <button class="watch-join" data-code="${esc(room.code)}" ${room.canJoinView ? "" : "disabled"}>观看</button>
      <button class="watch-copy" data-code="${esc(room.code)}">复制链接</button>
    </div>`;
  }).join("") : `<div class="subtle">暂无进行中的对局</div>`;
  return `<div class="panel watch-panel">
    <div class="watch-head"><b>观战室</b><div><button id="watchRefreshBtn">刷新</button><button id="watchCloseBtn">关闭</button></div></div>
    <div class="watch-list">${rows}</div>
  </div>`;
}

function watchPhaseLabel(phase) {
  if (phase === "lobby") return "大厅";
  if (phase === "setup") return "布阵";
  return "对局";
}

function seatsHTML() {
  const seats = state.room?.seats || activeSeats().map(seat => ({seat}));
  const canSwap = canSwapSeats();
  return seats.map(s => `
    <div class="seat ${state.room?.turn === s.seat ? "current" : ""}" data-seat="${s.seat}">
      <b>${seatNames[s.seat]}</b>
      <span class="subtle">${s.name ? esc(s.name) : "空位"} ${s.host ? "房主" : ""} ${s.ready ? "已提交" : ""}</span>
      ${canSwap && s.name && s.seat !== state.seat ? `<button type="button" class="seat-swap-btn" data-seat="${s.seat}">换位</button>` : ""}
    </div>
  `).join("");
}

function canSwapSeats() {
  return !state.viewer && currentMode() === "siguo" && state.room?.phase === "lobby" && (state.room.seats || []).filter(s => s.name).length === 4;
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

function viewCells() {
  return state.view?.Cells || state.view?.cells || [];
}

function pieceId(piece) {
  return piece?.ID ?? piece?.id ?? 0;
}

function boardHTML() {
  prunePieceMarks();
  const cells = new Map(viewCells().map(c => [`${c.Pos?.Row ?? c.pos.row},${c.Pos?.Col ?? c.pos.col}`, c]));
  const mode = currentMode();
  const rows = mode === "junqi" ? 13 : 17;
  const cols = mode === "junqi" ? 5 : 17;
  const stageClass = mode === "junqi" ? "board-stage board-stage-junqi" : "board-stage";
  const boardOpen = mode === "junqi"
    ? `<div class="board-surface-wrap board-surface-wrap-junqi">${playerTickersHTML()}<div class="board-clip board-clip-junqi"><div class="board board-${mode} board-junqi-rel-${state.seat}">${railOverlayHTML()}${turnFrontOverlayHTML()}${moveTrailOverlayHTML()}${boardChatOverlayHTML()}`
    : `<div class="board board-${mode} board-siguo-rel-${state.seat}">${railOverlayHTML()}${turnFrontOverlayHTML()}${moveTrailOverlayHTML()}${playerTickersHTML()}${boardChatOverlayHTML()}`;
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
      html += `<div class="cell ${cellClasses[type] || "off"} ${homeClass(row, col)} ${visualTrackClass(row, col)} ${piece ? "occupied" : ""} ${selected ? "selected" : ""} ${combat ? "combat-hit" : ""}" data-row="${row}" data-col="${col}" ${cellStyle}>`;
      if (piece) {
        const owner = piece.Owner ?? piece.owner;
        const rank = piece.Rank ?? piece.rank;
        const exposed = piece.Exposed ?? piece.exposed;
        const colorClass = pieceColorClass(owner);
        const exposedClass = exposed && rank !== 0 ? "flag-exposed" : "";
        const hiddenClass = rank === 0 ? "piece-hidden" : "";
        const orientClass = `piece-rel-${relativeSeat(owner)}`;
        const label = rank === 0 ? "" : (rankNames[rank] || "?");
        const marker = state.pieceMarks[pieceId(piece)];
        html += `<div class="piece ${colorClass} ${hiddenClass} ${orientClass} ${exposedClass}" data-piece="${pieceId(piece)}" data-owner="${owner}">${label}${pieceMarkHTML(marker)}</div>`;
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

function pieceMarkHTML(marker) {
  if (!marker) return "";
  const longClass = marker.length > 3 ? " piece-mark-long" : "";
  return `<span class="piece-mark${longClass}">${esc(marker)}</span>`;
}

function prunePieceMarks() {
  const alive = new Set();
  viewCells().forEach(c => {
    const id = pieceId(c?.Piece || c?.piece);
    if (id) alive.add(String(id));
  });
  Object.keys(state.pieceMarks).forEach(id => {
    if (!alive.has(id)) delete state.pieceMarks[id];
  });
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

function moveTrailOverlayHTML() {
  const trail = state.lastTrail;
  if (!trail || trail.mode !== currentMode() || !Array.isArray(trail.path) || trail.path.length < 2) return "";
  const points = trail.path.map(trailPoint).filter(Boolean);
  if (points.length < 2) return "";
  const polyline = points.map(p => `${p.x.toFixed(2)},${p.y.toFixed(2)}`).join(" ");
  const dots = points.map((p, idx) => {
    const cls = idx === points.length - 1 ? "move-trail-dot move-trail-dot-end" : "move-trail-dot";
    return `<circle class="${cls}" cx="${p.x.toFixed(2)}" cy="${p.y.toFixed(2)}" r="${idx === points.length - 1 ? "1.00" : ".72"}" />`;
  }).join("");
  return `<svg class="move-trail" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
    <defs>
      <marker id="moveTrailArrow" markerWidth="7" markerHeight="7" refX="5.6" refY="3.5" orient="auto" markerUnits="strokeWidth">
        <path d="M .6 .6 L 6 3.5 L .6 6.4 Z" />
      </marker>
    </defs>
    <polyline points="${polyline}" />
    ${dots}
  </svg>`;
}

function boardChatOverlayHTML() {
  pruneBoardChats();
  const items = state.boardChats.map(chat => {
    const elapsed = Date.now() - chat.createdAt;
    const top = `${chat.top.toFixed(1)}%`;
    const channelClass = chat.channel === "team" ? "team" : "public";
    const cls = chat.system ? `board-chat-item ${channelClass} system` : chat.viewer ? `board-chat-item ${channelClass} viewer` : `board-chat-item ${channelClass}`;
    return `<div class="${cls}" style="top:${top};animation-duration:${chat.durationMs}ms;animation-delay:-${Math.max(0, elapsed)}ms">${esc(chat.text)}</div>`;
  }).join("");
  return `<div class="board-chat-layer" aria-hidden="true">${items}</div>`;
}

function turnFrontOverlayHTML() {
  const geom = turnFrontGeometry();
  if (!geom) return "";
  return `<div class="turn-front-dots" style="left:${geom.x.toFixed(2)}%;top:${geom.y.toFixed(2)}%;width:${geom.width.toFixed(2)}%;height:${geom.height.toFixed(2)}%;transform:translate(-50%, -50%) rotate(${geom.angle.toFixed(2)}deg)"><span class="turn-front-dot"></span><span class="turn-front-dot"></span></div>`;
}

function turnFrontGeometry() {
  if (state.room?.phase !== "playing") return null;
  const turn = Number(state.room.turn);
  const spec = currentMode() === "junqi" ? junqiTurnFrontSpec(turn) : siguoTurnFrontSpec(turn);
  if (!spec) return null;
  const a = trailPoint(spec.a);
  const b = trailPoint(spec.b);
  const mid = trailPoint(spec.mid);
  const front = trailPoint(spec.front);
  if (!a || !b || !mid || !front) return null;

  const frontDx = front.x - mid.x;
  const frontDy = front.y - mid.y;
  const frontDistance = Math.hypot(frontDx, frontDy) || 1;
  const unitX = frontDx / frontDistance;
  const unitY = frontDy / frontDistance;
  const lineDx = b.x - a.x;
  const lineDy = b.y - a.y;
  const mode = currentMode();
  const dotDiameter = mode === "junqi" ? 3.2 : 2.95;
  const halfPieceHeight = mode === "junqi" ? 2.0 : 2.3;
  const offset = halfPieceHeight + dotDiameter / 2;

  return {
    x: mid.x + unitX * offset,
    y: mid.y + unitY * offset,
    width: dotDiameter * 4.2,
    height: dotDiameter,
    angle: Math.atan2(lineDy, lineDx) * 180 / Math.PI
  };
}

function siguoTurnFrontSpec(turn) {
  switch (turn) {
    case 0: return {a: {Row: 5, Col: 6}, b: {Row: 5, Col: 10}, mid: {Row: 5, Col: 8}, front: {Row: 6, Col: 8}};
    case 1: return {a: {Row: 6, Col: 11}, b: {Row: 10, Col: 11}, mid: {Row: 8, Col: 11}, front: {Row: 8, Col: 10}};
    case 2: return {a: {Row: 11, Col: 6}, b: {Row: 11, Col: 10}, mid: {Row: 11, Col: 8}, front: {Row: 10, Col: 8}};
    case 3: return {a: {Row: 6, Col: 5}, b: {Row: 10, Col: 5}, mid: {Row: 8, Col: 5}, front: {Row: 8, Col: 6}};
    default: return null;
  }
}

function junqiTurnFrontSpec(turn) {
  switch (turn) {
    case 0: return {a: {Row: 7, Col: 6}, b: {Row: 7, Col: 10}, mid: {Row: 7, Col: 8}, front: {Row: 8, Col: 8}};
    case 2: return {a: {Row: 9, Col: 6}, b: {Row: 9, Col: 10}, mid: {Row: 9, Col: 8}, front: {Row: 8, Col: 8}};
    default: return null;
  }
}

function trailPoint(pos) {
  const row = pos.Row ?? pos.row;
  const col = pos.Col ?? pos.col;
  if (!Number.isFinite(row) || !Number.isFinite(col)) return null;
  const display = toDisplay(row, col);
  if (currentMode() === "junqi") {
    const nudge = junqiBackRowNudge(row);
    return {
      x: junqiMap3Centers.x[display.col] + nudge.x,
      y: junqiMap3Centers.y[display.row] + nudge.y
    };
  }
  return {
    x: siguoMap8Centers.x[display.col],
    y: siguoMap8Centers.y[display.row]
  };
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

function markerPickerHTML(seat, isElim) {
  if (state.viewer || seat !== state.seat || isElim || !["setup", "playing"].includes(state.room?.phase)) return "";
  return `<div class="marker-picker" aria-label="棋子标记">
    ${pieceMarkerActions.map(marker => `<button type="button" class="marker-choice ${state.selectedMarker === marker ? "on" : ""}" data-marker="${esc(marker)}">${esc(marker)}</button>`).join("")}
  </div>`;
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
      ${markerPickerHTML(seat, isElim)}
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
  const rulesBtn = document.querySelector("#rulesBtn");
  if (rulesBtn) rulesBtn.onclick = () => { state.rulesOpen = true; render(); };
  const rulesClose = document.querySelector("#rulesClose");
  if (rulesClose) rulesClose.onclick = closeRules;
  const rulesOverlay = document.querySelector("#rulesOverlay");
  if (rulesOverlay) rulesOverlay.onclick = e => {
    if (e.target === rulesOverlay) closeRules();
  };
  const rulesLangHans = document.querySelector("#rulesLangHans");
  if (rulesLangHans) rulesLangHans.onclick = () => setRulesLang("hans");
  const rulesLangHant = document.querySelector("#rulesLangHant");
  if (rulesLangHant) rulesLangHant.onclick = () => setRulesLang("hant");
  const inviteLinkBtn = document.querySelector("#inviteLinkBtn");
  if (inviteLinkBtn) inviteLinkBtn.onclick = () => copyInviteLink(state.code);
  const viewerLinkBtn = document.querySelector("#viewerLinkBtn");
  if (viewerLinkBtn) viewerLinkBtn.onclick = () => copyViewerLink(state.code);
  const joinOfferView = document.querySelector("#joinOfferView");
  if (joinOfferView) joinOfferView.onclick = () => connectViewer(state.joinOffer?.code);
  const joinOfferClose = document.querySelector("#joinOfferClose");
  if (joinOfferClose) joinOfferClose.onclick = () => { state.joinOffer = null; render(); };
  const roomGoneRejoin = document.querySelector("#roomGoneRejoin");
  if (roomGoneRejoin) roomGoneRejoin.onclick = rejoinAfterGone;
  const roomGoneLeave = document.querySelector("#roomGoneLeave");
  if (roomGoneLeave) roomGoneLeave.onclick = dismissRoomGone;
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
  const reconnectBtn = document.querySelector("#reconnectBtn");
  if (reconnectBtn) reconnectBtn.onclick = () => {
    log("手动重连");
    forceReconnect(state.viewer);
  };
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
  document.querySelectorAll(".seat-swap-btn").forEach(btn => btn.onclick = () => swapSeat(btn.dataset.seat));
  document.querySelectorAll(".marker-choice").forEach(btn => btn.onclick = e => {
    e.stopPropagation();
    selectPieceMarker(btn.dataset.marker);
  });
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

function closeRules() {
  if (!state.rulesOpen) return;
  state.rulesOpen = false;
  render();
}

function setRulesLang(lang) {
  if (!rulesCopy[lang] || state.rulesLang === lang) return;
  state.rulesLang = lang;
  localStorage.setItem("siguo.rulesLang", lang);
  render();
}

function confirmAction(message, fn) {
  if (typeof window !== "undefined" && window.confirm) {
    if (!window.confirm(message)) return;
  }
  fn();
}

async function createRoom() {
  state.name = document.querySelector("#name").value.trim();
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
  const inviteName = window.prompt("给被邀请玩家取个名字（可留空，由系统分配）", "") || "";
  return shareURL("join", code, inviteName.trim() ? {name: inviteName.trim()} : null);
}

function shareURL(param, code, extras = null) {
  const origin = shareOrigin();
  if (!origin) return "";
  const url = new URL(location.pathname, origin);
  url.searchParams.set(param, code);
  if (extras) {
    for (const [key, value] of Object.entries(extras)) {
      url.searchParams.set(key, value);
    }
  }
  return url.href;
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
    log(`邀请链接已复制：${url}`);
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
    log(`观战链接已复制：${url}`);
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
  clearReconnectTimer();
  if (state.ws) {
    state.ws.onclose = null;
    state.ws.close();
  }
  state.ws = null;
  state.lastSocketMessageAt = 0;
  state.connectionStatus = "idle";
  state.connectionMessage = "";
  state.reconnectAttempts = 0;
  state.room = null;
  state.view = null;
  state.code = "";
  state.token = "";
  state.host = false;
  state.viewer = false;
  state.watchOpen = false;
  state.joinOffer = null;
  state.selected = null;
  state.selectedMarker = null;
  state.pieceMarks = {};
  state.combat = null;
  localStorage.removeItem("siguo.code");
  localStorage.removeItem("siguo.token");
  if (location.search.includes("watch=") || location.search.includes("join=")) history.replaceState(null, "", location.pathname);
  if (shouldRender) render();
}

async function joinRoom() {
  state.name = document.querySelector("#name").value.trim();
  const code = document.querySelector("#code").value.toUpperCase();
  const sessionToken = sessionTokenForRoom(code);
  state.code = code;
  const res = await fetch(`/api/rooms/${code}/join`, {method:"POST", body: JSON.stringify({name: state.name, sessionToken})});
  await acceptJoin(res);
}

function sessionTokenForRoom(code) {
  const normalizedCode = String(code || "").toUpperCase();
  if (!normalizedCode || state.viewer) return "";
  const savedCode = String(localStorage.getItem("siguo.code") || "").toUpperCase();
  return state.token && savedCode === normalizedCode ? state.token : "";
}

async function joinFromInvite(code) {
  code = String(code || "").toUpperCase();
  if (!code) return;
  const sessionToken = sessionTokenForRoom(code);
  state.code = code;
  const res = await fetch(`/api/rooms/${code}/join`, {method:"POST", body: JSON.stringify({name: state.name, sessionToken})});
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
  state.name = data.name || state.name;
  state.host = data.host;
  state.viewer = false;
  state.watchOpen = false;
  state.joinOffer = null;
  state.roomGoneOffer = null;
  state.selectedMarker = null;
  state.pieceMarks = {};
  localStorage.setItem("siguo.code", state.code);
  localStorage.setItem("siguo.token", state.token);
  localStorage.setItem("siguo.seat", String(state.seat));
  localStorage.setItem("siguo.name", state.name);
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
  openSocket(false);
}

function beginViewerSession(code) {
  if (state.ws) {
    state.ws.onclose = null;
    state.ws.close();
  }
  state.viewer = true;
  state.watchOpen = false;
  state.joinOffer = null;
  state.roomGoneOffer = null;
  state.code = code;
  state.token = "";
  state.host = false;
  state.seat = 2;
  state.selected = null;
  state.selectedMarker = null;
  state.pieceMarks = {};
  state.combat = null;
  localStorage.removeItem("siguo.code");
  localStorage.removeItem("siguo.token");
  history.replaceState(null, "", `?watch=${encodeURIComponent(state.code)}`);
  openSocket(true);
}

async function connectViewer(code) {
  code = String(code || "").toUpperCase();
  const nameInput = document.querySelector("#name");
  const status = await viewerRoomStatus(code);
  if (!status.ok) {
    log(status.message);
    return;
  }
  let promptMessage = "请输入观战昵称";
  let suggestedName = defaultViewerName(status.room, nameInput ? nameInput.value : state.name || "");
  while (true) {
    const input = window.prompt(promptMessage, suggestedName);
    if (input === null) return;
    const viewerName = input.trim();
    if (!viewerName) {
      promptMessage = "观战昵称不能为空，请重新输入";
      continue;
    }
    const viewerStatus = await viewerRoomStatus(code, viewerName);
    if (!viewerStatus.ok) {
      if (viewerStatus.message === "当前昵称已在该对局中落座，不能同时观战") {
        promptMessage = viewerStatus.message;
        suggestedName = viewerName;
        continue;
      }
      log(viewerStatus.message);
      return;
    }
    state.name = viewerName;
    if (nameInput) nameInput.value = viewerName;
    localStorage.setItem("siguo.name", state.name);
    break;
  }
  beginViewerSession(code);
}

function defaultViewerName(room, preferredName) {
  const viewerName = String(preferredName || "").trim();
  if (viewerName) return viewerName;
  return `观众${Math.max(1, Number(room?.viewers || 0) + 1)}`;
}

function openSocket(viewer, isReconnect = false) {
  clearReconnectTimer();
  if (state.ws) {
    state.ws.onclose = null;
    state.ws.close();
  }
  const socketGeneration = ++state.socketGeneration;
  const proto = location.protocol === "https:" ? "wss" : "ws";
  const query = viewer
    ? `room=${state.code}&viewer=1&name=${encodeURIComponent(state.name)}`
    : `room=${state.code}&token=${encodeURIComponent(state.token)}`;
  const ws = new WebSocket(`${proto}://${location.host}/ws?${query}`);
  state.ws = ws;
  state.lastSocketMessageAt = 0;
  state.connectionStatus = isReconnect ? "reconnecting" : "connecting";
  state.connectionMessage = isReconnect ? "正在重连" : "连接中";
  render();
  ws.onopen = () => {
    if (state.ws !== ws || state.socketGeneration !== socketGeneration) return;
    state.lastSocketMessageAt = Date.now();
    state.connectionStatus = "connected";
    state.connectionMessage = "";
    state.reconnectAttempts = 0;
    log(viewer ? (isReconnect ? "观战已重连" : "已进入观战") : (isReconnect ? "已重连" : "已连接"));
  };
  ws.onclose = () => {
    if (state.ws !== ws || state.socketGeneration !== socketGeneration) return;
    state.ws = null;
    state.lastSocketMessageAt = 0;
    state.selected = null;
    scheduleReconnect(viewer);
  };
  ws.onerror = () => {
    if (state.ws === ws && state.socketGeneration === socketGeneration) ws.close();
  };
  ws.onmessage = e => {
    if (state.ws !== ws || state.socketGeneration !== socketGeneration) return;
    state.lastSocketMessageAt = Date.now();
    onMessage(JSON.parse(e.data));
  };
}

function clearReconnectTimer() {
  if (!state.reconnectTimer) return;
  clearTimeout(state.reconnectTimer);
  state.reconnectTimer = null;
}

function scheduleReconnect(viewer) {
  if (state.roomGoneOffer) return;
  if (!state.code || (!viewer && !state.token)) return;
  const delay = reconnectDelaysMs[Math.min(state.reconnectAttempts, reconnectDelaysMs.length - 1)];
  const shouldLogDisconnect = state.connectionStatus !== "reconnecting";
  state.reconnectAttempts += 1;
  state.connectionStatus = "reconnecting";
  state.connectionMessage = state.reconnectAttempts > 6 ? "连接断开，请确认房间还在" : `连接断开，${Math.ceil(delay / 1000)}秒后重连`;
  if (shouldLogDisconnect) {
    log(viewer ? "观战已断开，正在重连" : "连接已断开，正在重连");
  }
  state.reconnectTimer = setTimeout(() => attemptReconnect(viewer), delay);
}

async function attemptReconnect(viewer) {
  state.reconnectTimer = null;
  const reconnectGeneration = state.socketGeneration;
  const reconnectCode = state.code;
  const reconnectToken = state.token;
  const reconnectViewer = state.viewer;
  if (state.reconnectAttempts >= 3) {
    const alive = await probeRoomAlive();
    if (state.socketGeneration !== reconnectGeneration || state.code !== reconnectCode || state.token !== reconnectToken || state.viewer !== reconnectViewer || state.roomGoneOffer) {
      return;
    }
    if (!alive) {
      handleRoomGone();
      return;
    }
  }
  if (state.socketGeneration !== reconnectGeneration || state.code !== reconnectCode || state.token !== reconnectToken || state.viewer !== reconnectViewer || state.roomGoneOffer) {
    return;
  }
  openSocket(viewer, true);
}

async function probeRoomAlive() {
  if (!state.code) return true;
  try {
    if (state.viewer) {
      const res = await fetch("/api/rooms");
      if (!res.ok) return true;
      const data = await res.json();
      return (data.rooms || []).some(r => String(r.code).toUpperCase() === String(state.code).toUpperCase());
    }
    if (!state.token) return true;
    const res = await fetch(`/api/rooms/${encodeURIComponent(state.code)}?token=${encodeURIComponent(state.token)}`);
    if (res.status === 404 || res.status === 401) return false;
    return true;
  } catch {
    return true;
  }
}

function handleRoomGone() {
  if (state.roomGoneOffer) return;
  clearReconnectTimer();
  if (state.ws) {
    state.ws.onclose = null;
    state.ws.close();
  }
  state.ws = null;
  state.lastSocketMessageAt = 0;
  state.reconnectAttempts = 0;
  state.selected = null;
  state.connectionStatus = "offline";
  state.connectionMessage = "对局已结束或不存在";
  state.roomGoneOffer = {code: state.code, isViewer: state.viewer, lastError: ""};
  log("对局已结束或不存在");
  render();
}

async function rejoinAfterGone() {
  const offer = state.roomGoneOffer;
  if (!offer) return;
  state.roomGoneOffer = null;
  state.connectionStatus = "connecting";
  state.connectionMessage = "正在重新进入";
  log("尝试重新进入对局");
  render();
  try {
    if (offer.isViewer) {
      const status = await viewerRoomStatus(offer.code);
      if (!status.ok) {
        offer.lastError = status.message;
        state.roomGoneOffer = offer;
        state.connectionStatus = "offline";
        state.connectionMessage = "对局已结束或不存在";
        log(`无法重新进入：${status.message}`);
        render();
        return;
      }
      beginViewerSession(offer.code);
      return;
    }
    const sessionToken = sessionTokenForRoom(offer.code);
    const res = await fetch(`/api/rooms/${encodeURIComponent(offer.code)}/join`, {method:"POST", body: JSON.stringify({name: state.name, sessionToken})});
    if (!res.ok) {
      const msg = await responseErrorText(res);
      offer.lastError = msg;
      state.roomGoneOffer = offer;
      state.connectionStatus = "offline";
      state.connectionMessage = "对局已结束或不存在";
      log(`无法重新进入：${msg}`);
      render();
      return;
    }
    await acceptJoin(res);
  } catch (e) {
    offer.lastError = String(e);
    state.roomGoneOffer = offer;
    state.connectionStatus = "offline";
    state.connectionMessage = "对局已结束或不存在";
    log(`无法重新进入：${e}`);
    render();
  }
}

function dismissRoomGone() {
  state.roomGoneOffer = null;
  leaveEndedRoom(true);
}

function connectionReady() {
  return state.ws && state.ws.readyState === WebSocket.OPEN;
}

function connectionBlockedText() {
  return state.connectionMessage || (state.connectionStatus === "reconnecting" ? "正在重连，请稍候" : "尚未连接，正在重连");
}

function checkSocketHealth() {
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN) return;
  if (Date.now() - state.lastSocketMessageAt < socketHeartbeatTimeoutMs) return;
  log("连接疑似卡死，正在重连");
  state.connectionStatus = "reconnecting";
  state.connectionMessage = "连接卡住，正在重连";
  render();
  state.ws.close();
}

async function viewerRoomStatus(code, viewerName = "") {
  viewerName = String(viewerName || "").trim();
  if (state.token && state.code === code && !state.viewer) {
    return {ok: false, message: "当前玩家不能同时观战自己的对局"};
  }
  const res = await fetch("/api/rooms");
  if (!res.ok) return {ok: false, message: `错误：${await responseErrorText(res)}`};
  const data = await res.json();
  const room = (data.rooms || []).find(r => String(r.code).toUpperCase() === code);
  if (!room) return {ok: false, message: "对局不在进行中"};
  if (viewerName && (room.seats || []).some(s => ((s.name || "").trim().toLowerCase() === viewerName.toLowerCase()))) {
    return {ok: false, message: "当前昵称已在该对局中落座，不能同时观战", room};
  }
  if (!room.canJoinView) return {ok: false, message: "观战席已满，请稍后再试", room};
  return {ok: true, room};
}

function onMessage(msg) {
  if (msg.type === "heartbeat") return;
  if (msg.type === "room.state") {
    state.room = msg.room;
    if (!state.viewer && typeof msg.room?.selfSeat === "number") {
      state.seat = msg.room.selfSeat;
      localStorage.setItem("siguo.seat", String(state.seat));
    }
    if (msg.room?.mode) {
      state.mode = msg.room.mode;
      localStorage.setItem("siguo.mode", state.mode);
    }
    state.view = msg.room?.view || state.view;
    clearInactiveMoveTrail();
  } else if (msg.type === "view") {
    state.view = msg.view;
    clearInactiveMoveTrail();
  } else if (msg.type === "chat.msg") {
    state.chat.push(msg.chat);
    enqueueBoardChat(msg.chat);
  } else if (msg.type === "error") {
    log(`错误：${msg.error?.message || msg.notice}`);
  } else if (msg.event) {
    handleEventEffect(msg.event);
    log(eventText(msg.event));
  }
  render();
}

function selectPieceMarker(marker) {
  state.selectedMarker = state.selectedMarker === marker ? null : marker;
  state.selected = null;
  render();
}

function handleMarkerClick(pieceId, owner) {
  const marker = state.selectedMarker;
  if (!marker) return false;
  if (!pieceId) {
    log("请选择要标记的对方棋子");
    return true;
  }
  if (marker === "unmark") {
    if (state.pieceMarks[pieceId]) {
      delete state.pieceMarks[pieceId];
      state.selectedMarker = null;
      render();
    } else {
      log("该棋子没有标记");
    }
    return true;
  }
  if (sameSide(owner, state.seat)) {
    log("只能标记对方棋子");
    return true;
  }
  state.pieceMarks[pieceId] = marker;
  state.selectedMarker = null;
  render();
  return true;
}

function clickCell(cell) {
  ensureAudio();
  if (state.roomGoneOffer) {
    log("请先选择重新进入或返回大厅");
    return;
  }
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
  if (["setup", "playing"].includes(state.room.phase)) {
    if (socketLooksStale()) {
      log("连接卡住，正在重连");
      forceReconnect(state.viewer);
      return;
    }
    if (!connectionReady()) {
      log(connectionBlockedText());
      if (!state.reconnectTimer) scheduleReconnect(state.viewer);
      return;
    }
  }
  if (state.selectedMarker && handleMarkerClick(pieceId, owner)) return;
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

function sendSocketMessage(msg, viewerRequest = false) {
  if (socketLooksStale()) {
    log("连接卡住，正在重连");
    forceReconnect(viewerRequest);
    return false;
  }
  if (!connectionReady()) {
    log(connectionBlockedText());
    if (!state.reconnectTimer) scheduleReconnect(viewerRequest);
    return false;
  }
  msg.seq = state.seq++;
  state.ws.send(JSON.stringify(msg));
  return true;
}

function send(msg) {
  ensureAudio();
  if (state.viewer) {
    log("观战中，不能操作对局");
    return;
  }
  if (socketLooksStale()) {
    log("连接卡住，正在重连");
    forceReconnect(false);
    return;
  }
  if (!connectionReady()) {
    log(connectionBlockedText());
    if (!state.reconnectTimer) scheduleReconnect(false);
    return;
  }
  msg.seq = state.seq++;
  state.ws.send(JSON.stringify(msg));
}

function sendChat(channel) {
  const input = document.querySelector("#chatText");
  if (!input) return;
  const effectiveChannel = state.viewer ? "all" : channel;
  sendSocketMessage({type:"chat.send", channel: effectiveChannel, text: input.value}, state.viewer);
  input.value = "";
}

function swapSeat(seat) {
  send({type:"seat.swap", seat: Number(seat)});
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
  if (type === "move" || type === "combat") setLastMoveTrail(ev);
  if (type === "gameEnded") state.lastTrail = null;
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

function clearInactiveMoveTrail() {
  if (state.room?.phase && state.room.phase !== "playing") {
    state.lastTrail = null;
  }
}

function setLastMoveTrail(ev) {
  const from = ev.From || ev.from;
  const to = ev.To || ev.to;
  const path = ev.Path || ev.path || [];
  const normalized = path.length >= 2 ? path : [from, to].filter(Boolean);
  state.lastTrail = {mode: currentMode(), path: normalized};
}

function shouldPlaySetupMusic() {
  return state.setupMusicEnabled && ["setup", "playing"].includes(state.room?.phase);
}

function ensureSetupMusic() {
  if (state.setupMusic) return state.setupMusic;
  const audio = new Audio("/song-small.mp3");
  audio.preload = "auto";
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

function warmSetupMusic() {
  if (!state.setupMusicEnabled) return;
  const audio = ensureSetupMusic();
  if (audio.networkState === HTMLMediaElement.NETWORK_EMPTY) {
    audio.load();
  }
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
  warmSetupMusic();
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

function pruneBoardChats() {
  const now = Date.now();
  state.boardChats = state.boardChats.filter(chat => now - chat.createdAt < chat.durationMs);
}

function scheduleBoardChatCleanup() {
  if (state.boardChatTimer) {
    clearTimeout(state.boardChatTimer);
    state.boardChatTimer = null;
  }
  if (!state.boardChats.length) return;
  const now = Date.now();
  let nextExpiry = Infinity;
  state.boardChats.forEach(chat => {
    nextExpiry = Math.min(nextExpiry, chat.createdAt + chat.durationMs - now);
  });
  state.boardChatTimer = setTimeout(() => {
    state.boardChatTimer = null;
    const before = state.boardChats.length;
    pruneBoardChats();
    if (state.boardChats.length !== before) render();
    scheduleBoardChatCleanup();
  }, Math.max(40, nextExpiry));
}

function chatBody(chat) {
  return String(chat.text || chat.emote || "").trim();
}

function isSystemChat(chat) {
  return chatBody(chat).startsWith("[系统]");
}

function enqueueBoardChat(chat) {
  if (!chat || (chat.channel !== "all" && chat.channel !== "team")) return;
  const text = boardChatText(chat);
  if (!text) return;
  pruneBoardChats();
  const lane = state.boardChatLane % 5;
  state.boardChatLane += 1;
  state.boardChats.push({
    text,
    channel: chat.channel,
    viewer: !!chat.viewer,
    system: isSystemChat(chat),
    top: 10 + lane * 15,
    createdAt: Date.now(),
    durationMs: Math.min(12000, 6400 + text.length * 90)
  });
  if (state.boardChats.length > 12) state.boardChats = state.boardChats.slice(-12);
  scheduleBoardChatCleanup();
}

function boardChatText(chat) {
  const body = chatBody(chat);
  if (!body) return "";
  if (isSystemChat(chat)) return body;
  if (chat.channel === "team") {
    const prefix = seatNames[Number(chat.from)] || "队友";
    const name = chat.name ? ` ${chat.name}` : "";
    return `${prefix}${name}：${body}（仅队友可见）`.trim();
  }
  if (chat.viewer) return `观战 ${chat.name || "观众"}：${body}`;
  const prefix = seatNames[Number(chat.from)] || "";
  const name = chat.name ? ` ${chat.name}` : "";
  return `${prefix}${name}：${body}`.trim();
}

function coord(p) {
  if (!p) return "";
  return `${p.Row ?? p.row},${p.Col ?? p.col}`;
}

function chatLine(c) {
  const label = c.channel === "team" ? "队伍" : "公屏";
  const body = chatBody(c);
  if (isSystemChat(c)) {
    return `<div class="line">[${label}] ${esc(body)}</div>`;
  }
  if (c.viewer) {
    return `<div class="line chat-line chat-line-viewer">[${label}] 观战 ${esc(c.name || "观众")}：${esc(body)}</div>`;
  }
  if (c.channel === "team") {
    return `<div class="line chat-line chat-line-team">[队伍（仅队友可见）] ${seatNames[Number(c.from)] || ""} ${esc(c.name)}：${esc(body)}</div>`;
  }
  return `<div class="line chat-line chat-line-public">[公屏] ${seatNames[Number(c.from)] || ""} ${esc(c.name)}：${esc(body)}</div>`;
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
setInterval(checkSocketHealth, socketWatchdogIntervalMs);
let lastVisibilityReconnectAt = 0;
const visibilityReconnectCooldownMs = 5000;
document.addEventListener("visibilitychange", () => {
  if (document.visibilityState !== "visible") return;
  if (state.roomGoneOffer) return;
  if (!state.code) return;
  if (!state.viewer && !state.token) return;
  const rs = state.ws && state.ws.readyState;
  const deadOrClosing = state.ws && (rs === WebSocket.CLOSING || rs === WebSocket.CLOSED);
  if (!(socketLooksStale() || deadOrClosing)) return;
  if (Date.now() - lastVisibilityReconnectAt < visibilityReconnectCooldownMs) return;
  lastVisibilityReconnectAt = Date.now();
  log("回到页面，检查连接");
  forceReconnect(state.viewer);
});
const initialParams = new URLSearchParams(location.search);
const initialWatchCode = initialParams.get("watch");
const initialJoinCode = initialParams.get("join");
const initialInviteName = initialParams.get("name");
if (initialJoinCode && initialInviteName) {
  state.name = initialInviteName;
  localStorage.setItem("siguo.name", state.name);
}
if (initialWatchCode) {
  connectViewer(initialWatchCode);
} else if (initialJoinCode) {
  joinFromInvite(initialJoinCode);
} else if (state.code && state.token) {
  restorePreviousSession();
}

async function restorePreviousSession() {
  try {
    const res = await fetch(`/api/rooms/${encodeURIComponent(state.code)}?token=${encodeURIComponent(state.token)}`);
    if (res.status === 404 || res.status === 401) {
      forgetSavedSession();
      return;
    }
  } catch {
    // Network probe failed — fall through and try WS anyway; the regular
    // reconnect machinery will handle real outages.
  }
  connect();
}

function forgetSavedSession() {
  state.code = "";
  state.token = "";
  state.host = false;
  state.seat = 0;
  state.room = null;
  state.view = null;
  state.viewer = false;
  state.selected = null;
  state.selectedMarker = null;
  state.pieceMarks = {};
  state.combat = null;
  state.connectionStatus = "idle";
  state.connectionMessage = "";
  state.reconnectAttempts = 0;
  localStorage.removeItem("siguo.code");
  localStorage.removeItem("siguo.token");
  render();
}
