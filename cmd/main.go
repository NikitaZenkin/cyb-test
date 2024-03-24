package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pressly/goose"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"cyb-test/docs"
	httpsrv "cyb-test/internal/http"
	"cyb-test/internal/repository"
	"cyb-test/internal/service"
	"cyb-test/pkg/dns"
)

// @title Cyb-test
func main() {
	config, err := Configure()
	if err != nil {
		panic(err)
	}

	docs.SwaggerInfo.BasePath = config.HTTPServer.BasePath
	docs.SwaggerInfo.Host = config.HTTPServer.Host + ":" + config.HTTPServer.Port

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("create logger")
	}

	defer logger.Sync()

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	db, err := repository.NewDB(ctx, config.DSN)
	if err != nil {
		logger.Fatal("connect db", zap.Error(err))
	}

	if err = goose.Up(db.DB, config.MigrationsPath); err != nil {
		logger.Fatal("run migrations", zap.Error(err))
	}

	rep := repository.New(db, logger)
	dnsServer := dns.New(config.DSNServerAddr)
	srv := service.New(ctx, logger, rep, dnsServer)

	httpController, err := httpsrv.New(logger, srv, config.HTTPServer.Port, config.HTTPServer.BasePath)
	if err != nil {
		logger.Fatal("create http service", zap.Error(err))
	}

	group := errgroup.Group{}

	group.Go(func() error {
		logger.Info("external http started")
		defer logger.Info("external http stopped")

		err = httpController.ListenAndServe()
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			return err
		}

		return nil
	})

	group.Go(func() error {
		sgnl := make(chan os.Signal, 1)
		signal.Notify(sgnl,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		stop := <-sgnl

		logger.Info("waiting for all processes to stop by signal", zap.Any("signal", stop))
		httpController.Close()

		return nil
	})

	logger.Info("application started")

	if err = group.Wait(); err != nil {
		logger.Fatal("group wait", zap.Error(err))
	}
}
