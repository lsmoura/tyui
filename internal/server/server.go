package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lsmoura/tyui/internal/manager"
	"github.com/lsmoura/tyui/pkg/database"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	db      *database.DB
	Logger  zerolog.Logger
	m       *manager.Manager
	version string
}

func New(version string, db *database.DB, logger zerolog.Logger) *Server {
	return &Server{
		db:      db,
		Logger:  logger,
		m:       manager.New(db),
		version: version,
	}
}

func (s Server) Start(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := s.db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Database is not available")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("tyui-version", s.version)

		infoRequest := false
		token := r.URL.Path
		for len(token) > 0 && token[len(token)-1] == '/' {
			token = token[:len(token)-1]
		}
		for len(token) > 0 && token[0] == '/' {
			token = token[1:]
		}
		if len(token) > 0 && token[len(token)-1] == '+' {
			infoRequest = true
			token = token[:len(token)-1]
		}

		if token == "" {
			fmt.Fprintf(w, "hi there")
			return
		}

		link, err := s.m.LinkWithToken(r.Context(), token)
		if err != nil {
			s.Logger.Error().Err(err).Str("token", token).Msg("error getting link")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Link not found")
			return
		}

		if link.ID == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not found")
			return
		}

		if infoRequest {
			data, err := json.Marshal(link)
			if err != nil {
				s.Logger.Error().Err(err).Msg("error marshalling json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "internal error")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		}

		w.WriteHeader(http.StatusPermanentRedirect)
		w.Header().Set("Location", link.URL)
		fmt.Fprintf(w, link.URL)
		return
	})

	if ctx == nil {
		ctx = context.Background()
	}

	addr := fmt.Sprintf(":%d", port)

	return s.run(ctx, addr, mux)
}

func (s Server) run(ctx context.Context, addr string, mux *http.ServeMux) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	srv.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	s.Logger.Info().Str("addr", addr).Msg("Starting Server")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "http.ListenAndServe")
	}

	s.Logger.Info().Msg("Graceful server shutdown")

	return nil
}
