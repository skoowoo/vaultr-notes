package view

// settingsModalCSS contains the modal overlay styles and all settings-specific CSS.
const settingsModalCSS = `
    /* ── Settings modal overlay ─────────────────────────────── */
    [x-cloak] { display: none !important; }
    .settings-modal-overlay {
      position: fixed; inset: 0; z-index: 1000;
      background: var(--overlay-bg);
      display: flex; align-items: center; justify-content: center;
    }
    .settings-modal-panel {
      width: 1040px; max-width: calc(100vw - 2rem);
      height: 720px; max-height: calc(100vh - 2rem);
      background: var(--bg); border: 1px solid var(--hr);
      border-radius: var(--radius-lg);
      display: flex; flex-direction: column; overflow: hidden;
      box-shadow: 0 24px 64px rgba(0,0,0,0.45);
    }
    html[data-theme="light"] .settings-modal-panel { box-shadow: 0 24px 64px rgba(0,0,0,0.18); }
    .settings-modal-bar {
      flex-shrink: 0; display: flex; align-items: center; justify-content: space-between;
      height: 40px; padding: 0 1rem; border-bottom: 1px solid var(--hr);
      user-select: none;
    }
    .settings-modal-title { font-size: var(--text-sm); font-weight: 600; color: var(--fg); }
    .settings-modal-close-btn {
      display: flex; align-items: center; justify-content: center;
      width: 28px; height: 28px; border-radius: var(--radius-sm);
      border: none; background: transparent; color: var(--muted);
      cursor: pointer; transition: background 120ms, color 120ms;
    }
    .settings-modal-close-btn:hover { background: var(--nav-act); color: var(--fg); }
    .settings-modal-close-btn svg { width: 15px; height: 15px; }
    .settings-modal-inner {
      flex: 1; min-height: 0; display: flex; overflow: hidden;
    }

    /* ── Settings inner layout ───────────────────────────────── */
    .settings-body { flex: 1; display: flex; min-height: 0; overflow: hidden; }

    /* ── Primary sidebar ──────────────────────────────────────── */
    .settings-sidebar {
      flex-shrink: 0; width: 196px; border-right: 1px solid var(--hr);
      background: var(--surface-soft);
      padding: 1rem 0.5rem; display: flex; flex-direction: column;
      gap: 2px; user-select: none;
    }
    .settings-sidebar-item {
      display: flex; align-items: center; gap: 0.5rem; width: 100%;
      padding: 0.42rem 0.75rem; border-radius: var(--radius-md); border: none;
      background: transparent; font-size: var(--text-base); font-weight: 500;
      color: var(--muted); cursor: pointer; text-align: left;
      transition: background 120ms, color 120ms;
    }
    .settings-sidebar-item:hover { color: var(--fg); background: var(--card-hov); }
    .settings-sidebar-item.active { color: var(--btn-primary-fg); background: var(--link); }
    .settings-sidebar-item svg { width: 14px; height: 14px; flex-shrink: 0; }

    /* ── Content area ─────────────────────────────────────────── */
    .settings-content { flex: 1; min-width: 0; display: flex; flex-direction: column; overflow: hidden; }

    /* ── Pane ─────────────────────────────────────────────────── */
    .settings-pane { flex: 1; overflow-y: auto; padding: 1.75rem 1.5rem 3rem; }
    ::-webkit-scrollbar { display: none; }

    /* ── Appearance fields ────────────────────────────────────── */
    .settings-fields { max-width: 640px; display: flex; flex-direction: column; gap: 1.75rem; }
    .settings-field-label {
      display: block; font-size: var(--text-sm); font-weight: 600;
      color: var(--fg); margin-bottom: 0.5rem;
    }
    .settings-field-desc { font-size: var(--text-xs); color: var(--muted); margin-top: 0.4rem; line-height: 1.5; }
    .settings-field-row { display: flex; gap: 0.5rem; align-items: center; }
    .settings-input {
      flex: 1; min-width: 0; background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-md); padding: 0.6rem 0.75rem; font-size: var(--text-sm); color: var(--fg);
      outline: none; font-family: var(--font-mono);
      transition: border-color 150ms;
    }
    .settings-input:focus { border-color: var(--link); }
    .settings-apply-btn, .cfg-save-btn, .agents-toolbar-btn {
      height: 32px; padding: 0 0.875rem; border-radius: var(--radius-md);
      border: 1px solid var(--card-bd); background: transparent; color: var(--fg);
      font-size: var(--text-sm); font-weight: 500; cursor: pointer; white-space: nowrap;
      transition: background 120ms, border-color 120ms;
      display: inline-flex; align-items: center; gap: 0.35rem;
    }
    .settings-apply-btn:hover:not(:disabled),
    .cfg-save-btn:hover:not(:disabled),
    .agents-toolbar-btn:hover:not(:disabled) { background: var(--card-hov); border-color: var(--muted); }
    .settings-apply-btn:disabled,
    .cfg-save-btn:disabled,
    .agents-toolbar-btn:disabled { opacity: 0.4; cursor: not-allowed; }
    .settings-apply-btn--danger { color: var(--s-err); border-color: rgba(217,112,112,0.25); }
    .settings-apply-btn--danger:hover:not(:disabled) { background: rgba(217,112,112,0.06); border-color: rgba(217,112,112,0.4); }
    .settings-error { margin-top: 0.4rem; font-size: var(--text-xs); color: var(--s-err); }

    /* ── Theme / effect segmented control ────────────────────── */
    .theme-seg {
      display: inline-flex; background: var(--surface-soft);
      border-radius: var(--radius-pill); padding: 4px; gap: 2px;
    }
    .theme-seg-btn {
      display: flex; align-items: center; gap: 0.375rem;
      padding: 0.28rem 0.85rem; border-radius: var(--radius-pill); border: none;
      background: transparent; font-size: var(--text-sm); font-weight: 500;
      color: var(--muted); cursor: pointer;
      transition: background 120ms, color 120ms, box-shadow 120ms;
    }
    .theme-seg-btn:hover { color: var(--fg); }
    .theme-seg-btn.active {
      background: var(--bg); color: var(--fg);
      box-shadow: 0 1px 3px rgba(0,0,0,0.18), 0 1px 2px rgba(0,0,0,0.10);
    }
    html[data-theme="light"] .theme-seg-btn.active {
      box-shadow: 0 1px 3px rgba(0,0,0,0.10), 0 1px 2px rgba(0,0,0,0.06);
    }
    .theme-seg-btn svg { width: 13px; height: 13px; flex-shrink: 0; }

    /* ── Server config ────────────────────────────────────────── */
    .cfg-content { flex: 1; min-width: 0; display: flex; flex-direction: column; overflow: hidden; }
    .cfg-action-bar {
      flex-shrink: 0; display: flex; align-items: center;
      justify-content: space-between; flex-wrap: wrap;
      gap: 0.75rem 1rem; margin-top: 2rem; padding-top: 1.25rem;
      border-top: 1px solid var(--hr); max-width: 640px;
    }
    .cfg-action-left { display: flex; align-items: center; gap: 0.75rem; flex: 1; min-width: 0; }
    .cfg-action-right { display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0; }
    .cfg-status-ok { font-size: var(--text-xs); color: var(--s-ok); font-weight: 500; }
    .cfg-status-err {
      font-size: var(--text-xs); color: var(--s-err); font-weight: 500;
      overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
    }
    .cfg-restart-note { font-size: var(--text-xs); color: var(--muted); }
    .cfg-dirty-badge {
      font-size: var(--text-xs); color: var(--muted); background: var(--code-bg);
      border: 1px solid var(--code-bd); border-radius: var(--radius-xs); padding: 0.12rem 0.5rem;
    }
    .cfg-discard-btn {
      height: 32px; padding: 0 0.875rem; border-radius: var(--radius-md);
      border: 1px solid var(--card-bd); background: transparent; color: var(--muted);
      font-size: var(--text-sm); font-weight: 500; cursor: pointer;
      transition: color 120ms, background 120ms, border-color 120ms;
    }
    .cfg-discard-btn:hover { color: var(--fg); background: var(--card-hov); border-color: var(--muted); }
    .cfg-pane-area { flex: 1; min-height: 0; position: relative; overflow: hidden; }
    .cfg-pane {
      position: absolute; inset: 0; overflow-y: auto;
      padding: 1.75rem 1.5rem 3rem;
    }
    .cfg-fields { max-width: 640px; display: flex; flex-direction: column; }
    .cfg-section {
      max-width: 640px; margin-bottom: 0.45rem;
      border-radius: var(--radius-lg); border: 1px solid var(--card-bd);
      background: transparent; overflow: hidden;
      transition: background 150ms, border-color 150ms;
    }
    .cfg-section.is-open { background: var(--surface-soft); border-color: var(--hr); }
    .cfg-section-head {
      display: flex; align-items: flex-start; justify-content: space-between;
      gap: 0.75rem; width: 100%; margin: 0; padding: 0.7rem 0.9rem;
      border: none; background: transparent; cursor: pointer;
      text-align: left; color: inherit; transition: background 120ms;
    }
    .cfg-section-head:hover { background: var(--nav-act); }
    .cfg-section.is-open .cfg-section-head { border-bottom: 1px solid var(--hr); }
    .cfg-section-head-text { min-width: 0; flex: 1; }
    .cfg-section-title {
      font-size: var(--text-sm); font-weight: 600; letter-spacing: -0.01em;
      color: var(--fg); margin: 0; padding: 0; display: block;
    }
    .cfg-section-head-desc {
      display: block; font-size: var(--text-xs); color: var(--muted);
      margin: 0.28rem 0 0; line-height: 1.5;
    }
    .cfg-section-head-meta {
      display: flex; align-items: center; gap: 0.4rem; flex-shrink: 0; margin-top: 0.1rem;
    }
    .cfg-section-dirty-dot {
      width: 6px; height: 6px; border-radius: 50%; background: var(--link); flex-shrink: 0;
    }
    .cfg-section-chev {
      flex-shrink: 0; color: var(--muted); display: flex; align-items: center; margin-top: 0.15rem;
    }
    .cfg-section-chev svg { width: 14px; height: 14px; transition: transform 150ms ease; }
    .cfg-section-chev.open svg { transform: rotate(90deg); }
    .cfg-section-body { padding: 0; }
    .cfg-section.is-open .cfg-section-body { padding: 0 0.9rem 0.9rem; }
    .cfg-field {
      display: grid;
      grid-template-columns: 1fr 220px;
      grid-template-areas: "meta ctrl" "desc desc";
      column-gap: 1rem; padding: 0.88rem 0;
      border-bottom: 1px solid var(--hr); align-items: center;
    }
    .cfg-field.multiline {
      grid-template-columns: 1fr;
      grid-template-areas: "meta" "ctrl" "desc";
      align-items: start;
    }
    .cfg-field.multiline .cfg-field-ctrl { justify-content: stretch; margin-top: 0.5rem; }
    .cfg-section.is-open .cfg-field:last-of-type { border-bottom: none; }
    .cfg-field.dirty .cfg-field-label::before {
      content: ''; display: inline-block; width: 5px; height: 5px;
      border-radius: 50%; background: var(--link);
      margin-right: 5px; vertical-align: middle; margin-bottom: 1px;
    }
    .cfg-field-meta { grid-area: meta; min-width: 0; }
    .cfg-field-ctrl { grid-area: ctrl; display: flex; align-items: center; justify-content: flex-end; }
    .cfg-field-label { font-size: var(--text-sm); font-weight: 500; color: var(--body); display: block; margin-bottom: 0.15rem; }
    .cfg-field-key {
      font-size: var(--text-xs); color: var(--muted); opacity: 0.65;
      font-family: var(--font-mono);
    }
    .cfg-field-desc { grid-area: desc; font-size: var(--text-xs); color: var(--muted); margin: 0.35rem 0 0; line-height: 1.5; }
    .cfg-wechat-auth { padding: 0.88rem 0 0; border-top: 1px solid var(--hr); }
    .cfg-wechat-auth-head {
      display: flex; align-items: center; justify-content: space-between;
      gap: 0.75rem; margin-bottom: 0.65rem;
    }
    .cfg-wechat-auth-title { font-size: var(--text-sm); font-weight: 500; color: var(--body); }
    .cfg-wechat-badge {
      font-size: var(--text-sm); font-weight: 500;
      padding: 3px 10px; border-radius: var(--radius-pill);
      color: var(--muted); background: var(--code-bg);
    }
    .cfg-wechat-badge.connected { color: var(--p3); background: var(--s-ok-bg); }
    .cfg-wechat-meta { font-size: var(--text-xs); color: var(--muted); line-height: 1.55; margin: 0 0 0.75rem; }
    .cfg-wechat-meta code {
      font-size: var(--text-2xs); padding: 0.05rem 0.3rem; border-radius: var(--radius-xs);
      background: var(--card-bg); border: 1px solid var(--code-bd);
    }
    .cfg-wechat-actions { display: flex; flex-wrap: wrap; gap: 0.5rem; align-items: center; }
    .cfg-wechat-qr {
      margin-top: 0.85rem; padding: 0.75rem; border: 1px solid var(--code-bd);
      border-radius: var(--radius-md); background: var(--code-bg); text-align: center;
    }
    .cfg-wechat-qr img { width: 180px; height: 180px; object-fit: contain; background: #fff; border-radius: var(--radius-sm); }
    .cfg-wechat-status { font-size: var(--text-xs); color: var(--muted); margin-top: 0.5rem; }
    .cfg-wechat-err { font-size: var(--text-xs); color: var(--s-err); margin-top: 0.5rem; }
    .cfg-wechat-ok { font-size: var(--text-xs); color: var(--s-ok); margin-top: 0.5rem; }
    .cfg-input {
      width: 100%; background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-md); padding: 0.6rem 0.65rem;
      font-size: var(--text-sm); color: var(--fg); outline: none;
      font-family: var(--font-mono);
      transition: border-color 150ms;
    }
    .cfg-input:focus { border-color: var(--link); }
    select.cfg-input { cursor: pointer; }
    input[type="number"].cfg-input { width: 110px; }
    .cfg-textarea {
      width: 100%; resize: vertical; min-height: 60px;
      background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-md); padding: 0.6rem 0.65rem;
      font-size: var(--text-sm); color: var(--fg); outline: none;
      font-family: var(--font-mono);
      transition: border-color 150ms; line-height: 1.5;
    }
    .cfg-textarea:focus { border-color: var(--link); }
    .cfg-reveal-wrap { display: flex; gap: 0.375rem; align-items: center; width: 100%; }
    .cfg-reveal-btn {
      flex-shrink: 0; width: 32px; height: 32px; padding: 0;
      display: flex; align-items: center; justify-content: center;
      background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-md); cursor: pointer; color: var(--muted);
      transition: color 120ms, background 120ms;
    }
    .cfg-reveal-btn:hover { color: var(--fg); background: var(--surface-soft); }
    .cfg-reveal-btn svg { width: 13px; height: 13px; }
    .cfg-toggle { display: inline-flex; align-items: center; cursor: pointer; }
    .cfg-toggle input[type="checkbox"] { display: none; }
    .cfg-toggle-pill {
      width: 36px; height: 20px; background: var(--code-bg);
      border: 1px solid var(--code-bd); border-radius: var(--radius-pill);
      position: relative; transition: background 150ms, border-color 150ms;
    }
    .cfg-toggle-pill::after {
      content: ''; position: absolute; width: 14px; height: 14px;
      border-radius: 50%; background: var(--muted); top: 2px; left: 2px;
      transition: transform 150ms, background 150ms;
    }
    .cfg-toggle input:checked + .cfg-toggle-pill { background: var(--btn-primary-active); border-color: var(--btn-primary-active); }
    .cfg-toggle input:checked + .cfg-toggle-pill::after { transform: translateX(16px); background: var(--btn-primary-fg); }
    .cfg-loader { font-size: var(--text-xs); color: var(--muted); padding: 2rem 0; }
    .cfg-err-msg { font-size: var(--text-xs); color: var(--s-err); padding: 2rem 0; }

    /* ── Agents tab ───────────────────────────────────────────── */
    .agents-pane { flex: 1; overflow-y: auto; padding: 1.75rem 1.5rem 3rem; }
    .agents-toolbar { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem; }
    .agents-toolbar-btn svg { width: 13px; height: 13px; flex-shrink: 0; }
    @keyframes spin { to { transform: rotate(360deg); } }
    .agents-toolbar-btn.spinning svg { animation: spin 0.6s linear infinite; }
    .agents-summary { font-size: var(--text-xs); color: var(--muted); }
    .agents-list { display: flex; flex-direction: column; gap: 0.45rem; }
    .agent-card {
      border: 1px solid var(--card-bd); border-radius: var(--radius-lg);
      background: var(--bg); padding: 1rem 1.25rem;
      display: flex; flex-direction: column; gap: 0.625rem;
      transition: border-color 120ms, background 120ms;
      min-width: 0; overflow: hidden;
    }
    .agent-card:hover { background: var(--surface-soft); border-color: rgba(255,255,255,0.11); }
    html[data-theme="light"] .agent-card:hover { background: var(--surface-soft); border-color: rgba(0,0,0,0.12); }
    .agent-card.unavailable { opacity: 0.48; }
    .agent-card-top {
      display: flex; align-items: center; gap: 0.6rem; min-width: 0;
    }
    .agent-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
    .agent-dot.ok { background: var(--s-ok); }
    .agent-dot.off { background: var(--muted); opacity: 0.55; }
    .agent-card-name {
      font-size: var(--text-sm); font-weight: 600; color: var(--fg);
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
      flex-shrink: 0; max-width: 220px;
    }
    .agent-card-id {
      font-family: var(--font-mono);
      font-size: var(--text-xs); color: var(--fg); opacity: 0.6;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
      flex: 1; min-width: 0;
    }
    .agent-status-badge {
      font-size: var(--text-xs); font-weight: 500;
      padding: 2px 8px; border-radius: var(--radius-pill); flex-shrink: 0;
      background: var(--code-bg); color: var(--muted); white-space: nowrap;
    }
    .agent-status-badge.ok { color: var(--p3); background: var(--s-ok-bg); }
    /* row 2: path · version · protocol — single line, no wrap */
    .agent-card-info {
      display: flex; align-items: center; gap: 0.85rem;
      padding-left: 1.1rem; min-width: 0; overflow: hidden;
    }
    .agent-col-mono {
      font-family: var(--font-mono);
      font-size: var(--text-xs); color: var(--muted);
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    }
    .agent-meta-path { flex: 1; min-width: 0; max-width: 340px; }
    /* row 3: model pills — always its own row */
    .agent-card-models {
      display: flex; flex-wrap: wrap; gap: 4px; align-items: center;
      padding-left: 1.1rem; min-width: 0; overflow: hidden;
    }
    .agent-model-pill {
      font-size: var(--text-2xs); font-weight: 500; padding: 1px 7px; border-radius: var(--radius-pill);
      background: var(--code-bg); color: var(--muted);
      white-space: nowrap; max-width: 160px; overflow: hidden; text-overflow: ellipsis;
    }
    .agent-model-more { font-size: var(--text-xs); color: var(--muted); opacity: 0.6; white-space: nowrap; }
    /* row 4: cli example */
    .agent-card-cli {
      display: flex; align-items: center; gap: 0.5rem;
      padding-left: 1.1rem; min-width: 0; overflow: hidden;
    }
    .agent-cli-label {
      font-size: var(--text-2xs); font-weight: 600;
      color: var(--muted); opacity: 0.55;
      flex-shrink: 0; white-space: nowrap;
    }
    .agent-cli-code {
      font-family: var(--font-mono); font-size: var(--text-2xs);
      color: var(--muted); opacity: 0.75;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
      flex: 1; min-width: 0;
    }
    .agent-cli-copy {
      flex-shrink: 0; height: 20px; padding: 0 7px;
      border: 1px solid var(--code-bd); border-radius: var(--radius-xs);
      background: transparent; color: var(--muted);
      font-size: var(--text-2xs); font-family: var(--font-mono); cursor: pointer;
      transition: color 100ms, background 100ms;
      white-space: nowrap;
    }
    .agent-cli-copy:hover { color: var(--fg); background: var(--card-hov); }
    .agent-cli-copy.copied { color: var(--p3); border-color: var(--s-ok-bd); }

    /* ── Editor effects ───────────────────────────────────────── */
    .effect-card {
      display: flex; flex-direction: column; align-items: flex-start;
      padding: 0.38rem 0.8rem; border-radius: var(--radius-md); border: 1px solid var(--code-bd);
      background: var(--code-bg); cursor: pointer; min-width: 86px;
      transition: border-color 130ms, background 130ms; text-align: left;
    }
    .effect-card:hover { border-color: var(--muted); background: var(--card-hov); }
    .effect-card.active { border-color: var(--link); background: var(--card-hov); }
    .effect-card-name { font-size: var(--text-sm); font-weight: 600; color: var(--fg); }
    .effect-card-desc { font-size: var(--text-xs); color: var(--muted); margin-top: 0.1rem; white-space: nowrap; }

    /* ── Shortcuts pane ───────────────────────────────────────── */
    .shortcuts-pane { flex: 1; overflow-y: auto; padding: 1.75rem 1.5rem 3rem; }
    .shortcuts-fields { max-width: 640px; }
    .shortcuts-list { display: flex; flex-direction: column; }
    .shortcuts-row {
      display: flex; align-items: center; justify-content: space-between;
      padding: 0.55rem 0; border-bottom: 1px solid var(--hr); gap: 1rem;
    }
    .shortcuts-list .shortcuts-row:last-child { border-bottom: none; }
    .shortcuts-row-meta { min-width: 0; flex: 1; }
    .shortcuts-label { font-size: var(--text-sm); font-weight: 600; color: var(--fg); display: block; }
    .shortcuts-desc { font-size: var(--text-xs); color: var(--muted); margin-top: 0.1rem; }
    .shortcuts-keys { display: flex; gap: 0.3rem; align-items: center; flex-shrink: 0; }
    .kbd {
      display: inline-flex; align-items: center;
      padding: 0.18rem 0.42rem; border-radius: var(--radius-xs);
      background: var(--code-bg); border: 1px solid var(--code-bd);
      font-family: var(--font-mono);
      font-size: var(--text-xs); color: var(--fg); white-space: nowrap; line-height: 1.4;
    }

    /* ── Mates pane ───────────────────────────────────────────── */
    .mates-pane { flex: 1; overflow-y: auto; padding: 1.75rem 1.5rem 3rem; }
    .mates-toolbar { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem; }
    .mates-list { display: flex; flex-direction: column; gap: 0.45rem; }
    .mates-empty { font-size: var(--text-sm); color: var(--muted); padding: 1.5rem 0; }
    .mate-card {
      border: 1px solid var(--card-bd); border-radius: var(--radius-lg);
      background: var(--bg); padding: 1rem 1.25rem;
      display: flex; flex-direction: column; gap: 0.625rem;
      transition: border-color 120ms, background 120ms;
    }
    .mate-card:hover { background: var(--surface-soft); border-color: rgba(255,255,255,0.11); }
    html[data-theme="light"] .mate-card:hover { background: var(--surface-soft); border-color: rgba(0,0,0,0.12); }
    .mate-card.disabled-card { opacity: 0.45; }
    /* row 1: dot · name · desc · actions */
    .mate-card-top { display: flex; align-items: center; gap: 0.65rem; min-width: 0; }
    .mate-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
    .mate-dot.on  { background: var(--s-ok); }
    .mate-dot.off { background: var(--muted); opacity: 0.55; }
    .mate-card-name { font-size: var(--text-sm); font-weight: 600; color: var(--fg); flex-shrink: 0; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mate-card-desc { font-size: var(--text-xs); color: var(--muted); flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mate-card-actions { display: flex; gap: 0.4rem; flex-shrink: 0; }
    /* row 2: model info + trigger badges */
    .mate-card-meta { display: flex; align-items: center; gap: 0.4rem; padding-left: 1.1rem; min-width: 0; flex-wrap: wrap; }
    .mate-badge {
      font-size: var(--text-xs); font-weight: 500; padding: 2px 8px; border-radius: var(--radius-pill);
      background: var(--code-bg); color: var(--muted);
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 200px;
    }
    .mate-badge.trigger { color: var(--p3); background: var(--s-ok-bg); }
    .mate-act-btn {
      height: 26px; padding: 0 0.6rem; border-radius: var(--radius-sm);
      border: 1px solid var(--code-bd); background: transparent;
      color: var(--muted); font-size: var(--text-xs); font-weight: 500;
      cursor: pointer; transition: color 100ms, background 100ms;
    }
    .mate-act-btn:hover { color: var(--fg); background: var(--card-hov); }
    .mate-act-btn.del:hover { color: var(--s-err); border-color: var(--s-err-bd); }
    .mate-form-wrap { max-width: 780px; }
    .mate-form-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 2rem; }
    .mate-back-btn {
      display: inline-flex; align-items: center; gap: 0.3rem;
      background: none; border: none; color: var(--muted); font-size: var(--text-sm);
      cursor: pointer; padding: 0; transition: color 100ms;
    }
    .mate-back-btn:hover { color: var(--fg); }
    .mate-back-btn svg { width: 14px; height: 14px; }
    .mate-form-title { font-size: var(--text-base); font-weight: 600; color: var(--fg); }
    .mate-form { display: flex; flex-direction: column; }
    .mate-form-section {
      padding: 1.65rem 0 1.5rem; border-top: 1px solid var(--hr);
      display: flex; flex-direction: column; gap: 1.25rem;
    }
    .mate-form-section:first-child { padding-top: 0; border-top: none; }
    .mate-form-section-triggers { gap: 1rem; }
    .mate-form-section-title {
      font-size: var(--text-2xs); font-weight: 600; letter-spacing: 0.07em;
      text-transform: uppercase; color: var(--muted); margin: 0;
    }
    .mate-trigger-section-top { display: flex; flex-direction: column; gap: 0.4rem; }
    .mate-section-desc { font-size: var(--text-sm); color: var(--muted); line-height: 1.5; margin: 0; }
    .mate-form-row { display: flex; gap: 1rem; }
    .mate-form-row > * { flex: 1; min-width: 0; }
    .mate-form-label { display: block; font-size: var(--text-sm); font-weight: 600; color: var(--fg); margin-bottom: 0.4rem; }
    .mate-form-input, .mate-form-select, .mate-form-textarea {
      width: 100%; background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-md); padding: 0.6rem 0.75rem;
      font-size: var(--text-sm); color: var(--fg); outline: none;
      font-family: inherit; transition: border-color 150ms;
    }
    .mate-form-input:focus, .mate-form-select:focus, .mate-form-textarea:focus { border-color: var(--fg); }
    .mate-form-textarea { resize: vertical; min-height: 80px; line-height: 1.55; }
    .mate-trigger-section-hdr {
      display: flex; align-items: center; justify-content: space-between; gap: 0.75rem;
    }
    .mate-trigger-add {
      font-size: var(--text-sm); font-weight: 600; color: var(--muted);
      background: none; border: none; cursor: pointer; padding: 0.15rem 0;
      transition: color 80ms; display: block;
    }
    .mate-trigger-add:hover { color: var(--fg); }
    .mate-triggers-empty {
      font-size: var(--text-sm); color: var(--muted); line-height: 1.5;
      padding: 1.1rem 1rem; border: 1px dashed var(--code-bd); border-radius: var(--radius-md);
      text-align: center; font-style: italic; opacity: 0.75;
    }
    .mate-var-panel { margin-bottom: 0.5rem; }
    .mate-var-panel-label {
      display: block; font-size: var(--text-xs); font-weight: 600;
      letter-spacing: 0.04em; text-transform: uppercase; color: var(--muted); margin-bottom: 0.4rem;
    }
    .mate-var-chips { display: flex; flex-wrap: wrap; gap: 0.35rem; }
    .mate-var-chip {
      display: inline-flex; align-items: center; height: 26px; padding: 0 0.6rem;
      border-radius: var(--radius-sm); border: 1px solid var(--card-bd); background: var(--bg);
      cursor: pointer; transition: border-color 100ms, background 100ms;
    }
    .mate-var-chip:hover { border-color: var(--muted); background: var(--surface-soft); }
    .mate-var-chip code {
      font-family: var(--font-mono);
      font-size: var(--text-xs); font-weight: 600; color: var(--fg); opacity: 0.8;
    }
    .mate-trigger-list { display: flex; flex-direction: column; gap: 1rem; }
    .mate-trigger-card {
      background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-lg); padding: 1rem 1.15rem 1.15rem;
      display: flex; flex-direction: column; gap: 1rem;
    }
    .mate-trigger-hdr {
      display: flex; align-items: center; justify-content: space-between;
      padding-bottom: 0.75rem; border-bottom: 1px solid var(--hr);
    }
    .mate-trigger-label { font-size: var(--text-sm); font-weight: 600; color: var(--fg); }
    .mate-trigger-hdr-actions { display: flex; align-items: center; gap: 0.65rem; }
    .mate-trigger-del {
      background: none; border: none; color: var(--s-err); font-size: var(--text-xs);
      font-weight: 500; cursor: pointer; padding: 0; transition: opacity 80ms;
      opacity: 0.55;
    }
    .mate-trigger-del:hover { opacity: 1; }
    .mate-trigger-body { display: flex; flex-direction: column; gap: 1.4rem; }
    .mate-block-label { margin-bottom: 0.65rem; }
    .mate-block-title { display: block; font-size: var(--text-sm); font-weight: 600; color: var(--fg); }
    .mate-block-hint { display: block; font-size: var(--text-xs); color: var(--muted); line-height: 1.5; margin-top: 0.2rem; }
    .mate-prompt-textarea { min-height: 120px; font-family: var(--font-mono); font-size: var(--text-sm); }
    .mate-schedule-presets { display: flex; flex-wrap: wrap; gap: 0.4rem; margin-bottom: 0.65rem; }
    .mate-schedule-preset {
      height: 28px; padding: 0 0.65rem; border-radius: var(--radius-sm);
      border: 1px solid var(--card-bd); background: transparent;
      color: var(--muted); font-size: var(--text-xs); font-weight: 500;
      cursor: pointer; transition: color 100ms, border-color 100ms, background 100ms;
    }
    .mate-schedule-preset:hover { color: var(--fg); border-color: var(--p3); background: var(--surface-soft); }
    .mate-schedule-preset.active { color: var(--p3); border-color: var(--s-ok-bd); background: var(--s-ok-bg); }
    .mate-schedule-custom-label { display: block; font-size: var(--text-xs); color: var(--muted); margin-bottom: 0.3rem; }
    .mate-color-palette { display: flex; gap: 0.4rem; flex-wrap: wrap; margin-top: 0.35rem; }
    .mate-color-swatch {
      width: 26px; height: 26px; border-radius: 50%; cursor: pointer; flex-shrink: 0;
      border: none;
      box-shadow: 0 0 0 2px transparent, 0 0 0 4px transparent;
      transition: transform 100ms, box-shadow 100ms;
    }
    .mate-color-swatch:hover { transform: scale(1.12); }
    .mate-color-swatch.active { box-shadow: 0 0 0 2px var(--bg), 0 0 0 4px var(--fg); }
    .mate-form-footer {
      display: flex; align-items: center; gap: 0.5rem;
      margin-top: 0.25rem; padding-top: 1.1rem; border-top: 1px solid var(--hr);
    }
    .mate-save-btn {
      height: 32px; padding: 0 1rem; border-radius: var(--radius-md);
      border: 1px solid var(--card-bd); background: transparent; color: var(--fg);
      font-size: var(--text-sm); font-weight: 600; cursor: pointer;
      transition: background 120ms, border-color 120ms;
    }
    .mate-save-btn:hover:not(:disabled) { background: var(--card-hov); border-color: var(--muted); }
    .mate-save-btn:disabled { opacity: 0.4; cursor: not-allowed; }
    .mate-cancel-btn {
      height: 32px; padding: 0 0.875rem; border-radius: var(--radius-md);
      border: 1px solid var(--card-bd); background: transparent;
      color: var(--muted); font-size: var(--text-sm); cursor: pointer;
      transition: color 100ms, background 100ms, border-color 100ms;
    }
    .mate-cancel-btn:hover { color: var(--fg); background: var(--card-hov); border-color: var(--muted); }
    .mate-form-err { flex: 1; font-size: var(--text-xs); color: var(--s-err); }

    /* ── Custom select ─────────────────────────────────────────── */
    .cselect { position: relative; width: 100%; }
    .cselect-btn {
      width: 100%; display: flex; align-items: center; justify-content: space-between;
      gap: 0.4rem; padding: 0.6rem 0.75rem;
      background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-md); cursor: pointer; text-align: left;
      font-size: var(--text-sm); font-family: inherit; color: var(--fg);
      transition: border-color 150ms;
    }
    .cselect-btn:focus { outline: none; border-color: var(--fg); }
    .cselect-btn.open { border-color: var(--fg); }
    .cselect-btn-text { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .cselect-btn svg { width: 11px; height: 11px; flex-shrink: 0; color: var(--muted); transition: transform 150ms; }
    .cselect-btn.open svg { transform: rotate(180deg); }
    .cselect-dropdown {
      position: absolute; top: calc(100% + 3px); left: 0; right: 0; z-index: 120;
      background: var(--bg); border: 1px solid var(--card-bd);
      border-radius: var(--radius-md); overflow: hidden;
      box-shadow: 0 4px 16px rgba(0,0,0,0.18);
      max-height: 220px; overflow-y: auto;
    }
    .cselect-option {
      width: 100%; display: flex; align-items: center; gap: 0.45rem;
      padding: 0.42rem 0.65rem; background: transparent; border: none; cursor: pointer;
      text-align: left; font-size: var(--text-sm); font-family: inherit;
      color: var(--fg); transition: background 80ms;
    }
    .cselect-option:hover { background: var(--card-hov); }
    .cselect-option.sel { color: var(--fg); }
    .cselect-option-dot {
      width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0;
      background: var(--fg); opacity: 0;
    }
    .cselect-option.sel .cselect-option-dot { opacity: 1; }

    /* ── Skills pane ─────────────────────────────────────────── */
    .skills-pane { flex: 1; overflow-y: auto; padding: 1.75rem 1.5rem 3rem; }
    .skills-desc {
      font-size: var(--text-sm); color: var(--muted); line-height: 1.5; margin: 0 0 1.25rem;
    }
    .skills-desc code {
      font-family: var(--font-mono); font-size: var(--text-xs);
      padding: 0.05rem 0.35rem; border-radius: var(--radius-xs);
      background: var(--code-bg); border: 1px solid var(--code-bd);
    }
    .skills-list { display: flex; flex-direction: column; gap: 0.45rem; }
    .skills-empty { font-size: var(--text-sm); color: var(--muted); padding: 1.5rem 0; }
    .skill-card {
      border: 1px solid var(--card-bd); border-radius: var(--radius-lg);
      background: var(--bg); padding: 0.75rem 1.25rem;
      display: flex; align-items: center; justify-content: space-between; gap: 1rem;
      transition: border-color 120ms, background 120ms;
    }
    .skill-card:hover { background: var(--surface-soft); }
    html[data-theme="light"] .skill-card:hover { background: var(--surface-soft); border-color: rgba(0,0,0,0.12); }
    .skill-card-left { display: flex; align-items: center; gap: 0.6rem; flex: 1; min-width: 0; }
    .skill-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
    .skill-dot.on  { background: var(--s-ok); }
    .skill-dot.off { background: var(--muted); opacity: 0.55; }
    .skill-name { font-size: var(--text-sm); font-weight: 600; color: var(--fg); }
    .skill-default-badge {
      font-size: var(--text-xs); font-weight: 500; padding: 2px 8px;
      border-radius: var(--radius-pill); background: var(--code-bg); color: var(--muted);
    }
    .skill-toggling { opacity: 0.55; pointer-events: none; }
    .skill-repo-link {
      font-size: var(--text-xs); color: var(--muted); font-family: var(--font-mono);
      text-decoration: none; opacity: 0.65;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 260px;
      transition: color 100ms, opacity 100ms;
    }
    .skill-repo-link:hover { color: var(--link); opacity: 1; }`

