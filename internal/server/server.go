package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lsmoura/tyui/internal/model"
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
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := s.DB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Database is not available")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		infoRequest := false
		token := r.URL.Path
		for token[len(token)-1] == '/' {
			token = token[:len(token)-1]
		}
		for token[0] == '/' {
			token = token[1:]
		}
		if token[len(token)-1] == '+' {
			infoRequest = true
			token = token[:len(token)-1]
		}

		var link model.Links
		rows, err := s.DB.QueryContext(r.Context(), "SELECT id, token, url, created_at, clicks FROM links WHERE token = $1", token)
		if err != nil {
			s.Logger.Error().Err(err).Msg("error querying database")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Database is not available")
			return
		}
		defer rows.Close()
		if rows.Next() {
			if err := rows.Scan(&link.ID, &link.Token, &link.URL, &link.CreatedAt, &link.Clicks); err != nil {
				s.Logger.Error().Err(err).Msg("error scanning database")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Database is not available")
				return
			}
		}

		if link.ID == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not found")
			return
		}

		if infoRequest {
			fmt.Printf("info for %s\n", token)

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
