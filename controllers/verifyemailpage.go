package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/nuric/go-web-app-template/auth"
	"github.com/nuric/go-web-app-template/models"
	"github.com/nuric/go-web-app-template/utils"
)

type VerifyEmailPage struct {
	BasePage
	Token      string `schema:"token"`
	TokenError error
	Error      error
	Message    string
}

func (p *VerifyEmailPage) Validate() bool {
	p.TokenError = ValidateToken(p.Token)
	return p.TokenError == nil
}

func sendEmailVerification(userID uint, email string) error {
	newToken := models.Token{
		UserID:    userID,
		Email:     email,
		Token:     utils.HumanFriendlyToken(),
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := db.Create(&newToken).Error; err != nil {
		slog.Error("could not create email verification token", "error", err)
		return errors.New("could not create verification token")
	}
	emailData := map[string]any{
		"Token": newToken.Token,
	}
	if err := sendTemplateEmail(email, "Email Verification", "verify_email.txt", emailData); err != nil {
		slog.Error("could not send verification email", "error", err)
		return errors.New("could not send verification email")
	}
	return nil
}

func checkEmailVerification(userID uint, email string, userToken string) error {
	// Get the last token that hasn't expired
	var token models.Token
	if err := db.Where("user_id = ?", userID).
		Where("token = ?", userToken).
		Where("email = ?", email).
		Where("purpose = ?", "email_verification").
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&token).Error; err != nil {
		slog.Error("could not find valid token", "error", err)
		return errors.New("invalid token or expired token")
	}
	// Check if the token matches
	if token.Token != userToken {
		return errors.New("invalid token or expired token")
	}
	// Delete token as it is now considered used
	if err := db.Delete(&token).Error; err != nil {
		slog.Error("could not delete token after verification", "error", err)
	}
	return nil
}

func (p *VerifyEmailPage) Handle(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	switch {
	case user.ID != 0 && user.EmailVerified:
		// User is logged in and email is verified, redirect to dashboard
		p.redirect = "/dashboard"
		return
	case user.ID != 0 && !user.EmailVerified:
		// We'll handle this case below
	default:
		p.redirect = "/login"
		return
	}
	if r.Method == http.MethodGet {
		return
	}
	// ---------------------------
	switch r.PostFormValue("_action") {
	case "resend_verification":
		if err := sendEmailVerification(user.ID, user.Email); err != nil {
			p.Error = err
			return
		}
		p.Message = "Verification email resent. Please check your inbox."
	case "verify_email":
		// Verify the user's email using the provided token
		if err := DecodeValidForm(p, r); err != nil {
			p.Error = err
			return
		}
		if err := checkEmailVerification(user.ID, user.Email, p.Token); err != nil {
			p.Error = err
			return
		}
		// Update the user's email verification status
		if err := db.Model(&user).Update("email_verified", true).Error; err != nil {
			slog.Error("could not update user email verification status", "error", err)
			p.Error = errors.New("could not verify email")
			return
		}
		// Redirect to dashboard after successful verification
		p.redirect = "/dashboard"
	default:
		p.notFound = true
	}
}
