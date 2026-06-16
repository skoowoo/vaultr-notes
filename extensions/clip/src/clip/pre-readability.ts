/**
 * Stage 1 — Pre-Readability DOM preparation.
 *
 * All functions here operate on the cloned document before Readability parses
 * it. The goal is to repair or promote content that Readability would otherwise
 * strip or mishandle.
 *
 * Public surface:
 *   captureCanvasDataUrls(doc)          — must be called on the LIVE document
 *   prepareDocument(clone, base, urls)  — run all fixes on the clone
 */

import { SITE_PATCHES } from "./site-patches";

// ── Lazy-load attribute list ─────────────────────────────────────────────────

/** Attributes that hold the real image URL for lazy-loaded <img> (WeChat, many CDNs). */
const LAZY_IMG_URL_ATTRS = ["data-src", "data-original", "data-lazy-src", "data-backup"] as const;

// ── Internal helpers ─────────────────────────────────────────────────────────

function resolveLazyCandidateUrl(raw: string, base: string): string | null {
  const t = raw.trim();
  if (!t) return null;
  try {
    const href = new URL(t, base).href;
    if (href.startsWith("http://") || href.startsWith("https://")) return href;
  } catch {
    /* ignore */
  }
  return null;
}

/**
 * True when src is a common 1×1 / transparent placeholder (not a real article image).
 */
function isLikelyDataUriPlaceholder(src: string): boolean {
  const s = src.trim().toLowerCase();
  if (!s.startsWith("data:")) return false;
  if (s.startsWith("data:image/svg+xml")) {
    if (s.includes("width='1px'") || s.includes('width="1px"') || s.includes("width=1px")) return true;
    if (s.includes("viewbox='0 0 1 1'") || s.includes('viewbox="0 0 1 1"')) return true;
    if (s.includes("viewbox=0%200%201%201") || s.includes("viewbox='0%200%201%201'")) return true;
  }
  if (s.startsWith("data:image/gif;base64,") && src.length < 160) return true;
  return false;
}

// ── Fix functions ─────────────────────────────────────────────────────────────

/**
 * [Zhihu] Fix invalid nested lists: <ul> as a direct child of <ul>.
 *
 * Zhihu emits <ul><li>Parent</li><ul><li>Child</li></ul></ul>.
 * The inner <ul> must sit inside a <li>, but Zhihu places it as a sibling.
 * Browsers render the visual indentation via CSS; Readability sees a flat list.
 */
function fixInvalidNestedLists(doc: Document): void {
  doc.querySelectorAll("ul > ul, ul > ol, ol > ul, ol > ol").forEach((nested) => {
    const parent = nested.parentElement!;
    let prev = nested.previousElementSibling;
    while (prev && prev.nodeName !== "LI") {
      prev = prev.previousElementSibling;
    }
    if (prev) {
      prev.appendChild(nested);
    } else {
      const li = doc.createElement("li");
      parent.insertBefore(li, nested);
      li.appendChild(nested);
    }
  });

  doc.querySelectorAll("ul > br, ol > br").forEach((br) => br.remove());
}

/**
 * Convert <object type="image/…"> to <img> before Readability removes them.
 */
function convertObjectsToImages(doc: Document, base: string): void {
  doc.querySelectorAll("object[data]").forEach((obj) => {
    const type = obj.getAttribute("type") || "";
    if (!type.includes("image") && !type.includes("svg")) return;

    const data = obj.getAttribute("data") || "";
    if (!data) return;

    const img = doc.createElement("img");
    if (!data.startsWith("data:") && !data.startsWith("http")) {
      try {
        img.setAttribute("src", new URL(data, base).href);
      } catch {
        img.setAttribute("src", data);
      }
    } else {
      img.setAttribute("src", data);
    }
    img.setAttribute("alt", obj.getAttribute("alt") || obj.getAttribute("title") || "Embedded image");
    obj.parentNode?.replaceChild(img, obj);
  });
}

/**
 * Resolve relative <video> src/poster and <source> src URLs to absolute before
 * Readability processes them.
 */
