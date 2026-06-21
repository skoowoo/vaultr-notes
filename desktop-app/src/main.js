const { app, BaseWindow, WebContentsView, shell, ipcMain, nativeImage, Menu, dialog, Notification } = require("electron");

app.setName("Vaultr");

const path = require("node:path");
const fs = require("node:fs");
const http = require("node:http");
const https = require("node:https");
const { getAutoStart, setAutoStart, setServerUrl, getMateNotifySettings, registerConfigIpcHandlers } = require("./config");

require("./server-manager").register({
  onRestartDone:  () => resetToStartScreen(),
  onServerStopped: () => resetToStartScreen(),
  onBeforeStop:   () => stopMateNotifications(),
  getAutoStart:   () => getAutoStart(),
  setAutoStart:   (v) => setAutoStart(v),
});

const SECTIONS = ["home", "agent"];

/** @returns {import("electron").NativeImage | null} */
function getAppIconImage() {
  const root = path.join(__dirname, "..");
  const png = path.join(root, "assets", "icon.png");
  const icns = path.join(root, "assets", "icon.icns");
  // Prefer PNG: Chromium often fails to decode minimal .icns (e.g. only ic07) for dock/window icons.
  const candidates = [png, icns];
  for (const p of candidates) {
    if (!fs.existsSync(p)) continue;
    try {
      const img = nativeImage.createFromPath(p);
      if (!img.isEmpty()) return img;
    } catch {
      /* ignore */
    }
  }
  return null;
}

/**
 * Path for {@link app.dock.setIcon}. Uses PNG first — same .icns decode limits as {@link getAppIconImage}.
 * @returns {string | null}
 */
function getMacDockIconPath() {
  if (process.platform !== "darwin") return null;
  const root = path.resolve(__dirname, "..", "assets");
  const png = path.join(root, "icon.png");
  const icns = path.join(root, "icon.icns");
  if (fs.existsSync(png)) return png;
  if (fs.existsSync(icns)) return icns;
  return null;
}

let win;       // BaseWindow — no built-in webContents, avoids event interception
let startView; // WebContentsView for the connection screen
let serverUrl = null;
const views = {};
let activeSection = null;
// Track the most-recently applied theme background so detached views can be
// pre-synced before they are made visible (prevents flash-of-wrong-theme).
let currentViewBgColor = '#0f0f0f';

// ── Mate run notifications (direct SSE from main process) ─────────────────────

/** @type {import("node:http").ClientRequest | null} */
let mateNotifReq = null;
let mateNotifReconnectTimer = null;
let mateNotifWatchdogTimer = null;
const MATE_NOTIF_WATCHDOG_MS = 45_000;

function stopMateNotifications() {
  if (mateNotifReconnectTimer) {
    clearTimeout(mateNotifReconnectTimer);
    mateNotifReconnectTimer = null;
  }
  if (mateNotifWatchdogTimer) {
    clearTimeout(mateNotifWatchdogTimer);
    mateNotifWatchdogTimer = null;
  }
  if (mateNotifReq) {
    try { mateNotifReq.destroy(); } catch (_) {}
    mateNotifReq = null;
  }
}

function scheduleMateNotifReconnect(url) {
  mateNotifReconnectTimer = setTimeout(() => {
    mateNotifReconnectTimer = null;
    if (serverUrl === url) startMateNotifications(url);
  }, 5000);
}

function resetMateNotifWatchdog(url) {
  if (mateNotifWatchdogTimer) clearTimeout(mateNotifWatchdogTimer);
  mateNotifWatchdogTimer = setTimeout(() => {
    mateNotifWatchdogTimer = null;
    if (serverUrl === url) startMateNotifications(url);
  }, MATE_NOTIF_WATCHDOG_MS);
}

/**
 * Play a sound by name or path.
 * Accepts: "beep" (system alert), "none"/falsy (silence),
 *          macOS system sound name (e.g. "Glass", "Ping"),
 *          or an absolute file path to a .aiff/.wav/.mp3.
 */
