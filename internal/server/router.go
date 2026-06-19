package server

import (
	"log/slog"
	"net/http"

	"github.com/hardhacker/vaultr/internal/agent"
	"github.com/hardhacker/vaultr/internal/config"
	"github.com/hardhacker/vaultr/internal/mate"
	"github.com/hardhacker/vaultr/internal/plugins/compile"
	"github.com/hardhacker/vaultr/internal/plugins/gitsync"
	"github.com/hardhacker/vaultr/internal/plugins/search"
	"github.com/hardhacker/vaultr/internal/server/handler"
	"github.com/hardhacker/vaultr/internal/server/middleware"
	"github.com/hardhacker/vaultr/internal/server/view"
	"github.com/hardhacker/vaultr/internal/skills"
	"github.com/hardhacker/vaultr/internal/storage"
)

// newRouter builds and returns the application HTTP router.
func newRouter(
	logger *slog.Logger,
	cfg *config.Config,
	vault *storage.Vault,
	cfgFileLoaded string,
	gitPlugin *gitsync.Plugin,
	searchPlugin *search.Plugin,
	compilePlugin *compile.Plugin,
	mateStore *mate.Store,
	mateRunner *mate.Runner,
	skillsMgr *skills.Manager,
	apiKey string) http.Handler {
	mux := http.NewServeMux()

	h := handler.New(logger, cfg)
	cfgHTTP := handler.NewConfigHTTP(logger, cfg, cfgFileLoaded)
	gh := handler.NewVault(vault)
	sh := search.NewHandler(searchPlugin)
	vh := view.NewView(vault, searchPlugin, cfg)

	agentHub := agent.NewHub()
	agentCache := agent.NewAgentCache(0, nil)
	agentCache.WarmUp()
	ah := handler.NewAgentAPI(logger, cfg, vault, agentHub, mateStore, agentCache)

	// Wire mate runner → agent API so trigger runs use the same execution path.
	if mateRunner != nil {
		mateRunner.SetRunFunc(ah.FireTriggerRun)
	}

	mux.HandleFunc("GET /api/agents", ah.AgentsGET)
	mux.HandleFunc("POST /api/chat", ah.ChatPOST)
	mux.HandleFunc("POST /api/runs", ah.RunsPOST)
	mux.HandleFunc("GET /api/runs/active", ah.RunsActiveGET)
	mux.HandleFunc("GET /api/runs/by-ref", ah.RunByRefGET)
	mux.HandleFunc("GET /api/runs/{id}/events", ah.RunEventsGET)
	mux.HandleFunc("GET /api/runs/{id}", ah.RunGET)
	mux.HandleFunc("POST /api/runs/{id}/cancel", ah.RunCancelPOST)

	if mateStore != nil {
		ch := handler.NewConversationAPI(mateStore)
		mux.HandleFunc("GET /api/conversations", ch.ConversationsGET)
		mux.HandleFunc("POST /api/conversations", ch.ConversationsPOST)
		mux.HandleFunc("GET /api/conversations/{id}", ch.ConversationGET)
		mux.HandleFunc("DELETE /api/conversations/{id}", ch.ConversationDELETE)
		mux.HandleFunc("PATCH /api/conversations/{id}/title", ch.ConversationTitlePATCH)

		mh := handler.NewMateAPI(logger, mateStore)
		mux.HandleFunc("GET /api/mate-events", mh.MateEventsGET)
		mux.HandleFunc("GET /api/mates", mh.MatesGET)
		mux.HandleFunc("POST /api/mates/reorder", mh.MatesReorderPOST)
		mux.HandleFunc("POST /api/mates", mh.MatesPOST)
		mux.HandleFunc("GET /api/mates/{id}", mh.MateGET)
		mux.HandleFunc("PUT /api/mates/{id}", mh.MatePUT)
		mux.HandleFunc("DELETE /api/mates/{id}", mh.MateDELETE)
	}

	skh := handler.NewSkillsHTTP(skillsMgr)
	mux.HandleFunc("GET /api/skills", skh.List)
	mux.HandleFunc("POST /api/skills/{name}/enable", skh.Enable)
	mux.HandleFunc("POST /api/skills/{name}/disable", skh.Disable)

	// System routes
	mux.HandleFunc("POST /healthz", h.HealthCheck)
	mux.HandleFunc("POST /version", h.Version)
mux.HandleFunc("POST /api/status", handler.NewStatus(vault, searchPlugin).Status)

	mux.HandleFunc("GET /api/config", cfgHTTP.Get)
	mux.HandleFunc("GET /api/config/schema", cfgHTTP.Schema)
	mux.HandleFunc("PATCH /api/config", cfgHTTP.Patch)

	wxHTTP := handler.NewWechatHTTP(logger, cfg, cfgFileLoaded)
	mux.HandleFunc("GET /api/wechat/status", wxHTTP.Status)
	mux.HandleFunc("POST /api/wechat/login/start", wxHTTP.LoginStart)
	mux.HandleFunc("GET /api/wechat/login/status", wxHTTP.LoginStatus)
	mux.HandleFunc("POST /api/wechat/logout", wxHTTP.Logout)

	// Website routes
	mux.HandleFunc("GET /notes/search", vh.SearchFragment)
	mux.HandleFunc("GET /library", vh.Library)
	mux.HandleFunc("GET /home", vh.Home)
	mux.HandleFunc("GET /home/refresh", vh.HomeRefresh)
	mux.HandleFunc("GET /library/refresh", vh.LibraryRefresh)
	mux.HandleFunc("GET /dir", vh.Dir)
	mux.HandleFunc("GET /dir/notes", vh.DirNotes)
	mux.HandleFunc("GET /dir/refresh", vh.DirRefresh)
	mux.HandleFunc("GET /folders", vh.Folders)
	mux.HandleFunc("GET /shorts", vh.Shorts)
	mux.HandleFunc("GET /shorts/day", vh.ShortsDay)
	mux.HandleFunc("GET /shorts/calendar", vh.ShortsCalendar)
	mux.HandleFunc("GET /shorts/stream", vh.ShortsStream)
	mux.HandleFunc("GET /images", vh.Images)
	mux.HandleFunc("GET /images/grid", vh.ImagesGrid)
	mux.HandleFunc("GET /agent", vh.AgentChat)
	mux.HandleFunc("GET /library/notes", vh.LibraryNotes)
	mux.HandleFunc("GET /library/tag", vh.LibraryTag)
	mux.HandleFunc("GET /library/index/select", vh.LibraryIndexSelect)
	mux.HandleFunc("GET /library/focus", vh.LibraryFocus)
	mux.HandleFunc("GET /library/unfocus", vh.LibraryUnfocus)
	mux.HandleFunc("GET /graph", vh.KnowledgeGraph)
	mux.HandleFunc("GET /api/graph/data", vh.KnowledgeGraphData)
	mux.HandleFunc("POST /api/graph/rebuild", vh.KnowledgeGraphRebuild)
	mux.HandleFunc("GET /notes/fragment", vh.NoteFragment)

	mux.Handle("POST /api/notes/resolve", handler.NewNoteResolve(vault))

	// Vault REST
	mux.HandleFunc("POST /api/vault/read", gh.Read)
	mux.HandleFunc("POST /api/vault/stat", gh.Stat)
	mux.HandleFunc("POST /api/vault/list", gh.List)
	mux.HandleFunc("POST /api/vault/list-dirs", gh.ListDirs)
	mux.HandleFunc("POST /api/vault/write", gh.Write)
	mux.HandleFunc("POST /api/vault/delete", gh.Delete)
	mux.HandleFunc("POST /api/vault/upload-image", gh.UploadImage)
	mux.HandleFunc("POST /api/vault/pin", gh.Pin)
mux.HandleFunc("POST /api/vault/shorts", gh.Short)
	mux.HandleFunc("POST /api/vault/shorts/list", gh.ShortList)
	mux.HandleFunc("GET /_assets/", gh.ServeAsset)
	mux.HandleFunc("GET /api/images/serve", gh.ServeImageByName)
	mux.HandleFunc("GET /api/images/at", gh.ServeImageAt)
	mux.HandleFunc("POST /api/images/delete", gh.DeleteGalleryImage)
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandler()))

	// Plugin routes
	mux.HandleFunc("POST /api/search", sh.Search)
	mux.HandleFunc("POST /api/tag/list", sh.TagList)
	mux.HandleFunc("POST /api/tag/count", sh.TagCount)
	mux.HandleFunc("POST /api/tag/delete", sh.TagDelete)

	if gitPlugin != nil {
		gsh := gitsync.NewHandler(gitPlugin)
		mux.HandleFunc("POST /api/git/sync", gsh.Sync)
	}
	if compilePlugin != nil {
		dh := compile.NewHandler(vault, compilePlugin)
		mux.HandleFunc("POST /api/compile/trigger", dh.Trigger)
	}

	return middleware.Chain(mux,
		middleware.Recoverer(logger),
		middleware.Logger(logger),
		middleware.Authenticator(apiKey, logger),
	)
}
