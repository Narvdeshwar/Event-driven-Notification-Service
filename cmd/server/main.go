package main

import (
	"context"
	"event-driven-notification-service/internal/api"
	"event-driven-notification-service/internal/config"
	"event-driven-notification-service/internal/metrics"
	"event-driven-notification-service/internal/migrations"
	"event-driven-notification-service/internal/model"
	"event-driven-notification-service/internal/notifier"
	"event-driven-notification-service/internal/service"
	"event-driven-notification-service/internal/store"
	"event-driven-notification-service/internal/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Load()

	// Register metrics
	metrics.Register()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Connect to database
	db, err := store.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatal("Could not connect to database after retries:", err)
	}
	defer db.Close()
	log.Println("Before the migrations start")

	// run migrations after db is connected
	migrations.Run(db)

	// Create repository
	repo := store.NewPostgresRepo(db)
	// Create service (inject repo)
	svc := service.New(repo)
	// Create handler (inject service)
	handler := api.New(svc)

	// Setup router
	router := gin.Default()
	api.RegisterRoutes(router, handler)

	// jobqueue
	jobQueue := make(chan model.Notification, 100)
	poller := worker.NewPoller(repo, jobQueue, 10, 5*time.Second)
	go poller.Start(ctx)
	log.Println("Logger goroutine is created")

	emailNotifier:=notifier.NewEmailNotifier()

	for i := 0; i < 5; i++ {
		w := worker.NewWorker(i, jobQueue, repo, emailNotifier)
		go w.Start(ctx)
	}

	// Start HTTP server
	log.Println("Server running on port", cfg.HTTPPort)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatal("server failed:", err)
	}
}
