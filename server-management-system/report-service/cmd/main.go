package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"

	"github.com/vcs-sms/report-service/config"
	"github.com/vcs-sms/report-service/internal/database"
	"github.com/vcs-sms/report-service/internal/email"
	"github.com/vcs-sms/report-service/internal/handler"
	"github.com/vcs-sms/report-service/internal/repository"
	"github.com/vcs-sms/report-service/internal/scheduler"
	"github.com/vcs-sms/report-service/internal/service"
	"github.com/vcs-sms/shared/logger"
	"github.com/vcs-sms/shared/middleware"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Init logger
	log := logger.NewLogger(cfg.App.Name, &logger.LogConfig{
		Level:      cfg.Log.Level,
		Dir:        cfg.Log.Dir,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	})

	log.Info().Msg("Starting report-service...")

	// 3. Connect PostgreSQL
	db := database.Connect(cfg.ReportDB)

	// 4. Connect Redis
	rdb := database.ConnectRedis(cfg.Redis)

	// 5. Connect Elasticsearch
	esAddrs := strings.Split(cfg.ES.Addresses, ",")
	for i, addr := range esAddrs {
		esAddrs[i] = strings.TrimSpace(addr)
		if !strings.HasPrefix(esAddrs[i], "http") {
			esAddrs[i] = "http://" + esAddrs[i] + ":9200"
		}
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: esAddrs,
		Username:  cfg.ES.Username,
		Password:  cfg.ES.Password,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Elasticsearch client")
	}

	// 6. Init repositories
	esUptimeRepo := repository.NewESUptimeRepo(esClient, cfg.ES.IndexName)
	serverCounterRepo := repository.NewServerCounterRepo(db)
	reportJobRepo := repository.NewReportJobRepo(db)
	snapshotRepo := repository.NewDailySnapshotRepo(db)

	// 7. Init email sender
	emailSender := email.NewGmailSender(cfg.SMTP)

	// 8. Init service
	reportSvc := service.NewReportService(
		esUptimeRepo,
		serverCounterRepo,
		reportJobRepo,
		snapshotRepo,
		emailSender,
		rdb,
		cfg.SMTP,
		log,
	)

	// 9. Init handler
	reportHandler := handler.NewReportHandler(reportSvc)

	// 10. Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": cfg.App.Name,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Report routes
	api := router.Group("/api/v1")
	{
		api.GET("/reports/summary", reportHandler.GetSummary)
		api.POST("/reports", reportHandler.SendReport)
	}

	// 11. Start cron scheduler in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dailyCron := scheduler.NewDailyReportCron(reportSvc, "0 8 * * *", log)
	go dailyCron.Start(ctx)

	// 12. Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Info().Msg("Shutting down report-service...")
		cancel()
	}()

	// 13. Start HTTP server
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	log.Info().Str("addr", addr).Msg("Report service started")

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP server")
	}
}