// settingsModalHTML returns the settings modal DOM. Include once per page that has navHTML.
func settingsModalHTML() string {
	return `
  <div id="vaultr-settings-modal"
       x-data="settingsCtrl()"
       x-show="$store.settingsModal.open"
       x-cloak
       class="settings-modal-overlay">
    <div class="settings-modal-panel" @mousedown.stop>
      <div class="settings-modal-bar">
        <span class="settings-modal-title">Settings</span>
        <button class="settings-modal-close-btn" @click="$store.settingsModal.open = false" type="button">
          <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6"/>
            <path stroke-linecap="round" stroke-linejoin="round" d="m6 6 12 12"/>
          </svg>
        </button>
      </div>
      <div class="settings-modal-inner">

        <!-- Primary sidebar -->
        <nav class="settings-sidebar">
          <button class="settings-sidebar-item"
                  :class="{active: tab === 'appearance'}"
                  @click="tab = 'appearance'">
            <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"/>
            </svg>
            Appearance
          </button>
          <button class="settings-sidebar-item"
                  :class="{active: tab === 'effect'}"
                  @click="tab = 'effect'">
            <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 21h8"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z"/>
            </svg>
            Editor
          </button>
          <button class="settings-sidebar-item"
                  :class="{active: tab === 'server'}"
                  @click="tab = 'server'">
            <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <rect width="20" height="8" x="2" y="2" rx="2"/>
              <rect width="20" height="8" x="2" y="14" rx="2"/>
              <path stroke-linecap="round" d="M6 6h.01"/>
              <path stroke-linecap="round" d="M6 18h.01"/>
            </svg>
            Server
          </button>
          <button class="settings-sidebar-item"
                  :class="{active: tab === 'mates'}"
                  @click="tab = 'mates'">
            <svg fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
              <path d="M12 6V2H8"/>
              <path d="M15 11v2"/>
              <path d="M2 12h2"/>
              <path d="M20 12h2"/>
              <path d="M20 16a2 2 0 0 1-2 2H8.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 4 20.286V8a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2z"/>
              <path d="M9 11v2"/>
            </svg>
            Mate Bots
          </button>
          <button class="settings-sidebar-item"
                  :class="{active: tab === 'skills'}"
                  @click="tab = 'skills'">
            <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M11.017 2.814a1 1 0 0 1 1.966 0l1.051 5.558a2 2 0 0 0 1.594 1.594l5.558 1.051a1 1 0 0 1 0 1.966l-5.558 1.051a2 2 0 0 0-1.594 1.594l-1.051 5.558a1 1 0 0 1-1.966 0l-1.051-5.558a2 2 0 0 0-1.594-1.594l-5.558-1.051a1 1 0 0 1 0-1.966l5.558-1.051a2 2 0 0 0 1.594-1.594z"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M20 2v4"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M22 4h-4"/>
              <circle cx="4" cy="20" r="2"/>
            </svg>
            Skills
          </button>
          <button class="settings-sidebar-item"
                  :class="{active: tab === 'agents'}"
                  @click="tab = 'agents'">
            <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 20v2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 2v2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 20v2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 2v2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M2 12h2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M2 17h2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M2 7h2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M20 12h2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M20 17h2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M20 7h2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 20v2"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 2v2"/>
              <rect x="4" y="4" width="16" height="16" rx="2"/>
              <rect x="8" y="8" width="8" height="8" rx="1"/>
            </svg>
            Agents
          </button>
          <button class="settings-sidebar-item"
                  :class="{active: tab === 'shortcuts'}"
                  @click="tab = 'shortcuts'">
            <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <path stroke-linecap="round" d="M10 8h.01"/>
              <path stroke-linecap="round" d="M12 12h.01"/>
              <path stroke-linecap="round" d="M14 8h.01"/>
              <path stroke-linecap="round" d="M16 12h.01"/>
              <path stroke-linecap="round" d="M18 8h.01"/>
              <path stroke-linecap="round" d="M6 8h.01"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 16h10"/>
              <path stroke-linecap="round" d="M8 12h.01"/>
              <rect width="20" height="16" x="2" y="4" rx="2"/>
            </svg>
            Shortcuts
          </button>
        </nav>

        <div class="settings-content">

          <!-- Appearance tab -->
          <div class="settings-pane" x-show="tab === 'appearance'">
            <div class="settings-fields">
              <div>
                <label class="settings-field-label">Theme</label>
                <div class="theme-seg">
                  <button class="theme-seg-btn" :class="{active: themePref==='auto'}" @click="setTheme('auto')">
                    <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                      <path stroke-linecap="round" d="M12 2v2"/>
                      <path stroke-linecap="round" stroke-linejoin="round" d="M14.837 16.385a6 6 0 1 1-7.223-7.222c.624-.147.97.66.715 1.248a4 4 0 0 0 5.26 5.259c.589-.255 1.396.09 1.248.715"/>
                      <path stroke-linecap="round" stroke-linejoin="round" d="M16 12a4 4 0 0 0-4-4"/>
                      <path stroke-linecap="round" d="m19 5-1.256 1.256"/>
                      <path stroke-linecap="round" d="M20 12h2"/>
                    </svg>
                    Auto
                  </button>
                  <button class="theme-seg-btn" :class="{active: themePref==='light'}" @click="setTheme('light')">
                    <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                      <circle cx="12" cy="12" r="4"/>
                      <path stroke-linecap="round" d="M12 2v2M12 20v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M2 12h2M20 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
                    </svg>
                    Light
                  </button>
                  <button class="theme-seg-btn" :class="{active: themePref==='dark'}" @click="setTheme('dark')">
                    <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"/>
                    </svg>
                    Dark
                  </button>
                </div>
                <p class="settings-field-desc">Choose a color theme, or follow your system setting automatically.</p>
              </div>
              <div>
                <label class="settings-field-label">Pixel Mode</label>
                <label class="cfg-toggle">
                  <input type="checkbox" :checked="pixelEnabled" @change="togglePixel()">
                  <span class="cfg-toggle-pill"></span>
                </label>
                <p class="settings-field-desc">Apply a retro pixel-art aesthetic to the UI chrome. Note content is unaffected.</p>
              </div>
            </div>
          </div>

          <!-- Editor tab -->
          <div class="settings-pane" x-show="tab === 'effect'">
            <div class="settings-fields">
              <div>
                <label class="settings-field-label">Enter Effect</label>
                <div class="theme-seg">
                  <button class="theme-seg-btn" :class="{active: effectPref==='none'}" @click="setEffect('none')">None</button>
                  <button class="theme-seg-btn" :class="{active: effectPref==='particles'}" @click="setEffect('particles')">Particles</button>
                </div>
                <p class="settings-field-desc">Visual effect when pressing Enter in the editor. Takes effect immediately.</p>
              </div>
            </div>
          </div>

          <!-- Server tab -->
          <div class="cfg-content" x-show="tab === 'server'">
            <div class="cfg-pane-area">
              <div class="cfg-pane">

                <div x-show="isElectron" style="max-width:640px; margin-bottom:1.75rem;">
                  <label class="settings-field-label">Connection</label>
                  <div class="settings-field-row">
                    <input type="url" class="settings-input"
                           x-model="serverUrl"
                           @keydown.enter="applyServerUrl()"
                           placeholder="http://localhost:54321"
                           spellcheck="false">
                    <button class="settings-apply-btn"
                            @click="applyServerUrl()"
                            :disabled="urlSaving"
                            x-text="urlSaving ? 'Applying…' : 'Apply'"></button>
                  </div>
                  <div class="settings-error" x-show="urlError" x-text="urlError"></div>
                  <p class="settings-field-desc">Vaultr server address used by the desktop app. Changes reload immediately.</p>
                </div>

                <div x-show="isElectron && serverManaged" style="max-width:640px; margin-bottom:1.75rem;">
                  <label class="settings-field-label">Process</label>
                  <div>
                    <button class="settings-apply-btn settings-apply-btn--danger"
                            @click="stopServer()"
                            :disabled="!serverRunning || serverStopping"
                            x-text="serverStopping ? 'Stopping…' : 'Stop Server'"></button>
                  </div>
                  <div class="settings-error" x-show="serverStopError" x-text="serverStopError"></div>
                  <p class="settings-field-desc">Server process started and managed by this desktop app.</p>
                </div>

                <label class="settings-field-label" x-show="!cfgLoading && !cfgError" style="margin-bottom:0.75rem;">Config</label>
                <div x-show="cfgLoading" class="cfg-loader">Loading configuration…</div>
                <div x-show="cfgError && !cfgLoading" class="cfg-err-msg" x-text="'Error: ' + cfgError"></div>

                <template x-for="section in sectionTabs" :key="section">
                  <section class="cfg-section" :class="{ 'is-open': openSection === section }" x-show="!cfgLoading && !cfgError">
                    <button type="button" class="cfg-section-head" @click="toggleSection(section)">
                      <div class="cfg-section-head-text">
                        <span class="cfg-section-title" x-text="sectionLabel(section)"></span>
                        <span class="cfg-section-head-desc"
                              x-show="sectionIntro(section)"
                              x-text="sectionIntro(section)"></span>
                      </div>
                      <div class="cfg-section-head-meta">
                        <span class="cfg-section-dirty-dot" x-show="sectionHasDirty(section)" title="Unsaved changes in this section"></span>
                        <span class="cfg-section-chev" :class="{open: openSection === section}">
                          <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7"/>
                          </svg>
                        </span>
                      </div>
                    </button>
                    <div class="cfg-section-body" x-show="openSection === section">
                      <div class="cfg-fields">
                        <template x-for="field in fieldsForSection(section)" :key="field.key">
                          <div class="cfg-field" :class="{dirty: isDirty(field.key), multiline: field.multiline}">
                            <div class="cfg-field-meta">
                              <span class="cfg-field-label" x-text="field.label"></span>
                              <code class="cfg-field-key" x-text="field.key"></code>
                            </div>
                            <div class="cfg-field-ctrl">
                              <template x-if="field.type === 'bool'">
                                <label class="cfg-toggle">
                                  <input type="checkbox"
                                         :checked="!!getVal(field.key)"
                                         @change="setVal(field.key, $event.target.checked)">
                                  <span class="cfg-toggle-pill"></span>
                                </label>
                              </template>
                              <template x-if="field.type === 'string' && field.enum && field.enum.length">
                                <select class="cfg-input" @change="setVal(field.key, $event.target.value)">
                                  <template x-for="opt in (field.enum || [])" :key="opt">
                                    <option :value="opt" :selected="getVal(field.key) === opt" x-text="opt"></option>
                                  </template>
                                </select>
                              </template>
                              <template x-if="field.type === 'int'">
                                <input type="number" class="cfg-input"
                                       :value="getVal(field.key)"
                                       @change="setVal(field.key, Number($event.target.value))"
                                       :min="field.constraints ? field.constraints.min : undefined"
                                       :max="field.constraints ? field.constraints.max : undefined">
                              </template>
                              <template x-if="field.type === 'string_list'">
                                <textarea class="cfg-textarea" rows="3" placeholder="One entry per line"
                                          :value="listToText(getVal(field.key))"
                                          @change="setVal(field.key, textToList($event.target.value))"></textarea>
                              </template>
                              <template x-if="field.sensitive && field.type !== 'bool' && field.type !== 'string_list'">
                                <div class="cfg-reveal-wrap">
                                  <input class="cfg-input"
                                         :type="revealed[field.key] ? 'text' : 'password'"
                                         :value="getVal(field.key) || ''"
                                         :placeholder="(secrets[field.key] && !getVal(field.key)) ? '••• set •••' : (field.default != null ? String(field.default) : '')"
                                         @input="setVal(field.key, $event.target.value)">
                                  <button type="button" class="cfg-reveal-btn"
                                          @click.stop="toggleReveal(field.key)"
                                          :title="revealed[field.key] ? 'Hide' : 'Reveal'">
                                    <svg x-show="!revealed[field.key]" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                                      <path stroke-linecap="round" stroke-linejoin="round" d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0"/><circle cx="12" cy="12" r="3"/>
                                    </svg>
                                    <svg x-show="revealed[field.key]" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                                      <path stroke-linecap="round" stroke-linejoin="round" d="M10.733 5.076a10.744 10.744 0 0 1 11.205 6.575 1 1 0 0 1 0 .696 10.747 10.747 0 0 1-1.444 2.49"/>
                                      <path stroke-linecap="round" stroke-linejoin="round" d="M14.084 14.158a3 3 0 0 1-4.242-4.242"/>
                                      <path stroke-linecap="round" stroke-linejoin="round" d="M17.479 17.499a10.75 10.75 0 0 1-15.417-5.151 1 1 0 0 1 0-.696 10.75 10.75 0 0 1 4.446-5.143"/>
                                      <path stroke-linecap="round" stroke-linejoin="round" d="m2 2 20 20"/>
                                    </svg>
                                  </button>
                                </div>
                              </template>
                              <template x-if="!field.sensitive && field.type === 'string' && field.multiline && !(field.enum && field.enum.length)">
                                <textarea class="cfg-textarea" rows="6"
                                          :value="getVal(field.key) ?? ''"
                                          :placeholder="field.default != null ? String(field.default) : ''"
                                          @input="setVal(field.key, $event.target.value)"></textarea>
                              </template>
                              <template x-if="!field.sensitive && (field.type === 'string' || field.type === 'duration') && !field.multiline && !(field.enum && field.enum.length)">
                                <div style="display:flex;gap:0.375rem;align-items:center;width:100%;">
                                  <input type="text" class="cfg-input" style="flex:1;min-width:0;"
                                         :value="getVal(field.key) ?? ''"
                                         :placeholder="field.default != null ? String(field.default) : ''"
                                         @input="setVal(field.key, $event.target.value)">
                                  <template x-if="field.key === 'vault.path' && isElectron">
                                    <button type="button" class="cfg-reveal-btn"
                                            title="Browse for folder"
                                            style="font-size:13px;letter-spacing:0.05em;padding:0 6px;width:auto;"
                                            @click="pickFolder(field.key, getVal(field.key))">···</button>
                                  </template>
                                </div>
                              </template>
                            </div>
                            <p class="cfg-field-desc" x-text="field.description"></p>
                          </div>
                        </template>
                      </div>

                      <div class="cfg-wechat-auth" x-show="section === 'plugins.wechat'">
                        <div class="cfg-wechat-auth-head">
                          <span class="cfg-wechat-auth-title">WeChat login</span>
                          <span class="cfg-wechat-badge"
                                :class="{ connected: wechatStatus.connected }"
                                x-text="wechatStatus.connected ? 'Connected' : 'Not connected'"></span>
                        </div>
                        <p class="cfg-wechat-meta">
                          Scan with WeChat to obtain an iLink bot token. Credentials are saved to
                          <code>config.toml</code>. Create a mate with a <code>wechat_message</code> trigger to handle replies.
                        </p>
                        <template x-if="wechatStatus.connected">
                          <div>
                            <p class="cfg-wechat-meta" x-show="wechatStatus.account_id">
                              Bot ID: <code x-text="wechatStatus.account_id"></code>
                              <span x-show="wechatStatus.saved_at"> · saved <span x-text="wechatStatus.saved_at"></span></span>
                            </p>
                            <div class="cfg-wechat-actions">
                              <button type="button" class="settings-apply-btn settings-apply-btn--danger"
                                      @click="wechatLogout()"
                                      :disabled="wechatAuthBusy"
                                      x-text="wechatAuthBusy ? 'Disconnecting…' : 'Disconnect'"></button>
                            </div>
                          </div>
                        </template>
                        <template x-if="!wechatStatus.connected">
                          <div>
                            <div class="cfg-wechat-actions">
                              <button type="button" class="settings-apply-btn"
                                      @click="startWechatLogin()"
                                      :disabled="wechatAuthBusy || wechatLoginBusy"
                                      x-text="wechatLoginBusy ? 'Waiting for scan…' : 'Scan QR to log in'"></button>
                              <button type="button" class="settings-apply-btn"
                                      x-show="wechatLoginBusy"
                                      @click="cancelWechatLogin()">Cancel</button>
                            </div>
                            <div class="cfg-wechat-qr" x-show="wechatQrcodeImg">
                              <img :src="wechatQrcodeImg" alt="WeChat login QR code">
                              <p class="cfg-wechat-status" x-show="wechatLoginStatus" x-text="wechatLoginStatus"></p>
                            </div>
                          </div>
                        </template>
                        <p class="cfg-wechat-err" x-show="wechatAuthError" x-text="wechatAuthError"></p>
                        <p class="cfg-wechat-ok" x-show="wechatLoginOk">Connected — restart the server to start the bridge.</p>
                      </div>
                    </div>
                  </section>
                </template>

                <div class="cfg-action-bar" x-show="!cfgLoading && !cfgError">
                  <div class="cfg-action-left">
                    <span class="cfg-status-ok" x-show="cfgSaveOk && !hasDirty">Saved — restart server to apply</span>
                    <span class="cfg-status-err" x-show="cfgSaveError" x-text="cfgSaveError"></span>
                    <span class="cfg-status-err" x-show="cfgRestartError" x-text="cfgRestartError"></span>
                    <span class="cfg-restart-note" x-show="cfgRestarting">Restarting server…</span>
                    <span class="cfg-dirty-badge" x-show="hasDirty"
                          x-text="Object.keys(patch).length + ' unsaved change' + (Object.keys(patch).length !== 1 ? 's' : '')"></span>
                    <span class="cfg-restart-note"
                          x-show="!hasDirty && !cfgSaveOk && !cfgSaveError && !cfgRestartError && !cfgRestarting">
                      Expand sections above to edit. Save once when done; restart server to apply.
                    </span>
                  </div>
                  <div class="cfg-action-right">
                    <button class="cfg-discard-btn" x-show="hasDirty" @click="discardAll()">Discard</button>
                    <button class="cfg-save-btn"
                            @click="saveConfig()"
                            :disabled="!hasDirty || cfgSaving || cfgRestarting"
                            x-text="cfgRestarting ? 'Restarting…' : cfgSaving ? 'Saving…' : 'Save all'"></button>
                  </div>
                </div>
              </div>
            </div>
          </div><!-- .cfg-content server -->

          <!-- Shortcuts tab -->
          <div class="shortcuts-pane" x-show="tab === 'shortcuts'">
            <div class="shortcuts-fields">
              <div class="shortcuts-list">
                <template x-for="s in shortcutDefs" :key="s.id">
                  <div class="shortcuts-row">
                    <div class="shortcuts-row-meta">
                      <span class="shortcuts-label" x-text="s.label"></span>
                      <span class="shortcuts-desc" x-text="s.desc"></span>
                    </div>
                    <div class="shortcuts-keys">
                      <template x-for="k in getEffectiveKeys(s)" :key="k">
                        <span class="kbd" x-text="k"></span>
                      </template>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </div>

          <!-- Mates tab -->
          <div class="mates-pane" x-show="tab === 'mates'">

            <template x-if="!mateFormMode && !matesSubPage">
              <div>
                <div class="mates-toolbar">
                  <button class="agents-toolbar-btn" @click="newMate()">
                    <svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14"/>
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 5v14"/>
                    </svg>
                    New Mate
                  </button>

` + toolbarRefreshBtnHTML("matesLoading", "loadMates()", "Loading…") + `
                </div>
                <div class="cfg-err-msg" x-show="matesError && !matesLoading" x-text="'Error: ' + matesError"></div>
                <div class="mates-list">
                  <template x-if="!matesLoading && matesList.length === 0">
                    <div class="mates-empty">No mates yet — click New Mate to create one.</div>
                  </template>
                  <template x-for="(m, mi) in matesList" :key="m.id">
                    <div class="mate-card" :class="m.enabled ? '' : 'disabled-card'">
                      <div class="mate-card-top">
                        <span class="mate-dot" :class="m.enabled ? 'on' : 'off'"
                              :style="m.enabled && m.color ? 'background:' + m.color : ''"></span>
                        <span class="mate-card-name" x-text="m.name"></span>
                        <span class="mate-card-desc" x-show="m.description" x-text="m.description"></span>
                        <div class="mate-card-actions">
                          <button class="mate-act-btn" @click.stop="moveMate(mi, -1)" :disabled="mi === 0" type="button" title="Move up">↑</button>
                          <button class="mate-act-btn" @click.stop="moveMate(mi, 1)" :disabled="mi === matesList.length - 1" type="button" title="Move down">↓</button>
                          <button class="mate-act-btn" @click.stop="openMateEdit(m.id)" type="button">Edit</button>
                          <button class="mate-act-btn del" @click.stop="deleteMate(m.id)" type="button">Delete</button>
                        </div>
                      </div>
                      <div class="mate-card-meta">
                        <span class="mate-badge" x-text="m.agentId || '—'"></span>
                        <template x-if="m.model">
                          <span class="mate-badge" x-text="m.model"></span>
                        </template>
                        <template x-if="m.triggerCount > 0">
                          <span class="mate-badge trigger"
                                x-text="m.triggerCount + (m.triggerCount === 1 ? ' trigger' : ' triggers')"></span>
                        </template>
                      </div>
                    </div>
                  </template>
                </div>
              </div>
            </template>


            <template x-if="mateFormMode">
              <div class="mate-form-wrap">
                <div class="mate-form-header">
` + mateBackBtnHTML("mateFormMode = null", "Mate Bots") + `
                  <span class="mate-form-title" x-text="mateFormMode === 'create' ? 'New Mate' : 'Edit Mate'"></span>
                </div>
                <div class="mate-form">
                  <section class="mate-form-section">
                    <h3 class="mate-form-section-title">Profile</h3>
                    <div class="mate-form-row">
                      <div>
                        <label class="mate-form-label">Name</label>
                        <input class="mate-form-input" type="text" x-model="mateDraft.name" placeholder="Quote Extractor" autofocus>
                      </div>
                      <div style="flex:0 0 auto;min-width:90px">
                        <label class="mate-form-label">Enabled</label>
                        <label class="cfg-toggle">
                          <input type="checkbox" :checked="mateDraft.enabled" @change="mateDraft.enabled = $event.target.checked">
                          <span class="cfg-toggle-pill"></span>
                        </label>
                      </div>
                    </div>
                    <div>
                      <label class="mate-form-label">Color</label>
                      <div class="mate-color-palette">
                        <template x-for="c in mateColors" :key="c">
                          <button type="button" class="mate-color-swatch"
                                  :class="mateDraft.color === c ? 'active' : ''"
                                  :style="'background:' + c"
                                  @click="mateDraft = Object.assign({}, mateDraft, {color: c})"
                                  :title="c"></button>
                        </template>
                      </div>
                    </div>
                    <div>
                      <label class="mate-form-label">Description <span style="font-weight:400;text-transform:none;letter-spacing:0;">(optional)</span></label>
                      <input class="mate-form-input" type="text" x-model="mateDraft.description" placeholder="Brief description of what this mate does">
                    </div>
                  </section>

                  <section class="mate-form-section">
                    <h3 class="mate-form-section-title">Agent</h3>
                    <div class="mate-form-row">
                      <div>
                        <label class="mate-form-label">Agent</label>
` + cselectHTML(
		`(agents.find(function(a){return a.id===mateDraft.agentId&&a.available;}) || {name: 'Select agent'}).name`,
		`<template x-for="a in agents.filter(function(a){return a.available;})" :key="a.id"><button type="button" class="cselect-option" :class="mateDraft.agentId===a.id?'sel':''" @click="mateDraft.agentId=a.id; onMateAgentChange(); csOpen=false"><span class="cselect-option-dot"></span><span x-text="a.name"></span></button></template>`,
	) + `
                      </div>
                      <div>
                        <label class="mate-form-label">Model</label>
` + cselectHTML(
		`mateDraft.model || 'Default'`,
		`<button type="button" class="cselect-option" :class="mateDraft.model===''?'sel':''" @click="mateDraft.model=''; csOpen=false"><span class="cselect-option-dot"></span><span>Default</span></button>`+
			`<template x-for="m in mateModelsForAgent(mateDraft.agentId)" :key="m.id"><button type="button" class="cselect-option" :class="mateDraft.model===m.id?'sel':''" @click="mateDraft.model=m.id; csOpen=false"><span class="cselect-option-dot"></span><span :title="(m.label && m.label !== m.id) ? m.label : ''" x-text="m.id"></span></button></template>`,
	) + `
                      </div>
                    </div>
                    <div>
                      <label class="mate-form-label">Working Directory <span style="font-weight:400;text-transform:none;letter-spacing:0;">(vault root if blank)</span></label>
                      <input class="mate-form-input" type="text" x-model="mateDraft.cwd" placeholder="/absolute/path or leave blank">
                    </div>
                    <div>
                      <label class="mate-form-label">System Prompt</label>
                      <textarea class="mate-form-textarea" x-model="mateDraft.systemPrompt" rows="5"
                                placeholder="Instructions prepended to every message…"></textarea>
                    </div>
                  </section>

                  <section class="mate-form-section mate-form-section-triggers">
                    <div class="mate-trigger-section-top">
                      <div class="mate-trigger-section-hdr">
                        <h3 class="mate-form-section-title">Triggers</h3>
                      </div>
                      <p class="mate-section-desc">Automatically run this mate on vault events or on a schedule. Each trigger sends a prompt template to the agent.</p>
                    </div>
                    <template x-if="mateTriggers.length === 0">
                      <div class="mate-triggers-empty">No triggers — mate responds only to manual chat.</div>
                    </template>
                    <div class="mate-trigger-list">
                      <template x-for="(t, ti) in mateTriggers" :key="ti">
                        <div class="mate-trigger-card">
                          <div class="mate-trigger-hdr">
                            <span class="mate-trigger-label">Trigger <span x-text="ti+1"></span></span>
                            <div class="mate-trigger-hdr-actions">
                              <button class="mate-trigger-del" type="button" @click="removeMateTrigger(ti)">Remove</button>
                            </div>
                          </div>
                          <div class="mate-trigger-body">
                            <div class="mate-trigger-block mate-trigger-events">
                              <div class="mate-block-label">
                                <span class="mate-block-title">Event</span>
                                <span class="mate-block-hint" x-text="isScheduledTrigger(t) ? 'Scheduled triggers run on a timer — choose one schedule below' : (isWechatTrigger(t) ? 'WeChat DM trigger — use Content and WechatUserID in the prompt' : (isCompileTrigger(t) ? 'Fires when the user manually triggers compilation via the API — Path carries the note' : 'Vault event that activates this trigger'))"></span>
                              </div>
` + cselectHTML(
		`(mateEventDefs.find(function(d){return d.type===(t.eventTypes[0]||'');}) || {label: 'Select event…'}).label`,
		`<template x-for="def in mateEventDefs" :key="def.type"><button type="button" class="cselect-option" :class="(t.eventTypes[0]||'')===def.type ? 'sel' : ''" :title="def.description" @click="setMateET(t, def.type); csOpen=false"><span class="cselect-option-dot"></span><span x-text="def.label"></span></button></template>`,
	) + `
                            </div>
                            <template x-if="isScheduledTrigger(t)">
                              <div class="mate-trigger-block mate-trigger-schedule">
                                <div class="mate-block-label">
                                  <span class="mate-block-title">Schedule</span>
                                  <span class="mate-block-hint">Server local time. Minimum interval: 15 minutes.</span>
                                </div>
                                <div class="mate-schedule-presets">
                                  <template x-for="p in mateSchedulePresets" :key="p.value">
                                    <button type="button" class="mate-schedule-preset"
                                            :class="(t.schedule || '') === p.value ? 'active' : ''"
                                            @click="t.schedule = p.value"
                                            x-text="p.label"></button>
                                  </template>
                                </div>
                                <label class="mate-schedule-custom-label">Custom</label>
                                <input class="mate-form-input" type="text" x-model="t.schedule"
                                       placeholder="every 1h  or  daily 09:00">
                              </div>
                            </template>
                            <template x-if="!isScheduledTrigger(t) && !isWechatTrigger(t)">
                              <div class="mate-trigger-block mate-trigger-paths">
                                <div class="mate-block-label">
                                  <span class="mate-block-title">Path Prefixes <span style="font-weight:400;opacity:0.6;">(optional)</span></span>
                                  <span class="mate-block-hint">Only fire when the event path starts with one of these prefixes. Leave empty to match all paths. One prefix per line, e.g. <code style="font-size:var(--text-2xs);padding:0 3px;background:var(--code-bg);border-radius:2px;">/journal/</code></span>
                                </div>
                                <textarea class="mate-form-textarea"
                                          rows="3"
                                          :value="(t.pathPrefixes || []).join('\n')"
                                          @change="t.pathPrefixes = $event.target.value.split('\n').map(function(s){return s.trim();}).filter(Boolean)"
                                          placeholder="/journal/&#10;/projects/work/"></textarea>
                              </div>
                            </template>
                            <div class="mate-trigger-block mate-trigger-prompt">
                              <div class="mate-block-label">
                                <span class="mate-block-title">Prompt template</span>
                                <span class="mate-block-hint">Message sent to the agent when this trigger fires. Click a variable below to insert at cursor.</span>
                              </div>
                              <div class="mate-var-panel">
                                <span class="mate-var-panel-label">Variables</span>
                                <div class="mate-var-chips">
                                  <template x-for="v in matePromptVarsForTrigger(t)" :key="v.token">
                                    <button type="button" class="mate-var-chip" :title="v.desc"
                                            @click="insertMateVar(ti, v.token, $event)">
                                      <code x-text="v.token"></code>
                                    </button>
                                  </template>
                                </div>
                              </div>
                              <textarea class="mate-form-textarea mate-prompt-textarea" x-model="t.prompt" rows="5"
                                        @focus="mateActivePromptIdx = ti"
                                        :placeholder="matePromptPlaceholder(t)"></textarea>
                            </div>
                          </div>
                        </div>
                      </template>
                    </div>
                    <button class="mate-trigger-add" type="button" @click="addMateTrigger()" style="margin-top:0.25rem;">+ Add trigger</button>
                  </section>

                  <div class="mate-form-footer">
                    <button class="mate-save-btn" type="button"
                            :disabled="!mateDraft.name.trim() || mateSaving"
                            @click="saveMate()"
                            x-text="mateSaving ? 'Saving…' : 'Save'"></button>
                    <button class="mate-cancel-btn" type="button" @click="mateFormMode = null">Cancel</button>
                    <span class="mate-form-err" x-text="mateSaveError"></span>
                  </div>
                </div>
              </div>
            </template>

          </div><!-- .mates-pane -->

          <!-- Skills tab -->
          <div class="skills-pane" x-show="tab === 'skills'">
            <div class="mates-toolbar">
` + toolbarRefreshBtnHTML("skillsLoading", "loadSkills()", "Loading…") + `
            </div>
            <p class="skills-desc">
              Enable or disable skills for agents. Source directory: <code>~/.vaultr/skills/</code>
            </p>
            <div class="cfg-err-msg" x-show="skillsError && !skillsLoading" x-text="'Error: ' + skillsError"></div>
            <div class="cfg-loader" x-show="skillsLoading">Loading skills…</div>
            <div x-show="!skillsLoading">
              <template x-if="skillsList.length === 0">
                <div class="skills-empty">No skills found in ~/.vaultr/skills/</div>
              </template>
              <div class="skills-list">
                <template x-for="s in sortedSkillsList" :key="s.name">
                  <div class="skill-card">
                    <div class="skill-card-left">
                      <span class="skill-dot" :class="s.enabled ? 'on' : 'off'"></span>
                      <span class="skill-name" x-text="s.name"></span>
                      <span class="skill-default-badge" x-show="s.default">default</span>
                      <template x-if="s.repoUrl">
                        <a class="skill-repo-link" :href="s.repoUrl" target="_blank" rel="noopener noreferrer"
                           @click.stop
                           x-text="s.repoUrl.replace('https://github.com/', '')"></a>
                      </template>
                    </div>
                    <template x-if="!s.default">
                      <label class="cfg-toggle" :class="skillsToggling[s.name] ? 'skill-toggling' : ''">
                        <input type="checkbox"
                               :checked="s.enabled"
                               :disabled="!!skillsToggling[s.name]"
                               @change="toggleSkill(s.name, $event.target.checked)">
                        <span class="cfg-toggle-pill"></span>
                      </label>
                    </template>
                  </div>
                </template>
              </div>
            </div>
          </div><!-- .skills-pane -->

          <!-- Agents tab -->
          <div class="agents-pane" x-show="tab === 'agents'">
            <div class="agents-toolbar">
` + toolbarRefreshBtnHTML("agentsLoading", "loadAgents(true)", "Detecting…") + `
              <span class="agents-summary" x-show="!agentsLoading && agents.length">
                <span x-text="agents.filter(a=>a.available).length + ' available · ' + agents.filter(a=>!a.available).length + ' not installed'"></span>
                <span x-show="agentsFromCache && agentsCachedAt"
                      x-text="' · cached ' + Math.round((Date.now() - agentsCachedAt) / 60000) + 'm ago'"></span>
              </span>
            </div>
            <div class="cfg-loader" x-show="agentsLoading">Detecting agents on PATH — this may take a moment…</div>
            <div class="cfg-err-msg" x-show="agentsError && !agentsLoading" x-text="'Error: ' + agentsError"></div>
            <div class="agents-list" x-show="!agentsLoading && !agentsError && agents.length">
              <template x-for="ag in sortedAgents" :key="ag.id">
                <div class="agent-card" :class="{unavailable: !ag.available}">
                  <div class="agent-card-top">
                    <span class="agent-dot" :class="ag.available ? 'ok' : 'off'"></span>
                    <span class="agent-card-name" x-text="ag.name"></span>
                    <span class="agent-card-id" x-text="ag.id"></span>
                    <span class="agent-status-badge" :class="ag.available ? 'ok' : ''"
                          x-text="ag.available ? 'available' : 'not installed'"></span>
                  </div>
                  <div class="agent-card-info">
                    <span class="agent-col-mono agent-meta-path"
                          :title="ag.path || ag.bin"
                          x-text="ag.path || ag.bin || '—'"></span>
                    <span class="agent-col-mono"
                          x-show="ag.version"
                          x-text="'v' + ag.version"></span>
                    <span class="agent-col-mono"
                          x-show="ag.streamFormat"
                          x-text="ag.streamFormat"></span>
                  </div>
                  <div class="agent-card-models">
                    <template x-for="m in agentDisplayModels(ag)" :key="m.id">
                      <span class="agent-model-pill" :title="(m.label && m.label !== m.id) ? m.id + ' — ' + m.label : m.id" x-text="m.id"></span>
                    </template>
                    <span class="agent-model-more"
                          x-show="ag.models && ag.models.length > 4"
                          x-text="'+' + (ag.models.length - 4) + ' more'"></span>
                  </div>
                  <template x-if="ag.cliExample">
                    <div class="agent-card-cli">
                      <span class="agent-cli-label">Example:</span>
                      <span class="agent-cli-code" :title="ag.cliExample" x-text="ag.cliExample"></span>
                      <button type="button" class="agent-cli-copy"
                              :class="agentCopied === ag.id ? 'copied' : ''"
                              @click.stop="copyCliExample(ag)"
                              x-text="agentCopied === ag.id ? 'copied' : 'copy'"></button>
                    </div>
                  </template>
                </div>
              </template>
            </div>
          </div><!-- .agents-pane -->

        </div><!-- .settings-content -->
      </div><!-- .settings-modal-inner -->
    </div><!-- .settings-modal-panel -->
  </div><!-- #vaultr-settings-modal -->`
}

