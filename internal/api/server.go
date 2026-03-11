package api

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sxwebdev/donejournal"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/auth/v1/authv1connect"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/inbox/v1/inboxv1connect"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/notes/v1/notesv1connect"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/todos/v1/todosv1connect"
	"github.com/sxwebdev/donejournal/internal/config"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/internal/tmanager"
	"github.com/sxwebdev/tokenmanager"
	"github.com/tkcrm/mx/logger"
)

type Server struct {
	logger     logger.Logger
	config     *config.Config
	httpServer *http.Server
}

func New(
	l logger.Logger,
	conf *config.Config,
	baseService *baseservices.BaseServices,
	st *store.Store,
	taskManager *tmanager.Manager,
	tokenMgr *tokenmanager.Manager[TokenData],
) *Server {
	mux := http.NewServeMux()

	// Connect-RPC interceptors
	interceptors := connect.WithInterceptors(newAuthInterceptor(tokenMgr))

	// Auth service
	authPath, authHandler := authv1connect.NewAuthServiceHandler(
		NewAuthHandler(l, conf, tokenMgr),
		interceptors,
	)

	// Inbox service
	inboxPath, inboxHandler := inboxv1connect.NewInboxServiceHandler(
		NewInboxHandler(l, baseService, st),
		interceptors,
	)

	// Todos service
	todosPath, todosHandler := todosv1connect.NewTodoServiceHandler(
		NewTodosHandler(l, baseService, st),
		interceptors,
	)

	// Notes service
	notesPath, notesHandler := notesv1connect.NewNoteServiceHandler(
		NewNotesHandler(l, baseService, st),
		interceptors,
	)

	// Mount Connect-RPC handlers under /api/v1
	const apiPrefix = "/api/v1"
	mux.Handle(apiPrefix+authPath, http.StripPrefix(apiPrefix, authHandler))
	mux.Handle(apiPrefix+inboxPath, http.StripPrefix(apiPrefix, inboxHandler))
	mux.Handle(apiPrefix+todosPath, http.StripPrefix(apiPrefix, todosHandler))
	mux.Handle(apiPrefix+notesPath, http.StripPrefix(apiPrefix, notesHandler))

	// Serve frontend SPA from embedded filesystem
	frontendFS, err := fs.Sub(donejournal.FrontendFS, "frontend/dist")
	if err != nil {
		l.Warnf("failed to load frontend assets: %v", err)
	} else {
		mux.Handle("/", spaHandler(http.FS(frontendFS), conf))
	}

	s := &Server{
		logger: l,
		config: conf,
		httpServer: &http.Server{
			Addr:    conf.Server.Addr,
			Handler: mux,
		},
	}

	return s
}

// Name returns the service name
func (s *Server) Name() string {
	return "api"
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	var url string
	if strings.HasPrefix(s.config.Server.Addr, ":") {
		url = "http://localhost" + s.config.Server.Addr
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Infof("Starting server on %s", url)
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// spaHandler serves static files and falls back to index.html for SPA routing
func spaHandler(filesystem http.FileSystem, conf *config.Config) http.Handler {
	fileServer := http.FileServer(filesystem)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to open the file
		f, err := filesystem.Open(path)
		if err != nil {
			// File not found — serve index.html for SPA routing
			serveIndexHTML(w, filesystem, conf)
			return
		}
		f.Close()

		if path == "/" || path == "/index.html" {
			serveIndexHTML(w, filesystem, conf)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}

func serveIndexHTML(w http.ResponseWriter, filesystem http.FileSystem, conf *config.Config) {
	f, err := filesystem.Open("/index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "failed to read index.html", http.StatusInternalServerError)
		return
	}

	script := fmt.Sprintf(
		`<script>window.__ENV__={"telegramBotUsername":%q}</script>`,
		conf.Telegram.BotUsername,
	)
	injected := strings.Replace(string(content), "</head>", script+"</head>", 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(injected))
}
