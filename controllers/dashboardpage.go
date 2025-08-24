package controllers

import (
	"net/http"

	"github.com/nuric/go-web-app-template/auth"
	"github.com/nuric/go-web-app-template/models"
)

type DashboardPage struct {
	BasePage
	User models.User
}

func (p *DashboardPage) Handle(w http.ResponseWriter, r *http.Request) {
	p.User = auth.GetCurrentUser(r)
}
