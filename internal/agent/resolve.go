package agent

import (
	"os"
	"path/filepath"
	"strings"
)

var agentBinEnvKeys = map[string]string{
	"codex": "CODEX_BIN",
}

// expandHomePath supports ~ and ~/
func expandHomePath(value string) (string, error) {
	if value == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(value, "~/") || strings.HasPrefix(value, "~\\") {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(h, value[2:]), nil
	}
	return value, nil
}

// ConfiguredExecutableOverride returns absolute path from env if set (CODEX_BIN).
func ConfiguredExecutableOverride(def *AgentDef, configuredEnv map[string]string) string {
	if def == nil {
		return ""
	}
	key, ok := agentBinEnvKeys[def.ID]
	if !ok {
		return ""
	}
	raw, ok := configuredEnv[key]
	if !ok {
		return ""
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	exp, err := expandHomePath(raw)
	if err != nil {
		return ""
	}
	if !filepath.IsAbs(exp) {
		return ""
	}
	if st, err := os.Stat(exp); err == nil && !st.IsDir() {
		return exp
	}
	return ""
}

// ResolveAgentExecutable returns absolute path to the agent binary or "".
func ResolveAgentExecutable(def *AgentDef, configuredEnv map[string]string) string {
	if def == nil || def.Bin == "" {
		return ""
	}
	if p := ConfiguredExecutableOverride(def, configuredEnv); p != "" {
		return p
	}
	candidates := []string{def.Bin}
	candidates = append(candidates, def.FallbackBins...)
	for _, bin := range candidates {
		if p := ResolveOnPath(bin); p != "" {
			return p
		}
	}
	return ""
}
