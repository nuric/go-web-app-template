package components

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"

	"github.com/gorilla/csrf"
	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog/log"
)

type SignUpPage struct {
	Email                string `schema:"email"`
	EmailError           error
	Password             string `schema:"password"`
	PasswordError        error
	ConfirmPassword      string `schema:"confirmPassword"`
	ConfirmPasswordError error
	CSRF                 template.HTML
	GeneralError         error
}

func (p *SignUpPage) Validate() (ok bool) {
	ok = true
	if p.Email == "" {
		p.EmailError = fmt.Errorf("email is required")
		ok = false
	}
	if p.Password == "" {
		p.PasswordError = fmt.Errorf("password is required")
		ok = false
	}
	if p.Password != p.ConfirmPassword {
		p.ConfirmPasswordError = fmt.Errorf("passwords do not match")
		ok = false
	}
	if len(p.Password) < 8 ||
		!regexp.MustCompile(`[a-z]`).MatchString(p.Password) ||
		!regexp.MustCompile(`[A-Z]`).MatchString(p.Password) ||
		!regexp.MustCompile(`\d`).MatchString(p.Password) ||
		!regexp.MustCompile(`[@$!%*?&=]`).MatchString(p.Password) {
		p.PasswordError = fmt.Errorf("password must be at least 8 characters long, contain at least one lowercase letter, one uppercase letter, one digit, and one special character")
		ok = false
	}
	return
}

func (p *SignUpPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.CSRF = csrf.TemplateField(r)
	if r.Method == http.MethodGet {
		render(w, "signup.html", p)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	if r.PostFormValue("_action") != "signup" {
		// Not our action, ignore
		return
	}
	// ---------------------------
	if err := DecodeValidForm(p, r); err != nil {
		p.GeneralError = err
		render(w, "signup.html", p)
		return
	}
	newUser := models.User{
		Email:    p.Email,
		Password: utils.HashPassword(p.Password),
		Role:     "basic", // Default role
	}

	if err := db.Create(&newUser).Error; err != nil {
		log.Error().Err(err).Msg("could not create user")
		p.GeneralError = fmt.Errorf("could not create user")
		render(w, "signup.html", p)
		return
	}

	if err := auth.LogUserIn(w, r, newUser.ID, ss); err != nil {
		log.Error().Err(err).Msg("could not log user in after signup")
		http.Error(w, "could not log user in after signup", http.StatusInternalServerError)
		return
	}
	// Here you would typically save the user to the database.
	log.Debug().Str("email", p.Email).Msg("User signed up successfully")
	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
