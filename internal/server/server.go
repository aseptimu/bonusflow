package server

import (
	"github.com/aseptimu/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func Serve() {
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
	http.ListenAndServe(":3000", r)
}
