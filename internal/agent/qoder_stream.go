package agent

import (
	"encoding/json"
	"strings"
)

// QoderStream parses qoder stream-json JSONL.
type QoderStream struct {
	buffer              string
	on                  func(map[string]any)
	emittedThinkingStart bool
}

func NewQoderStream(on func(map[string]any)) *QoderStream {
	return &QoderStream{on: on}
}

func (q *QoderStream) Feed(chunk string) { q.feedString(chunk) }
func (q *QoderStream) Flush() {
	s := strings.TrimSpace(q.buffer)
	q.buffer = ""
	if s != "" {
		q.handleLine(s)
	}
}

func (q *QoderStream) feedString(chunk string) {
	q.buffer += chunk
	for {
		nl := strings.IndexByte(q.buffer, '\n')
		if nl < 0 {
			return
		}
		line := strings.TrimSpace(q.buffer[:nl])
		q.buffer = q.buffer[nl+1:]
		if line == "" {
			continue
		}
		q.handleLine(line)
	}
}

func (q *QoderStream) handleLine(line string) {
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		q.on(map[string]any{"type": "raw", "line": line})
		return
	}
	typ, _ := obj["type"].(string)
	switch typ {
	case "system":
		if obj["subtype"] == "init" {
			ev := map[string]any{"type": "status", "label": "initializing", "model": obj["model"]}
			q.on(ev)
			if sid := toString(obj["session_id"]); sid != "" {
				q.on(sessionAgentEvent(sid))
			}
		}
	case "assistant":
		msg, _ := obj["message"].(map[string]any)
		if msg == nil {
			return
		}
		if arr, ok := msg["content"].([]any); ok {
			for _, x := range arr {
				b, ok := x.(map[string]any)
				if !ok {
					continue
				}
				if b["type"] == "text" {
					if t := toString(b["text"]); t != "" {
						q.on(map[string]any{"type": "text_delta", "delta": t})
					}
				}
				if b["type"] == "thinking" {
					if !q.emittedThinkingStart {
						q.emittedThinkingStart = true
						q.on(map[string]any{"type": "thinking_start"})
					}
					if t := toString(b["thinking"]); t != "" {
						q.on(map[string]any{"type": "thinking_delta", "delta": t})
					}
				}
			}
		} else if s, ok := msg["content"].(string); ok && s != "" {
			q.on(map[string]any{"type": "text_delta", "delta": s})
		}
	case "result":
		q.on(map[string]any{"type": "usage", "usage": obj["usage"], "costUsd": obj["total_cost_usd"], "durationMs": obj["duration_ms"], "stopReason": obj["stop_reason"], "isError": obj["is_error"]})
		if isTrue(obj["is_error"]) {
			q.on(map[string]any{"type": "error", "message": qoderErrMsg(obj), "raw": line})
		}
	default:
		q.on(map[string]any{"type": "raw", "line": line})
	}
}

func isTrue(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func qoderErrMsg(obj map[string]any) string {
	if e, ok := obj["error"].(map[string]any); ok {
		if m, ok := e["message"].(string); ok {
			return m
		}
	}
	return "Qoder run failed"
}
