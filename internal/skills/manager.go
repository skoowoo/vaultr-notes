package skills

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DefaultSkills are always enabled and cannot be disabled by the user.
var DefaultSkills = []string{
	"vaultr-compile-note",
	"vaultr-index-knowledge",
	"vaultr-memory",
	"vaultr-notes",
}

// state persists only the non-default skills the user has explicitly enabled.
type state struct {
	EnabledExtra []string `json:"enabled_extra"`
}

// SkillInfo describes one skill found in the source directory.
type SkillInfo struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
	Enabled bool   `json:"enabled"`
	RepoURL string `json:"repoUrl,omitempty"`
}

// Manager tracks which skills are enabled and manages their vault symlinks.
type Manager struct {
	mu        sync.Mutex
	sourceDir string
	targets   []string
	stateFile string
	st        state
}

// Open creates a Manager rooted at vaultRoot. State is loaded from disk;
// errors (e.g. missing state file) are silently ignored so the manager is
// always usable.
func Open(vaultRoot string) *Manager {
	home, _ := os.UserHomeDir()
	m := &Manager{
		sourceDir: filepath.Join(home, ".vaultr", "skills"),
		targets: []string{
			filepath.Join(vaultRoot, ".agents", "skills"),
			filepath.Join(vaultRoot, ".claude", "skills"),
		},
		stateFile: filepath.Join(vaultRoot, ".vaultr", "skills_state.json"),
	}
	_ = m.loadState()
	return m
}

func (m *Manager) loadState() error {
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &m.st)
}

func (m *Manager) saveState() error {
	if err := os.MkdirAll(filepath.Dir(m.stateFile), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(&m.st)
	if err != nil {
		return err
	}
	return os.WriteFile(m.stateFile, data, 0644)
}

func isDefault(name string) bool {
	for _, n := range DefaultSkills {
		if n == name {
			return true
		}
	}
	return false
}

// isExtraEnabled reads m.st without locking; callers must hold m.mu.
func (m *Manager) isExtraEnabled(name string) bool {
	for _, n := range m.st.EnabledExtra {
		if n == name {
			return true
		}
	}
	return false
}

func (m *Manager) isEnabled(name string) bool {
	if isDefault(name) {
		return true
	}
	return m.isExtraEnabled(name)
}

// parseExternalSkills reads ~/.vaultr/skills/external-skills.txt and returns a
// map of skill-dir-name → GitHub repo URL. Errors are silently ignored.
func (m *Manager) parseExternalSkills() map[string]string {
	result := make(map[string]string)
	f, err := os.Open(filepath.Join(m.sourceDir, "external-skills.txt"))
	if err != nil {
		return result
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Format: GitHubUser/RepoName:path/in/repo:skill-dir-name
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}
		githubPath := strings.TrimSpace(parts[0])
		skillName := strings.TrimSpace(parts[2])
		if githubPath == "" || skillName == "" {
			continue
		}
		result[skillName] = "https://github.com/" + githubPath
	}
	return result
}

// List returns all skills found in the source directory with their current status.
func (m *Manager) List() ([]SkillInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SkillInfo{}, nil
		}
		return nil, err
	}

	external := m.parseExternalSkills()

	out := make([]SkillInfo, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		out = append(out, SkillInfo{
			Name:    name,
			Default: isDefault(name),
			Enabled: m.isEnabled(name),
			RepoURL: external[name],
		})
	}
	return out, nil
}

// Enable enables a skill by creating symlinks and persisting state.
// Default skills are always enabled; this is a no-op for them.
func (m *Manager) Enable(name string) error {
	if isDefault(name) {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	src := filepath.Join(m.sourceDir, name)
	if _, err := os.Stat(src); err != nil {
		return err
	}

	for _, t := range m.targets {
		if err := os.MkdirAll(t, 0755); err != nil {
			return err
		}
		dst := filepath.Join(t, name)
		_ = os.Remove(dst)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}

	if !m.isExtraEnabled(name) {
		m.st.EnabledExtra = append(m.st.EnabledExtra, name)
	}
	return m.saveState()
}

// Disable disables a skill by removing symlinks and persisting state.
// Returns an error for default skills.
func (m *Manager) Disable(name string) error {
	if isDefault(name) {
		return fmt.Errorf("skill %q is a default skill and cannot be disabled", name)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range m.targets {
		_ = os.Remove(filepath.Join(t, name))
	}

	var extra []string
	for _, n := range m.st.EnabledExtra {
		if n != name {
			extra = append(extra, n)
		}
	}
	m.st.EnabledExtra = extra
	return m.saveState()
}

// LinkEnabled creates symlinks for all currently-enabled skills.
// Called once at server startup.
func (m *Manager) LinkEnabled(logger *slog.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.sourceDir)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Warn("skills: read source dir", "path", m.sourceDir, "err", err)
		}
		return
	}

	for _, t := range m.targets {
		if err := os.MkdirAll(t, 0755); err != nil {
			logger.Warn("skills: create target dir", "path", t, "err", err)
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !m.isEnabled(name) {
			continue
		}
		src := filepath.Join(m.sourceDir, name)
		for _, t := range m.targets {
			dst := filepath.Join(t, name)
			_ = os.Remove(dst)
			if err := os.Symlink(src, dst); err != nil {
				logger.Warn("skills: symlink failed", "src", src, "dst", dst, "err", err)
			}
		}
	}
	logger.Info("skills linked", "src", m.sourceDir)
}
