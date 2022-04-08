package main

import (
	"context"
	"github.com/lsmoura/env"
	"github.com/lsmoura/tyui/internal/server"
	"github.com/lsmoura/tyui/pkg/database"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
)

const (
	envDBHost = "DB_HOST"
	envDBPort = "DB_PORT"
	envDBUser = "DB_USER"
	envDBPass = "DB_PASSWORD"
	envDBName = "DB_DATABASE"
	envDBSSL  = "DB_SSL"

	envPort = "PORT"
)

func dbConfig() (database.Params, error) {
	var dbParams database.Params
	var err error
	dbParams.Host, err = env.Str(envDBHost, "")
	if err != nil {
		return dbParams, errors.Wrap(err, envDBHost)
	}
	dbParams.Port, err = env.Int(envDBPort, 0)
	if err != nil {
		return dbParams, errors.Wrap(err, envDBPort)
	}
	dbParams.User, err = env.Str(envDBUser, "")
	if err != nil {
		return dbParams, errors.Wrap(err, envDBUser)
	}
	dbParams.Password, err = env.Str(envDBPass, "")
	if err != nil {
		return dbParams, errors.Wrap(err, envDBPass)
	}
	dbParams.Database, err = env.Str(envDBName, "")
	if err != nil {
		return dbParams, errors.Wrap(err, envDBName)
	}
	dbParams.SSL, err = env.Bool(envDBSSL, false)
	if err != nil {
		return dbParams, errors.Wrap(err, envDBSSL)
	}

	return dbParams, nil
}

func initDB() (*database.DB, error) {
	// read config
	dbParams, err := dbConfig()
	if err != nil {
		return nil, errors.Wrap(err, "dbConfig")
	}

	db, err := database.New(dbParams)
	if err != nil {
		return nil, errors.Wrap(err, "database.New")
	}

	return db, nil
}

func signalListen(cancel context.CancelFunc) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	cancel()
}

func main() {
	logger := Logger()
	logger.Info().Msg("tyui.me")

	db, err := initDB()
	if err != nil {
		logger.Fatal().Err(err).Msg("initDB")
	}
	defer db.Close()

	// start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go signalListen(cancel)

	port, err := env.Int(envPort, 13000)
	if err != nil {
		logger.Fatal().Err(err).Msg("env.Int")
	}
	s := server.New(db, logger)
	if err := s.Start(ctx, port); err != nil {
		logger.Fatal().Err(err).Msg("server.Start")
	}
}
