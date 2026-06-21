const { app } = require("electron");
const path = require("node:path");
const fs = require("node:fs");

const DEFAULT_SERVER_URL = "http://127.0.0.1:54321";

function getConfigPath() {
  return path.join(app.getPath("userData"), "config.json");
}

function loadConfig() {
  try {
    return JSON.parse(fs.readFileSync(getConfigPath(), "utf8"));
  } catch {
    return { serverUrl: DEFAULT_SERVER_URL };
  }
}

function saveConfig(config) {
  try {
    fs.writeFileSync(getConfigPath(), JSON.stringify(config, null, 2));
  } catch (e) {
    console.error("Failed to save config:", e);
  }
}

function getServerUrl() {
  return loadConfig().serverUrl;
}

function setServerUrl(url) {
  const c = loadConfig();
  c.serverUrl = url;
  c.autoStartServer = true;
  saveConfig(c);
}

function getAutoStart() {
  return loadConfig().autoStartServer !== false;
}

function setAutoStart(v) {
  const c = loadConfig();
  c.autoStartServer = v;
  saveConfig(c);
}

const DEFAULT_MATE_NOTIFY = {
  textEnabled: true,
  soundEnabled: true,
  startSound: "beep",
  doneSound: "beep",
};

function getMateNotifySettings() {
  const c = loadConfig();
  return { ...DEFAULT_MATE_NOTIFY, ...(c.mateNotify || {}) };
}

function setMateNotifySettings(settings) {
  const c = loadConfig();
  c.mateNotify = { ...getMateNotifySettings(), ...settings };
  saveConfig(c);
}

function registerConfigIpcHandlers(ipcMain) {
  ipcMain.handle("get-server-url", () => getServerUrl());
  ipcMain.handle("mate-notify:get-settings", () => getMateNotifySettings());
  ipcMain.handle("mate-notify:set-settings", (_e, settings) => {
    setMateNotifySettings(settings);
    return getMateNotifySettings();
  });
}

module.exports = {
  DEFAULT_SERVER_URL,
  getServerUrl,
  setServerUrl,
  getAutoStart,
  setAutoStart,
  getMateNotifySettings,
  setMateNotifySettings,
  registerConfigIpcHandlers,
};
