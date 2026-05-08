package static

import (
	"bytes"
	"image/jpeg"
	"strings"
	"testing"
)

func TestJunqiRailOverlayStretchesWithWideBoard(t *testing.T) {
	data, err := FS.ReadFile("dist/app.js")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.js) error = %v", err)
	}
	js := string(data)
	if !strings.Contains(js, `class="rails rails-junqi" viewBox="0 0 5 13" preserveAspectRatio="none"`) {
		t.Fatalf("junqi rail overlay must use preserveAspectRatio=\"none\" so it stretches with the wider board")
	}
	if !strings.Contains(js, `addVertical(2.5, 5.5, 7.5);`) {
		t.Fatalf("junqi rail overlay must draw the center mountain crossing without extending it into the camp lattice")
	}
}

func TestJunqiSpecialCellsDoNotShrinkPieces(t *testing.T) {
	data, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(data)
	for _, want := range []string{
		".board-junqi .camp {\n  border-radius: 0;\n  margin: 0;",
		".board-junqi .hq {\n  margin: 0;",
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("junqi special cells must keep full grid size; missing %q", want)
		}
	}
}

func TestPlayerTickersTouchTerritoryEdgesWithoutOverlap(t *testing.T) {
	data, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(data)
	for _, want := range []string{
		`.board-siguo {
  --siguo-cell-w: 5.2%;
  --siguo-cell-h: 4.6%;
  display: block;
  overflow: visible;`,
		`box-sizing: border-box;`,
		`width: 31.7647058824%;`,
		`height: 11.7647058824%;`,
		`.player-ticker.ticker-rel-0 { left: 68.2352941176%; bottom: 0; }`,
		`.player-ticker.ticker-rel-2 { left: 0; top: 0; }`,
		`left: 89.4117647059%;`,
		`top: 11.7647058824%;`,
		`height: 8.2352941176%;`,
		`left: -21.1764705882%;`,
		`top: 80%;`,
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("player ticker should use exact non-overlapping territory edge placement; missing %q", want)
		}
	}
}

func TestBoardWrapKeepsBoardsNearTopControls(t *testing.T) {
	data, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(data)
	for _, want := range []string{
		".board-wrap {\n  position: relative;\n  display: grid;\n  place-items: start center;",
		"padding: 8px 0 0;",
		".victory-layer {\n  position: absolute;\n  inset: 8px 0 0;",
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("board wrapper should keep 2:2 and 1:1 boards close to the top controls; missing %q", want)
		}
	}
}

func TestDeadTrayRendersBelowBoardAsHorizontalList(t *testing.T) {
	jsData, err := FS.ReadFile("dist/app.js")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.js) error = %v", err)
	}
	js := string(jsData)
	if !strings.Contains(js, "let html = `<div class=\"${stageClass}\">${boardOpen}`") ||
		!strings.Contains(js, "return html + `${boardClose}${deadTrayHTML()}</div>`;") {
		t.Fatalf("dead tray should render after the board inside the board stage")
	}

	cssData, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(cssData)
	for _, want := range []string{
		".board-stage {\n  display: flex;\n  flex-direction: column;\n  align-items: center;",
		"grid-template-columns: auto minmax(0, 1fr);",
		".dead-list {\n  display: flex;\n  align-items: center;",
		"flex: 0 0 auto;",
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("dead tray should sit below the board as a horizontal strip; missing %q", want)
		}
	}
}

func TestSideActivityPanelsShowNewestFiveFirst(t *testing.T) {
	jsData, err := FS.ReadFile("dist/app.js")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.js) error = %v", err)
	}
	js := string(jsData)
	chatHeader := strings.Index(js, `<b>聊天</b>`)
	logHeader := strings.Index(js, `<b>战况</b>`)
	if chatHeader < 0 || logHeader < 0 || chatHeader > logHeader {
		t.Fatalf("chat panel should render above battle log panel")
	}
	for _, want := range []string{
		`state.chat.slice(-80).reverse().map(chatLine)`,
		`state.log.slice(-80).reverse().map(x =>`,
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("activity streams should render newest entries first; missing %q", want)
		}
	}

	cssData, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(cssData)
	for _, want := range []string{
		`grid-template-rows: auto auto auto auto minmax(0, 1fr);`,
		`max-height: calc(5 * 36px + 4 * 6px);`,
		`min-height: 36px;`,
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("activity streams should show five rows before scrolling; missing %q", want)
		}
	}
}

