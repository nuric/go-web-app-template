package components

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/rs/zerolog/log"
)

type ForgotPasswordForm struct {
	Email          string `schema:"resetEmail"`
	EmailError     error
	GeneralError   error
	SuccessMessage string
	DialogOpen     bool
	CSRF           template.HTML
}

func (f *ForgotPasswordForm) Validate() (ok bool) {
	ok = true
	if f.Email == "" {
		f.EmailError = fmt.Errorf("email is required")
		ok = false
	}
	return
}

func (f *ForgotPasswordForm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.CSRF = csrf.TemplateField(r)
	if r.Method == http.MethodGet {
		return
	}
	// ---------------------------
	r.ParseForm()
	if r.PostFormValue("_action") != "forgot_password" {
		// Not our action, ignore
		return
	}
	// ---------------------------
	f.DialogOpen = true
	if err := DecodeValidForm(f, r); err != nil {
		f.GeneralError = err
		return
	}
	log.Debug().Str("email", f.Email).Msg("Forgot password request")
	f.SuccessMessage = "If an account with that email exists, a reset link has been sent."
}
