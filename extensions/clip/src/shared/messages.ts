export type ClipRequestMessage = { type: "CLIP_REQUEST" };
export type ClipResponseMessage =
  | { ok: true; title: string; url: string; markdown: string }
  | { ok: false; error: string };

export type ClipAndSaveMessage = { type: "CLIP_AND_SAVE"; tabId: number };

export type ClipAndSaveResult =
  | { ok: true; vaultPath: string }
  | { ok: false; error: string };

/** Extract markdown only — does not save to vault. */
export type ClipExtractMessage = { type: "CLIP_EXTRACT"; tabId: number };

export type ClipExtractResult =
  | { ok: true; title: string; url: string; markdown: string }
  | { ok: false; error: string };

/** Save a pre-extracted markdown string to the vault. */
export type SaveMarkdownMessage = {
  type: "SAVE_MARKDOWN";
  markdown: string;
  title: string;
};

export type SaveMarkdownResult =
  | { ok: true; vaultPath: string }
  | { ok: false; error: string };
