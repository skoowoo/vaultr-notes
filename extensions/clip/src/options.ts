import { loadSettings, saveSettings, normalizeBaseUrl, normalizeSaveDir, type ClipSettings } from "./shared/settings.js";

function byId<T extends HTMLElement>(id: string): T {
  return document.getElementById(id) as T;
}

function setSaveFeedback(
  kind: "hidden" | "success" | "error",
  label: string,
  detail: string,
) {
  const box = byId<HTMLElement>("saveFeedback");
  const labelEl = byId<HTMLElement>("saveFeedbackLabel");
  const detailEl = byId<HTMLElement>("saveFeedbackDetail");

  if (kind === "hidden") {
    box.classList.add("is-hidden");
    box.classList.remove("feedback--success", "feedback--error");
    labelEl.textContent = "";
    detailEl.textContent = "";
    detailEl.className = "feedback__detail";
    return;
  }

  box.classList.remove("is-hidden");
  box.classList.toggle("feedback--success", kind === "success");
  box.classList.toggle("feedback--error", kind === "error");
  labelEl.textContent = label;
  detailEl.textContent = detail;
  detailEl.className = "feedback__detail";
}

async function ensureOriginForBase(baseUrl: string): Promise<void> {
  const origin = new URL(normalizeBaseUrl(baseUrl)).origin;
  const localhost =
    origin.startsWith("http://127.0.0.1") ||
    origin.startsWith("http://localhost") ||
    origin.startsWith("http://[::1]");
  if (localhost) return;

  const perm = { origins: [`${origin}/*`] } as chrome.permissions.Permissions;
  const has = await chrome.permissions.contains(perm);
  if (has) return;
  const granted = await chrome.permissions.request(perm);
  if (!granted) {
    throw new Error("Permission for this API origin was denied.");
  }
}

document.addEventListener("DOMContentLoaded", async () => {
  const s = await loadSettings();
  byId<HTMLInputElement>("baseUrl").value = s.baseUrl;
  byId<HTMLInputElement>("apiKey").value = s.apiKey;
  byId<HTMLInputElement>("saveDir").value = s.saveDir;

  byId<HTMLFormElement>("form").addEventListener("submit", async (e) => {
    e.preventDefault();
    setSaveFeedback("hidden", "", "");

    let next: ClipSettings;
    try {
      const baseUrl = normalizeBaseUrl(byId<HTMLInputElement>("baseUrl").value);
      await ensureOriginForBase(baseUrl);
      const saveDir = normalizeSaveDir(byId<HTMLInputElement>("saveDir").value || "Web Clips");
      next = {
        baseUrl,
        apiKey: byId<HTMLInputElement>("apiKey").value,
        saveDir,
      };
    } catch (err) {
      setSaveFeedback(
        "error",
        "Not saved",
        err instanceof Error ? err.message : String(err),
      );
      return;
    }

    try {
      await saveSettings(next);
      setSaveFeedback("success", "Settings saved", "Your preferences will apply to the next clip.");
    } catch (err) {
      setSaveFeedback(
        "error",
        "Not saved",
        err instanceof Error ? err.message : String(err),
      );
    }
  });
});
