  window.handleSearchResultSelection = function(el) {
    if (!el || !el.dataset) return false;
    var focusURL = el.dataset.focusUrl || '';
    if (!focusURL) return false;
    try {
      var u = new URL(focusURL, window.location.origin);
      var p = u.searchParams.get('path');
      if (!p) return false;
      var nameEl = el.querySelector('.sr-name');
      var title = nameEl ? nameEl.textContent.trim() : (p.split('/').pop().replace(/\.md$/, '') || 'Note');
      if (window.__vaultrDrawer) {
        void window.__vaultrDrawer.openNoteInDrawer(p, title,
          el.dataset.noteIsKnowledge === 'true', false, el.dataset.noteIsIndex === 'true',
          el.dataset.noteCanCompile === 'true');
        return true;
      }
    } catch(e) { return false; }
    return false;
  };

  function agentChatCtrl() {
    var ctrl = Object.assign(drawerCtrl(), {
      mates: [],
      selectedMateId: '',
      conversationId: '',
      messages: [],
      mateEventDefs: [],
      inputText: '',
      isRunning: false,
      currentRunId: null,
      timeTick: 0,
      isMac: /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent),

      _timeTicker: null,
      toastText: '',
      toastKind: 'ok',
      toastVisible: false,
      _toastTimer: null,
      _runPollerTimer: null,
      _runPollerSeq: 0,
      _syncTimer: null,
      _syncSeq: 0,
      _lastMsgMs: 0,
      _mdCache: new Map(),
      convType: 'chat',
      convTypes: [
        { value: 'chat',    label: 'Chat'    },
        { value: 'trigger', label: 'Trigger' },
      ],

      async init() {
        this.initDrawer();
        await Promise.all([this.loadMates(), this.loadMateEvents()]);
        // Restore last selected mate.
        var lastMateId = sessionStorage.getItem('vaultr_mate_id');
        var target = lastMateId ? this.mates.find(function(m) { return m.id === lastMateId; }) : null;
        this.selectedMateId = (target || this.mates[0] || {}).id || '';
        if (this.selectedMateId) void this.refreshMateConversation(this.selectedMateId);

        this.timeTick = Date.now();
        this._timeTicker = setInterval(function() { self.timeTick = Date.now(); }, 30000);
        this.$watch('timeTick', () => this._refreshTimes());

        // Delegate wiki-link clicks in chat messages to the drawer editor.
        var chatScroll = document.getElementById('chat-scroll');
        if (chatScroll) {
          chatScroll.addEventListener('click', function(e) {
            var a = e.target.closest('a');
            if (!a) return;
            var href = a.getAttribute('href') || '';
            if (!href.startsWith('/notes?')) return;
            e.preventDefault();
            try {
              var name = new URLSearchParams(href.split('?')[1] || '').get('name') || '';
              if (name && typeof __vaultrDrawerOpenWikiLink === 'function') {
                void __vaultrDrawerOpenWikiLink(name.replace(/\.md$/, ''));
              }
            } catch(_) {}
          });
        }
      },

      async loadMates() {
        try {
          var resp = await fetch('/api/mates');
          if (!resp.ok) return;
          this.mates = ((await resp.json()).mates || []).filter(function(m) { return m.enabled; });
        } catch(_) {}
      },

      async loadMateEvents() {
        try {
          var resp = await fetch('/api/mate-events');
          if (!resp.ok) return;
          this.mateEventDefs = (await resp.json()).events || [];
        } catch(_) {}
      },

      triggerEventLabel(type) {
        var d = this.mateEventDefs.find(function(e) { return e.type === type; });
        return d ? d.label : type;
      },

      async refreshMateConversation(mateId) {
        if (this.isRunning || !mateId) return;
        try {
          var resp = await fetch('/api/conversations?mateId=' + encodeURIComponent(mateId) + '&type=' + encodeURIComponent(this.convType), { cache: 'no-store' });
          if (!resp.ok) return;
          var convs = (await resp.json()).conversations || [];
          this.conversationId = convs.length > 0 ? convs[0].id : '';
          if (this.conversationId) {
            await this.loadMessages(this.conversationId);
          } else {
            this.messages = [];
          }
        } catch(_) {}
      },

      formatStoredMessages(msgs) {
        var out = [];
        for (var i = 0; i < msgs.length; i++) {
          var m = msgs[i];
          if (m.role === 'user') {
            out.push({
              id: m.id,
              role: 'user', content: m.content,
              createdAt: m.createdAt ? new Date(m.createdAt).getTime() : 0,
            });
          } else {
            var at = m.updatedAt ? new Date(m.updatedAt).getTime() : (m.createdAt ? new Date(m.createdAt).getTime() : 0);
            out.push({
              id: m.id,
              role: 'assistant', agentId: m.agentId, mateId: m.mateId,
              triggerEvent: m.triggerEvent || '',
              segments: m.content ? [{ type: 'text', content: m.content }] : [],
              status: m.status || 'succeeded',
              startTime: 0, duration: 0,
              createdAt: m.createdAt ? new Date(m.createdAt).getTime() : 0,
              completedAt: at,
              copied: false,
            });
          }
        }
        return out;
      },

      async loadMessages(convId) {
        try {
          var resp = await fetch('/api/conversations/' + convId, { cache: 'no-store' });
          if (!resp.ok) return;
          var msgs = (await resp.json()).messages || [];
          this.messages = this.formatStoredMessages(msgs);
          this._refreshTimes();
          this.$nextTick(function() { this.scrollToBottom(); }.bind(this));
          var lastMs = 0;
          for (var i = 0; i < msgs.length; i++) {
            var t = msgs[i].updatedAt ? new Date(msgs[i].updatedAt).getTime() : 0;
            if (t > lastMs) lastMs = t;
          }
          this._lastMsgMs = lastMs;
          this._startSyncPoller(convId);
        } catch(_) {}
      },

      selectMate(id) {
        if (this.isRunning) return;
        if (id !== this.selectedMateId) {
          this._cancelPoller();
          this.conversationId = '';
          this.messages = [];
        }
        this.selectedMateId = id;
        sessionStorage.setItem('vaultr_mate_id', id);
        void this.refreshMateConversation(id);
      },

      async newChat() {
        if (this.isRunning || !this.selectedMateId) return;
        // Don't create a new conversation when the current one is already empty.
        if (this.messages.length === 0) return;
        this._cancelPoller();
        var d = new Date();
        var title = d.getFullYear() + '-' +
          String(d.getMonth() + 1).padStart(2, '0') + '-' +
          String(d.getDate()).padStart(2, '0') + ' ' +
          String(d.getHours()).padStart(2, '0') + ':' +
          String(d.getMinutes()).padStart(2, '0');
        try {
          var resp = await fetch('/api/conversations', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ mateId: this.selectedMateId, title: title }),
          });
          if (resp.ok) {
            var data = await resp.json();
            this.conversationId = data.conversation.id;
            this.messages = [];
          }
        } catch(_) {}
        var ta = document.getElementById('chat-textarea');
        if (ta) { ta.style.height = ''; ta.focus(); }
      },

      convTypeLabel(t) {
        var found = this.convTypes.find(function(c) { return c.value === t; });
        return found ? found.label : t;
      },

      setConvType(t) {
        if (t === this.convType) return;
        this._cancelPoller();
        this.convType = t;
        this.conversationId = '';
        this.messages = [];
        if (this.selectedMateId) void this.refreshMateConversation(this.selectedMateId);
      },

      refreshPage() { window.location.reload(); },

      async send() {
        var text = this.inputText.trim();
        if (!text || this.isRunning || !this.selectedMateId) return;
        var mate = this.selectedMate;
        if (!mate) return;

        // Lazy-create conversation on first message.
        if (!this.conversationId) {
          try {
            var cresp = await fetch('/api/conversations', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ mateId: mate.id, title: '' }),
            });
            if (cresp.ok) {
              this.conversationId = (await cresp.json()).conversation.id;
            }
          } catch(_) {}
          if (!this.conversationId) return;
        }

        this.inputText = '';
        var ta = document.getElementById('chat-textarea');
        if (ta) ta.style.height = '';
        this.isRunning = true;

        var _genId = function() {
          return (typeof crypto !== 'undefined' && crypto.randomUUID)
            ? crypto.randomUUID()
            : (Date.now().toString(36) + Math.random().toString(36).slice(2));
        };
        var _userMsgId = _genId();
        var _assistantMsgId = _genId();

        var _userTs = Date.now();
        this.messages.push({ id: _userMsgId, role: 'user', content: text, createdAt: _userTs, _fmtTime: this.formatTime(_userTs) });
        this.messages.push({
          id: _assistantMsgId,
          role: 'assistant', agentId: mate.agentId, mateId: mate.id,
          segments: [], status: 'running',
          startTime: Date.now(), duration: 0, completedAt: 0, copied: false,
        });
        var msgIdx = this.messages.length - 1;
        this.$nextTick(function() { this.scrollToBottom(); }.bind(this));

        var sseGotEnd = true;
        try {
          var resp = await fetch('/api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ mateId: mate.id, message: text, conversationId: this.conversationId, userMessageId: _userMsgId, assistantMessageId: _assistantMsgId }),
          });
          if (!resp.ok) {
            var errText = await resp.text();
            this.messages[msgIdx].segments.push({ type: 'error', message: errText || 'Request failed' });
            this.messages[msgIdx].status = 'failed';
            return;
          }
          sseGotEnd = await this.consumeSSE(resp, msgIdx);
          // For cursor-agent: reload from DB after streaming ends to guarantee the
          // displayed text matches the authoritative DB content (text_snapshot), catching
          // any remaining divergence edge cases not handled during streaming.
          if (sseGotEnd && this.conversationId && mate.agentId === 'cursor-agent') {
            await this.syncLastMsgFromDB(this.conversationId, msgIdx);
          }
        } catch(e) {
          if (msgIdx < this.messages.length) {
            this.messages[msgIdx].segments.push({ type: 'error', message: e.message || 'Connection error' });
            this.messages[msgIdx].status = 'failed';
          }
        } finally {
          // SSE dropped without an end event — agent is still running on server; poll for completion.
          if (!sseGotEnd && this.currentRunId) this.spawnRunPoller(this.currentRunId, msgIdx);
          this.isRunning = false;
          this.currentRunId = null;
          this.scrollToBottom();
        }
      },

      // syncLastMsgFromDB replaces the streamed text segments of the assistant message at
      // msgIdx with the authoritative text from the DB, correcting any streaming artifacts.
      // Non-text segments (tool_use, thinking) are preserved in place.
      async syncLastMsgFromDB(convId, msgIdx) {
        try {
          var resp = await fetch('/api/conversations/' + convId, { cache: 'no-store' });
          if (!resp.ok) return;
          var dbMsgs = (await resp.json()).messages || [];
          var dbMsg = null;
          for (var i = dbMsgs.length - 1; i >= 0; i--) {
            if (dbMsgs[i].role === 'assistant' && dbMsgs[i].content) { dbMsg = dbMsgs[i]; break; }
          }
          if (!dbMsg) return;
          var cur = this.messages[msgIdx];
          if (!cur || cur.role !== 'assistant') return;
          var nonText = (cur.segments || []).filter(function(s) { return s.type !== 'text'; });
          cur.segments = nonText.concat([{ type: 'text', content: dbMsg.content }]);
        } catch(_) {}
      },

      async consumeSSE(resp, msgIdx) {
        var reader = resp.body.getReader();
        var decoder = new TextDecoder();
        var buf = '';
        var gotEnd = false;
        try {
          while (true) {
            var chunk = await reader.read();
            if (chunk.done) break;
            buf += decoder.decode(chunk.value, { stream: true });
            var blocks = buf.split('\n\n');
            buf = blocks.pop();
            for (var bi = 0; bi < blocks.length; bi++) {
              var block = blocks[bi];
              if (!block.trim()) continue;
              var event = '', rawData = '';
              var lines = block.split('\n');
              for (var li = 0; li < lines.length; li++) {
                var line = lines[li];
                if (line.startsWith('event: ')) event = line.slice(7).trim();
                else if (line.startsWith('data: ')) rawData = line.slice(6);
              }
              if (event === 'end') gotEnd = true;
              if (event && rawData) { this.handleSSEEvent(event, rawData, msgIdx); this.scrollToBottom(); }
            }
          }
        } finally {
          try { reader.releaseLock(); } catch(_) {}
        }
        return gotEnd;
      },

      handleSSEEvent(event, rawData, msgIdx) {
        var data;
        try { data = JSON.parse(rawData); } catch(_) { return; }
        var msg = this.messages[msgIdx];
        if (!msg) return;
        switch (event) {
          case 'start': if (data.runId) this.currentRunId = data.runId; break;
          case 'heartbeat': break;
          case 'agent': this.handleAgentSegment(data, msgIdx); break;
          case 'stdout': case 'stderr': {
            var chunk = (data.chunk || '').replace(/\n$/, '');
            if (!chunk) break;
            var segs = msg.segments, last = segs.length ? segs[segs.length-1] : null;
            if (last && last.type === 'console') { last.content += '\n' + chunk; }
            else { segs.push({ type: 'console', content: chunk }); }
            break;
          }
          case 'error': msg.segments.push({ type: 'error', message: data.message || 'Error' }); break;
          case 'end':
            msg.status = data.status || 'succeeded';
            msg.duration = Date.now() - msg.startTime;
            msg.completedAt = Date.now();
            msg._fmtTime = this.formatTime(msg.completedAt);
            if (document.hidden) this.showCompletionToast(msg.status, this.getMateNameForMsg(msg));
            break;
        }
      },

      handleAgentSegment(data, msgIdx) {
        var segs = this.messages[msgIdx].segments;
        var last = segs.length ? segs[segs.length-1] : null;
        switch (data.type) {
          case 'text_delta': {
            var d = data.delta || ''; if (!d) break;
            if (last && last.type === 'text') { last.content += d; }
            else { segs.push({ type: 'text', content: d }); }
            break;
          }
          case 'text_replace': {
            // cursor-agent reformatted mid-stream: update last text segment in-place
            // (avoids a DOM remove+create flash) and remove any earlier text segments.
            var newText = data.text || '';
            var lastTi = -1;
            for (var rti = segs.length - 1; rti >= 0; rti--) {
              if (segs[rti].type === 'text') { lastTi = rti; break; }
            }
            if (lastTi >= 0) {
              for (var rti = lastTi - 1; rti >= 0; rti--) {
                if (segs[rti].type === 'text') { segs.splice(rti, 1); lastTi--; }
              }
              segs[lastTi].content = newText;
            } else if (newText) {
              segs.push({ type: 'text', content: newText });
            }
            break;
          }
          case 'thinking_start': {
            if (!last || last.type !== 'thinking') {
              segs.push({ type: 'thinking', content: '', open: false });
            }
            break;
          }
          case 'thinking_delta': {
            var td = data.delta || ''; if (!td) break;
            if (last && last.type === 'thinking') { last.content += td; }
            else { segs.push({ type: 'thinking', content: td, open: false }); }
            break;
          }
          case 'tool_use': {
            var toolName = data.name || 'tool', mergeTarget = null;
            for (var k = segs.length-1; k >= 0; k--) {
              var sk = segs[k];
              if (sk.type === 'status') continue;
              if (sk.type === 'tool_use' && sk.name === toolName && sk.results.length >= sk.count) mergeTarget = sk;
              break;
            }
            if (mergeTarget) { mergeTarget.count++; }
            else { segs.push({ type: 'tool_use', name: toolName, count: 1, results: [], open: false }); }
            break;
          }
          case 'tool_result': {
            var trContent = data.content || '';
            if (typeof trContent !== 'string') trContent = JSON.stringify(trContent);
            var pending = null;
            for (var ti = segs.length-1; ti >= 0; ti--) {
              if (segs[ti].type === 'tool_use' && segs[ti].results.length < segs[ti].count) { pending = segs[ti]; break; }
              if (segs[ti].type === 'text' || segs[ti].type === 'error') break;
            }
            if (pending) { pending.results.push(trContent); }
            else { segs.push({ type: 'tool_result', content: trContent, open: false }); }
            break;
          }
          case 'status': {
            var label = (data.label || '').trim(); if (!label || label === 'running' || label === 'requesting') break;
            if (last && last.type === 'status') { last.label = label; }
            else { segs.push({ type: 'status', label: label }); }
            break;
          }
          case 'error': segs.push({ type: 'error', message: data.message || 'Agent error' }); break;
          case 'raw': {
            var rawLine = (data.line || '').replace(/\n$/, ''); if (!rawLine) break;
            if (last && last.type === 'console') { last.content += '\n' + rawLine; }
            else { segs.push({ type: 'console', content: rawLine }); }
            break;
          }
        }
      },

      async cancel() {
        var id = this.currentRunId;
        if (!id) return;
        try { await fetch('/api/runs/' + id + '/cancel', { method: 'POST' }); } catch(_) {}
      },

      renderMarkdown(text) {
        if (!text || typeof text !== 'string') return '';
        // Skip cache while streaming — content changes every delta; only cache settled text.
        var useCache = !this.isRunning;
        if (useCache) {
          var cached = this._mdCache.get(text);
          if (cached !== undefined) return cached;
        }
        // Pre-process wiki links: [[stem]] and [[stem|text]] → standard markdown links.
        var processed = text.replace(/\[\[([^\]\[|]+?)(?:\|([^\]\[]+?))?\]\]/g, function(_, target, display) {
          target = target.trim();
          display = (display || target).trim();
          var name = target.endsWith('.md') ? target : target + '.md';
          return '[' + display + '](/notes?name=' + encodeURIComponent(name) + ')';
        });
        // Escape lone ~ to prevent casual tildes in LLM output from triggering
        // GFM strikethrough. Double-tilde (~~text~~) is preserved intentionally.
        processed = processed.replace(/(?<!~)~(?!~)/g, '\\~');
        var html;
        if (typeof marked === 'undefined') {
          html = processed.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
        } else {
          html = marked.parse(processed);
          if (typeof DOMPurify !== 'undefined') {
            html = DOMPurify.sanitize(html, {
              ALLOWED_TAGS: ['p','br','strong','em','s','del','code','pre','h1','h2','h3','h4','h5','h6',
                             'ul','ol','li','blockquote','a','hr','table','thead','tbody','tr','th','td','img'],
              ALLOWED_ATTR: ['href','title','src','alt'],
            });
          }
        }
        if (useCache) {
          if (this._mdCache.size >= 300) {
            // Evict oldest 50 entries (Map preserves insertion order) instead of
            // clearing all at once to avoid re-rendering every visible message.
            var iter = this._mdCache.keys();
            for (var ei = 0; ei < 50; ei++) {
              var nxt = iter.next();
              if (nxt.done) break;
              this._mdCache.delete(nxt.value);
            }
          }
          this._mdCache.set(text, html);
        }
        return html;
      },

      scrollToBottom() { var el = document.getElementById('chat-scroll'); if (el) el.scrollTop = el.scrollHeight; },

      handleKeydown(e) {
        if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') { e.preventDefault(); void this.send(); }
      },

      autoResize(el) { el.style.height = 'auto'; el.style.height = Math.min(el.scrollHeight, 140) + 'px'; },

      formatDuration(ms) {
        if (!ms) return '';
        return ms < 1000 ? ms + 'ms' : (ms / 1000).toFixed(1) + 's';
      },

      msgAt(msg) {
        return msg.completedAt || msg.createdAt || 0;
      },

      _refreshTimes() {
        for (var i = 0; i < this.messages.length; i++) {
          var m = this.messages[i];
          var ts = this.msgAt(m);
          if (ts) m._fmtTime = this.formatTime(ts);
        }
      },

      formatTime(ts) {
        if (!ts) return '';
        var now = Date.now();
        var diff = now - ts;
        if (diff < 0) diff = 0;
        var sec = Math.floor(diff / 1000);
        if (sec < 45) return 'now';
        var min = Math.floor(sec / 60);
        if (min < 60) return min + 'm ago';
        var hr = Math.floor(min / 60);
        if (hr < 24) return hr + 'h ago';

        var d = new Date(ts);
        var clock = this.formatClock(d);
        var today = new Date(now);
        var yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1);
        if (this.sameCalendarDay(d, yesterday)) return 'Yesterday, ' + clock;

        var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        var datePart = months[d.getMonth()] + ' ' + d.getDate();
        if (d.getFullYear() !== today.getFullYear()) datePart += ', ' + d.getFullYear();
        return datePart + ', ' + clock;
      },

      formatClock(d) {
        var h = d.getHours(), m = d.getMinutes();
        var ap = h >= 12 ? 'PM' : 'AM';
        h = h % 12;
        if (h === 0) h = 12;
        return h + ':' + String(m).padStart(2, '0') + ' ' + ap;
      },

      sameCalendarDay(a, b) {
        return a.getFullYear() === b.getFullYear() &&
          a.getMonth() === b.getMonth() &&
          a.getDate() === b.getDate();
      },

      async copyMateText(msg) {
        var text = (msg.segments || [])
          .filter(function(s) { return s.type === 'text'; })
          .map(function(s) { return s.content || ''; })
          .join('').trim();
        if (!text) return;
        try {
          await navigator.clipboard.writeText(text);
          msg.copied = true;
          var self = this;
          setTimeout(function() { msg.copied = false; }, 2000);
        } catch(_) {}
      },

      getMateNameForMsg(msg) {
        var m = this.mates.find(function(m) { return m.id === msg.mateId; });
        return m ? m.name : (msg.agentId || 'Agent');
      },

      getMateColor(mateId) {
        var m = this.mates.find(function(m) { return m.id === mateId; });
        return (m && m.color) ? m.color : '';
      },

      mateInitials(name) {
        if (!name) return '?';
        var parts = name.trim().split(/\s+/);
        if (parts.length >= 2) return (parts[0][0] + parts[parts.length-1][0]).toUpperCase();
        return parts[0].slice(0, 1).toUpperCase();
      },

      insertPath(e) {
        var path = (e && e.detail && e.detail.path) || '';
        if (!path) return;
        var ta = document.getElementById('chat-textarea');
        var start = ta ? ta.selectionStart : this.inputText.length;
        var end   = ta ? ta.selectionEnd   : start;
        var bt = String.fromCharCode(96);
        var before = this.inputText.slice(0, start);
        var after  = this.inputText.slice(end);
        var leadSp = before.length > 0 && before[before.length - 1] !== ' ' ? ' ' : '';
        var tailSp = after.length  > 0 && after[0] !== ' ' ? ' ' : '';
        var insert = leadSp + bt + path + bt + tailSp;
        this.inputText = this.inputText.slice(0, start) + insert + this.inputText.slice(end);
        this.$nextTick(function() {
          if (ta) {
            ta.selectionStart = ta.selectionEnd = start + insert.length;
            ta.focus();
            this.autoResize(ta);
          }
        }.bind(this));
      },

      showCompletionToast(status, mateName) {
        this.toastText = (mateName || 'Agent') + (status === 'succeeded' ? ' finished' : ' failed');
        this.toastKind = status === 'succeeded' ? 'ok' : 'err';
        this.toastVisible = true;
        if (this._toastTimer) clearTimeout(this._toastTimer);
        var self = this;
        this._toastTimer = setTimeout(function() { self.toastVisible = false; }, 6000);
      },

      _cancelPoller() {
        if (this._runPollerTimer) { clearTimeout(this._runPollerTimer); this._runPollerTimer = null; }
        this._runPollerSeq++;
      },

      _cancelSyncPoller() {
        if (this._syncTimer) { clearTimeout(this._syncTimer); this._syncTimer = null; }
        this._syncSeq++;
      },

      _startSyncPoller(convId) {
        this._cancelSyncPoller();
        var self = this;
        var mySeq = self._syncSeq;
        function poll() {
          if (self._syncSeq !== mySeq) return;
          if (self.isRunning || !convId || document.visibilityState !== 'visible') {
            self._syncTimer = setTimeout(poll, 5000);
            return;
          }
          var mateId = self.selectedMateId;
          var convType = self.convType;
          // Check if a newer conversation was created (e.g. WeChat /new)
          fetch('/api/conversations?mateId=' + encodeURIComponent(mateId) + '&type=' + encodeURIComponent(convType), { cache: 'no-store' })
            .then(function(r) { return r.ok ? r.json() : null; })
            .then(function(data) {
              if (self._syncSeq !== mySeq) return;
              var convs = (data && data.conversations) || [];
              if (convs.length > 0 && convs[0].id !== convId) {
                // Active conversation switched; refreshMateConversation will restart the poller
                void self.refreshMateConversation(mateId);
                return;
              }
              // Same conversation — fetch new messages only
              fetch('/api/conversations/' + convId + '?since=' + self._lastMsgMs, { cache: 'no-store' })
                .then(function(r2) { return r2.ok ? r2.json() : null; })
                .then(function(data2) {
                  if (self._syncSeq !== mySeq) return;
                  var msgs = (data2 && data2.messages) || [];
                  if (msgs.length > 0) {
                    var newMsgs = self.formatStoredMessages(msgs);
                    var changed = false;
                    for (var i = 0; i < newMsgs.length; i++) {
                      var nm = newMsgs[i];
                      var existingIdx = -1;
                      if (nm.id) {
                        for (var k = 0; k < self.messages.length; k++) {
                          if (self.messages[k].id === nm.id) { existingIdx = k; break; }
                        }
                      }
                      if (existingIdx >= 0) {
                        var ex = self.messages[existingIdx];
                        var prevStatus = ex.status;
                        ex.content = nm.content;
                        ex.status = nm.status;
                        // If the message was already terminal in memory with rich live-streamed
                        // segments (thinking, tool_use, etc.), preserve them — DB only stores
                        // final text, so blindly replacing would wipe all streaming artifacts.
                        // Only replace segments when the message was still running (just completed)
                        // or when there are no rich segments to preserve.
                        var hasRichSegs = (ex.segments || []).some(function(s) { return s.type !== 'text'; });
                        if (prevStatus !== 'running' && hasRichSegs) {
                          var richSegs = (ex.segments || []).filter(function(s) { return s.type !== 'text'; });
                          var dbTextSegs = (nm.segments || []).filter(function(s) { return s.type === 'text'; });
                          ex.segments = richSegs.concat(dbTextSegs);
                        } else {
                          ex.segments = nm.segments;
                        }
                        ex.completedAt = nm.completedAt;
                        ex._fmtTime = nm._fmtTime;
                        if (nm.triggerEvent) ex.triggerEvent = nm.triggerEvent;
                      } else {
                        self.messages.push(nm);
                      }
                      changed = true;
                    }
                    var lastTs = 0;
                    for (var j = 0; j < msgs.length; j++) {
                      var t = msgs[j].updatedAt ? new Date(msgs[j].updatedAt).getTime() : 0;
                      if (t > lastTs) lastTs = t;
                    }
                    if (lastTs > self._lastMsgMs) self._lastMsgMs = lastTs;
                    if (changed) {
                      self._refreshTimes();
                      self.$nextTick(function() { self.scrollToBottom(); });
                    }
                  }
                  self._syncTimer = setTimeout(poll, 5000);
                })
                .catch(function() {
                  if (self._syncSeq !== mySeq) return;
                  self._syncTimer = setTimeout(poll, 5000);
                });
            })
            .catch(function() {
              if (self._syncSeq !== mySeq) return;
              self._syncTimer = setTimeout(poll, 5000);
            });
        }
        self._syncTimer = setTimeout(poll, 5000);
      },

      spawnRunPoller(runId, msgIdx) {
        this._cancelPoller();
        var self = this;
        var mySeq = self._runPollerSeq;
        var n = 0;
        function poll() {
          if (self._runPollerSeq !== mySeq) return;
          if (n++ > 720) return; // 1 h at 5 s intervals
          fetch('/api/runs/' + runId, { cache: 'no-store' })
            .then(function(r) { return r.ok ? r.json() : null; })
            .then(function(run) {
              if (self._runPollerSeq !== mySeq) return;
              if (!run) { self._runPollerTimer = setTimeout(poll, 5000); return; }
              var s = run.status;
              if (s === 'succeeded' || s === 'failed' || s === 'canceled') {
                self._runPollerTimer = null;
                var st = s === 'succeeded' ? 'succeeded' : 'failed';
                var msg = msgIdx < self.messages.length ? self.messages[msgIdx] : null;
                if (msg && msg.status === 'running') {
                  msg.status = st;
                  msg.completedAt = run.updatedAt || Date.now();
                  msg._fmtTime = self.formatTime(msg.completedAt);
                  self.scrollToBottom();
                }
                self.showCompletionToast(st, msg ? self.getMateNameForMsg(msg) : null);
              } else {
                self._runPollerTimer = setTimeout(poll, 5000);
              }
            })
            .catch(function() {
              if (self._runPollerSeq !== mySeq) return;
              self._runPollerTimer = setTimeout(poll, 5000);
            });
        }
        self._runPollerTimer = setTimeout(poll, 3000);
      },

      isLastThinkingInMsg(msg, j) {
        var segs = msg.segments || [];
        for (var k = segs.length - 1; k >= 0; k--) {
          if (segs[k].type === 'thinking') return k === j;
        }
        return false;
      },

      destroy() {
        if (this._timeTicker) { clearInterval(this._timeTicker); this._timeTicker = null; }
        if (this._toastTimer) { clearTimeout(this._toastTimer); this._toastTimer = null; }
        this._cancelPoller();
        this._cancelSyncPoller();
      },
    });

    // Object.assign flattens getters — define computed props properly so Alpine tracks them.
    Object.defineProperties(ctrl, {
      selectedMate: {
        get() { var id = this.selectedMateId; return this.mates.find(function(m) { return m.id === id; }) || null; },
        configurable: true, enumerable: true,
      },
    });
    return ctrl;
  }
