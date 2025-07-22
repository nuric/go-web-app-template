package components

import (
	"net/http"

	"github.com/nuric/go-api-template/auth"
)

type LoginPage struct {
	LoginForm *LoginForm
}

func (p *LoginPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)
	if user.ID != 0 {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	// ---------------------------
	p.LoginForm = &LoginForm{}
	p.LoginForm.ServeHTTP(w, r)
	render(w, "login.html", p)
}
