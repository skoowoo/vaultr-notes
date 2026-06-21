const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("vaultrDesktop", {
  platform: process.platform,
  checkServer: (url) => ipcRenderer.invoke("check-server", url),
  vaultrOnPath: () => ipcRenderer.invoke("vaultr-on-path"),
  startVaultrServerDetached: (opts) => ipcRenderer.invoke("start-vaultr-server-detached", opts),
  stopServer: () => ipcRenderer.invoke("stop-vaultr-server"),
  getServerProcessStatus: () => ipcRenderer.invoke("get-server-process-status"),
  getShellDebugPaths: () => ipcRenderer.invoke("get-shell-debug-paths"),
  getServerUrl: () => ipcRenderer.invoke("get-server-url"),
  setServerUrl: (url) => ipcRenderer.invoke("set-server-url", url),
  restartServer: () => ipcRenderer.invoke("restart-vaultr-server"),
  navigateToLibrary: () => ipcRenderer.invoke("navigate-library"),
  syncVaultDataAcrossSections: () => ipcRenderer.invoke("sync-vault-data-across-sections"),
  setViewBgColor: (color) => ipcRenderer.send("set-view-bg-color", color),
  setWindowButtonVisibility: (visible) => ipcRenderer.send("set-window-button-visibility", visible),
  drafts: {
    list:   ()         => ipcRenderer.invoke("draft:list"),
    read:   (id)       => ipcRenderer.invoke("draft:read", id),
    write:  (id, data) => ipcRenderer.invoke("draft:write", id, data),
    delete: (id)       => ipcRenderer.invoke("draft:delete", id),
  },
  pickFolder: (opts) => ipcRenderer.invoke("pick-folder", opts),
  installCli: () => ipcRenderer.invoke("install-cli"),
  mateNotify: {
    getSettings: ()           => ipcRenderer.invoke("mate-notify:get-settings"),
    setSettings: (settings)   => ipcRenderer.invoke("mate-notify:set-settings", settings),
    previewSound: (sound)     => ipcRenderer.invoke("mate-notify:preview-sound", sound),
  },
});