function playSound(sound) {
  if (!sound || sound === "none") return;
  if (sound === "beep") { shell.beep(); return; }
  if (process.platform === "darwin") {
    const { spawn } = require("node:child_process");
    const soundPath = sound.startsWith("/")
      ? sound
      : `/System/Library/Sounds/${sound}.aiff`;
    const child = spawn("/usr/bin/afplay", [soundPath], { stdio: "ignore" });
    child.on("error", () => shell.beep()); // fall back if afplay is unavailable
    child.unref();
  } else {
    shell.beep();
  }
}

function showMateNotification(type, payload) {
  const settings = getMateNotifySettings();
  const name = payload.mateName || "Mate";
  const isDone = type === "run_done";

  if (settings.textEnabled && Notification.isSupported()) {
    const title = name;
    const status = isDone ? (payload.success ? "Finished" : "Failed") : "Started";
    const msg = (payload.lastMessage || "").trim().slice(0, 100);
    const body = msg ? `${status} — ${msg}` : status;
    new Notification({ title, body, silent: true }).show();
  }

  if (settings.soundEnabled) {
    playSound(isDone ? settings.doneSound : settings.startSound);
  }
}

function startMateNotifications(url) {
  stopMateNotifications();
  if (!url) return;

  let parsed;
  try { parsed = new URL("/api/mate/run-notifications", url); } catch { return; }
  const mod = parsed.protocol === "https:" ? https : http;

  let buffer = "";
  const req = mod.request(parsed, (res) => {
    if (res.statusCode !== 200) {
      res.resume();
      scheduleMateNotifReconnect(url);
      return;
    }
    resetMateNotifWatchdog(url);
    res.on("data", (chunk) => {
      resetMateNotifWatchdog(url);
      buffer += chunk.toString();
      const blocks = buffer.split("\n\n");
      buffer = blocks.pop(); // keep the incomplete trailing fragment
      for (const block of blocks) {
        if (!block.trim()) continue;
        const evMatch   = block.match(/^event: (.+)$/m);
        const dataMatch = block.match(/^data: (.+)$/m);
        if (!dataMatch) continue;
        const type = evMatch?.[1]?.trim() || "";
        try {
          showMateNotification(type, JSON.parse(dataMatch[1].trim()));
        } catch { /* ignore malformed events */ }
      }
    });
    res.on("end",   () => scheduleMateNotifReconnect(url));
    res.on("error", () => scheduleMateNotifReconnect(url));
  });
  req.on("error", () => scheduleMateNotifReconnect(url));
  req.end();
  mateNotifReq = req;
}

// ─────────────────────────────────────────────────────────────────────────────

function getContentBounds() {
  const [w, h] = win.getContentSize();
  return { x: 0, y: 0, width: w, height: h };
}

function resizeViews() {
  const bounds = getContentBounds();
  if (activeSection && views[activeSection]) {
    views[activeSection].setBounds(bounds);
  } else if (startView) {
    startView.setBounds(bounds);
  }
}

function makeWebPrefs() {
  return {
    preload: path.join(__dirname, "preload.js"),
    contextIsolation: true,
    nodeIntegration: false,
    sandbox: true,
  };
}

// opts: { fullUrl?: string, reload?: boolean }
function showSection(name, opts = {}) {
  const { fullUrl = null, reload = false } = opts;
  if (!views[name]) return;

  if (activeSection && activeSection !== name && views[activeSection]) {
    win.contentView.removeChildView(views[activeSection]);
  }

  if (activeSection !== name) {
    activeSection = name;
    // Pre-sync the native background colour before the view is composited so
    // the OS-level layer never flashes the wrong colour on first paint.
    try { views[name].setBackgroundColor(currentViewBgColor); } catch (_) {}
    win.contentView.addChildView(views[name]);
    views[name].setBounds(getContentBounds());
  }

  views[name].webContents.focus();

  if (reload) {
    views[name].webContents.reload();
    return;
  }

  if (fullUrl) {
    const sectionBase = serverUrl + "/" + name;
    const currentUrl = views[name].webContents.getURL();
    if (fullUrl !== sectionBase && fullUrl !== currentUrl) {
      views[name].webContents.loadURL(fullUrl);
    }
  }
}