func TestJunqiUsesMap3BoardAsset(t *testing.T) {
	cssData, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(cssData)
	if !strings.Contains(css, `url("/map3.jpg")`) {
		t.Fatalf("junqi board should use map3.jpg as the board surface")
	}
	if !strings.Contains(css, `aspect-ratio: 27 / 37;`) {
		t.Fatalf("junqi board should keep the full map3.jpg board proportions")
	}
	if !strings.Contains(css, `.board-surface-wrap-junqi {
  position: relative;`) ||
		!strings.Contains(css, `.board-clip-junqi > .board-junqi {
  position: absolute;`) {
		t.Fatalf("junqi board should render inside the junqi surface wrapper")
	}
	if !strings.Contains(css, `.board-junqi {
    width: auto;
    height: min(92vh, 100vw);
  }`) {
		t.Fatalf("junqi board should override the generic mobile width rule")
	}
	if !strings.Contains(css, `display: block;`) ||
		!strings.Contains(css, `.board-junqi .cell {
  position: absolute;`) {
		t.Fatalf("junqi board cells should be absolute overlays matched to map3.jpg")
	}
	if !strings.Contains(css, `--junqi-cell-w: 9.1%;`) ||
		!strings.Contains(css, `--junqi-cell-h: 4.0%;`) {
		t.Fatalf("junqi board cells should use the measured map3 footprint")
	}
	if !strings.Contains(css, `.board-junqi-rel-0::before { transform: rotate(180deg); }`) ||
		!strings.Contains(css, `.board-junqi-rel-2::before { transform: rotate(0deg); }`) {
		t.Fatalf("junqi board image should rotate with the 1:1 viewer seat")
	}
	jsData, err := FS.ReadFile("dist/app.js")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.js) error = %v", err)
	}
	js := string(jsData)
	if !strings.Contains(js, `board-surface-wrap board-surface-wrap-junqi`) ||
		!strings.Contains(js, `board-clip board-clip-junqi`) ||
		!strings.Contains(js, `board-junqi-rel-${state.seat}`) {
		t.Fatalf("junqi board markup should wrap the board in a clipping surface")
	}
	if !strings.Contains(js, `const junqiMap3Centers = {`) ||
		!strings.Contains(js, `x: [13.35, 31.48, 50.00, 67.94, 86.57]`) ||
		!strings.Contains(js, `y: [8.45, 15.10, 21.75, 28.38, 35.03, 41.72, 50.08, 58.11, 64.78, 71.45, 78.04, 84.65, 91.55]`) {
		t.Fatalf("junqi cell centers should be measured from map3.jpg")
	}
	if !strings.Contains(js, `displayRow === 6 && displayCol === 1`) ||
		!strings.Contains(js, `displayRow === 6 && displayCol === 3`) {
		t.Fatalf("junqi mountain labels should line up with the two map3 山界 positions")
	}
	if strings.Contains(js, `row === 8 && [6, 8, 10].includes(col)) return "turn-front"`) {
		t.Fatalf("junqi should not draw central green turn markers over the map image")
	}
	if _, err := FS.ReadFile("dist/map3.jpg"); err != nil {
		t.Fatalf("ReadFile(dist/map3.jpg) error = %v", err)
	}
	data, err := FS.ReadFile("dist/map3.jpg")
	if err != nil {
		t.Fatalf("ReadFile(dist/map3.jpg) error = %v", err)
	}
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeConfig(dist/map3.jpg) error = %v", err)
	}
	if cfg.Width != 864 || cfg.Height != 1184 {
		t.Fatalf("dist/map3.jpg size = %dx%d, want regenerated 864x1184 board", cfg.Width, cfg.Height)
	}
}

func TestSiguoUsesMap8BoardAsset(t *testing.T) {
	cssData, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(cssData)
	if !strings.Contains(css, `url("/map8.jpg")`) {
		t.Fatalf("siguo board should use map8.jpg as the 2v2 board surface")
	}
	if !strings.Contains(css, `.board-siguo .cell {
  position: absolute;`) {
		t.Fatalf("siguo board cells should be absolute overlays matched to map8.jpg")
	}
	if !strings.Contains(css, `--siguo-cell-w: 5.2%;`) ||
		!strings.Contains(css, `--siguo-cell-h: 4.6%;`) {
		t.Fatalf("siguo board cells should use the measured map8 footprint")
	}

	jsData, err := FS.ReadFile("dist/app.js")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.js) error = %v", err)
	}
	js := string(jsData)
	if !strings.Contains(js, `board-siguo-rel-${state.seat}`) ||
		!strings.Contains(js, `const siguoMap8Centers = {`) {
		t.Fatalf("siguo board should rotate map8 by viewer seat and use measured centers")
	}
	data, err := FS.ReadFile("dist/map8.jpg")
	if err != nil {
		t.Fatalf("ReadFile(dist/map8.jpg) error = %v", err)
	}
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeConfig(dist/map8.jpg) error = %v", err)
	}
	if cfg.Width != 800 || cfg.Height != 800 {
		t.Fatalf("dist/map8.jpg size = %dx%d, want cropped 800x800 board", cfg.Width, cfg.Height)
	}
}

