package routes

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/components"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, ss sessions.Store) http.Handler {
	mux := http.NewServeMux()
	components.Set(db, ss)
	mux.Handle("/login", &components.LoginPage{})
	mux.Handle("GET /logout", &components.LogoutPage{})
	mux.Handle("/signup", &components.SignUpPage{})
	authBlock := http.NewServeMux()
	authBlock.Handle("GET /dashboard", &components.DashboardPage{})
	mux.Handle("/", auth.AuthenticatedOnly(authBlock))
	return mux
}
