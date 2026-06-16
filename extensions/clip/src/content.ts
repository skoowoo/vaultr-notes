import type { ClipRequestMessage, ClipResponseMessage } from "./shared/messages.js";
import { clipPageToMarkdown } from "./clip/markdown.js";

chrome.runtime.onMessage.addListener(
  (msg: ClipRequestMessage, _sender, sendResponse: (r: ClipResponseMessage) => void) => {
    if (msg?.type !== "CLIP_REQUEST") {
      return;
    }

    try {
      const { title, url, markdown } = clipPageToMarkdown(document, window.location.href);
      sendResponse({ ok: true, title, url, markdown });
    } catch (e) {
      const err = e instanceof Error ? e.message : String(e);
      sendResponse({ ok: false, error: err });
    }
    return true;
  },
);
