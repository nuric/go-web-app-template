package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
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

func (p *SignUpPage) Handle(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	if user.ID != 0 {
		p.redirect = "/dashboard"
		return
	}
	// ---------------------------
	if r.Method == http.MethodGet {
		return
	}
	r.ParseForm()
	if r.PostFormValue("_action") != "signup" {
		p.notFound = true
		return
	}
	// ---------------------------
	if err := DecodeValidForm(p, r); err != nil {
		p.Error = err
		return
	}
	newUser := models.User{
		Email:    p.Email,
		Password: utils.HashPassword(p.Password),
		Role:     "basic", // Default role
	}

	if err := db.Create(&newUser).Error; err != nil {
		slog.Error("could not create user", "error", err)
		p.Error = errors.New("could not create user")
		return
	}

	if err := auth.LogUserIn(w, r, newUser.ID, ss); err != nil {
		slog.Error("could not log user in after signup", "error", err)
		http.Error(w, "could not log user in after signup", http.StatusInternalServerError)
		return
	}

	if err := sendEmailVerification(newUser.ID, newUser.Email); err != nil {
		slog.Error("could not send new user email verification", "error", err)
	}
	slog.Debug("User signed up successfully", "email", p.Email)
	// Redirect to dashboard
	p.redirect = "/dashboard"
}
