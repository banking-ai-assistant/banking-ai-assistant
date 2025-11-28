package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"banking-ai-assistant/internal/ai"
	"banking-ai-assistant/internal/db"
	"banking-ai-assistant/internal/logger"
	"banking-ai-assistant/internal/server"
	"banking-ai-assistant/internal/tickets"
	"banking-ai-assistant/pkg/utils"
)

func main() {
	dsn := utils.MustEnv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/correspondence?sslmode=disable")
	port := utils.MustEnv("PORT", "8080")

	appLog := logger.NewStdLogger(os.Stdout, "debug")

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		appLog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	for i := 0; i < 10; i++ {
		if err := dbConn.Ping(); err != nil {
			appLog.Info("Waiting for DB to be ready...", "attempt", i+1)
			time.Sleep(1 * time.Second)
			continue
		}
		appLog.Info("DB is ready")
		break
	}

	repos := db.NewRepositories(dbConn)
	aiClient := ai.NewStubClient()
	notifier := tickets.NewStubNotifier(appLog)

	service := tickets.NewService(repos.Tickets, aiClient, notifier, appLog)

	router := server.NewRouter(service, appLog)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
	}

	ctxWorkers, cancelWorkers := context.WithCancel(context.Background())
	service.StartWorkers(ctxWorkers)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		appLog.Info("server starting", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	appLog.Info("shutdown signal received")

	service.StopWorkers()
	cancelWorkers()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		appLog.Error("server forced to shutdown", "error", err)
	}

	if err := dbConn.Close(); err != nil {
		appLog.Error("failed to close database", "error", err)
	}

	appLog.Info("server exited properly")
}
