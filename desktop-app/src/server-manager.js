"use strict";

const { app, ipcMain } = require("electron");
const path = require("node:path");
const fs = require("node:fs");
const http = require("node:http");
const https = require("node:https");
const { spawn, spawnSync } = require("node:child_process");
const os = require("node:os");

// ── Diagnostics ───────────────────────────────────────────────────────────────

/** @type {{ [key: string]: number }} */
const shellDiagThrottle = {};
const CHECK_FAIL_LOG_GAP_MS = 12000;

function getShellDiagPaths() {
  const base = app.getPath("userData");
  return {
    userData: base,
    diagnostics: path.join(base, "vaultr-shell-diagnostics.log"),
    serverOutput: path.join(base, "vaultr-server-output.log"),
  };
}

function appendDiagFile(line) {
  try {
    const { diagnostics } = getShellDiagPaths();
    fs.appendFileSync(diagnostics, line);
  } catch (_) { /* noop */ }
}

function diagLog(...parts) {
  const text = `[${new Date().toISOString()}] ${parts.join(" ")}\n`;
  console.error("[vaultr-shell]", ...parts);
  appendDiagFile(text);
}

function diagLogCheckFailThrottled(url, reason) {
  const now = Date.now();
  const key = `${reason}|${url}`;
  const prev = shellDiagThrottle[key];
  if (prev != null && now - prev < CHECK_FAIL_LOG_GAP_MS) return;
  shellDiagThrottle[key] = now;
  diagLog("check-server", reason, url);
}

// ── PATH expansion for GUI launch ────────────────────────────────────────────
// On macOS, apps launched from Finder/Dock don't source ~/.zshrc, so
// process.env.PATH is the minimal system PATH.  Expand it with common
// user-level tool directories so Electron can locate the vaultr binary.
// Full shell env (http_proxy, tokens, etc.) is captured by the Go server
// at startup via agent.WarmShellEnv — no need to do it here.

function expandedEnv() {
  const home = os.homedir();
  const extra = [
    path.join(home, ".local", "bin"),
    path.join(home, "bin"),
    path.join(home, ".local", "share", "pnpm"),
    path.join(home, ".npm-global", "bin"),
    path.join(home, ".opencode", "bin"),
    path.join(home, ".volta", "bin"),
    "/opt/homebrew/bin",
    "/opt/homebrew/sbin",
    "/usr/local/bin",
  ];
  const current = process.env.PATH || "";
  const parts = new Set(current.split(path.delimiter).filter(Boolean));
  for (const p of extra) parts.add(p);
  return { ...process.env, PATH: [...parts].join(path.delimiter) };
}

// ── Vaultr binary resolution ──────────────────────────────────────────────────

const VAULTR_FALLBACK_BINS = [
  "/usr/local/bin/vaultr",
  path.join(os.homedir(), ".local", "bin", "vaultr"),
];

/** Returns the first working vaultr binary path, or null if none found. */
function resolveVaultrBin() {
  for (const bin of ["vaultr", ...VAULTR_FALLBACK_BINS]) {
    try {
      const r = spawnSync(bin, ["--version"], {
        encoding: "utf8", timeout: 8000, windowsHide: true, env: expandedEnv(),
      });
      if (!r.error && r.status === 0) return bin;
    } catch { /* try next */ }
  }
  return null;
}

// ── PID / process helpers ─────────────────────────────────────────────────────

function getServerPIDFilePath() {
  return path.join(app.getPath("userData"), "vaultr-server.pid");
}

function readServerPID() {
  try {
    const content = fs.readFileSync(getServerPIDFilePath(), "utf8").trim();
    const pid = parseInt(content, 10);
    return isNaN(pid) || pid <= 0 ? null : pid;
  } catch {
    return null;
  }
}

function isProcessAlive(pid) {
  try { process.kill(pid, 0); return true; } catch { return false; }
}

// ── Server spawn ──────────────────────────────────────────────────────────────

/** PID of the vaultr server process most recently spawned by this Electron instance. */
let managedServerChildPid = 0;

/** @returns {{ available: true } | { available: false, detail: string }} */
function probeVaultrCliSync() {
  try {
    const bin = resolveVaultrBin();
    if (bin) {
      diagLog("vaultr found at:", bin);
      return { available: true };
    }
    return {
      available: false,
      detail: "`vaultr` not found in PATH, /usr/local/bin, or ~/.local/bin.",
    };
  } catch (e) {
    diagLog("probeVaultrCli sync exception:", e);
    return { available: false, detail: String(e.message || e).slice(0, 320) };
  }
}