func TestVictoryAndSetupCultureUIAreBundled(t *testing.T) {
	jsData, err := FS.ReadFile("dist/app.js")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.js) error = %v", err)
	}
	js := string(jsData)
	for _, want := range []string{
		"function victoryHTML()",
		"function setupCountdownText()",
		"setupDeadlineMs",
		"秒后自动开局",
		"观战室",
		"function watchRoomPanelHTML()",
		"function inviteLinkButtonHTML()",
		"function inviteURL(code)",
		"function shareOrigin()",
		"siguo.shareOrigin",
		"请输入可分享的服务器地址",
		"http://YOUR_LAN_IP:8080",
		"未设置可分享的服务器地址",
		"function copyInviteLink(code)",
		"function joinFromInvite(code)",
		"function offerViewerAfterJoinFailure(code, message)",
		"邀请链接已复制：${url}",
		"观战链接已复制：${url}",
		"玩家座位已满，可选择观战",
		"joinOfferView",
		"function connectViewer(code)",
		"function viewerRoomStatus(code)",
		"观战席已满，请稍后再试",
		"function shareURL(param, code, extras = null)",
		"url.searchParams.set(key, value);",
		`const initialInviteName = initialParams.get("name");`,
		"viewer=1",
		"function responseErrorText(res)",
		"return data.error || text;",
		"const quickChatPhrases = [",
		"不怕神一样的对手，就怕猪一样的队友",
		"function quickChatHTML()",
		"function fillQuickChat(text)",
		"function canSwapSeats()",
		`send({type:"seat.swap", seat: Number(seat)})`,
		`msg.room?.selfSeat`,
		"function junqiSeatColor(seat)",
		`Number(seat) === 0 ? "red" : "blue"`,
		"function sameTeam(a, b)",
		"const seatA = Number(a);",
		`return junqiSeatColor(winners[0]) === "red" ? "红方获胜" : "蓝方获胜"`,
		"return `victory-${junqiSeatColor(winners[0])}`",
		"return `piece-${junqiSeatColor(owner)}`",
		"function victoryActionsHTML()",
		"restartRoomBtn",
		"leaveEndedRoom",
		`!["lobby", "ended"].includes(state.room.phase)`,
		"另一方获胜",
		"state.room?.Winner",
		"function setupCultureHTML()",
		"siguo.setupMusic",
		`!["setup", "playing"].includes(state.room.phase)`,
		`state.setupMusicEnabled && ["setup", "playing"].includes(state.room?.phase)`,
		"四方列阵，战局正酣",
		"poem-title",
		"明·杨慎",
		"滚滚长江东逝水",
		"/song.mp3",
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("app.js missing %q", want)
		}
	}

	cssData, err := FS.ReadFile("dist/app.css")
	if err != nil {
		t.Fatalf("ReadFile(dist/app.css) error = %v", err)
	}
	css := string(cssData)
	for _, want := range []string{".victory-layer", ".victory-red", ".victory-blue", ".victory-ew", ".victory-actions", "pointer-events: auto;", ".beauty", ".petal", ".culture-panel", ".watch-panel", ".watch-row", ".join-offer", ".quick-chat", ".quick-chat-btn", ".seat-swap-btn", `url("/picture02.png")`, "@keyframes poem-cross"} {
		if !strings.Contains(css, want) {
			t.Fatalf("app.css missing %q", want)
		}
	}
	if !strings.Contains(js, `seats seats-${currentMode()}`) ||
		!strings.Contains(css, `.seats-junqi .seat[data-seat="2"] b`) {
		t.Fatalf("junqi sidebar seats should render South as blue instead of the 2v2 yellow")
	}
	for _, want := range []string{`"STXingkai"`, `"AR PL UKai CN"`, `"FandolKai"`, `"Noto Serif CJK SC"`} {
		if !strings.Contains(css, want) {
			t.Fatalf("poem font stack missing %q", want)
		}
	}

	if _, err := FS.ReadFile("dist/setup-music.ogg"); err != nil {
		t.Fatalf("ReadFile(dist/setup-music.ogg) error = %v", err)
	}
	if _, err := FS.ReadFile("dist/song.mp3"); err != nil {
		t.Fatalf("ReadFile(dist/song.mp3) error = %v", err)
	}
	if _, err := FS.ReadFile("dist/picture02.png"); err != nil {
		t.Fatalf("ReadFile(dist/picture02.png) error = %v", err)
	}
}
