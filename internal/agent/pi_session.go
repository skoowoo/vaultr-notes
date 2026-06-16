package agent

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PiSession wraps Pi `--mode rpc` stdin/stdout (apps/daemon/src/pi-rpc.ts).
type PiSession struct {
	fatal bool
}

func (p *PiSession) HasFatalError() bool { return p.fatal }

const (
	piMaxImages     = 10
	piMaxImageBytes = 20 * 1024 * 1024
)

var piImageExt = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".gif": "image/gif", ".webp": "image/webp",
}

// AttachPiRPC sends `prompt` over Pi's line protocol and maps stdout events.
func AttachPiRPC(
	stdin io.WriteCloser,
	stdout io.Reader,
	prompt, model string,
	imagePaths []string,
	uploadRoot string,
	emit func(event string, data any),
	onDone func(),
) *PiSession {
	s := &PiSession{}
	emit("agent", map[string]any{"type": "status", "label": "initializing", "model": model})

	nextID := 1

	var images []map[string]any
	var total int
	for _, pth := range imagePaths {
		if len(images) >= piMaxImages {
			break
		}
		pth = strings.TrimSpace(pth)
		if pth == "" {
			continue
		}
		real, err := filepath.EvalSymlinks(pth)
		if err != nil {
			real = pth
		}
		st, err := os.Stat(real)
		if err != nil || st.IsDir() || !st.Mode().IsRegular() {
			continue
		}
		if uploadRoot != "" {
			root, _ := filepath.EvalSymlinks(uploadRoot)
			if root != "" {
				rp, _ := filepath.EvalSymlinks(real)
				if !strings.HasPrefix(rp, root+string(filepath.Separator)) && rp != root {
					continue
				}
			}
		}
		ext := strings.ToLower(filepath.Ext(real))
		mime, ok := piImageExt[ext]
		if !ok {
			continue
		}
		if total+int(st.Size()) > piMaxImageBytes {
			continue
		}
		data, err := os.ReadFile(real)
		if err != nil {
			continue
		}
		total += len(data)
		images = append(images, map[string]any{
			"type": "image", "data": base64.StdEncoding.EncodeToString(data), "mimeType": mime,
		})
	}

	msg := map[string]any{"id": nextID, "type": "prompt", "message": prompt}
	if len(images) > 0 {
		msg["images"] = images
	}
	b, _ := json.Marshal(msg)
	_, _ = stdin.Write(append(b, '\n'))

	started := time.Now()
	sentFirst := false

	go func() {
		sc := bufio.NewScanner(stdout)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 8<<20)
		for sc.Scan() {
			line := sc.Bytes()
			var raw map[string]any
			if err := json.Unmarshal(line, &raw); err != nil {
				continue
			}
			if raw["type"] == "response" {
				continue
			}
			if t, _ := raw["type"].(string); t == "agent_end" {
				_ = stdin.Close()
				if onDone != nil {
					onDone()
				}
				return
			}
			mapPi(raw, emit, started, &sentFirst)
		}
	}()
	return s
}

func mapPi(raw map[string]any, emit func(string, any), started time.Time, sentFirst *bool) {
	t, _ := raw["type"].(string)
	switch t {
	case "agent_start":
		emit("agent", map[string]any{"type": "status", "label": "working"})
	case "turn_start":
		emit("agent", map[string]any{"type": "status", "label": "thinking"})
	case "turn_end":
		if msg, ok := raw["message"].(map[string]any); ok {
			if u, ok := msg["usage"].(map[string]any); ok {
				emit("agent", map[string]any{"type": "usage", "usage": u, "durationMs": time.Since(started).Milliseconds()})
			}
		}
	case "message_update":
		ev, _ := raw["assistantMessageEvent"].(map[string]any)
		if ev == nil {
			return
		}
		et, _ := ev["type"].(string)
		switch et {
		case "text_delta":
			if d, ok := ev["delta"].(string); ok && d != "" {
				if !*sentFirst {
					*sentFirst = true
					emit("agent", map[string]any{"type": "status", "label": "streaming", "ttftMs": time.Since(started).Milliseconds()})
				}
				emit("agent", map[string]any{"type": "text_delta", "delta": d})
			}
		case "thinking_delta":
			if d, ok := ev["delta"].(string); ok {
				emit("agent", map[string]any{"type": "thinking_delta", "delta": d})
			}
		case "thinking_start":
			emit("agent", map[string]any{"type": "thinking_start"})
		case "error":
			reason, _ := ev["reason"].(string)
			emit("agent", map[string]any{"type": "error", "message": reason})
		}
	case "tool_execution_start":
		emit("agent", map[string]any{"type": "tool_use", "id": raw["toolCallId"], "name": raw["toolName"], "input": raw["args"]})
	case "tool_execution_end":
		emit("agent", map[string]any{"type": "tool_result", "toolUseId": raw["toolCallId"], "content": ""})
	case "extension_error":
		emit("agent", map[string]any{"type": "error", "message": raw["error"]})
	}
}
