package components

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/csrf"
	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
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

func (p *VerifyEmailPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	if r.Method == http.MethodGet {
		switch {
		case user.ID == 0:
			// User is not logged in, redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		case user.ID != 0 && !user.EmailVerified:
			// User is logged in but email is not verified, show verification page
			p.CSRF = csrf.TemplateField(r)
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
			Token:     utils.GenerateRandomString(),
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		db.Create(&newToken)
		p.Message = "Verification email resent. Please check your inbox."
		render(w, "verify_email.html", p)
	case "verify_email":
		// Verify the user's email using the provided token
	}
	// ---------------------------
	// It's not our action, ignore
	// ---------------------------
}
