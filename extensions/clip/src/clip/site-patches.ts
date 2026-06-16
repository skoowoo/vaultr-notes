/**
 * Site-specific DOM patches applied before Readability runs.
 *
 * Each patch targets a specific platform or site family. Add new entries here
 * whenever a site requires structural DOM fixes that are not general enough to
 * live in pre-readability.ts.
 *
 * Interface:
 *   site     — human-readable platform name, used in comments / debug logs
 *   matches  — quick check run on the clone; return false to skip the patch
 *   apply    — DOM mutations on the clone (called only when matches() is true)
 */

export interface SitePatch {
  site: string;
  matches(clone: Document, url: string): boolean;
  apply(clone: Document): void;
}

// ── Patches ───────────────────────────────────────────────────────────────────

/**
 * [Squarespace] Fix deeply-nested Post Body layout that causes intro paragraphs
 * to be dropped by Readability.
 *
 * Problem: Squarespace blog posts nest content 5 levels deep inside
 *   sqs-layout > sqs-row > sqs-col > sqs-block > sqs-block-content > sqs-html-content > <p>
 * Readability propagates content scores upward with a 1/3 factor per level, so
 * a 400-char intro block scores ~5–10 at the sqs-row level — far below the
 * sibling-inclusion threshold of topCandidateScore × 0.2.
 *
 * Fix: flatten the Post Body into a single shallow <div> so all paragraphs and
 * figures are scored as one unit. Walk .sqs-block elements in document order to
 * preserve the original text/image interleaving.
 *
 * Verified on: tobyord.com (Squarespace 7.1 blog)
 */
const squarespacePost: SitePatch = {
  site: "Squarespace",

  matches: (clone) => !!clone.querySelector("[data-layout-label='Post Body']"),

  apply: (clone) => {
    clone.querySelectorAll("[data-layout-label='Post Body']").forEach((layout) => {
      const container = clone.createElement("div");

      layout.querySelectorAll(".sqs-block").forEach((block) => {
        // Text block: extract sqs-html-content children.
        const textContent = block.querySelector(".sqs-html-content");
        if (textContent) {
          const wrapper = clone.createElement("div");
          Array.from(textContent.childNodes).forEach((child) =>
            wrapper.appendChild(child.cloneNode(true)),
          );
          if (wrapper.textContent?.trim()) container.appendChild(wrapper);
          return;
        }
        // Image block: extract the figure element.
        const fig = block.querySelector("figure.sqs-block-image-figure");
        if (fig) {
          container.appendChild(fig.cloneNode(true));
        }
      });

      if (container.textContent?.trim()) {
        layout.parentNode?.replaceChild(container, layout);
      }
    });
  },
};

/**
 * [X / Twitter] Remove UI chrome from tweet/status pages so Readability only
 * sees the conversation thread content.
 *
 * Problem: X renders its UI almost entirely with SVG icons and React portals.
 * The sidebar, top navigation bar, and tweet action bars (like / retweet /
 * reply rows) are rendered alongside the article content in the DOM. Readability
 * cannot reliably distinguish them from real text, so it pulls them in.
 *
 * Fix: strip the known UI chrome containers by their data-testid attributes
 * before Readability scores the document.
 */
const twitterXPost: SitePatch = {
  site: "X / Twitter",

  matches: (_clone, url) => {
    try {
      const h = new URL(url).hostname;
      return h === "x.com" || h === "twitter.com" || h === "mobile.twitter.com";
    } catch {
      return false;
    }
  },

  apply: (clone) => {
    // Top navigation bar.
    clone.querySelectorAll('[data-testid="TopNavBar"]').forEach((el) => el.remove());

    // Right-hand sidebar (trends, "Who to follow", etc.).
    clone.querySelectorAll('[data-testid="sidebarColumn"]').forEach((el) => el.remove());

    // Tweet action bars: the like / retweet / reply / views / bookmark row.
    // These are <div role="group"> elements that contain only <button> children.
    clone.querySelectorAll('[role="group"]').forEach((el) => {
      const hasButtons = el.querySelector("button") !== null;
      const hasText = (el.textContent ?? "").trim().length > 0;
      // Remove action bars that are purely interactive (buttons only, no prose).
      if (hasButtons && !hasText) el.remove();
    });

    // "More" / "..." contextual menus rendered as top-level portal layers.
    clone.querySelectorAll('[data-testid="Dropdown"]').forEach((el) => el.remove());

    // Follow-suggestion banners injected between tweets.
    clone.querySelectorAll('[data-testid="UserCell"]').forEach((el) => el.remove());

    // "Show more replies" / "Show more tweets" buttons that add noise.
    clone.querySelectorAll('[data-testid="tweet_activity"]').forEach((el) => el.remove());
  },
};

// ── Registry ──────────────────────────────────────────────────────────────────

export const SITE_PATCHES: SitePatch[] = [
  squarespacePost,
  twitterXPost,
  // Add new site patches here.
];
