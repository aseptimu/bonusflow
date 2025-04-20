package server

import (
	"database/sql"
	"github.com/aseptimu/internal/middlewares"
	"github.com/aseptimu/internal/migrations"
	"github.com/aseptimu/internal/repository"
	"github.com/aseptimu/internal/services"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"

	"github.com/aseptimu/internal/config"
	"github.com/aseptimu/internal/handlers"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Serve() {
	conf := config.NewConfig()

	migrations.RunMigrations(conf.DSN)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	slog.Info("Connecting db", "dsn", conf.DSN)
	db, err := sql.Open("pgx", conf.DSN)
	if err != nil {
		slog.Error("Failed to open DB", "error", err)
		return
	}
	defer db.Close()

	userRepo := repository.NewPostgresUserRepository(db)
	userService := services.NewUserService(userRepo, conf)
	userHandler := handlers.NewUserHandler(userService)

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, conf.AccrualSystemAddress)
	orderHandler := handlers.NewOrderHandler(orderService)

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", userHandler.RegisterUser)
		r.Post("/api/user/login", userHandler.LoginUser)
	})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWTAuthMiddleware(conf.SecretKey))

		r.Route("/api/user", func(r chi.Router) {
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orderHandler.UploadOrder)
				r.Get("/", orderHandler.GetUserOrders)
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", handlers.GetUserBalance)
				r.Post("/withdraw", handlers.WithdrawBalance)
			})

			r.Get("/withdrawals", handlers.GetUserWithdrawals)
		})
	})

	slog.Info("Starting server", "address", "http://"+conf.ServerAddress)
	slog.Info("Starting accrual server", "address", "http://"+conf.AccrualSystemAddress)
	if err := http.ListenAndServe(conf.ServerAddress, r); err != nil {
		slog.Error("Server failed", "error", err)
	}
}
