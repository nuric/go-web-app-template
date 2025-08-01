package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/email"
	"github.com/nuric/go-api-template/templates"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// helper to reduce boilerplate in controllers
func render(w http.ResponseWriter, name string, data any) {
	templates.RenderHTML(w, name, data)
}

func sendTemplateEmail(to, subject, templateName string, data any) error {
	// Render the template to a string
	body, err := templates.RenderEmail(templateName, data)
	if err != nil {
		log.Error().Err(err).Msg("could not render email template")
		return errors.New("could not render email template")
	}
	// Send the email using the emailer
	if err := em.SendEmail(to, subject, body); err != nil {
		log.Error().Err(err).Msg("could not send email")
		return errors.New("could not send email")
	}
	return nil
}

// This is the global database connection exposed to the controllers. It should
// be thought of as a dependency of the controllers. Because the nested structure
// can be complex, we are using a global variable. During unit testing it may be
// difficult to inject these dependencies to run tests in parallel, but the
// quality of life improvements are worth it i.e. no need to pass the database
// connection to every component.
var db *gorm.DB
var ss sessions.Store
var em email.Emailer

// SetDB sets the global database connection
func Set(database *gorm.DB, store sessions.Store, emailer email.Emailer) {
	db = database
	ss = store
	em = emailer
	log.Debug().Str("database", db.Name()).Type("store", store).Type("emailer", emailer).Msg("Database and session store set")
}

type Validator interface {
	Validate() bool
}

func DecodeValidForm[T Validator](v T, r *http.Request) error {
	// First, parse the form data from the request body.
	// This populates r.PostForm.
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	// Decode the form data from r.PostForm into the struct.
	// We use r.PostForm to ensure we only get data from the request body,
	// not from the URL query parameters.
	newSchemaDecoder := schema.NewDecoder()
	newSchemaDecoder.IgnoreUnknownKeys(true) // Ignore any unknown keys in the form data
	if err := newSchemaDecoder.Decode(v, r.PostForm); err != nil {
		return fmt.Errorf("failed to decode form: %w", err)
	}
	if !v.Validate() {
		return fmt.Errorf("please correct the errors in the form")
	}
	return nil
}
