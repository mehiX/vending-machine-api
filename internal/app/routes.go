package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"

	_ "github.com/mehiX/vending-machine-api/docs"
	"github.com/mehiX/vending-machine-api/internal/app/model"
	swg "github.com/swaggo/http-swagger"
)

func (a *app) SetupRoutes() {

	if a.Router == nil {
		a.Router = chi.NewMux()
	}

	a.Router.Use(middleware.RequestID)
	a.Router.Use(middleware.RealIP)
	a.Router.Use(middleware.Logger)
	a.Router.Use(middleware.Timeout(60 * time.Second))

	//protected routes
	a.Router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(a.JwtAuth))
		r.Use(jwtauth.Authenticator)
		r.Use(a.UserCtx)
		r.Get("/user", a.handleShowCurrentUser())
		r.Group(func(r chi.Router) {
			r.Use(a.SellerCtx)
			r.Post("/product", a.handleCreateProduct())
			r.Group(func(r chi.Router) {
				r.Use(a.ProductCtx)
				r.Put("/product/{productID:[a-zA-Z0-9-]+}", a.handleUpdateProduct())
				r.Delete("/product/{productID:[a-zA-Z0-9-]+}", a.handleDeleteProduct())
			})
		})
	})

	// public routes
	a.Router.Group(func(r chi.Router) {
		r.Get("/health", a.handleHealth)
		r.Post("/login", a.handleLogin())
		r.Post("/user", a.handleAddUser())
		r.Get("/product/list", a.handleListProducts())
		r.Group(func(r chi.Router) {
			r.Use(a.ProductCtx)
			r.Get("/product/{productID:[a-zA-Z0-9-]+}", a.handleProductDetails())
		})
	})

	a.Router.Mount("/swagger", swg.WrapHandler)
}

// @Summary 	Health endpoing
// @Description Validate the application is running
// @Tags		public
// @Produces	text/plain
// @Success		200 {string} string "OK"
// @Router 		/health [get]
func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (a *app) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, "authentication error", http.StatusUnauthorized)
			return
		}

		username := claims["user"].(string)

		usr, err := a.dbFindOneByUsername(r.Context(), username)
		if err != nil {
			http.Error(w, "authentication error", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", usr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SellerCtx only allows seller accounts to access successive endpoints
// Requires a "user" object in current request context
func (a *app) SellerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usr, ok := r.Context().Value("user").(*model.User)
		if !ok || usr.Role != model.ROLE_SELLER {
			http.Error(w, "authentication error", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *app) ProductCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		productID := chi.URLParam(r, "productID")
		product, err := a.dbFindProductByID(r.Context(), productID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), "product", product)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
