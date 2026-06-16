package handler

import (
	"net/http"
	"strings"
)

// InfoResponse is the JSON body returned by GET /api/info.
type InfoResponse struct {
	Vault   VaultInfo   `json:"vault"`
	Server  ServerInfo  `json:"server"`
	Plugins PluginsInfo `json:"plugins"`
}

type VaultInfo struct {
	Path string `json:"path"`
}

type ServerInfo struct {
	Addr string `json:"addr"`
	TLS  bool   `json:"tls"`
}

type PluginsInfo struct {
	Search     SearchPluginInfo     `json:"search"`
	GitSync    GitSyncPluginInfo    `json:"git_sync"`
	Compile    CompilePluginInfo    `json:"compile"`
	ImageFetch ImageFetchPluginInfo `json:"image_fetch"`
}

type SearchPluginInfo struct {
	Enabled  bool `json:"enabled"`
	UseJieba bool `json:"use_jieba"`
}

type GitSyncPluginInfo struct {
	Enabled      bool   `json:"enabled"`
	Remote       string `json:"remote,omitempty"`
	Branch       string `json:"branch,omitempty"`
	AutoCommit   bool   `json:"auto_commit,omitempty"`
	SyncInterval string `json:"sync_interval,omitempty"`
}

type CompilePluginInfo struct {
	Enabled bool `json:"enabled"`
}

// ImageFetchPluginInfo describes the remote image download plugin.
type ImageFetchPluginInfo struct {
	Enabled   bool   `json:"enabled"`
	AssetsDir string `json:"assets_dir,omitempty"`
}

// Info handles POST /api/info.
func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	cfg := h.cfg

	info := InfoResponse{
		Vault: VaultInfo{
			Path: cfg.Vault.Path,
		},
		Server: ServerInfo{
			Addr: cfg.Server.TCPAddr(),
			TLS:  cfg.Server.TLSEnabled(),
		},
		Plugins: PluginsInfo{
			Search: SearchPluginInfo{Enabled: true, UseJieba: cfg.Plugins.Search.UseJieba},
			GitSync: GitSyncPluginInfo{
				Enabled: cfg.Plugins.GitSync.Enabled,
			},
			Compile: CompilePluginInfo{
				Enabled: cfg.Plugins.Compile.Enabled,
			},
			ImageFetch: ImageFetchPluginInfo{
				Enabled: cfg.Plugins.ImageFetch.Enabled,
			},
		},
	}

	if cfg.Plugins.GitSync.Enabled {
		info.Plugins.GitSync.Remote = cfg.Plugins.GitSync.Remote
		info.Plugins.GitSync.Branch = cfg.Plugins.GitSync.Branch
		info.Plugins.GitSync.AutoCommit = cfg.Plugins.GitSync.AutoCommit
		info.Plugins.GitSync.SyncInterval = cfg.Plugins.GitSync.SyncInterval
	}

	if cfg.Plugins.ImageFetch.Enabled {
		ad := strings.TrimSpace(cfg.Plugins.ImageFetch.AssetsDir)
		if ad == "" {
			ad = "_assets"
		}
		info.Plugins.ImageFetch.AssetsDir = ad
	}

	respondJSON(w, http.StatusOK, info)
}
