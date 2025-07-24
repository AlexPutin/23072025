package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexputin/downloader/config"
	"github.com/alexputin/downloader/internal/downloader"
	"github.com/alexputin/downloader/internal/handler"
	"github.com/alexputin/downloader/internal/task"
	"github.com/labstack/echo/v4"
)

func gracefulShutdown(app *echo.Echo, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true
}

func main() {
	config := config.MustLoad("config.yaml")
	taskRepo := task.NewInMemoryTaskRepository()
	downloadService := downloader.NewTaskDownloader(taskRepo, config)
	taskService := task.NewTaskService(taskRepo, downloadService, config)
	handler := handler.NewApiHandler(*taskService, config)

	app := echo.New()
	{
		app.Server.IdleTimeout = config.Server.IdleTimeout
		app.Server.ReadTimeout = config.Server.ReadTimeout
		app.Server.WriteTimeout = config.Server.WriteTimeout
	}
	handler.RegisterRoutes(app)

	done := make(chan bool, 1)
	go gracefulShutdown(app, done)

	err := app.Start(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port))
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
