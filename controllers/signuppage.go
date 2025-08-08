package controllers

import (
	"errors"
	"net/http"

	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog/log"
)

type SignUpPage struct {
	BasePage
	Email                string `schema:"email"`
	EmailError           error
	Password             string `schema:"password"`
	PasswordError        error
	ConfirmPassword      string `schema:"confirmPassword"`
	ConfirmPasswordError error
	Error                error
}

func (p *SignUpPage) Validate() bool {
	p.EmailError = ValidateEmail(p.Email)
	p.PasswordError = ValidatePassword(p.Password)
	if p.Password != p.ConfirmPassword {
		p.ConfirmPasswordError = errors.New("passwords do not match")
	}
	return p.EmailError == nil &&
		p.PasswordError == nil &&
		p.ConfirmPasswordError == nil
}

func (p SignUpPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	if user.ID != 0 {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	// ---------------------------
	if r.Method == http.MethodGet {
		render(r, w, &p)
		return
	}
	r.ParseForm()
	if r.PostFormValue("_action") != "signup" {
		// Not our action, ignore
		return
	}
	// ---------------------------
	if err := DecodeValidForm(&p, r); err != nil {
		p.Error = err
		render(r, w, &p)
		return
	}
	newUser := models.User{
		Email:    p.Email,
		Password: utils.HashPassword(p.Password),
		Role:     "basic", // Default role
	}

	if err := db.Create(&newUser).Error; err != nil {
		log.Error().Err(err).Msg("could not create user")
		p.Error = errors.New("could not create user")
		render(r, w, &p)
		return
	}

	if err := auth.LogUserIn(w, r, newUser.ID, ss); err != nil {
		log.Error().Err(err).Msg("could not log user in after signup")
		http.Error(w, "could not log user in after signup", http.StatusInternalServerError)
		return
	}

	if err := sendEmailVerification(newUser.ID, newUser.Email); err != nil {
		log.Error().Err(err).Msg("could not send new user email verification")
	}
	log.Debug().Str("email", p.Email).Msg("User signed up successfully")
	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
