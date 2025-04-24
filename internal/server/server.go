package server

import (
	"context"
	"database/sql"
	"github.com/aseptimu/internal/config"
	"github.com/aseptimu/internal/handlers"
	"github.com/aseptimu/internal/middlewares"
	"github.com/aseptimu/internal/migrations"
	"github.com/aseptimu/internal/repository"
	"github.com/aseptimu/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func Serve() error {
	conf := config.NewConfig()

	migrations.RunMigrations(conf.DSN)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	slog.Info("Connecting db", "dsn", conf.DSN)
	db, err := sql.Open("pgx", conf.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo, conf)
	userHandler := handlers.NewUserHandler(userService)

	orderRepo := repository.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, conf.AccrualSystemAddress)
	orderHandler := handlers.NewOrderHandler(orderService)

	balanceRepo := repository.NewBalanceRepository(db)
	balanceService := services.NewBalanceService(balanceRepo)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	worker := services.NewAccrualWorker(orderRepo, orderService, 10*time.Second)
	go worker.Start(ctx)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", userHandler.RegisterUser)
		r.Post("/login", userHandler.LoginUser)

		r.Group(func(r chi.Router) {
			r.Use(middlewares.JWTAuthMiddleware(conf.SecretKey))

			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orderHandler.UploadOrder)
				r.Get("/", orderHandler.GetUserOrders)
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balanceHandler.GetUserBalance)
				r.Post("/withdraw", balanceHandler.WithdrawBalance)
			})

			r.Get("/withdrawals", balanceHandler.GetUserWithdrawals)
		})
	})

	srv := &http.Server{
		Addr:    conf.ServerAddress,
		Handler: r,
	}

	go func() {
		slog.Info("Starting server", "address", "http://"+conf.ServerAddress)
		slog.Info("Starting accrual server", "address", "http://"+conf.AccrualSystemAddress)
		if err := http.ListenAndServe(conf.ServerAddress, r); err != nil {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutdown signal received, stopping...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}
	slog.Info("Server stopped gracefully")

	return nil
}