function resolveVideoUrls(doc: Document, base: string): void {
  doc.querySelectorAll("video").forEach((video) => {
    const src = video.getAttribute("src");
    if (src && !src.startsWith("data:") && !src.startsWith("http")) {
      try {
        video.setAttribute("src", new URL(src, base).href);
      } catch {
        // keep original
      }
    }
    const poster = video.getAttribute("poster");
    if (poster && !poster.startsWith("data:") && !poster.startsWith("http")) {
      try {
        video.setAttribute("poster", new URL(poster, base).href);
      } catch {
        // keep original
      }
    }
    // Also resolve relative src on <source> children.
    video.querySelectorAll("source[src]").forEach((source) => {
      const ssrc = source.getAttribute("src")!;
      if (!ssrc.startsWith("data:") && !ssrc.startsWith("http")) {
        try {
          source.setAttribute("src", new URL(ssrc, base).href);
        } catch {
          // keep original
        }
      }
    });
  });
}

/**
 * Resolve relative <img> src URLs to absolute before Readability processes them.
 */
function resolveImageUrls(doc: Document, base: string): void {
  doc.querySelectorAll("img").forEach((img) => {
    const src = img.getAttribute("src");
    if (src && !src.startsWith("data:") && !src.startsWith("http")) {
      try {
        img.setAttribute("src", new URL(src, base).href);
      } catch {
        // keep original if URL construction fails
      }
    }
  });
}

/**
 * [WeChat & lazy-load sites] Promote real URL from data-src (etc.) when src is
 * empty, a data-URI placeholder, or a 1×1 SVG.
 */
function promoteLazyImageSources(doc: Document, base: string): void {
  doc.querySelectorAll("img").forEach((img) => {
    let lazyUrl: string | null = null;
    for (const attr of LAZY_IMG_URL_ATTRS) {
      const raw = img.getAttribute(attr);
      if (!raw) continue;
      lazyUrl = resolveLazyCandidateUrl(raw, base);
      if (lazyUrl) break;
    }
    if (!lazyUrl) return;

    const src = (img.getAttribute("src") || "").trim();
    if (!src) {
      img.setAttribute("src", lazyUrl);
      return;
    }
    if (isLikelyDataUriPlaceholder(src)) {
      img.setAttribute("src", lazyUrl);
      return;
    }
    if (src.toLowerCase().startsWith("data:image/svg+xml") && lazyUrl.includes("://")) {
      img.setAttribute("src", lazyUrl);
    }
  });
}

/**
 * Replace each <canvas> in the clone with an <img> element using the
 * pre-captured data URLs.
 */
function substituteCanvasImages(clone: Document, dataUrls: string[]): void {
  const canvases = Array.from(clone.querySelectorAll("canvas"));
  canvases.forEach((canvas, i) => {
    const dataUrl = dataUrls[i] ?? "";
    const label =
      canvas.getAttribute("aria-label") ||
      (canvas.id ? `Chart: ${canvas.id}` : "Chart");
    const img = clone.createElement("img");
    img.setAttribute("alt", label);
    if (dataUrl) {
      img.setAttribute("src", dataUrl);
    }
    canvas.parentNode?.replaceChild(img, canvas);
  });
}

/**
 * Replace interactive <input> elements with inline <span> elements containing
 * their current value.
 */
function inlineInputValues(clone: Document): void {
  const SKIP_TYPES = new Set(["hidden", "submit", "button", "image", "reset", "checkbox", "radio"]);
  clone.querySelectorAll("input").forEach((input) => {
    const type = (input.getAttribute("type") || "text").toLowerCase();
    if (SKIP_TYPES.has(type)) return;
    const value = input.getAttribute("value") ?? "";
    if (!value.trim()) return;
    const span = clone.createElement("span");
    span.textContent = value;
    input.parentNode?.replaceChild(span, input);
  });
}

/**
 * Convert "card deck" div layouts into <ul> or <table> elements so Readability
 * doesn't strip them as low-scoring short-text blocks.
 */
