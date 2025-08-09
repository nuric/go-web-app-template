package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog/log"
)

type LoginPage struct {
	BasePage
	LoginForm          LoginForm
	ForgotPasswordForm ForgotPasswordForm
}

func (p *LoginPage) Handle(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	switch {
	case user.ID != 0 && user.EmailVerified:
		p.redirect = "/dashboard"
		return
	case user.ID != 0 && !user.EmailVerified:
		p.redirect = "/verify-email"
		return
	}
	// ---------------------------
	if r.Method == http.MethodGet {
		return
	}
	// ---------------------------
	r.ParseForm()
	switch r.PostFormValue("_action") {
	case "login":
		f := &p.LoginForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			return
		}
		var user models.User
		if err := db.Where("email = ?", f.Email).First(&user).Error; err != nil {
			log.Debug().Err(err).Uint("userId", user.ID).Msg("could not find user")
			f.Error = errors.New("invalid email or password")
			return
		}
		if !utils.VerifyPassword(user.Password, f.Password) {
			log.Debug().Uint("userId", user.ID).Msg("password verification failed")
			f.Error = errors.New("invalid email or password")
			return
		}
		if err := auth.LogUserIn(w, r, user.ID, ss); err != nil {
			log.Error().Err(err).Uint("userId", user.ID).Msg("could not log user in")
			f.Error = errors.New("could not log user in")
			return
		}
		log.Debug().Uint("userId", user.ID).Str("email", f.Email).Msg("User logged in successfully")
		// Redirect to dashboard
		p.redirect = "/dashboard"
	case "forgot_password":
		f := &p.ForgotPasswordForm
		f.DialogOpen = true
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			return
		}
		resetToken := models.Token{
			Email:     f.Email,
			Token:     utils.HumanFriendlyToken(),
			Purpose:   "reset_password",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		if err := db.Create(&resetToken).Error; err != nil {
			log.Error().Err(err).Msg("could not create password reset token")
			f.Error = errors.New("could not send password reset email")
			return
		}
		emailData := map[string]any{
			"Token": resetToken.Token,
		}
		if err := sendTemplateEmail(f.Email, "Password Reset", "reset_password.txt", emailData); err != nil {
			log.Error().Err(err).Msg("could not send password reset email")
			f.Error = err
			return
		}
		log.Debug().Str("email", f.Email).Msg("Forgot password request")
		p.redirect = fmt.Sprintf("/reset-password?email=%s", f.Email)
	default:
		p.notFound = true
	}
}

type LoginForm struct {
	Email         string `schema:"email"`
	EmailError    error
	Password      string `schema:"password"`
	PasswordError error
	Error         error
}

func (f *LoginForm) Validate() bool {
	f.EmailError = ValidateEmail(f.Email)
	f.PasswordError = ValidatePassword(f.Password)
	return f.EmailError == nil && f.PasswordError == nil
}

type ForgotPasswordForm struct {
	Email      string `schema:"resetEmail"`
	EmailError error
	Error      error
	DialogOpen bool
}

func (f *ForgotPasswordForm) Validate() bool {
	f.EmailError = ValidateEmail(f.Email)
	return f.EmailError == nil
}
