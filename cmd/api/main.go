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

	"github.com/luanjr10/shoptech-go-api/internal/config"
	"github.com/luanjr10/shoptech-go-api/internal/database"
	"github.com/luanjr10/shoptech-go-api/internal/router"
	"github.com/luanjr10/shoptech-go-api/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DBDSN)
	if err != nil {
		log.Fatalf("không kết nối được database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate thất bại: %v", err)
	}
	if err := service.NewAuthService(db, cfg).EnsureAdmin(cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("tạo tài khoản admin thất bại: %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router.New(db, cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server đang chạy tại http://localhost:%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server lỗi: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown lỗi: %v", err)
	}
	log.Println("server đã dừng")
}
