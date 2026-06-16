const SAFE = /[^a-zA-Z0-9\u4e00-\u9fff._-]+/g;

const DEFAULT_SAVE_DIR = "Web Clips";

export function slugify(title: string, maxLen = 80): string {
  const s = title
    .trim()
    .toLowerCase()
    .replace(SAFE, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
  if (!s) return "clip";
  return s.length > maxLen ? s.slice(0, maxLen) : s;
}

function pad2(n: number): string {
  return n < 10 ? `0${n}` : String(n);
}

export function formatPathTemplate(
  template: string,
  title: string,
  now = new Date(),
): string {
  const date = `${now.getFullYear()}-${pad2(now.getMonth() + 1)}-${pad2(now.getDate())}`;
  const time = `${pad2(now.getHours())}${pad2(now.getMinutes())}${pad2(now.getSeconds())}`;
  const slug = slugify(title);
  return template
    .replaceAll("{date}", date)
    .replaceAll("{time}", time)
    .replaceAll("{slug}", slug);
}

/** Reject path traversal and absolute paths */
export function assertSafeRelativePath(p: string): void {
  const norm = p.replace(/\\/g, "/").trim();
  if (!norm || norm.startsWith("/")) {
    throw new Error("Path must be relative to the vault (no leading slash).");
  }
  for (const part of norm.split("/")) {
    if (part === ".." || part === ".") {
      throw new Error("Path segments cannot be . or ..");
    }
  }
  if (!/\.(md|markdown)$/i.test(norm)) {
    throw new Error("Path must end with .md or .markdown");
  }
}

export function resolveNotePath(pageTitle: string, saveDir?: string): string {
  const dir = (saveDir ?? DEFAULT_SAVE_DIR).trim().replace(/^\/+|\/+$/g, "") || DEFAULT_SAVE_DIR;
  const template = `${dir}/{slug}.md`;
  const raw = formatPathTemplate(template, pageTitle);
  assertSafeRelativePath(raw);
  return raw;
}