// Show the connection screen, discarding any active section views.
function resetToStartScreen() {
  if (!serverUrl) return; // already reset, guard against concurrent did-fail-load calls
  serverUrl = null;
  agentCache = null;
  stopMateNotifications();

  if (activeSection && views[activeSection]) {
    try { win.contentView.removeChildView(views[activeSection]); } catch { }
  }
  activeSection = null;
  for (const section of SECTIONS) {
    if (views[section]) {
      try { views[section].webContents.destroy(); } catch { }
      delete views[section];
    }
  }

  showStartScreen();
}

function showStartScreen() {
  startView = new WebContentsView({ webPreferences: makeWebPrefs(), backgroundColor: '#181715' });

  startView.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: "deny" };
  });

  // Reload start.html if the file itself somehow fails (defensive)
  startView.webContents.on("did-fail-load", (_event, errorCode, _desc, validatedURL) => {
    if (validatedURL && !validatedURL.startsWith("file://") && errorCode < 0) {
      startView.webContents.loadFile(path.join(__dirname, "start.html"));
    }
  });

  win.contentView.addChildView(startView);
  startView.setBounds(getContentBounds());
  startView.webContents.loadFile(path.join(__dirname, "start.html"));
}

function setupSectionView(view, sectionName) {
  view.webContents.setWindowOpenHandler(({ url }) => {
    // Cmd+Click on an internal link — navigate within the app instead of opening the browser.
    if (serverUrl && url.startsWith(serverUrl)) {
      try {
        const parsed = new URL(url);
        const firstSeg = parsed.pathname.split("/").filter(Boolean)[0];
        if (SECTIONS.includes(firstSeg)) {
          setImmediate(() => showSection(firstSeg, { fullUrl: url }));
          return { action: "deny" };
        }
      } catch { /* ignore malformed URLs */ }
      setImmediate(() => view.webContents.loadURL(url));
      return { action: "deny" };
    }
    shell.openExternal(url);
    return { action: "deny" };
  });

  // When the server stops, the active view will fail to load on any navigation.
  // ERR_ABORTED (-3) is a normal cancellation (e.g. redirect); ignore it.
  view.webContents.on("did-fail-load", (_event, errorCode, _desc, validatedURL) => {
    if (validatedURL && !validatedURL.startsWith("file://") && errorCode < 0 && errorCode !== -3) {
      resetToStartScreen();
    }
  });

  view.webContents.on("will-navigate", (event, url) => {
    try {
      const parsed = new URL(url);
      const firstSeg = parsed.pathname.split("/").filter(Boolean)[0];
      if (SECTIONS.includes(firstSeg) && firstSeg !== sectionName) {
        event.preventDefault();
        showSection(firstSeg, { fullUrl: url });
      }
    } catch { }
  });
}

function createSectionViews(url) {
  serverUrl = url;
  for (const section of SECTIONS) {
    const view = new WebContentsView({ webPreferences: makeWebPrefs(), backgroundColor: '#181715' });
    view.webContents.loadURL(url + "/" + section);
    setupSectionView(view, section);
    views[section] = view;
  }
  // Remove start screen, show home
  if (startView) {
    win.contentView.removeChildView(startView);
    try { startView.webContents.destroy(); } catch { }
    startView = null;
  }
  showSection("home");
}

function createWindow() {
  const icon = getAppIconImage();
  win = new BaseWindow({
    width: 1440,
    height: 960,
    minWidth: 960,
    minHeight: 640,
    title: "Vaultr",
    backgroundColor: "#191919",
    titleBarStyle: process.platform === "darwin" ? "hiddenInset" : "default",
    ...(icon ? { icon } : {}),
  });

  win.on("resize", resizeViews);

  // After switching away and back, refresh section views that are safe to reload
  // (external vault changes while the app was in the background).
  let windowHadBlur = false;
  win.on("blur", () => { windowHadBlur = true; });
  win.on("focus", () => {
    if (!windowHadBlur) return;
    windowHadBlur = false;
    void scheduleSyncVaultDataAcrossSectionViews();
  });

  // Reset module-level state when the window is closed so that re-opening
  // (via macOS dock activate) starts with a clean slate.
  win.on("closed", () => {
    activeSection = null;
    serverUrl = null;
    if (startView) {
      try { startView.webContents.destroy(); } catch { }
      startView = null;
    }
    for (const section of SECTIONS) {
      if (views[section]) {
        try { views[section].webContents.destroy(); } catch { }
        delete views[section];
      }
    }
    win = null;
  });

  showStartScreen();
}

