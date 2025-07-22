package components

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog/log"
)

type LoginForm struct {
	Email          string `schema:"email"`
	EmailError     error
	Password       string `schema:"password"`
	PasswordError  error
	GeneralError   error
	CSRF           template.HTML
	ForgotPassword *ForgotPasswordForm
}

func (f *LoginForm) Validate() (ok bool) {
	ok = true
	if f.Email == "" {
		f.EmailError = fmt.Errorf("email is required")
		ok = false
	}
	if f.Password == "" {
		f.PasswordError = fmt.Errorf("password is required")
		ok = false
	}
	return
}

func (f *LoginForm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.CSRF = csrf.TemplateField(r)
	f.ForgotPassword = &ForgotPasswordForm{}
	f.ForgotPassword.ServeHTTP(w, r)
	if r.Method == http.MethodGet {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// ---------------------------
	r.ParseForm()
	if r.PostFormValue("_action") != "login" {
		// Not our action, ignore
		return
	}
	// ---------------------------
	if err := DecodeValidForm(f, r); err != nil {
		f.GeneralError = err
		return
	}
	var user models.User
	if err := db.Where("email = ?", f.Email).First(&user).Error; err != nil {
		log.Debug().Err(err).Uint("userId", user.ID).Msg("could not find user")
		f.GeneralError = fmt.Errorf("invalid email or password")
		return
	}
	if !utils.VerifyPassword(user.Password, f.Password) {
		log.Debug().Uint("userId", user.ID).Msg("password verification failed")
		f.GeneralError = fmt.Errorf("invalid email or password")
		return
	}
	if err := auth.LogUserIn(w, r, user.ID, ss); err != nil {
		log.Error().Err(err).Uint("userId", user.ID).Msg("could not log user in")
		f.GeneralError = fmt.Errorf("could not log user in")
		return
	}
	log.Debug().Uint("userId", user.ID).Str("email", f.Email).Msg("User logged in successfully")
	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
