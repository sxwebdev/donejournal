package api

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sxwebdev/donejournal"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/auth/v1/authv1connect"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/inbox/v1/inboxv1connect"
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

	// Mount Connect-RPC handlers under /api/v1
	const apiPrefix = "/api/v1"
	mux.Handle(apiPrefix+authPath, http.StripPrefix(apiPrefix, authHandler))
	mux.Handle(apiPrefix+inboxPath, http.StripPrefix(apiPrefix, inboxHandler))
	mux.Handle(apiPrefix+todosPath, http.StripPrefix(apiPrefix, todosHandler))

	// Serve frontend SPA from embedded filesystem
	frontendFS, err := fs.Sub(donejournal.FrontendFS, "frontend/dist")
	if err != nil {
		l.Warnf("failed to load frontend assets: %v", err)
	} else {
		mux.Handle("/", spaHandler(http.FS(frontendFS)))
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
func spaHandler(filesystem http.FileSystem) http.Handler {
	fileServer := http.FileServer(filesystem)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to open the file
		f, err := filesystem.Open(path)
		if err != nil {
			// File not found — serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}