// ── IPC: server connection ────────────────────────────────────────────────────

registerConfigIpcHandlers(ipcMain);

ipcMain.handle("set-server-url", (_event, url) => {
  agentCache = null;
  setServerUrl(url);
  startMateNotifications(url);
  if (win) {
    if (Object.keys(views).length === 0) {
      // First successful connection — create all section views
      createSectionViews(url);
    } else {
      // Server URL changed — reload all views with the new base URL
      serverUrl = url;
      for (const section of SECTIONS) {
        views[section].webContents.loadURL(url + "/" + section);
      }
      showSection("home");
    }
  }
});

/** @returns {Promise<string[]>} section names reloaded */
async function syncVaultDataAcrossSectionViews() {
  const reloaded = [];
  if (!win || win.isDestroyed() || !serverUrl) return reloaded;
  for (const section of SECTIONS) {
    const view = views[section];
    if (!view) continue;
    const wc = view.webContents;
    if (wc.isDestroyed()) continue;
    try {
      // Returns 'reload' | 'custom' | 'skip'.
      // If the page defines __vaultrBackgroundRefresh, call it instead of wc.reload().
      const action = await wc.executeJavaScript(`(function(){
        if(!window.__vaultrShellSafeForBackgroundReload||!window.__vaultrShellSafeForBackgroundReload())return'skip';
        if(typeof window.__vaultrBackgroundRefresh==='function'){window.__vaultrBackgroundRefresh();return'custom';}
        return'reload';
      })()`, true);
      if (action === 'reload') { wc.reload(); reloaded.push(section); }
      else if (action === 'custom') { reloaded.push(section); }
    } catch {
      /* blank, unloading, or JS error */
    }
  }
  return reloaded;
}

const SYNC_VAULT_DEBOUNCE_MS = 500;
let syncVaultDebounceTimer = null;
/** @type {Promise<{ reloaded: string[] }> | null} */
let syncVaultDebouncePromise = null;
/** @type {((value: { reloaded: string[] }) => void) | null} */
let syncVaultDebounceResolve = null;

/**
 * Global debounce for vault-driven section reloads: IPC, app foreground, drawer
 * close, etc. share one window — only the last trigger within SYNC_VAULT_DEBOUNCE_MS runs.
 * Concurrent ipcRenderer.invoke callers share the same Promise until flush.
 */
function scheduleSyncVaultDataAcrossSectionViews() {
  if (!win || win.isDestroyed() || !serverUrl) {
    return Promise.resolve({ reloaded: [] });
  }
  if (!syncVaultDebouncePromise) {
    syncVaultDebouncePromise = new Promise((resolve) => {
      syncVaultDebounceResolve = resolve;
    });
  }
  if (syncVaultDebounceTimer) clearTimeout(syncVaultDebounceTimer);
  syncVaultDebounceTimer = setTimeout(async () => {
    syncVaultDebounceTimer = null;
    const resolve = syncVaultDebounceResolve;
    syncVaultDebouncePromise = null;
    syncVaultDebounceResolve = null;
    try {
      const reloaded = await syncVaultDataAcrossSectionViews();
      resolve({ reloaded });
    } catch {
      resolve({ reloaded: [] });
    }
  }, SYNC_VAULT_DEBOUNCE_MS);
  return syncVaultDebouncePromise;
}

