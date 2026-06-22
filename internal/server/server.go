package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/hardhacker/vaultr/internal/config"
	"github.com/hardhacker/vaultr/internal/mate"
	"github.com/hardhacker/vaultr/internal/plugin"
	"github.com/hardhacker/vaultr/internal/plugins/compile"
	"github.com/hardhacker/vaultr/internal/plugins/gitsync"
	"github.com/hardhacker/vaultr/internal/plugins/imgfetch"
	"github.com/hardhacker/vaultr/internal/plugins/search"
	discordplugin "github.com/hardhacker/vaultr/internal/plugins/discord"
	wechatplugin "github.com/hardhacker/vaultr/internal/plugins/wechat"
	"github.com/hardhacker/vaultr/internal/skills"
	"github.com/hardhacker/vaultr/internal/storage"
)

const shutdownTimeout = 15 * time.Second

// Server manages a TCP listener for HTTP.
type Server struct {
	handler   http.Handler
	logger    *slog.Logger
	cfg       *config.Config
	vault     *storage.Vault
	pluginMgr *plugin.Manager
}

// New creates a Server from the given config, logger, and Vault.
// It initialises any enabled plugins and wires the vault event hook.
func New(cfg *config.Config, cfgFileLoaded string, logger *slog.Logger, vault *storage.Vault) *Server {
	skillsMgr := skills.Open(vault.Root())
	skillsMgr.LinkEnabled(logger)

	mgr := plugin.NewManager(logger)

	// The search plugin is always active: it drives all full-text index updates
	// in response to vault events and backtracks unindexed notes on startup.
	searchPlugin := search.New(cfg.Plugins.Search, vault, logger)
	mgr.Register(searchPlugin)

	var gitPlugin *gitsync.Plugin
	if cfg.Plugins.GitSync.Enabled {
		gs, err := gitsync.New(cfg.Plugins.GitSync, vault.Root(), logger)
		if err != nil {
			logger.Warn("git_sync plugin could not be created", "err", err)
		} else {
			mgr.Register(gs)
			gitPlugin = gs
			logger.Info("git_sync plugin registered")
		}
	}

	var compilePlugin *compile.Plugin
	if cfg.Plugins.Compile.Enabled {
		dp, err := compile.New(cfg.Vault.KnowledgeDir, vault, logger)
		if err != nil {
			logger.Warn("compile plugin could not be created", "err", err)
		} else {
			dp.SetDispatch(mgr.Dispatch)
			mgr.Register(dp)
			compilePlugin = dp
			logger.Info("compile plugin registered", "knowledge_dir", cfg.Vault.KnowledgeDir)
		}
	}

	if cfg.Plugins.ImageFetch.Enabled {
		mgr.Register(imgfetch.New(cfg.Plugins.ImageFetch, vault, logger))
		logger.Info("image_fetch plugin registered", "assets_dir", cfg.Plugins.ImageFetch.AssetsDir)
	}

	// Mate runner: fires agent runs on vault events matching trigger configs.
	var mateStore *mate.Store
	var mateRunner *mate.Runner
	if ms, err := mate.Open(vault.Root()); err != nil {
		logger.Warn("mate store unavailable", "err", err)
	} else {
		mateStore = ms
		if err := ms.MarkStalledRunMessagesAsFailed(); err != nil {
			logger.Warn("mate: mark stalled messages failed", "err", err)
		}
		mateRunner = mate.NewRunner(ms, logger)
		mgr.Register(mateRunner)
		logger.Info("mate_runner plugin registered")
	}

	if cfg.Plugins.Wechat.Enabled {
		wp := wechatplugin.New(cfg.Plugins.Wechat, logger)
		wp.SetDispatch(mgr.Dispatch)
		mgr.Register(wp)
		logger.Info("wechat plugin registered")
	}

	if cfg.Plugins.Discord.Enabled {
		dp := discordplugin.New(cfg.Plugins.Discord, logger)
		dp.SetDispatch(mgr.Dispatch)
		mgr.Register(dp)
		logger.Info("discord plugin registered")
	}

	vault.SetShortsDir(cfg.Vault.ShortsDir)

	// Wire vault mutations → plugin event bus.
	vault.SetEventHook(func(e plugin.Event) {
		mgr.Dispatch(e)
	})

	return &Server{
		handler:   newRouter(logger, cfg, vault, cfgFileLoaded, gitPlugin, searchPlugin, compilePlugin, mateStore, mateRunner, skillsMgr, cfg.Server.APIKey),
		logger:    logger,
		cfg:       cfg,
		vault:     vault,
		pluginMgr: mgr,
	}
}

func writeServerPIDFile(path string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	line := strconv.Itoa(os.Getpid()) + "\n"
	return os.WriteFile(path, []byte(line), 0600)
}

// Run starts the configured TCP listener and blocks until a shutdown signal is
// received, then performs a graceful shutdown. If pidFile is non-empty, the
// process ID is written after Listen succeeds and the file is removed when Run
// returns (including after shutdown or early serve errors).
func (s *Server) Run(pidFile string) error {
	addr := s.cfg.Server.TCPAddr()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("tcp listen %s: %w", addr, err)
	}

	if pidFile != "" {
		if err := writeServerPIDFile(pidFile); err != nil {
			_ = l.Close()
			return fmt.Errorf("write pid file: %w", err)
		}
		defer func() {
			if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
				s.logger.Warn("remove pid file", "path", pidFile, "err", err)
			}
		}()
	}

	protocol := "http"
	if s.cfg.Server.TLSEnabled() {
		protocol = "https"
	}
	s.logger.Info("listening", "protocol", protocol, "addr", addr)

	srv := &http.Server{
		Handler:      s.handler,
		ReadTimeout:  time.Duration(s.cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.Server.WriteTimeout) * time.Second,
	}

	// Start the vault watcher so that files dropped directly into the vault
	// directory (outside the API) are picked up and indexed automatically.
	// The watcher is best-effort: failure is logged but does not prevent serving.
	watchCtx, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch()
	if err := s.vault.StartWatcher(watchCtx, s.logger); err != nil {
		s.logger.Warn("vault watcher could not start", "err", err)
	} else {
		s.logger.Info("vault watcher started", "root", s.vault.Root())
	}

	// Start plugins.
	pluginCtx, cancelPlugins := context.WithCancel(context.Background())
	defer cancelPlugins()
	s.pluginMgr.Start(pluginCtx)

	errCh := make(chan error, 1)
	go func() {
		var err error
		if s.cfg.Server.TLSEnabled() {
			err = srv.ServeTLS(l, s.cfg.Server.CertFile, s.cfg.Server.KeyFile)
		} else {
			err = srv.Serve(l)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve %s: %w", l.Addr(), err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		s.logger.Info("shutdown signal received", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	s.logger.Info("graceful shutdown", "timeout", shutdownTimeout)

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	// Stop plugins before closing the vault so any final commits can complete.
	s.pluginMgr.Stop()

	if err := s.vault.Close(); err != nil {
		s.logger.Warn("vault close error", "err", err)
	}
	s.logger.Info("server stopped")
	return nil
}
