package controllers

import (
	"net/http"

	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/models"
)

type DashboardPage struct {
	BasePage
	User models.User
}

func (p DashboardPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.User = auth.GetCurrentUser(r)
	render(w, "dashboard.html", p)
}
