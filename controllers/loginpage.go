package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
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

func (p LoginPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	switch {
	case user.ID != 0 && user.EmailVerified:
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	case user.ID != 0 && !user.EmailVerified:
		http.Redirect(w, r, "/verify-email", http.StatusSeeOther)
		return
	}
	p.LoginForm.CSRF = csrf.TemplateField(r)
	p.ForgotPasswordForm.CSRF = csrf.TemplateField(r)
	// ---------------------------
	if r.Method == http.MethodGet {
		render(w, "login.html", p)
		return
	}
	// ---------------------------
	r.ParseForm()
	switch r.PostFormValue("_action") {
	case "login":
		f := &p.LoginForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			render(w, "login.html", p)
			return
		}
		var user models.User
		if err := db.Where("email = ?", f.Email).First(&user).Error; err != nil {
			log.Debug().Err(err).Uint("userId", user.ID).Msg("could not find user")
			f.Error = errors.New("invalid email or password")
			render(w, "login.html", p)
			return
		}
		if !utils.VerifyPassword(user.Password, f.Password) {
			log.Debug().Uint("userId", user.ID).Msg("password verification failed")
			f.Error = errors.New("invalid email or password")
			render(w, "login.html", p)
			return
		}
		if err := auth.LogUserIn(w, r, user.ID, ss); err != nil {
			log.Error().Err(err).Uint("userId", user.ID).Msg("could not log user in")
			f.Error = errors.New("could not log user in")
			render(w, "login.html", p)
			return
		}
		log.Debug().Uint("userId", user.ID).Str("email", f.Email).Msg("User logged in successfully")
		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	case "forgot_password":
		f := &p.ForgotPasswordForm
		f.DialogOpen = true
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			render(w, "login.html", p)
			return
		}
		resetToken := models.Token{
			Email:     f.Email,
			Token:     utils.HumanFriendlyToken(),
			Purpose:   "reset_password",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if err := db.Create(&resetToken).Error; err != nil {
			log.Error().Err(err).Msg("could not create password reset token")
			f.Error = errors.New("could not send password reset email")
			render(w, "login.html", p)
			return
		}
		emailData := map[string]any{
			"Token": resetToken.Token,
		}
		if err := sendTemplateEmail(f.Email, "Password Reset", "reset_password.txt", emailData); err != nil {
			log.Error().Err(err).Msg("could not send password reset email")
			f.Error = err
			render(w, "login.html", p)
			return
		}
		log.Debug().Str("email", f.Email).Msg("Forgot password request")
		redirect := fmt.Sprintf("/reset-password?email=%s", f.Email)
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	default:
		http.NotFound(w, r)
	}
}

type LoginForm struct {
	Email         string `schema:"email"`
	EmailError    error
	Password      string `schema:"password"`
	PasswordError error
	Error         error
	CSRF          template.HTML
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
	CSRF       template.HTML
}

func (f *ForgotPasswordForm) Validate() bool {
	f.EmailError = ValidateEmail(f.Email)
	return f.EmailError == nil
}
