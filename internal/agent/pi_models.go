package agent

import (
	"strings"
)

// ParsePiModels parses `pi --list-models` table output (from stderr or combined).
func ParsePiModels(text string) []ModelOption {
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		cleaned = append(cleaned, l)
	}
	if len(cleaned) == 0 {
		return nil
	}
	seen := map[string]struct{}{"default": {}}
	out := []ModelOption{DefaultModelOption}
	// skip header row
	start := 0
	if len(cleaned) > 0 && strings.Contains(strings.ToLower(cleaned[0]), "provider") {
		start = 1
	}
	for i := start; i < len(cleaned); i++ {
		parts := strings.Fields(cleaned[i])
		if len(parts) < 2 {
			continue
		}
		provider := parts[0]
		modelID := parts[1]
		fullID := provider + "/" + modelID
		if _, ok := seen[fullID]; ok {
			continue
		}
		seen[fullID] = struct{}{}
		out = append(out, ModelOption{ID: fullID, Label: fullID})
	}
	if len(out) <= 1 {
		return nil
	}
	return out
}
