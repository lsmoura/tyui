package server

import (
	"context"
	"fmt"
	"github.com/lsmoura/tyui/pkg/database"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	DB     *database.DB
	Logger zerolog.Logger
}

func (s Server) Start(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "tyui.me")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := s.DB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Database is not available")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
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
