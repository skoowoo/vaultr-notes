package agent

import "sync"

// DefaultModelOption matches agents.ts DEFAULT_MODEL_OPTION.
var DefaultModelOption = ModelOption{ID: "default", Label: "Default (CLI config)"}

var (
	capsMu     sync.RWMutex
	capability = map[string]map[string]bool{} // agent id -> capability key -> true if flag present in help
)

// SetCapabilities stores probe results from --help (agents.ts agentCapabilities).
func SetCapabilities(agentID string, caps map[string]bool) {
	capsMu.Lock()
	defer capsMu.Unlock()
	c := make(map[string]bool, len(caps))
	for k, v := range caps {
		c[k] = v
	}
	capability[agentID] = c
}

// Capabilities returns a copy of probed keys for agentID.
func Capabilities(agentID string) map[string]bool {
	capsMu.RLock()
	defer capsMu.RUnlock()
	src := capability[agentID]
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]bool, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func capTrue(c map[string]bool, key string) bool {
	return c != nil && c[key]
}
