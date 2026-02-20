package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/heth/STM/internal/config"
	"github.com/heth/STM/internal/controller"
	"github.com/heth/STM/internal/middleware"
	"github.com/heth/STM/internal/model"
	"github.com/heth/STM/internal/repository"
	"github.com/heth/STM/internal/router"
	"github.com/heth/STM/internal/service"
	"github.com/heth/STM/internal/utils"
	"github.com/heth/STM/proto"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		slog.Error("JWT_SECRET is required")
		os.Exit(1)
	}

	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	slog.SetDefault(logger)

	// Ensure data directory exists for SQLite
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	// Database
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Task{}, &model.RefreshToken{}); err != nil {
		slog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// JWT & Services
	jwtService := utils.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiry, cfg.RefreshExpiry)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)
	taskService := service.NewTaskService(taskRepo)

	// Real-time task notifications (gRPC)
	broadcaster := service.NewTaskEventBroadcaster()
	taskService.SetTaskNotifier(broadcaster)
	grpcNotificationSrv := service.NewNotificationGrpcServer(broadcaster)

	// Controllers
	authCtrl := controller.NewAuthController(authService)
	userCtrl := controller.NewUserController(userRepo)
	taskCtrl := controller.NewTaskController(taskService)
	adminCtrl := controller.NewAdminController(taskService)

	// Router
	r := router.Setup(cfg, authCtrl, userCtrl, taskCtrl, adminCtrl, jwtService)

	// gRPC server (streaming task notifications)
	grpcLis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("gRPC listen failed", "error", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(middleware.AuthStreamInterceptor(jwtService)),
	)
	proto.RegisterNotificationServiceServer(grpcServer, grpcNotificationSrv)
	go func() {
		slog.Info("gRPC server listening", "addr", ":50051")
		if err := grpcServer.Serve(grpcLis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("starting HTTP server", "addr", addr)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}