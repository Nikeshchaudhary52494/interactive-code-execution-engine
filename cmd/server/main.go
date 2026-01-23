package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"execution-engine/internal/api"
	"execution-engine/internal/engine"
	"execution-engine/internal/executor"
)

func main() {
	// ---- bootstrap docker executor ----
	dockerExec, err := executor.NewDockerExecutor()
	if err != nil {
		panic(err)
	}

	// ---- preload docker images ----
	// Scope the context for preloading
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		if err := dockerExec.PreloadImages(ctx); err != nil {
			cancel()
			log.Fatalf("❌ failed to preload images: %v", err)
		}
		cancel()
	}

	// ---- engine ----
	eng := engine.New(dockerExec)

	// ---- router ----
	r := api.New(eng)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Println("🚀 Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 1. Shutdown HTTP server first (stop accepting new requests)
	// Give it 5 seconds to finish current HTTP requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	// 2. Shutdown Engine (wait for running code to finish)
	// Give it 5 minutes to drain pending jobs
	log.Println("Waiting for active sessions to finish...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer shutdownCancel()

	if err := eng.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Engine forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
