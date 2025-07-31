package components

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/csrf"
	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog/log"
)

type VerifyEmailPage struct {
	Token      string `schema:"token"`
	TokenError error
	CSRF       template.HTML
	Error      error
	Message    string
}

func (p *VerifyEmailPage) Validate() (ok bool) {
	ok = true
	if p.Token == "" {
		p.TokenError = fmt.Errorf("token is required")
		ok = false
	} else {
		// Validate token format (simple regex for demonstration)
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, p.Token); !matched {
			p.TokenError = fmt.Errorf("invalid token format")
			ok = false
		}
	}
	return
}

func (p VerifyEmailPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	p.CSRF = csrf.TemplateField(r)
	if r.Method == http.MethodGet {
		switch {
		case user.ID == 0:
			// User is not logged in, redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		case user.ID != 0 && !user.EmailVerified:
			// User is logged in but email is not verified, show verification page
			render(w, "verify_email.html", p)
		case user.ID != 0 && user.EmailVerified:
			// User is logged in and email is verified, redirect to dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
		return
	}
	// ---------------------------
	r.ParseForm()
	switch r.PostFormValue("_action") {
	case "resend_verification":
		// Todo: Implement resend verification logic
		newToken := models.Token{
			UserID:    user.ID,
			Token:     utils.HumanFriendlyToken(),
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if err := db.Create(&newToken).Error; err != nil {
			log.Error().Err(err).Msg("could not create verification token")
			p.Error = errors.New("could not create verification token")
			render(w, "verify_email.html", p)
			return
		}
		emailData := map[string]any{
			"Token": newToken.Token,
		}
		if err := sendTemplateEmail(user.Email, "Email Verification", "verify_email.txt", emailData); err != nil {
			log.Error().Err(err).Msg("could not send email")
			p.Error = err
			render(w, "verify_email.html", p)
			return
		}
		p.Message = "Verification email resent. Please check your inbox."
		render(w, "verify_email.html", p)
	case "verify_email":
		// Verify the user's email using the provided token
		if err := DecodeValidForm(&p, r); err != nil {
			p.Error = err
			render(w, "verify_email.html", p)
			return
		}
		// Get the last token that hasn't expired
		var token models.Token
		if err := db.Where("user_id = ?", user.ID).
			Where("token = ?", p.Token).
			Where("purpose = ?", "email_verification").
			Where("expires_at > ?", time.Now()).
			Order("created_at DESC").
			First(&token).Error; err != nil {
			log.Error().Err(err).Msg("could not find valid token")
			p.Error = fmt.Errorf("invalid token")
			render(w, "verify_email.html", p)
			return
		}
		// Check if the token matches
		if token.Token != p.Token {
			log.Error().Msg("token mismatch")
			p.Error = errors.New("invalid token")
			render(w, "verify_email.html", p)
			return
		}
		// Update the user's email verification status
		if err := db.Model(&user).Update("email_verified", true).Error; err != nil {
			log.Error().Err(err).Msg("could not update user email verification status")
			p.Error = errors.New("could not verify email")
			render(w, "verify_email.html", p)
			return
		}
		// Delete the token after successful verification
		if err := db.Delete(&token).Error; err != nil {
			log.Error().Err(err).Msg("could not delete token after verification")
		}
		// Redirect to dashboard after successful verification
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