// settingsCtrlJS is the Alpine.js controller for the settings modal.
// It uses lazy init — data is only loaded when the modal is first opened.
const settingsCtrlJS = `
  function settingsCtrl() {
    return {
      _inited: false,
      isElectron: !!window.vaultrDesktop,
      tab: 'appearance',
      serverUrl: '',
      urlSaving: false,
      urlError: '',
      serverManaged: false,
      serverRunning: false,
      serverStopping: false,
      serverStopError: '',

      schema: [],
      values: {},
      secrets: {},
      cfgLoading: false,
      cfgError: '',
      cfgSaving: false,
      cfgSaveError: '',
      cfgSaveOk: false,
      cfgRestarting: false,
      cfgRestartError: '',
      patch: {},
      revealed: {},
      openSection: '',

      wechatStatus: { connected: false, account_id: '', saved_at: '', enabled: false },
      wechatAuthBusy: false,
      wechatAuthError: '',
      wechatLoginBusy: false,
      wechatLoginStatus: '',
      wechatQrcode: '',
      wechatQrcodeImg: '',
      wechatPollTimer: null,
      wechatLoginOk: false,

      agents: [],
      agentsLoading: false,
      agentsError: '',
      agentsLoaded: false,
      agentsFromCache: false,
      agentsCachedAt: 0,
      agentCopied: '',

      matesList: [],
      matesLoading: false,
      matesError: '',
      mateFormMode: null,
      matesSubPage: null,
      mateEditId: '',
      mateDraft: {},
      mateTriggers: [],
      mateActivePromptIdx: -1,
      mateSaving: false,
      mateSaveError: '',
      mateEventDefs: [],
      matePromptVarsVault: [
        { token: '\x7b\x7b.Path\x7d\x7d', desc: 'Full vault path of the affected note' },
        { token: '\x7b\x7b.Name\x7d\x7d', desc: 'Filename without extension' },
        { token: '\x7b\x7b.Content\x7d\x7d', desc: 'Appended short-note text (short_note_created only)' },
      ],
      matePromptVarsCompile: [
        { token: '\x7b\x7b.Path\x7d\x7d', desc: 'Vault path of the note to compile' },
        { token: '\x7b\x7b.Name\x7d\x7d', desc: 'Filename without extension' },
      ],
      matePromptVarsWechat: [
        { token: '\x7b\x7b.Content\x7d\x7d', desc: 'Incoming WeChat DM text' },
        { token: '\x7b\x7b.WechatUserID\x7d\x7d', desc: 'Sender WeChat user ID' },
      ],
      matePromptVarsScheduled: [
        { token: '\x7b\x7b.Now\x7d\x7d', desc: 'Trigger time (RFC3339)' },
        { token: '\x7b\x7b.Date\x7d\x7d', desc: 'Date YYYY-MM-DD' },
        { token: '\x7b\x7b.Time\x7d\x7d', desc: 'Time HH:MM' },
      ],
      mateSchedulePresets: [
        { label: 'Every hour', value: 'every 1h' },
        { label: 'Every 6 hours', value: 'every 6h' },
        { label: 'Daily 09:00', value: 'daily 09:00' },
        { label: 'Daily 21:00', value: 'daily 21:00' },
      ],
      mateColors: ['var(--p0)','var(--p1)','var(--p2)','var(--p3)'],

      skillsList: [],
      skillsLoading: false,
      skillsError: '',
      skillsToggling: {},

      get sortedSkillsList() {
        return [...this.skillsList].sort(function(a, b) {
          const rank = function(s) { return s.default ? 0 : s.enabled ? 1 : 2; };
          return rank(a) - rank(b);
        });
      },

      get themePref() { return Alpine.store('theme').pref; },
      setTheme(v) { Alpine.store('theme').set(v); },
      get pixelEnabled() { return Alpine.store('pixel').enabled; },
      togglePixel() { Alpine.store('pixel').toggle(); },

      effectPref: localStorage.getItem('vaultr-editor-effect') || 'particles',
      setEffect(key) { this.effectPref = key; localStorage.setItem('vaultr-editor-effect', key); },

      isMac: /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent),
      customKeys: JSON.parse(localStorage.getItem('vaultr-custom-keys') || '{}'),
      shortcutDefs: [
        { id: 'dismiss',        label: 'Dismiss',                desc: 'Close any open overlay or dialog',          mac: 'Esc',  win: 'Esc' },
        { id: 'toggle-search',  label: 'Search',                 desc: 'Open the quick search overlay',             mac: '⌘K',   win: 'Ctrl+K' },
        { id: 'new-note',       label: 'New Note',               desc: 'Open a new blank note in the editor',       mac: '⌘N',   win: 'Ctrl+N' },
        { id: 'quick-note',     label: 'Quick Note',             desc: 'Open the quick-capture note dialog',        mac: '⌘.',   win: 'Ctrl+.' },
        { id: 'toggle-editor',  label: 'Toggle Editor',          desc: 'Open or close the editor panel',            mac: '⌘E',   win: 'Ctrl+E' },
        { id: 'close-tab',      label: 'Close Editor Tab',       desc: 'Close the active tab in the editor',        mac: '⌘W',   win: 'Ctrl+W' },
        { id: 'expand-editor',  label: 'Expand / Shrink Editor', desc: 'Toggle editor between 80% and 100% width',  mac: '⌘\\',  win: 'Ctrl+\\' },
        { id: 'nav-home',       label: 'Go to Notes',            desc: 'Navigate to the Notes page',                mac: '⌘1',   win: 'Ctrl+1' },
        { id: 'nav-agent',      label: 'Go to Agent Chat',       desc: 'Navigate to the Agent Chat page',           mac: '⌘2',   win: 'Ctrl+2' },
        { id: 'refresh',        label: 'Refresh',                desc: 'Reload the current page',                   mac: '⌘R',   win: 'Ctrl+R' },
        { id: 'open-settings',  label: 'Settings',               desc: 'Open the settings dialog',                  mac: '⌘,',   win: 'Ctrl+,' },
      ],
      getEffectiveKeys(s) {
        const custom = this.customKeys[s.id];
        const key = this.isMac ? (custom?.mac ?? s.mac) : (custom?.win ?? s.win);
        return key.split('/').map(k => k.trim());
      },

      get sectionTabs() {
        const excluded = new Set(['log', 'server']);
        const seen = new Set(), tabs = [];
        for (const f of this.schema) {
          if (!seen.has(f.section) && !excluded.has(f.section)) {
            seen.add(f.section);
            tabs.push(f.section);
          }
        }
        return tabs;
      },

      sectionLabel(s) {
        const m = {
          'server': 'Server', 'vault': 'Vault', 'agent': 'Agent',
          'plugins.search': 'Search', 'plugins.git_sync': 'Git Sync',
          'plugins.compile': 'Compile',
          'plugins.image_fetch': 'Image Fetch',
          'plugins.wechat': 'WeChat',
        };
        return m[s] || s;
      },

      sectionIntro(s) {
        const m = {
          'vault': 'Paths, layout, and core vault behavior.',
          'agent': 'Agent CLI integration and global prompt settings.',
          'plugins.search': 'Indexing and search quality.',
          'plugins.git_sync': 'Automatic git push/pull for the vault.',
          'plugins.compile': 'AI knowledge compilation and related options.',
          'plugins.image_fetch': 'Fetching and storing remote images.',
          'plugins.wechat': 'WeChat iLink bridge — poll DMs and emit wechat_message mate events.',
          'server.listen': 'HTTP listen address and port.',
          'server': 'HTTP listen address and port.',
        };
        return m[s] || '';
      },

      fieldsForSection(s) { return this.schema.filter(f => f.section === s); },

      toggleSection(s) {
        const opening = this.openSection !== s;
        this.openSection = opening ? s : '';
        if (opening && s === 'plugins.wechat') this.loadWechatStatus();
      },

      sectionHasDirty(s) { return this.fieldsForSection(s).some(f => this.isDirty(f.key)); },

      getVal(key) {
        if (key in this.patch) return this.patch[key];
        const parts = key.split('.');
        let v = this.values;
        for (const p of parts) {
          if (v == null || typeof v !== 'object') return null;
          v = v[p];
        }
        return v ?? null;
      },

      setVal(key, val) {
        this.patch = { ...this.patch, [key]: val };
        this.cfgSaveOk = false;
        this.cfgSaveError = '';
        this.cfgRestartError = '';
      },

      isDirty(key) { return key in this.patch; },
      get hasDirty() { return Object.keys(this.patch).length > 0; },

      discardAll() {
        this.patch = {};
        this.cfgSaveOk = false;
        this.cfgSaveError = '';
        this.cfgRestartError = '';
      },

      toggleReveal(key) { this.revealed = { ...this.revealed, [key]: !this.revealed[key] }; },
      listToText(v) { if (!Array.isArray(v)) return ''; return v.join('\n'); },
      textToList(s) { return s.split('\n').map(x => x.trim()).filter(Boolean); },

      buildNested(flat) {
        const out = {};
        for (const [key, val] of Object.entries(flat)) {
          const parts = key.split('.');
          let o = out;
          for (let i = 0; i < parts.length - 1; i++) {
            if (!(parts[i] in o)) o[parts[i]] = {};
            o = o[parts[i]];
          }
          o[parts[parts.length - 1]] = val;
        }
        return out;
      },

      async init() {
        window.__vaultrSettingsShell = this;
        window.__vaultrHotkeys.register('open-settings', ',', function() {
          Alpine.store('settingsModal').open = true;
        });
        this.$watch('$store.settingsModal.open', async (open) => {
          if (!open) {
            if (window.__vaultrEscPop) window.__vaultrEscPop('settings');
            this.cancelWechatLogin();
            return;
          }
          if (window.__vaultrEscPush) window.__vaultrEscPush('settings', () => { Alpine.store('settingsModal').open = false; });
          if (!this._inited) {
            this._inited = true;
            if (window.vaultrDesktop) {
              this.serverUrl = await window.vaultrDesktop.getServerUrl();
            }
            this.$watch('tab', val => {
              if (val === 'server') { this.loadServerStatus(); this.loadConfig(); }
              if (val === 'mates') {
                this.matesSubPage = null;
                this.loadMates();
                if (!this.agentsLoaded) this.loadAgents();
                if (!this.mateEventDefs.length) this.loadMateEvents();
              }
              if (val === 'skills') this.loadSkills();
              if (val === 'agents') { if (!this.agentsLoaded) this.loadAgents(); }
            });
          }
          await Promise.all([this.loadConfig(), this.loadServerStatus()]);
          if (this.tab === 'mates') {
            this.loadMates();
            if (!this.agentsLoaded) this.loadAgents();
            if (!this.mateEventDefs.length) this.loadMateEvents();
          }
          if (this.tab === 'skills') this.loadSkills();
          if (this.tab === 'agents') { if (!this.agentsLoaded) this.loadAgents(); }
        });
      },

      async loadServerStatus() {
        if (!window.vaultrDesktop?.getServerProcessStatus) return;
        try {
          const s = await window.vaultrDesktop.getServerProcessStatus();
          this.serverManaged = s.managed;
          this.serverRunning = s.alive;
        } catch { /* noop */ }
      },

      async stopServer() {
        this.serverStopping = true;
        this.serverStopError = '';
        try {
          const r = await window.vaultrDesktop.stopServer();
          if (!r.ok) { this.serverStopError = r.error || 'Stop failed'; this.serverStopping = false; }
        } catch (e) { this.serverStopError = e.message; this.serverStopping = false; }
      },

      async loadAgents(force = false) {
        this.agentsLoading = true;
        this.agentsError = '';
        try {
          const url = force ? '/api/agents?force=true' : '/api/agents';
          const r = await fetch(url);
          if (!r.ok) throw new Error('HTTP ' + r.status);
          const d = await r.json();
          this.agents = d.agents || [];
          this.agentsFromCache = d.fromCache || false;
          this.agentsCachedAt = d.fetchedAt || 0;
          this.agentsLoaded = true;
        } catch (e) { this.agentsError = e.message; }
        finally { this.agentsLoading = false; }
      },

      get sortedAgents() {
        return [...this.agents].sort((a, b) => (b.available ? 1 : 0) - (a.available ? 1 : 0));
      },

      agentDisplayModels(ag) { return (ag.models || []).slice(0, 4); },

      copyCliExample(ag) {
        if (!ag.cliExample) return;
        navigator.clipboard.writeText(ag.cliExample).catch(function() {});
        this.agentCopied = ag.id;
        setTimeout(() => { if (this.agentCopied === ag.id) this.agentCopied = ''; }, 1500);
      },

      async loadMates() {
        this.matesLoading = true;
        this.matesError = '';
        try {
          const r = await fetch('/api/mates');
          if (!r.ok) throw new Error('HTTP ' + r.status);
          const d = await r.json();
          this.matesList = d.mates || [];
        } catch(e) { this.matesError = e.message; }
        finally { this.matesLoading = false; }
      },

      async loadMateEvents() {
        try {
          const r = await fetch('/api/mate-events');
          if (!r.ok) return;
          const d = await r.json();
          this.mateEventDefs = d.events || [];
        } catch(_) {}
      },

      newMate() {
        const first = this.agents.find(function(a){ return a.available; });
        const firstModel = (first && first.models && first.models.length) ? first.models[0].id : '';
        this.mateDraft = { name: '', description: '', agentId: first ? first.id : '', model: firstModel, color: this.mateColors[0], cwd: '', systemPrompt: '', enabled: true };
        this.mateTriggers = [];
        this.mateEditId = '';
        this.mateSaveError = '';
        this.mateFormMode = 'create';
      },

      async openMateEdit(id) {
        try {
          const [r] = await Promise.all([
            fetch('/api/mates/' + id),
            this.agentsLoaded ? Promise.resolve() : this.loadAgents(),
          ]);
          if (!r.ok) throw new Error('HTTP ' + r.status);
          const d = await r.json();
          const m = d.mate;
          const validColor = this.mateColors.includes(m.color) ? m.color : this.mateColors[0];
          const savedModel = m.model || '';
          const knownModels = this.mateModelsForAgent(m.agentId);
          const modelValid = !savedModel || savedModel === 'default' ||
            knownModels.length === 0 ||
            knownModels.some(function(x) { return x.id === savedModel; });
          this.mateDraft = { name: m.name, description: m.description || '', agentId: m.agentId, model: modelValid ? savedModel : '', color: validColor, cwd: m.cwd || '', systemPrompt: m.systemPrompt || '', enabled: m.enabled };
          this.mateTriggers = (m.triggers || []).map(function(t) {
            return Object.assign({}, t, {
              eventTypes: (t.eventTypes || []).map(function(et) {
                return et === 'weixin_message' ? 'wechat_message' : et;
              }),
              schedule: t.schedule || '',
              pathPrefixes: t.pathPrefixes || [],
            });
          });
          this.mateEditId = m.id;
          this.mateSaveError = '';
          this.mateFormMode = 'edit';
        } catch(e) { window.showError('Load failed: ' + e.message, 'Load failed'); }
      },

      mateModelsForAgent(agentId) {
        const a = this.agents.find(function(x) { return x.id === agentId; });
        return (a && a.models) ? a.models : [];
      },

      onMateAgentChange() {
        const models = this.mateModelsForAgent(this.mateDraft.agentId);
        this.mateDraft.model = models.length ? models[0].id : '';
      },

      async saveMate() {
        if (this.mateSaving || !this.mateDraft.name.trim()) return;
        this.mateSaving = true;
        this.mateSaveError = '';
        try {
          const payload = Object.assign({}, this.mateDraft, { triggers: this.mateTriggers });
          const url = this.mateFormMode === 'create' ? '/api/mates' : '/api/mates/' + this.mateEditId;
          const method = this.mateFormMode === 'create' ? 'POST' : 'PUT';
          const r = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
          if (!r.ok) { this.mateSaveError = await r.text(); return; }
          const d = await r.json();
          const saved = Object.assign({}, d.mate, { triggerCount: (d.mate.triggers || []).length });
          if (this.mateFormMode === 'create') {
            this.matesList.push(saved);
          } else {
            this.matesList = this.matesList.map(function(m) { return m.id === saved.id ? saved : m; });
          }
          this.mateFormMode = null;
        } catch(e) { this.mateSaveError = e.message; }
        finally { this.mateSaving = false; }
      },

      async deleteMate(id) {
        const ok = (typeof window.showConfirm === 'function')
          ? await window.showConfirm({ title: 'Delete mate', message: 'This mate and all its data will be permanently deleted.', confirmLabel: 'Delete', danger: true })
          : window.confirm('Delete this mate? This cannot be undone.');
        if (!ok) return;
        try {
          const r = await fetch('/api/mates/' + id, { method: 'DELETE' });
          if (!r.ok) { window.showError('Delete failed (server error)', 'Delete failed'); return; }
          this.matesList = this.matesList.filter(function(m) { return m.id !== id; });
        } catch(e) { window.showError('Delete failed: ' + e.message, 'Delete failed'); }
      },

      async moveMate(idx, dir) {
        var newIdx = idx + dir;
        if (newIdx < 0 || newIdx >= this.matesList.length) return;
        var tmp = this.matesList[idx];
        this.matesList[idx] = this.matesList[newIdx];
        this.matesList[newIdx] = tmp;
        this.matesList = this.matesList.slice();
        try {
          await fetch('/api/mates/reorder', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ids: this.matesList.map(function(m) { return m.id; }) }),
          });
        } catch(_) {}
      },

      addMateTrigger() {
        this.mateTriggers.push({ id: '', mateId: '', eventTypes: ['note_created'], schedule: '', prompt: '', pathPrefixes: [], enabled: true });
      },

      removeMateTrigger(idx) { this.mateTriggers.splice(idx, 1); },

      isScheduledTrigger(t) { return (t.eventTypes || []).indexOf('scheduled') >= 0; },
      isWechatTrigger(t) { return (t.eventTypes || []).indexOf('wechat_message') >= 0; },
      isCompileTrigger(t) { return (t.eventTypes || []).indexOf('compile_requested') >= 0; },

      matePromptVarsForTrigger(t) {
        if (this.isScheduledTrigger(t)) return this.matePromptVarsScheduled;
        if (this.isWechatTrigger(t)) return this.matePromptVarsWechat;
        if (this.isCompileTrigger(t)) return this.matePromptVarsCompile;
        return this.matePromptVarsVault;
      },

      matePromptPlaceholder(t) {
        if (this.isScheduledTrigger(t)) return 'Review my vault and write a daily digest. Time: \x7b\x7b.Now\x7d\x7d';
        if (this.isWechatTrigger(t)) return 'Reply to this WeChat message:\n\n\x7b\x7b.Content\x7d\x7d';
        if (this.isCompileTrigger(t)) return 'Compile \x7b\x7b.Path\x7d\x7d into knowledge units.';
        return 'Summarize the key points in \x7b\x7b.Path\x7d\x7d';
      },

      setMateET(trigger, et) {
        if (et === 'scheduled') {
          trigger.eventTypes = ['scheduled'];
          if (!trigger.schedule) trigger.schedule = 'daily 09:00';
        } else {
          trigger.eventTypes = [et];
          trigger.schedule = '';
        }
      },

      insertMateVar(ti, token, event) {
        this.mateActivePromptIdx = ti;
        const promptBlock = event.target.closest('.mate-trigger-prompt');
        const ta = promptBlock && promptBlock.querySelector('textarea');
        if (ta && typeof ta.selectionStart === 'number') {
          const start = ta.selectionStart;
          const end = ta.selectionEnd;
          const val = ta.value || '';
          ta.value = val.slice(0, start) + token + val.slice(end);
          ta.dispatchEvent(new Event('input', { bubbles: true }));
          const pos = start + token.length;
          ta.focus();
          ta.setSelectionRange(pos, pos);
          return;
        }
        const t = this.mateTriggers[ti];
        if (!t) return;
        t.prompt = (t.prompt || '') + token;
      },

      async loadSkills() {
        this.skillsLoading = true;
        this.skillsError = '';
        try {
          const r = await fetch('/api/skills');
          if (!r.ok) throw new Error('HTTP ' + r.status);
          const d = await r.json();
          this.skillsList = d.skills || [];
        } catch(e) { this.skillsError = e.message; }
        finally { this.skillsLoading = false; }
      },

      async toggleSkill(name, enabled) {
        this.skillsList = this.skillsList.map(function(s) {
          return s.name === name ? Object.assign({}, s, { enabled: enabled }) : s;
        });
        this.skillsToggling = Object.assign({}, this.skillsToggling, { [name]: true });
        try {
          const action = enabled ? 'enable' : 'disable';
          const r = await fetch('/api/skills/' + encodeURIComponent(name) + '/' + action, { method: 'POST' });
          if (!r.ok) {
            this.skillsList = this.skillsList.map(function(s) {
              return s.name === name ? Object.assign({}, s, { enabled: !enabled }) : s;
            });
          }
        } catch(_) {
          this.skillsList = this.skillsList.map(function(s) {
            return s.name === name ? Object.assign({}, s, { enabled: !enabled }) : s;
          });
        } finally {
          const t = Object.assign({}, this.skillsToggling);
          delete t[name];
          this.skillsToggling = t;
        }
      },

      async loadConfig() {
        this.cfgLoading = true;
        this.cfgError = '';
        try {
          const [sr, vr] = await Promise.all([
            fetch('/api/config/schema'),
            fetch('/api/config'),
          ]);
          this.schema = (await sr.json()).fields || [];
          const vd = await vr.json();
          this.values = vd.values || {};
          this.secrets = vd.secrets || {};
          this.openSection = '';
          await this.loadWechatStatus();
        } catch (e) { this.cfgError = e.message; }
        finally { this.cfgLoading = false; }
      },

      async loadWechatStatus() {
        try {
          const r = await fetch('/api/wechat/status');
          if (!r.ok) return;
          const d = await r.json();
          this.wechatStatus = {
            connected: !!d.connected,
            account_id: d.account_id || '',
            saved_at: d.saved_at || '',
            enabled: !!d.enabled,
          };
        } catch { /* noop */ }
      },

      cancelWechatLogin() {
        if (this.wechatPollTimer) { clearInterval(this.wechatPollTimer); this.wechatPollTimer = null; }
        this.wechatLoginBusy = false;
        this.wechatLoginStatus = '';
        this.wechatQrcode = '';
        this.wechatQrcodeImg = '';
      },

      wechatStatusLabel(status) {
        const m = { wait: 'Open WeChat and scan the QR code', scaned: 'QR scanned — confirm login in WeChat', expired: 'QR expired — fetching a new code…', confirmed: 'Login successful' };
        return m[status] || status;
      },

      async startWechatLogin() {
        this.wechatAuthError = '';
        this.wechatLoginOk = false;
        this.wechatAuthBusy = true;
        try {
          const r = await fetch('/api/wechat/login/start', { method: 'POST' });
          const d = await r.json();
          if (!r.ok) throw new Error(d.error || 'Failed to start login');
          this.wechatQrcode = d.qrcode || '';
          this.wechatQrcodeImg = d.qrcode_image || '';
          this.wechatLoginBusy = true;
          this.wechatLoginStatus = this.wechatStatusLabel('wait');
          if (this.wechatPollTimer) clearInterval(this.wechatPollTimer);
          this.wechatPollTimer = setInterval(() => this.pollWechatLogin(), 1500);
          await this.pollWechatLogin();
        } catch (e) { this.wechatAuthError = e.message; this.cancelWechatLogin(); }
        finally { this.wechatAuthBusy = false; }
      },

      async pollWechatLogin() {
        if (!this.wechatQrcode) return;
        try {
          const r = await fetch('/api/wechat/login/status?qrcode=' + encodeURIComponent(this.wechatQrcode));
          const d = await r.json();
          if (!r.ok) throw new Error(d.error || 'Login poll failed');
          if (d.qrcode && d.qrcode !== this.wechatQrcode) this.wechatQrcode = d.qrcode;
          if (d.qrcode_image) this.wechatQrcodeImg = d.qrcode_image;
          if (d.status) this.wechatLoginStatus = this.wechatStatusLabel(d.status);
          if (d.status === 'confirmed') {
            this.cancelWechatLogin();
            this.wechatLoginOk = true;
            await this.loadWechatStatus();
            await this.loadConfig();
          }
        } catch (e) { this.wechatAuthError = e.message; this.cancelWechatLogin(); }
      },

      async wechatLogout() {
        this.wechatAuthError = '';
        this.wechatLoginOk = false;
        this.wechatAuthBusy = true;
        try {
          const r = await fetch('/api/wechat/logout', { method: 'POST' });
          const d = await r.json();
          if (!r.ok) throw new Error(d.error || 'Logout failed');
          this.wechatStatus = { connected: false, account_id: '', saved_at: '', enabled: false };
          await this.loadConfig();
        } catch (e) { this.wechatAuthError = e.message; }
        finally { this.wechatAuthBusy = false; }
      },

      async saveConfig() {
        if (!this.hasDirty || this.cfgSaving || this.cfgRestarting) return;
        this.cfgSaving = true;
        this.cfgSaveError = '';
        this.cfgSaveOk = false;
        this.cfgRestartError = '';
        let saved = false;
        try {
          const res = await fetch('/api/config', {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ patch: this.buildNested(this.patch) }),
          });
          const data = await res.json();
          if (!res.ok) {
            const detail = Array.isArray(data.errors) ? data.errors.join('; ') : (data.error || 'Save failed');
            this.cfgSaveError = detail;
          } else {
            this.patch = {};
            const prevSection = this.openSection;
            await this.loadConfig();
            this.openSection = prevSection;
            saved = true;
          }
        } catch (e) { this.cfgSaveError = e.message; }
        finally { this.cfgSaving = false; }

        if (saved && window.vaultrDesktop?.restartServer) {
          this.cfgRestarting = true;
          try {
            const r = await window.vaultrDesktop.restartServer();
            if (r.ok) return;
            if (r.reason === 'no_pid') { this.cfgSaveOk = true; }
            else { this.cfgRestartError = r.error || 'Restart failed'; }
          } catch (e) { this.cfgRestartError = e.message; }
          finally { this.cfgRestarting = false; }
        } else if (saved) {
          this.cfgSaveOk = true;
        }
      },

      async applyServerUrl() {
        this.urlError = '';
        const raw = this.serverUrl.trim().replace(/\/$/, '');
        try {
          const p = new URL(raw);
          if (!['http:', 'https:'].includes(p.protocol)) throw new Error('Must use http:// or https://');
          this.urlSaving = true;
          await window.vaultrDesktop.setServerUrl(raw);
        } catch (e) { this.urlError = e.message; this.urlSaving = false; }
      },

      async pickFolder(key, currentVal) {
        if (!window.vaultrDesktop?.pickFolder) return;
        const result = await window.vaultrDesktop.pickFolder({
          title: 'Select Vault Root Folder',
          defaultPath: currentVal || undefined,
        });
        if (!result.canceled && result.path) this.setVal(key, result.path);
      },
    };
  }
`

