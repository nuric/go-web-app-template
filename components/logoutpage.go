package components

import (
	"net/http"

	"github.com/nuric/go-api-template/auth"
	"github.com/rs/zerolog/log"
)

type LogoutPage struct {
}

func (p LogoutPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := auth.LogUserOut(w, r, ss); err != nil {
		log.Error().Err(err).Msg("could not log user out")
		http.Error(w, "could not log user out, please try again", http.StatusInternalServerError)
		return
	}
	log.Debug().Msg("User logged out successfully")
	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