function convertCardGroups(clone: Document): void {
  const NUMERIC = /^[€$£¥₹+\-]?[\d][0-9,. ]*[%€$£¥₹kmbt]?$/i;
  const processed = new WeakSet<Element>();

  function leafTexts(el: Element): string[] {
    const out: string[] = [];
    (function walk(node: Node): void {
      if (node.nodeType === Node.TEXT_NODE) {
        const t = (node.textContent ?? "").trim();
        if (t) out.push(t);
      } else if (node.nodeType === Node.ELEMENT_NODE) {
        const tag = (node as Element).nodeName;
        if (tag === "SCRIPT" || tag === "STYLE") return;
        node.childNodes.forEach(walk);
      }
    })(el);
    return out;
  }

  function insideTable(el: Element): boolean {
    let p: Element | null = el.parentElement;
    while (p) {
      if (p.nodeName === "TABLE") return true;
      p = p.parentElement;
    }
    return false;
  }

  clone.querySelectorAll("div, ul, ol").forEach((parent) => {
    if (processed.has(parent) || insideTable(parent)) return;

    const children = Array.from(parent.children).filter(
      (c) => c.nodeName === "DIV" || c.nodeName === "LI",
    );
    if (children.length < 2) return;

    const cardLike = children.filter((c) => {
      const texts = leafTexts(c as Element);
      return texts.length >= 1 && texts.length <= 3;
    });
    if (cardLike.length < 2) return;

    const withNumeric = cardLike.filter((c) =>
      leafTexts(c as Element).some((t) => NUMERIC.test(t)),
    );
    if (withNumeric.length < 2) return;

    processed.add(parent);

    const rowData: string[][] = [];
    children.forEach((card) => {
      const texts = leafTexts(card as Element).filter((t) => t.length > 0);
      if (texts.length) rowData.push(texts);
    });
    if (rowData.length < 2) return;

    const maxCols = Math.max(...rowData.map((r) => r.length));

    if (maxCols <= 2) {
      const ul = clone.createElement("ul");
      rowData.forEach(([label, value]) => {
        const li = clone.createElement("li");
        if (value !== undefined) {
          const strong = clone.createElement("strong");
          strong.textContent = label;
          li.appendChild(strong);
          li.appendChild(clone.createTextNode(`: ${value}`));
        } else {
          li.textContent = label;
        }
        ul.appendChild(li);
      });
      parent.parentNode?.replaceChild(ul, parent);
    } else {
      const paddedRows = rowData.map((r) => {
        const padded = [...r];
        while (padded.length < maxCols) padded.push("");
        return padded;
      });

      const table = clone.createElement("table");
      const tbody = clone.createElement("tbody");
      table.appendChild(tbody);
      paddedRows.forEach((texts) => {
        const tr = clone.createElement("tr");
        texts.forEach((text) => {
          const td = clone.createElement("td");
          td.textContent = text;
          tr.appendChild(td);
        });
        tbody.appendChild(tr);
      });
      parent.parentNode?.replaceChild(table, parent);
    }
  });
}

/**
 * Convert inline <svg> elements to <img> tags with data: URIs so that
 * Readability preserves them and Turndown can emit Markdown image syntax.
 *
 * Alt text is sourced from (in priority order):
 *   1. aria-label / aria-labelledby on the <svg> itself
 *   2. A <title> child element inside the SVG
 *   3. A sibling/parent label element (.diagram-label, figcaption, etc.)
 *   4. Fallback: "Diagram"
 */
