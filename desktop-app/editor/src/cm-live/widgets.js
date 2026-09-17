import { WidgetType } from '@codemirror/view';

// Clickable [[wikilink]] chip; onClick injected by drawer.js (no Vaultr routing here).
export class WikiLinkWidget extends WidgetType {
  constructor(target, alias, onClick) {
    super();
    this.target = target;
    this.alias = alias;
    this.onClick = onClick;
  }

  eq(other) {
    return other.target === this.target && other.alias === this.alias;
  }

  toDOM() {
    const span = document.createElement('span');
    span.className = 'cm-lp-wikilink';
    span.textContent = this.alias || this.target;
    span.title = this.target;
    span.setAttribute('data-wl-value', this.target);
    if (this.onClick) {
      span.addEventListener('mousedown', (e) => {
        e.preventDefault();
        this.onClick(this.target, this.alias, e);
      });
    }
    return span;
  }

  // Let our listener run instead of CM6 caret-near-widget default.
  ignoreEvent() {
    return false;
  }
}

// ![[wikiimage]] → <img>; resolveSrc from drawer.js.
export class WikiImageWidget extends WidgetType {
  constructor(filename, resolveSrc) {
    super();
    this.filename = filename;
    this.resolveSrc = resolveSrc;
  }

  eq(other) {
    return other.filename === this.filename;
  }

  toDOM() {
    const wrap = document.createElement('span');
    wrap.className = 'cm-lp-wikiimage';
    const img = document.createElement('img');
    img.src = this.resolveSrc ? this.resolveSrc(this.filename) : this.filename;
    img.alt = this.filename;
    img.draggable = false;
    wrap.appendChild(img);
    return wrap;
  }

  // Click places caret nearby (no navigation).
  ignoreEvent() {
    return true;
  }
}

// GFM "[ ]"/"[x]" → checkbox; fixed 3-char replace at `pos`. No raw-edit mode.
export class TaskCheckboxWidget extends WidgetType {
  constructor(checked, pos) {
    super();
    this.checked = checked;
    this.pos = pos;
  }

  eq(other) {
    return other.checked === this.checked && other.pos === this.pos;
  }

  toDOM(view) {
    const input = document.createElement('input');
    input.type = 'checkbox';
    input.className = 'cm-lp-task-checkbox';
    input.checked = this.checked;
    input.addEventListener('mousedown', (e) => {
      e.preventDefault();
      const insert = this.checked ? '[ ]' : '[x]';
      view.dispatch({ changes: { from: this.pos, to: this.pos + 3, insert } });
    });
    return input;
  }

  ignoreEvent() {
    return false;
  }
}

// ![alt](url) — URL from source (not wiki resolveSrc).
export class MarkdownImageWidget extends WidgetType {
  constructor(url, alt) {
    super();
    this.url = url;
    this.alt = alt;
  }

  eq(other) {
    return other.url === this.url && other.alt === this.alt;
  }

  toDOM() {
    const wrap = document.createElement('span');
    wrap.className = 'cm-lp-wikiimage';
    const img = document.createElement('img');
    img.src = this.url;
    img.alt = this.alt || '';
    img.draggable = false;
    wrap.appendChild(img);
    return wrap;
  }

  ignoreEvent() {
    return true;
  }
}

// "-"/"*"/"+" → •. Ordered markers stay literal (real content).
export class BulletWidget extends WidgetType {
  eq() {
    return true;
  }

  toDOM() {
    const span = document.createElement('span');
    span.className = 'cm-lp-bullet';
    span.textContent = '•';
    return span;
  }

  ignoreEvent() {
    return true;
  }
}

// Thematic break as <hr>. Single-line block only — multi-line block replace
// corrupts CM6 height-map (old frontmatter card bug).
export class HorizontalRuleWidget extends WidgetType {
  // Blank after hr: drop bottom padding so it doesn't stack on blank-line height.
  constructor(followedByBlank) {
    super();
    this.followedByBlank = !!followedByBlank;
  }

  eq(other) {
    return other.followedByBlank === this.followedByBlank;
  }

  toDOM() {
    const hr = document.createElement('hr');
    hr.className = 'cm-lp-hr' + (this.followedByBlank ? ' cm-lp-hr-tight-bottom' : '');
    return hr;
  }

  ignoreEvent() {
    return true;
  }
}

// Empty "| |" cell — no TableCell node; nbsp keeps box height.
export class TableEmptyCellWidget extends WidgetType {
  constructor(className, style) {
    super();
    this.className = className;
    this.style = style || null;
  }

  eq(other) {
    return other.className === this.className && other.style === this.style;
  }

  toDOM() {
    const span = document.createElement('span');
    span.className = this.className;
    if (this.style) span.setAttribute('style', this.style);
    span.innerHTML = '&nbsp;';
    return span;
  }

  ignoreEvent() {
    return true;
  }
}

// Delimiter row hairline — needs inline-block + line font-size shrink (theme.js).
export class TableDelimiterWidget extends WidgetType {
  eq() {
    return true;
  }

  toDOM() {
    const span = document.createElement('span');
    span.className = 'cm-lp-table-delim-widget';
    return span;
  }

  ignoreEvent() {
    return true;
  }
}

// Frontmatter collapsed line — same hairline trick as TableDelimiterWidget.
export class FrontmatterCollapsedLineWidget extends WidgetType {
  eq() {
    return true;
  }

  toDOM() {
    const span = document.createElement('span');
    span.className = 'cm-lp-fm-collapsed-widget';
    return span;
  }

  ignoreEvent() {
    return true;
  }
}

// "+N more" — selection-only into first hidden line; decorateFrontmatter expands.
export class FrontmatterMoreWidget extends WidgetType {
  constructor(count, revealPos) {
    super();
    this.count = count;
    this.revealPos = revealPos;
  }

  eq(other) {
    return other.count === this.count && other.revealPos === this.revealPos;
  }

  toDOM(view) {
    const span = document.createElement('span');
    span.className = 'cm-lp-fm-more';
    span.textContent = '+' + this.count + ' more';
    span.addEventListener('mousedown', (e) => {
      e.preventDefault();
      view.dispatch({ selection: { anchor: this.revealPos } });
    });
    return span;
  }

  ignoreEvent() {
    return false;
  }
}
