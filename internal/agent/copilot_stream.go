package agent

import (
	"encoding/json"
	"strings"
)

// CopilotStream parses GitHub Copilot CLI JSONL (--output-format json).
type CopilotStream struct {
	buffer string
	on     func(map[string]any)
}

func NewCopilotStream(on func(map[string]any)) *CopilotStream {
	return &CopilotStream{on: on}
}

func (c *CopilotStream) Feed(chunk string) {
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
		c.handle(obj)
	}
}

func (c *CopilotStream) Flush() {
	s := strings.TrimSpace(c.buffer)
	c.buffer = ""
	if s == "" {
		return
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		c.on(map[string]any{"type": "raw", "line": s})
		return
	}
	c.handle(obj)
}

func (c *CopilotStream) handle(obj map[string]any) {
	typ, _ := obj["type"].(string)
	data, _ := obj["data"].(map[string]any)
	switch typ {
	case "session.tools_updated":
		if m, ok := data["model"].(string); ok {
			c.on(map[string]any{"type": "status", "label": "initializing", "model": m})
		}
	case "assistant.turn_start":
		c.on(map[string]any{"type": "status", "label": "streaming"})
	case "assistant.reasoning_delta":
		if dc, ok := data["deltaContent"].(string); ok {
			c.on(map[string]any{"type": "thinking_delta", "delta": dc})
		}
	case "assistant.message_delta":
		if dc, ok := data["deltaContent"].(string); ok {
			c.on(map[string]any{"type": "text_delta", "delta": dc})
		}
	case "tool.execution_start":
		c.on(map[string]any{"type": "tool_use", "id": data["toolCallId"], "name": data["toolName"], "input": data["arguments"]})
	case "tool.execution_complete":
		c.on(map[string]any{
			"type": "tool_result", "toolUseId": data["toolCallId"],
			"content": copilotResult(data["result"]), "isError": data["success"] == false,
		})
	case "result":
		exit := obj["exitCode"]
		ok := obj["success"] == true || exit == float64(0) || exit == 0
		var sr string
		if ok {
			sr = "completed"
		} else {
			sr = "error"
		}
		c.on(map[string]any{"type": "usage", "usage": obj["usage"], "stopReason": sr})
	}
}

func copilotResult(r any) string {
	if r == nil {
		return ""
	}
	if s, ok := r.(string); ok {
		return s
	}
	m, ok := r.(map[string]any)
	if !ok {
		return jsonStringify(r)
	}
	if s, ok := m["content"].(string); ok {
		return s
	}
	if s, ok := m["detailedContent"].(string); ok {
		return s
	}
	return jsonStringify(m)
}