function convertInlineSvgsToImages(clone: Document): void {
  clone.querySelectorAll("svg").forEach((svg) => {
    // Skip decorative/icon SVGs — aria-hidden="true" is the standard marker for
    // purely visual elements (button icons, UI decorations) that carry no content.
    if (svg.getAttribute("aria-hidden") === "true") return;

    // 1. aria-label on the svg element
    let alt = svg.getAttribute("aria-label") || "";

    // 2. <title> child inside the SVG
    if (!alt) {
      const titleEl = svg.querySelector("title");
      if (titleEl) alt = titleEl.textContent?.trim() || "";
    }

    // 3. Sibling/parent label element
    if (!alt) {
      const parent = svg.parentElement;
      if (parent) {
        const label = parent.querySelector(
          ".diagram-label, figcaption, [class*='label'], [class*='caption']",
        );
        if (label) alt = label.textContent?.trim() || "";
        if (!alt) alt = parent.getAttribute("aria-label") || "";
      }
    }

    // Serialize SVG to string and encode as base64 data URI.
    // unescape(encodeURIComponent(...)) converts the UTF-16 JS string to a
    // UTF-8 byte sequence that btoa can handle without throwing on CJK chars.
    let dataUri: string;
    try {
      const serializer = new XMLSerializer();
      const svgStr = serializer.serializeToString(svg);
      const b64 = btoa(unescape(encodeURIComponent(svgStr)));
      dataUri = `data:image/svg+xml;base64,${b64}`;
    } catch {
      return; // serialisation failed — leave the svg in place
    }

    const img = clone.createElement("img");
    img.setAttribute("src", dataUri);
    img.setAttribute("alt", alt || "Diagram");
    svg.parentNode?.replaceChild(img, svg);
  });
}

/**
 * Replace <button> elements that wrap images (lightbox triggers, gallery cards,
 * etc.) with plain <div> elements. Readability strips every <button> it finds,
 * which silently removes any <img> nested inside.
 *
 * A button qualifies if it contains at least one <img> and no non-whitespace
 * text of its own (i.e. it is a purely visual/interactive wrapper, not a
 * labelled action button).
 */
function rescueImagesInButtons(clone: Document): void {
  clone.querySelectorAll("button").forEach((btn) => {
    if (!btn.querySelector("img")) return;
    // Keep only purely visual wrappers — skip buttons with their own label text.
    const ownText = Array.from(btn.childNodes)
      .filter((n) => n.nodeType === Node.TEXT_NODE)
      .map((n) => n.textContent ?? "")
      .join("")
      .trim();
    if (ownText) return;
    const div = clone.createElement("div");
    while (btn.firstChild) div.appendChild(btn.firstChild);
    btn.parentNode?.replaceChild(div, btn);
  });
}

// ── Public API ────────────────────────────────────────────────────────────────

/**
 * Capture rendered canvas pixel data from the LIVE document as data URLs.
 * Must be called BEFORE cloneNode() — canvas drawing buffers are not copied.
 */
export function captureCanvasDataUrls(doc: Document): string[] {
  const urls: string[] = [];
  doc.querySelectorAll("canvas").forEach((el) => {
    const canvas = el as HTMLCanvasElement;
    try {
      const url = canvas.toDataURL("image/png");
      // Blank canvas → tiny fixed URL (~170 chars). Cap at 1 MB.
      if (url.length > 2000 && url.length < 1_048_576) {
        urls.push(url);
      } else {
        urls.push("");
      }
    } catch {
      urls.push("");
    }
  });
  return urls;
}

/**
 * Apply all pre-Readability DOM fixes to the cloned document.
 *
 * @param clone        - Cloned document (not the live page).
 * @param base         - Base URL for resolving relative URLs.
 * @param canvasUrls   - Data URLs captured from the live document's canvases.
 */
export function prepareDocument(clone: Document, base: string, canvasUrls: string[], url = ""): void {
  fixInvalidNestedLists(clone);
  convertObjectsToImages(clone, base);
  promoteLazyImageSources(clone, base);
  resolveImageUrls(clone, base);
  resolveVideoUrls(clone, base);
  substituteCanvasImages(clone, canvasUrls);
  convertInlineSvgsToImages(clone);
  rescueImagesInButtons(clone);
  for (const patch of SITE_PATCHES) {
    if (patch.matches(clone, url)) patch.apply(clone);
  }
  convertCardGroups(clone);
  inlineInputValues(clone);
}
