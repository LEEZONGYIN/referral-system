package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"referral-system/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := app.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	container, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer func() { _ = container.Close() }()

	server := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: container.Handler(),
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	log.Printf("referral system service started on %s", cfg.Server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