func cselectHTML(labelExpr, body string) string {
	return `<div class="cselect" x-data="{ csOpen: false }" @click.outside="csOpen = false">` +
		`<button type="button" class="cselect-btn" :class="{open: csOpen}" @click="csOpen = !csOpen" @keydown.escape="csOpen = false">` +
		`<span class="cselect-btn-text" x-text="` + labelExpr + `"></span>` +
		`<svg fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"/></svg>` +
		`</button><div class="cselect-dropdown" x-show="csOpen">` + body + `</div></div>`
}

func mateBackBtnHTML(onclick, label string) string {
	return `<button class="mate-back-btn" type="button" @click="` + onclick + `">` +
		`<svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/></svg>` +
		label + `</button>`
}

func toolbarRefreshBtnHTML(loadingExpr, onclick, busyLabel string) string {
	return `<button class="agents-toolbar-btn" :class="{spinning: ` + loadingExpr + `}" @click="` + onclick + `" :disabled="` + loadingExpr + `">` +
		`<svg fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path stroke-linecap="round" stroke-linejoin="round" d="M3 3v5h5"/><path stroke-linecap="round" stroke-linejoin="round" d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/><path stroke-linecap="round" stroke-linejoin="round" d="M16 16h5v5"/></svg>` +
		`<span x-text="` + loadingExpr + ` ? '` + busyLabel + `' : 'Refresh'"></span></button>`
}
