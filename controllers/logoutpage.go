package controllers

import (
	"log/slog"
	"net/http"

	"github.com/nuric/go-api-template/auth"
)

type LogoutPage struct {
}

func (p LogoutPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := auth.LogUserOut(w, r, ss); err != nil {
		slog.Error("could not log user out", "error", err)
		http.Error(w, "could not log user out, please try again", http.StatusInternalServerError)
		return
	}
	slog.Debug("User logged out successfully")
	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
