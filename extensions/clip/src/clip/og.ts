function metaContent(el: Element): string {
  return (el.getAttribute("content") || el.getAttribute("value") || "").trim();
}

function resolveHttpUrl(raw: string, base: string): string | null {
  try {
    const href = new URL(raw, base).href;
    if (href.startsWith("http://") || href.startsWith("https://")) return href;
  } catch {
    /* ignore */
  }
  return null;
}

function firstMeta(
  doc: Document,
  names: string[],
): string {
  for (const name of names) {
    const lower = name.toLowerCase();
    for (const el of Array.from(doc.querySelectorAll("meta"))) {
      const prop = (
        el.getAttribute("property") ||
        el.getAttribute("name") ||
        ""
      )
        .trim()
        .toLowerCase();
      if (prop !== lower) continue;
      const content = metaContent(el);
      if (content) return content;
    }
  }
  return "";
}

export interface OgPreview {
  description?: string;
  site_name?: string;
  image_url?: string;
}

export function extractOgPreview(doc: Document, base: string): OgPreview {
  const description =
    firstMeta(doc, ["og:description", "twitter:description", "description"]) ||
    undefined;
  const site_name = firstMeta(doc, ["og:site_name"]) || undefined;
  const imageRaw = firstMeta(doc, [
    "og:image",
    "og:image:url",
    "og:image:secure_url",
    "twitter:image",
    "twitter:image:src",
  ]);
  const image_url = imageRaw ? resolveHttpUrl(imageRaw, base) || undefined : undefined;
  return { description, site_name, image_url };
}

export function appendOgFrontMatter(
  fm: string[],
  og: OgPreview,
  quote: (s: string) => string,
): void {
  if (og.description) fm.push(`description: ${quote(og.description)}`);
  if (og.site_name) fm.push(`site_name: ${quote(og.site_name)}`);
  if (og.image_url) fm.push(`image_url: ${quote(og.image_url)}`);
}
