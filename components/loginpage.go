package components

import "net/http"

type LoginPage struct {
	LoginForm *LoginForm
}

func (p *LoginPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.LoginForm = &LoginForm{}
	p.LoginForm.ServeHTTP(w, r)
	render(w, "login.html", p)
}
