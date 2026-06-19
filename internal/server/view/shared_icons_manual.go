package view

// svgPxPanel is a hand-crafted pixel layout icon for the reading drawer toggle.
// Not from pixelarticons — kept here so it isn't overwritten by `make icons`.
const svgPxPanel = `<svg fill="currentColor" viewBox="0 0 24 24"><path d="M4 2h16v2H4zm0 18h16v2H4zM2 4h2v16H2zm18 0h2v16h-2zm-6 0h2v16h-2z"/></svg>`

// svgPxBack is a hand-crafted pixelart left-pointing chevron for the back navigation button.
const svgPxBack = `<svg fill="currentColor" viewBox="0 0 24 24"><path d="M14 4h2v2h-2zM12 6h2v2h-2zM10 8h2v2h-2zM8 10h2v4H8zM10 14h2v2h-2zM12 16h2v2h-2zM14 18h2v2h-2z"/></svg>`

// svgBack is the Lucide "ChevronLeft" icon for the back navigation button.
const svgBack = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="m15 18-6-6 6-6"/></svg>`

// topbarIconBack bundles the smooth back chevron (shown by default) with the pixel variant
// (class lib-ai-px, hidden by default). pixelCSS swaps them when html[data-pixel="on"] is active.
const topbarIconBack = svgBack + `<svg class="lib-ai-px" fill="currentColor" viewBox="0 0 24 24"><path d="M14 4h2v2h-2zM12 6h2v2h-2zM10 8h2v2h-2zM8 10h2v4H8zM10 14h2v2h-2zM12 16h2v2h-2zM14 18h2v2h-2z"/></svg>`

// Smooth (non-pixel) variants of the three shared topbar action icons — Lucide icons.
const svgReload = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path stroke-linecap="round" stroke-linejoin="round" d="M3 3v5h5"/><path stroke-linecap="round" stroke-linejoin="round" d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/><path stroke-linecap="round" stroke-linejoin="round" d="M16 16h5v5"/></svg>`
const svgPanel = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24"><rect x="3" y="3" width="18" height="18" rx="2"/><path stroke-linecap="round" d="M15 3v18"/></svg>`
const svgSearch = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"/><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-4.34-4.34"/></svg>`

// topbarIconReload / topbarIconPanel / topbarIconSearch bundle a smooth icon (shown by
// default) and a pixel icon (class lib-ai-px, hidden by default).  pixelCSS swaps them
// when html[data-pixel="on"] is active.
const topbarIconReload = svgReload + `<svg class="lib-ai-px" fill="currentColor" viewBox="0 0 24 24"><path d="M16 4h2v6h-2zm-2-2h2v2h-2zm0 2h2v8h-2zM4 8H2v5h2z"/><path d="M4 6h16v2H4zm4 14H6v-6h2zm2 2H8v-2h2zm0-2H8v-8h2zm10-4h2v-5h-2z"/><path d="M20 18H4v-2h16z"/></svg>`
const topbarIconPanel = svgPanel + `<svg class="lib-ai-px" fill="currentColor" viewBox="0 0 24 24"><path d="M4 2h16v2H4zm0 18h16v2H4zM2 4h2v16H2zm18 0h2v16h-2zm-6 0h2v16h-2z"/></svg>`
const topbarIconSearch = svgSearch + `<svg class="lib-ai-px" fill="currentColor" viewBox="0 0 24 24"><path d="M22 22h-2v-2h2v2Zm-2-2h-2v-2h2v2Zm-6-2H6v-2h8v2Zm4 0h-2v-2h2v2ZM6 16H4v-2h2v2Zm10 0h-2v-2h2v2ZM4 14H2V6h2v8Zm14 0h-2V6h2v8ZM6 6H4V4h2v2Zm10 0h-2V4h2v2Zm-2-2H6V2h8v2Z"/></svg>`

// svgGraph is the Lucide "Network" icon used for the Knowledge Graph nav item.
const svgGraph = `<svg fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24">
        <rect x="16" y="16" width="6" height="6" rx="1"/>
        <rect x="2" y="16" width="6" height="6" rx="1"/>
        <rect x="9" y="2" width="6" height="6" rx="1"/>
        <path stroke-linecap="round" stroke-linejoin="round" d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/>
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 12V8"/>
      </svg>`

// svgShort is the Lucide "Zap" icon used in shortTriggerButton.
const svgShort = `<svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z"/></svg>`

// topbarIconShort bundles the smooth short-note icon with its pixel-art variant (svgPxZap).
const topbarIconShort = svgShort + `<svg class="lib-ai-px" fill="currentColor" viewBox="0 0 24 24"><path d="M4 13h8v6h2v2h-2v2h-2v-8H2v-4h2v2Zm12 6h-2v-2h2v2Zm2-2h-2v-2h2v2Zm2-2h-2v-2h2v2Zm-6-6h8v4h-2v-2h-8V5h-2V3h2V1h2v8Zm-8 2H4V9h2v2Zm2-2H6V7h2v2Zm2-2H8V5h2v2Z"/></svg>`
