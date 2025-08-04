package controllers

import (
	"errors"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ResetPasswordPage struct {
	Email                string `schema:"email"`
	EmailError           error
	Token                string `schema:"token"`
	TokenError           error
	NewPassword          string `schema:"newPassword"`
	NewPasswordError     error
	ConfirmPassword      string `schema:"confirmPassword"`
	ConfirmPasswordError error
	Error                error
	CSRF                 template.HTML
}

func (p *ResetPasswordPage) Validate() bool {
	p.EmailError = ValidateEmail(p.Email)
	p.TokenError = ValidateToken(p.Token)
	p.NewPasswordError = ValidatePassword(p.NewPassword)
	if p.NewPassword != p.ConfirmPassword {
		p.ConfirmPasswordError = errors.New("passwords do not match")
	}
	return p.EmailError == nil &&
		p.TokenError == nil &&
		p.NewPasswordError == nil &&
		p.ConfirmPasswordError == nil
}

func (p ResetPasswordPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	if user.ID != 0 {
		// User is logged in, redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	// ---------------------------
	p.CSRF = csrf.TemplateField(r)
	p.Email = r.URL.Query().Get("email")
	if r.Method == http.MethodGet {
		render(w, "reset_password.html", p)
		return
	}
	// ---------------------------
	r.ParseForm()
	if r.PostFormValue("_action") != "reset_password" {
		http.NotFound(w, r)
		return
	}

	if err := DecodeValidForm(&p, r); err != nil {
		p.Error = err
		render(w, "reset_password.html", p)
		return
	}
	var token models.Token
	if err := db.Where("token = ?", p.Token).
		Where("email = ?", p.Email).
		Where("purpose = ?", "reset_password").
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&token).Error; err != nil {
		log.Error().Err(err).Msg("could not find valid token")
		p.Error = errors.New("invalid token")
		render(w, "reset_password.html", p)
		return
	}
	if p.Token != token.Token {
		log.Error().Msg("password reset token mismatch")
		p.Error = errors.New("invalid token")
		render(w, "reset_password.html", p)
		return
	}
	// Reset the user's password
	rows, err := gorm.G[models.User](db).Where("email = ?", p.Email).Update(r.Context(), "password", utils.HashPassword(p.NewPassword))
	if err != nil || rows == 0 {
		log.Error().Err(err).Msg("could not update user password")
		p.Error = errors.New("could not update password")
		render(w, "reset_password.html", p)
		return
	}
	// Delete the token after successful password reset
	if err := db.Delete(&token).Error; err != nil {
		log.Error().Err(err).Msg("could not delete reset token")
	}

	// Jobs done, they can now login with the new password
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
