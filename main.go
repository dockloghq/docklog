package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"docklog/audit"
	appcli "docklog/cli"
	"docklog/config"
	"docklog/containers"
	"docklog/db"
	"docklog/middleware"
	"docklog/server"
	"docklog/services"
	"docklog/stats"
)

func main() {
	exit, code, appRuntime := appcli.Dispatch(os.Args)
	if exit {
		os.Exit(code)
	}

	appcli.LogRunMode(appRuntime)

	config.LoadDebugMode()
	config.LoadRuntimeConfig()
	config.LoadAuthDisabled()
	config.InitSecretKey()
	config.LoadActionFlags()
	middleware.InitClientAccess()
	middleware.InitWSUpgrader()
	containers.Init()

	if err := db.InitDB(db.ServerPath(config.AuthDisabled)); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	initNotificationModule()
	services.StartVersionUpdateMonitor(appcli.Version, notificationService)
	audit.OnLogged = dispatchAuditNotification
	initAlertModule()

	srv := server.New(server.Deps{
		Notifications: notificationService,
		Alerts:        alertEngine,
	})

	// Setup graceful shutdown listener to persist in-memory ringbuffer before exit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		log.Println("Shutting down DockLog gracefully...")

		// Snapshot in-memory ringbuffer metrics to disk
		if err := stats.SaveSnapshotToFile(); err != nil {
			log.Printf("Failed to save metrics snapshot: %v", err)
		} else {
			log.Println("Metrics ringbuffer successfully snapshotted to disk")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Echo().Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := srv.Run(appRuntime); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