// Reload each section WebContentsView whose page reports it is safe to full-reload
// (see window.__vaultrShellSafeForBackgroundReload in server HTML).
ipcMain.handle("sync-vault-data-across-sections", () => {
  return scheduleSyncVaultDataAcrossSectionViews();
});

// Update the calling WebContentsView's background color when the renderer theme changes.
// This prevents a flash of the wrong background color during wc.reload() from main process.
ipcMain.on("set-window-button-visibility", (_event, visible) => {
  try { if (win) win.setWindowButtonVisibility(!!visible); } catch (_) {}
});

ipcMain.on("set-view-bg-color", (event, color) => {
  // Always track the latest theme bg so showSection() can pre-sync detached views.
  currentViewBgColor = color;

  // Update every view's native background colour immediately — including detached
  // ones — so the OS layer never shows a stale colour when a section is revealed.
  const allViews = [...Object.values(views), startView].filter(Boolean);
  for (const v of allViews) {
    if (!v.webContents.isDestroyed()) {
      try { v.setBackgroundColor(color); } catch (_) {}
    }
  }

  // Detached WebContentsViews may not receive localStorage storage events or
  // matchMedia change events from Chromium while they are off-screen.  Push a
  // lightweight data-theme sync into every view that didn't originate this call
  // so their HTML is already correct when they are next made visible.
  const senderWc = event.sender;
  for (const v of Object.values(views)) {
    if (v.webContents.isDestroyed() || v.webContents === senderWc) continue;
    v.webContents.executeJavaScript(`(function(){
      try{
        var pref=localStorage.getItem('theme')||'auto';
        var dark=pref==='auto'?window.matchMedia('(prefers-color-scheme: dark)').matches:pref==='dark';
        document.documentElement.setAttribute('data-theme',dark?'dark':'light');
      }catch(_){}
    })()`).catch(() => {});
  }
});

// ── IPC: drafts ───────────────────────────────────────────────────────────────

const MAX_DRAFTS = 10;
function getDraftsDir() { return path.join(app.getPath("userData"), "drafts"); }
const DRAFT_ID_RE = /^[A-Za-z0-9._-]{1,128}$/;

function assertDraftId(id) {
  if (typeof id !== "string" || !DRAFT_ID_RE.test(id)) {
    throw new Error("invalid draft id");
  }
  return id;
}

function ensureDraftsDir() {
  const dir = getDraftsDir();
  fs.mkdirSync(dir, { recursive: true, mode: 0o700 });
  return dir;
}

function getDraftPath(id) {
  id = assertDraftId(id);
  const dir = ensureDraftsDir();
  const file = path.join(dir, id + ".json");
  if (path.dirname(file) !== dir) throw new Error("invalid draft path");
  return file;
}

ipcMain.handle("draft:list", () => {
  const dir = ensureDraftsDir();
  try {
    return fs.readdirSync(dir)
      .filter(f => f.endsWith(".json"))
      .map(f => {
        try {
          const d = JSON.parse(fs.readFileSync(path.join(dir, f), "utf8"));
          return {
            id: f.slice(0, -5),
            path: d.path || d.pathInput || "",
            title: d.title || "",
            updatedAt: d.updatedAt || 0,
          };
        } catch { return null; }
      })
      .filter(Boolean)
      .sort((a, b) => b.updatedAt - a.updatedAt);
  } catch { return []; }
});

ipcMain.handle("draft:read", (_e, id) => {
  try {
    return JSON.parse(fs.readFileSync(getDraftPath(id), "utf8"));
  } catch { return null; }
});

ipcMain.handle("draft:write", (_e, id, data) => {
  const dir = ensureDraftsDir();
  const target = getDraftPath(id);
  const tmp = path.join(dir, `${id}.${process.pid}.${Date.now()}.tmp`);
  const payload = JSON.stringify({ ...data, updatedAt: Date.now() });
  fs.writeFileSync(tmp, payload, { encoding: "utf8", mode: 0o600 });
  fs.renameSync(tmp, target);
  try {
    const files = fs.readdirSync(dir).filter(f => f.endsWith(".json"));
    if (files.length > MAX_DRAFTS) {
      const sorted = files.map(f => {
        try { return { f, updatedAt: JSON.parse(fs.readFileSync(path.join(dir, f), "utf8")).updatedAt || 0 }; }
        catch { return { f, updatedAt: 0 }; }
      }).sort((a, b) => a.updatedAt - b.updatedAt);
      for (let i = 0; i < files.length - MAX_DRAFTS; i++) {
        fs.rmSync(path.join(dir, sorted[i].f), { force: true });
      }
    }
  } catch { }
});

