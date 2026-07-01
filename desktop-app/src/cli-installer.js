"use strict";

const path = require("node:path");
const fs = require("node:fs");
const os = require("node:os");
const { spawnSync } = require("node:child_process");

const BINARY_NAME = process.platform === "win32" ? "vaultr.exe" : "vaultr";

// Sentinel file: records which app version last installed the CLI.
// If it matches the current app version, installation is skipped.
const SENTINEL_FILE = path.join(os.homedir(), ".vaultr", ".cli_app_version");

function getBundledArchive() {
  const { app } = require("electron");
  if (!app.isPackaged) return null; // dev mode: no bundled archive
  const ext = process.platform === "win32" ? "zip" : "tar.gz";
  return path.join(process.resourcesPath, "cli", `vaultr.${ext}`);
}

function getSystemInstallDir() {
  return path.join(os.homedir(), ".local", "bin");
}

/**
 * Install the bundled vaultr CLI by extracting the tar.gz from app resources.
 *
 * Skipped when the sentinel file matches the current app version.
 * Binary is always overwritten (keeps CLI in sync with app version).
 * config.toml and skills/ files are only written if absent.
 *
 * @param {(msg: string) => void} [log]
 * @returns {Promise<{ ok: boolean, skipped?: boolean, installDir?: string, error?: string }>}
 */
async function installCli(log = () => {}) {
  const { app } = require("electron");
  const appVersion = app.getVersion();
  const fsp = fs.promises;

  // Skip if this app version has already installed the CLI
  try {
    const installed = (await fsp.readFile(SENTINEL_FILE, "utf8")).trim();
    if (installed === appVersion) {
      log(`cli-installer: already up to date (v${appVersion}), skipping`);
      return { ok: true, skipped: true };
    }
  } catch { /* sentinel absent or unreadable → proceed */ }

  const archive = getBundledArchive();
  if (!archive) {
    log("cli-installer: no bundled archive (dev mode), skipping");
    return { ok: true, skipped: true };
  }

  try { await fsp.access(archive); } catch {
    log(`cli-installer: bundled archive not found at ${archive}`);
    return { ok: false, error: "Bundled CLI archive not found in app resources" };
  }

  // Extract archive to a temp directory
  let tmpDir;
  try {
    tmpDir = await fsp.mkdtemp(path.join(os.tmpdir(), "vaultr-install-"));
  } catch (e) {
    log(`cli-installer: mkdtemp failed: ${e.message}`);
    return { ok: false, error: e.message };
  }

  try {
    let r;
    if (process.platform === "win32") {
      r = spawnSync("powershell", [
        "-NoProfile", "-Command",
        `Expand-Archive -Path "${archive}" -DestinationPath "${tmpDir}" -Force`,
      ], { encoding: "utf8" });
    } else {
      r = spawnSync("tar", ["xzf", archive, "-C", tmpDir], { encoding: "utf8" });
    }
    if (r.error || r.status !== 0) {
      const reason = r.error?.message || r.stderr || `exit code ${r.status}`;
      log(`cli-installer: extraction failed: ${reason}`);
      return { ok: false, error: `archive extraction failed: ${reason}` };
    }
  } catch (e) {
    log(`cli-installer: extraction spawn failed: ${e.message}`);
    return { ok: false, error: e.message };
  }

  // Install binary (always overwrite — keeps CLI version in sync with app)
  const installDir = getSystemInstallDir();
  try {
    await fsp.mkdir(installDir, { recursive: true });
  } catch (e) {
    log(`cli-installer: mkdir ${installDir} failed: ${e.message}`);
    return { ok: false, error: `Cannot create install dir: ${e.message}` };
  }

  const destBin = path.join(installDir, BINARY_NAME);
  try {
    await fsp.copyFile(path.join(tmpDir, BINARY_NAME), destBin);
    if (process.platform !== "win32") await fsp.chmod(destBin, 0o755);
    log(`cli-installer: installed to ${destBin}`);
  } catch (e) {
    log(`cli-installer: binary copy failed: ${e.message}`);
    return { ok: false, error: e.message };
  }

  // config.toml — only write if not already present (never overwrite user config)
  try {
    const vaultrDir = path.join(os.homedir(), ".vaultr");
    await fsp.mkdir(vaultrDir, { recursive: true });

    const configDst = path.join(vaultrDir, "config.toml");
    try {
      await fsp.access(configDst);
      // file exists — skip
    } catch {
      try {
        await fsp.copyFile(path.join(tmpDir, "config.example.toml"), configDst);
        log(`cli-installer: installed default config to ${configDst}`);
      } catch { /* config.example.toml absent in archive */ }
    }

    // skills/ — only write files that do not already exist
    const skillsSrc = path.join(tmpDir, "skills");
    try {
      await fsp.access(skillsSrc);
      fs.cpSync(skillsSrc, path.join(vaultrDir, "skills"), {
        recursive: true,
        force: false,
        errorOnExist: false,
      });
      log(`cli-installer: installed built-in skills to ${path.join(vaultrDir, "skills")}`);
    } catch { /* skills dir absent in archive */ }
  } catch (e) {
    log(`cli-installer: config/skills copy failed (non-fatal): ${e.message}`);
  }

  // Clean up temp dir
  try { fs.rmSync(tmpDir, { recursive: true, force: true }); } catch { /* noop */ }

  // Write sentinel so subsequent launches with the same app version skip install
  try {
    await fsp.mkdir(path.dirname(SENTINEL_FILE), { recursive: true });
    await fsp.writeFile(SENTINEL_FILE, appVersion, "utf8");
  } catch (e) {
    log(`cli-installer: sentinel write failed (non-fatal): ${e.message}`);
  }

  return { ok: true, installDir };
}

module.exports = { installCli, BINARY_NAME };
