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
)

type AccountPage struct {
	User               models.User
	ChangeEmailForm    ChangeEmailForm
	ChangePasswordForm ChangePasswordForm
}

type ChangeEmailForm struct {
	Action     string `schema:"_action"`
	Email      string `schema:"email"`
	EmailError error
	Token      string `schema:"token"`
	TokenError error
	CSRF       template.HTML
	Error      error
}

func (f *ChangeEmailForm) Validate() bool {
	f.EmailError = ValidateEmail(f.Email)
	if f.Action == "change_email" {
		f.TokenError = ValidateToken(f.Token)
	}
	return f.EmailError == nil && f.TokenError == nil
}

type ChangePasswordForm struct {
	CurrentPassword      string `schema:"currentPassword"`
	CurrentPasswordError error
	NewPassword          string `schema:"newPassword"`
	NewPasswordError     error
	ConfirmPassword      string `schema:"confirmPassword"`
	ConfirmPasswordError error
	CSRF                 template.HTML
	Error                error
}

func (f *ChangePasswordForm) Validate() bool {
	f.CurrentPasswordError = ValidatePassword(f.CurrentPassword)
	f.NewPasswordError = ValidatePassword(f.NewPassword)
	if f.NewPassword != f.ConfirmPassword {
		f.ConfirmPasswordError = errors.New("passwords do not match")
	}
	return f.CurrentPasswordError == nil && f.NewPasswordError == nil && f.ConfirmPasswordError == nil
}

func (p AccountPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.User = auth.GetCurrentUser(r)
	p.ChangeEmailForm.CSRF = csrf.TemplateField(r)
	p.ChangePasswordForm.CSRF = csrf.TemplateField(r)
	// ---------------------------
	if r.Method == http.MethodGet {
		render(w, "account.html", p)
		return
	}
	// ---------------------------
	r.ParseForm()
	switch r.PostFormValue("_action") {
	case "request_email_change_token":
		f := &p.ChangeEmailForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			render(w, "account.html", p)
			return
		}
		verificationToken := models.Token{
			UserID:    p.User.ID,
			Email:     f.Email,
			Token:     utils.HumanFriendlyToken(),
			Purpose:   "change_email",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		if err := db.Create(&verificationToken).Error; err != nil {
			log.Error().Err(err).Msg("could not create email change verification token")
			f.Error = errors.New("could not create email change verification token")
			render(w, "account.html", p)
			return
		}
		emailData := map[string]any{
			"Token": verificationToken.Token,
		}
		if err := sendTemplateEmail(f.Email, "Email Change Verification", "verify_email.txt", emailData); err != nil {
			log.Error().Err(err).Msg("could not send email")
			f.Error = errors.New("could not send verification email")
			render(w, "verify_email.html", p)
			return
		}
		// Switch to next action
		f.Action = "change_email"
		render(w, "account.html", p)
	case "change_email":
		f := &p.ChangeEmailForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			render(w, "account.html", p)
			return
		}
		// Verify the token
		var verificationToken models.Token
		if err := db.Where("user_id = ? AND email = ? AND token = ? AND purpose = ?", p.User.ID, f.Email, f.Token, "change_email").First(&verificationToken).Error; err != nil {
			log.Error().Err(err).Msg("could not find verification token")
			f.Error = errors.New("invalid or expired token")
			render(w, "account.html", p)
			return
		}
		// Update the email of the user
		if err := db.Model(&p.User).Update("email", verificationToken.Email).Error; err != nil {
			log.Error().Err(err).Msg("could not update user email")
			f.Error = errors.New("could not change user email")
			render(w, "account.html", p)
			return
		}
		// Delete the verification token
		if err := db.Delete(&verificationToken).Error; err != nil {
			log.Error().Err(err).Msg("could not delete verification token")
		}
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	case "change_password":
		f := &p.ChangePasswordForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			render(w, "account.html", p)
			return
		}
		if !utils.VerifyPassword(p.User.Password, f.CurrentPassword) {
			f.CurrentPasswordError = errors.New("please enter your current password")
			render(w, "account.html", p)
			return
		}
		hashedPassword := utils.HashPassword(f.NewPassword)
		if err := db.Model(&p.User).Update("password", hashedPassword).Error; err != nil {
			log.Error().Err(err).Msg("could not change user password")
			f.Error = errors.New("could not change user password")
			render(w, "account.html", p)
			return
		}
		// Redirect to GET current page
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	default:
		http.NotFound(w, r)
	}
}