ipcMain.handle("draft:delete", (_e, id) => {
  fs.rmSync(getDraftPath(id), { force: true });
});


ipcMain.handle("mate-notify:preview-sound", (_e, sound) => {
  playSound(sound);
});

ipcMain.handle("pick-folder", async (_event, opts = {}) => {
  const result = await dialog.showOpenDialog(win, {
    title: opts.title || "Select Folder",
    properties: ["openDirectory", "createDirectory"],
    defaultPath: opts.defaultPath || app.getPath("home"),
  });
  if (result.canceled || !result.filePaths.length) return { canceled: true };
  return { canceled: false, path: result.filePaths[0] };
});

// ── App lifecycle ─────────────────────────────────────────────────────────────

// macOS: BaseWindow has no default application menu, so Cmd+Z/X/C/V/A don't
// reach contenteditable elements unless an Edit menu is explicitly registered.
// role:"undo" click works because Electron calls webContents.undo() explicitly,
// but the keyboard accelerator goes through the native macOS undo: responder chain
// which doesn't reach WebContentsView in a BaseWindow. So we use an explicit
// click handler for undo/redo so both click AND keyboard call webContents.undo().
function getActiveWebContents() {
  if (activeSection && views[activeSection] && !views[activeSection].webContents.isDestroyed()) {
    return views[activeSection].webContents;
  }
  return null;
}

function buildAppMenu() {
  const template = [
    ...(process.platform === "darwin" ? [{ role: "appMenu" }] : []),
    {
      label: "Edit",
      submenu: [
        {
          label: "Undo",
          accelerator: "CmdOrCtrl+Z",
          // registerAccelerator:false → Cmd+Z is NOT intercepted by the OS menu;
          // it reaches the renderer as a keydown event so ProseMirror / CodeMirror
          // can handle it via their own Mod-z keymap bindings.
          // The click handler covers the mouse-click path via executeJavaScript.
          registerAccelerator: false,
          click: () => {
            const wc = getActiveWebContents();
            if (wc) wc.executeJavaScript("window.__vaultrUndo && window.__vaultrUndo()").catch(() => { });
          },
        },
        {
          label: "Redo",
          accelerator: "Shift+CmdOrCtrl+Z",
          registerAccelerator: false,
          click: () => {
            const wc = getActiveWebContents();
            if (wc) wc.executeJavaScript("window.__vaultrRedo && window.__vaultrRedo()").catch(() => { });
          },
        },
        { type: "separator" },
        { role: "cut" },
        { role: "copy" },
        { role: "paste" },
        { role: "pasteAndMatchStyle" },
        { role: "delete" },
        { role: "selectAll" },
      ],
    },
    { label: "Window", role: "windowMenu" },
  ];
  Menu.setApplicationMenu(Menu.buildFromTemplate(template));
}

// Must be called before app is ready
app.commandLine.appendSwitch("enable-gpu-rasterization");
app.commandLine.appendSwitch("enable-zero-copy");

app.whenReady().then(() => {
  buildAppMenu();
  const dockIconPath = getMacDockIconPath();
  if (app.dock && dockIconPath) {
    try {
      app.dock.setIcon(dockIconPath);
    } catch (e) {
      console.error("[vaultr-shell] app.dock.setIcon failed:", e);
    }
  }

  fs.mkdirSync(getDraftsDir(), { recursive: true });
  createWindow();

  app.on("activate", () => {
    if (!win || win.isDestroyed()) {
      createWindow();
    } else if (process.platform === "darwin") {
      void scheduleSyncVaultDataAcrossSectionViews();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