/** Detached subprocess; stdout/stderr go to userData log. Survives after Electron exits. */
function startVaultrServerDetached() {
  const { serverOutput } = getShellDiagPaths();
  try {
    fs.mkdirSync(path.dirname(serverOutput), { recursive: true });
    fs.appendFileSync(
      serverOutput,
      `\n--- ${new Date().toISOString()} spawn vaultr start server ---\n`
    );
  } catch (e) {
    diagLog("server output log preamble failed:", e.message);
  }

  return new Promise((resolve) => {
    let settled = false;
    const finish = (/** @type {{ ok: boolean, error?: string, pid?: number }} */ out) => {
      if (settled) return;
      settled = true;
      resolve(out);
    };

    const bin = resolveVaultrBin();
    if (!bin) {
      finish({ ok: false, error: "`vaultr` not found in PATH, /usr/local/bin, or ~/.local/bin." });
      return;
    }

    /** @type {number | undefined} */
    let logFd;
    try {
      logFd = fs.openSync(serverOutput, "a");
    } catch (e) {
      diagLog("open server output log:", e.message);
      finish({ ok: false, error: `cannot open log file (${e.message})` });
      return;
    }

    let child;
    try {
      child = spawn(bin, ["start", "server", "--pid-file", getServerPIDFilePath()], {
        detached: true,
        stdio: ["ignore", logFd, logFd],
        windowsHide: true,
        env: expandedEnv(),
      });
    } catch (e) {
      try { fs.closeSync(logFd); } catch (_) { /* noop */ }
      diagLog("spawn vaultr start server threw:", e.message);
      finish({ ok: false, error: e.message });
      return;
    }

    try { fs.closeSync(logFd); } catch (_) { /* noop */ }

    child.once("error", (err) => {
      diagLog("spawn vaultr start server event error:", err.message);
      finish({ ok: false, error: err.message });
    });
    child.unref();

    setImmediate(() => {
      if (settled) return;
      const pid = child.pid;
      if (typeof pid === "number" && pid > 0) {
        diagLog("started vaultr server, pid=", pid, "logs append to", serverOutput);
        finish({ ok: true, pid });
      } else {
        diagLog("spawn produced no pid");
        finish({ ok: false, error: "spawn failed (no process id)" });
      }
    });
  });
}

// ── IPC registration ──────────────────────────────────────────────────────────

/**
 * @param {{
 *   onRestartDone?: () => void,
 *   onServerStopped?: () => void,
 *   getAutoStart?: () => boolean,
 *   setAutoStart?: (v: boolean) => void,
 * }} [opts]
 */
