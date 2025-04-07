package server

import (
	"github.com/aseptimu/internal/config"
	"github.com/aseptimu/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
)

func Serve() {
	conf := config.NewConfig()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.RegisterUser)
		r.Post("/login", handlers.LoginUser)

		r.Route("/orders", func(r chi.Router) {
			r.Post("/", handlers.UploadOrder)
			r.Get("/", handlers.GetUserOrders)
		})

		r.Route("/balance", func(r chi.Router) {
			r.Get("/", handlers.GetUserBalance)
			r.Post("/withdraw", handlers.WithdrawBalance)
		})

		r.Get("/withdrawals", handlers.GetUserWithdrawals)
	})

	slog.Info("Starting server", "address", conf.ServerAddress)
	if err := http.ListenAndServe(conf.ServerAddress, r); err != nil {
		slog.Error("Server failed", "error", err)
	}
}
