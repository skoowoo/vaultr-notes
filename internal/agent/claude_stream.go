package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ClaudeStream feeds Claude Code stream-json lines and emits UI-shaped maps (type=key).
type ClaudeStream struct {
	buffer           string
	on               func(map[string]any)
	blocks           map[string]map[string]any
	currentMessageID string
	textStreamed     map[string]struct{}
}

func NewClaudeStream(on func(map[string]any)) *ClaudeStream {
	return &ClaudeStream{
		on:           on,
		blocks:       make(map[string]map[string]any),
		textStreamed: make(map[string]struct{}),
	}
}

func (c *ClaudeStream) Feed(chunk string) {
	c.buffer += chunk
	for {
		nl := strings.IndexByte(c.buffer, '\n')
		if nl < 0 {
			return
		}
		line := strings.TrimSpace(c.buffer[:nl])
		c.buffer = c.buffer[nl+1:]
		if line == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			c.on(map[string]any{"type": "raw", "line": line})
			continue
		}
		c.handleRoot(obj)
	}
}

func (c *ClaudeStream) Flush() {
	rem := strings.TrimSpace(c.buffer)
	c.buffer = ""
	if rem == "" {
		return
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(rem), &obj); err != nil {
		c.on(map[string]any{"type": "raw", "line": rem})
		return
	}
	c.handleRoot(obj)
}

func (c *ClaudeStream) handleRoot(obj map[string]any) {
	typ, _ := obj["type"].(string)
	switch typ {
	case "system":
		sub, _ := obj["subtype"].(string)
		if sub == "init" {
			c.on(map[string]any{"type": "status", "label": "initializing", "model": obj["model"]})
		}
		if sub == "status" {
			c.on(map[string]any{"type": "status", "label": obj["status"]})
		}
	case "stream_event":
		if ev, ok := obj["event"].(map[string]any); ok {
			c.handleStreamEvent(ev)
		}
	case "assistant":
		c.handleAssistant(obj)
	case "user":
		c.handleUser(obj)
	case "result":
		c.on(map[string]any{"type": "usage", "usage": obj["usage"], "costUsd": obj["total_cost_usd"], "stopReason": obj["stop_reason"]})
	}
}

func (c *ClaudeStream) blockKey(idx float64) string {
	return fmt.Sprintf("%s:%.0f", c.currentMessageID, idx)
}

func (c *ClaudeStream) handleAssistant(obj map[string]any) {
	msg, _ := obj["message"].(map[string]any)
	if msg == nil {
		return
	}
	if mid, ok := msg["id"].(string); ok {
		c.currentMessageID = mid
	}
	content, _ := msg["content"].([]any)
	_, streamed := c.textStreamed[c.currentMessageID]
	for _, blk := range content {
		b, ok := blk.(map[string]any)
		if !ok {
			continue
		}
		bt, _ := b["type"].(string)
		if bt == "tool_use" {
			c.on(map[string]any{"type": "tool_use", "id": b["id"], "name": b["name"], "input": b["input"]})
		}
		if !streamed && bt == "text" {
			if t, ok := b["text"].(string); ok && t != "" {
				c.on(map[string]any{"type": "text_delta", "delta": t})
			}
		}
		if !streamed && bt == "thinking" {
			if t, ok := b["thinking"].(string); ok && t != "" {
				c.on(map[string]any{"type": "thinking_delta", "delta": t})
			}
		}
	}
}

func (c *ClaudeStream) handleUser(obj map[string]any) {
	msg, _ := obj["message"].(map[string]any)
	if msg == nil {
		return
	}
	content, _ := msg["content"].([]any)
	for _, blk := range content {
		b, ok := blk.(map[string]any)
		if !ok {
			continue
		}
		if b["type"] == "tool_result" {
			c.on(map[string]any{
				"type": "tool_result", "toolUseId": b["tool_use_id"],
				"content": stringifyToolResult(b["content"]), "isError": b["is_error"],
			})
		}
	}
}

func stringifyToolResult(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		var parts []string
		for _, x := range t {
			m, ok := x.(map[string]any)
			if !ok {
				parts = append(parts, jsonStringify(x))
				continue
			}
			if m["type"] == "text" {
				parts = append(parts, toString(m["text"]))
				continue
			}
			parts = append(parts, jsonStringify(m))
		}
		return strings.Join(parts, "\n")
	default:
		return jsonStringify(v)
	}
}

func jsonStringify(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func (c *ClaudeStream) handleStreamEvent(ev map[string]any) {
	et, _ := ev["type"].(string)
	switch et {
	case "message_start":
		if m, ok := ev["message"].(map[string]any); ok {
			if id, ok := m["id"].(string); ok {
				c.currentMessageID = id
			}
		}
	case "content_block_start":
		idx, _ := ev["index"].(float64)
		cb, _ := ev["content_block"].(map[string]any)
		if cb != nil {
			c.blocks[c.blockKey(idx)] = map[string]any{
				"type": cb["type"], "name": cb["name"], "id": cb["id"], "input": "",
			}
			if cb["type"] == "thinking" {
				c.on(map[string]any{"type": "thinking_start"})
			}
		}
	case "content_block_delta":
		idx, _ := ev["index"].(float64)
		delta, _ := ev["delta"].(map[string]any)
		st := c.blocks[c.blockKey(idx)]
		if delta == nil || st == nil {
			return
		}
		dt, _ := delta["type"].(string)
		if dt == "text_delta" {
			if t, ok := delta["text"].(string); ok {
				c.textStreamed[c.currentMessageID] = struct{}{}
				c.on(map[string]any{"type": "text_delta", "delta": t})
			}
		}
		if dt == "thinking_delta" {
			if t, ok := delta["thinking"].(string); ok {
				c.textStreamed[c.currentMessageID] = struct{}{}
				c.on(map[string]any{"type": "thinking_delta", "delta": t})
			}
		}
		if dt == "input_json_delta" {
			pj, _ := delta["partial_json"].(string)
			if st["type"] == "tool_use" {
				st["input"] = toString(st["input"]) + pj
			}
		}
	case "content_block_stop":
		idx, _ := ev["index"].(float64)
		delete(c.blocks, c.blockKey(idx))
	}
}
