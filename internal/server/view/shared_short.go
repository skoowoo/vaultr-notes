package view

// shortDialogCSS, shortDialogHTML, and shortDialogJS form a self-contained
// quick-capture dialog for short notes. Drop all three into any page shell to
// get a modal reachable via the toolbar button or the keyboard shortcut
// Ctrl+Shift+Space (⌘+Shift+Space on macOS).
//
// The dialog POSTs to POST /api/vault/shorts and calls
// window.__vaultrAfterVaultMutation (if defined) on success.

// shortTriggerButton is the toolbar icon button that opens the short note dialog.
// title reflects the keyboard shortcut.
const shortTriggerButton = `<button type="button" class="lib-action-btn"
              title="New short note (Ctrl+.)"
              onclick="window.openShortDialog && window.openShortDialog()">
        ` + topbarIconShort + `
      </button>`
