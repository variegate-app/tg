package main

import (
	"context"
	"os"
	"syscall"
	"telegram/cmd/api/application"
	"telegram/cmd/api/settings"
	"telegram/internal/graceful"
	"telegram/internal/logger"
	"time"
)

func main() {
	conf := settings.NewConfig()
	conf.WithFlag()
	conf.WithEnv()

	ctx := context.Background()
	l := logger.NewLogger(logger.Debug)
	ctx = l.WithContextFields(ctx,
		logger.Field{Key: "pid", Value: os.Getpid()},
	)
	defer l.Sync()

	l.InfoCtx(ctx, "server running with options", logger.Field{Key: "config", Value: conf})

	server := application.NewServer(l, conf.Address)

	gs := graceful.NewGracefulShutdown(ctx, 1*time.Second)
	gs.AddTask(server)
	err := gs.Wait(syscall.SIGTERM, syscall.SIGINT)

	if err != nil {
		l.ErrorCtx(ctx, "server finished", logger.Field{Key: "error", Value: err})
	}
}