function register(opts = {}) {
  const { onRestartDone, onServerStopped, getAutoStart, setAutoStart } = opts;

  /** Schedule a UI-transition callback after the current IPC promise resolves. */
  function deferUiTransition(fn) {
    if (fn) setTimeout(() => { try { fn(); } catch { /* noop */ } }, 0);
  }

  /** Shared SIGTERM → wait → optional SIGKILL logic. Returns true when the process is gone. */
  async function killProcess(pid) {
    if (!isProcessAlive(pid)) return true;
    try { process.kill(pid, "SIGTERM"); } catch { return false; }
    const deadline = Date.now() + 8000;
    while (Date.now() < deadline) {
      await new Promise((r) => setTimeout(r, 150));
      if (!isProcessAlive(pid)) return true;
    }
    diagLog("kill-process: still alive after 8s, sending SIGKILL to pid", pid);
    try { process.kill(pid, "SIGKILL"); } catch { /* noop */ }
    await new Promise((r) => setTimeout(r, 300));
    return !isProcessAlive(pid);
  }

  // ── check-server ────────────────────────────────────────────────────────────

  ipcMain.handle("check-server", async (_event, url) => {
    return new Promise((resolve) => {
      try {
        const parsed = new URL(url);
        const mod = parsed.protocol === "https:" ? https : http;
        const req = mod.request(
          {
            method: "GET",
            hostname: parsed.hostname,
            port: parsed.port || (parsed.protocol === "https:" ? 443 : 80),
            path: "/library",
            timeout: 3000,
          },
          (res) => { res.resume(); resolve(true); }
        );
        req.on("error", (err) => {
          diagLogCheckFailThrottled(url, `http error: ${err?.message || err}`);
          resolve(false);
        });
        req.on("timeout", () => {
          diagLogCheckFailThrottled(url, "timeout GET /library (3s)");
          req.destroy();
          resolve(false);
        });
        req.end();
      } catch (err) {
        diagLog("check-server bad URL:", url, err);
        resolve(false);
      }
    });
  });

  // ── vaultr-on-path ──────────────────────────────────────────────────────────

  ipcMain.handle("vaultr-on-path", () => probeVaultrCliSync());

  // ── start-vaultr-server-detached ────────────────────────────────────────────

  ipcMain.handle("start-vaultr-server-detached", async (_event, invokeOpts = {}) => {
    const { userInitiated = false } = invokeOpts;

    // Block automatic start when the user intentionally stopped the server.
    // A user-initiated attempt (clicking Connect) re-enables auto-start.
    if (!userInitiated && getAutoStart && !getAutoStart()) {
      diagLog("start-server: autoStart disabled, skip spawn");
      return { ok: false, reason: "stopped_by_user" };
    }
    if (userInitiated && setAutoStart) {
      setAutoStart(true);
    }

    // If Electron already spawned a server that is still initialising, skip the
    // duplicate spawn and let the start screen keep polling.
    if (managedServerChildPid > 0 && isProcessAlive(managedServerChildPid)) {
      diagLog("start-server: managed server pid=%d alive, skip duplicate spawn", managedServerChildPid);
      return { ok: true };
    }
    const result = await startVaultrServerDetached();
    if (result.ok && result.pid) managedServerChildPid = result.pid;
    return result;
  });

  // ── get-shell-debug-paths ───────────────────────────────────────────────────

  ipcMain.handle("get-shell-debug-paths", () => getShellDiagPaths());

  // ── get-server-process-status ───────────────────────────────────────────────

  ipcMain.handle("get-server-process-status", () => {
    const pid = readServerPID();
    return {
      managed: pid !== null,
      pid,
      alive: pid !== null && isProcessAlive(pid),
    };
  });

  // ── stop-vaultr-server ──────────────────────────────────────────────────────

  ipcMain.handle("stop-vaultr-server", async () => {
    const pid = readServerPID();
    if (!pid) {
      return { ok: false, reason: "no_pid", error: "No managed server process found" };
    }

    diagLog("stop-server: stopping managed server pid", pid);
    await killProcess(pid);
    managedServerChildPid = 0;
    if (setAutoStart) setAutoStart(false);
    deferUiTransition(onServerStopped);
    return { ok: true };
  });

  // ── install-cli ─────────────────────────────────────────────────────────────

  ipcMain.handle("install-cli", () => {
    return new Promise((resolve) => {
      // Use a login shell so the user's profile (.zprofile, .bash_profile, etc.) is sourced,
      // which brings in proxy env vars (http_proxy, https_proxy, etc.) and extended PATH —
      // both are absent from Electron's minimal process.env when launched from the Dock.
      const shell = process.env.SHELL || "/bin/sh";
      const child = spawn(
        shell,
        ["-l", "-i", "-c", "curl -sL https://raw.githubusercontent.com/skoowoo/vaultr-notes/main/install-cli.sh | sh"],
        { stdio: ["ignore", "pipe", "pipe"], windowsHide: true }
      );

      let output = "";
      child.stdout.on("data", (chunk) => { output += chunk.toString(); });
      child.stderr.on("data", (chunk) => { output += chunk.toString(); });

      child.on("close", (code) => {
        diagLog("install-cli: exit code", code);
        if (code !== 0) {
          resolve({ ok: false, output: output.slice(-2000).trim() });
          return;
        }
        const bin = resolveVaultrBin();
        if (bin) {
          diagLog("install-cli: verified vaultr at:", bin);
          resolve({ ok: true, output: output.slice(-2000).trim() });
        } else {
          diagLog("install-cli: verify failed, vaultr not found after install");
          resolve({ ok: false, output: output.slice(-2000).trim(), error: "installed but vaultr not found in PATH, /usr/local/bin, or ~/.local/bin" });
        }
      });

      child.on("error", (err) => {
        diagLog("install-cli: spawn error:", err.message);
        resolve({ ok: false, error: err.message, output: output.trim() });
      });
    });
  });

  // ── restart-vaultr-server ───────────────────────────────────────────────────

  ipcMain.handle("restart-vaultr-server", async () => {
    const pid = readServerPID();
    if (!pid) {
      return { ok: false, reason: "no_pid", error: "No valid PID file — server was not started by the desktop app" };
    }

    if (isProcessAlive(pid)) {
      diagLog("restart-server: stopping pid", pid);
      const killed = await killProcess(pid);
      if (!killed) {
        diagLog("restart-server: could not confirm process death, proceeding anyway");
      }
    } else {
      diagLog("restart-server: pid", pid, "already not running, spawning fresh");
    }

    diagLog("restart-server: spawning new vaultr server");
    managedServerChildPid = 0;
    const spawnResult = await startVaultrServerDetached();
    if (!spawnResult.ok) {
      return { ok: false, reason: "spawn_failed", error: spawnResult.error };
    }
    managedServerChildPid = spawnResult.pid;
    deferUiTransition(onRestartDone);
    return { ok: true, pid: spawnResult.pid };
  });
}

module.exports = { register };
