package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/auth"
	"github.com/nuric/go-api-template/email"
	"github.com/nuric/go-api-template/middleware"
	"github.com/nuric/go-api-template/static"
	"github.com/nuric/go-api-template/storage"
	"github.com/nuric/go-api-template/templates"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// This is the global database connection exposed to the controllers. It should
// be thought of as a dependency of the controllers. Because the nested structure
// can be complex, we are using a global variable. During unit testing it may be
// difficult to inject these dependencies to run tests in parallel, but the
// quality of life improvements are worth it i.e. no need to pass the database
// connection to every component.
var db *gorm.DB
var ss sessions.Store
var em email.Emailer
var st storage.Storer

type Config struct {
	Mux        *http.ServeMux
	Database   *gorm.DB
	Session    sessions.Store
	Emailer    email.Emailer
	Storer     storage.Storer
	CSRFSecret string
	Debug      bool
}

// SetDB sets the global database connection
func Setup(c Config) http.Handler {
	db = c.Database
	ss = c.Session
	em = c.Emailer
	st = c.Storer
	log.Debug().Str("database", db.Name()).Type("session", ss).Type("emailer", em).Str("storer", st.Name()).Msg("Database and session store set")
	// ---------------------------
	// Handle static files
	mux := c.Mux
	if mux == nil {
		mux = http.NewServeMux()
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	// Our routes
	mux.Handle("/login", PageHandler(func() AppPager {
		return &LoginPage{BasePage: BasePage{Title: "Login", Template: "login.html"}}
	}))
	mux.Handle("GET /logout", LogoutPage{})
	mux.Handle("/signup", PageHandler(func() AppPager {
		return &SignUpPage{BasePage: BasePage{Title: "Sign Up", Template: "signup.html"}}
	}))
	mux.Handle("/verify-email", PageHandler(func() AppPager {
		return &VerifyEmailPage{BasePage: BasePage{Title: "Verify Email", Template: "verify_email.html"}}
	}))
	mux.Handle("/reset-password", PageHandler(func() AppPager {
		return &ResetPasswordPage{BasePage: BasePage{Title: "Reset Password", Template: "reset_password.html"}}
	}))
	mux.Handle("GET /dashboard", auth.VerifiedOnly(PageHandler(func() AppPager {
		return &DashboardPage{BasePage: BasePage{Title: "Dashboard", Template: "dashboard.html"}}
	})))
	mux.Handle("/account", auth.VerifiedOnly(PageHandler(func() AppPager {
		return &AccountPage{BasePage: BasePage{Title: "Account", Template: "account.html"}}
	})))
	mux.Handle("GET /uploads/", auth.VerifiedOnly(http.StripPrefix("/uploads/", http.FileServerFS(st))))
	mux.Handle("GET /{$}", http.RedirectHandler("/dashboard", http.StatusSeeOther))
	// Middleware
	var handler http.Handler = mux
	// https://github.com/gorilla/csrf/issues/190
	handler = auth.UserMiddleware(handler, db, ss)
	handler = csrf.Protect([]byte(c.CSRFSecret), csrf.Secure(!c.Debug), csrf.TrustedOrigins([]string{"localhost:8080"}))(handler)
	handler = middleware.NotFoundRenderer(handler)
	return handler
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

type AppPager interface {
	TemplateName() string
	PreHandle(r *http.Request)
	Handle(w http.ResponseWriter, r *http.Request)
	PostHandle(w http.ResponseWriter, r *http.Request)
	Redirect() string
	NotFound() bool
}

type FlashMessage struct {
	Level   string
	Message string
}

type BasePage struct {
	// Page title used in <title>
	Title string
	// Template file name to render
	Template string
	// Used to protect form submissions. Note that all forms on the same page
	// share the same CSRF token.
	CSRF template.HTML
	// Flash messages to be displayed on the page
	FlashMessages []FlashMessage
	// Used for redirects after form submissions
	redirect string
	// Indicates whether the page or action was not found
	notFound bool
}

func (p BasePage) TemplateName() string {
	return p.Template
}

func (p BasePage) Redirect() string {
	return p.redirect
}

func (p BasePage) NotFound() bool {
	return p.notFound
}

const (
	FlashInfo    = "info"
	FlashSuccess = "success"
	FlashWarning = "warning"
	FlashError   = "error"
)

func (p *BasePage) PreHandle(r *http.Request) {
	p.CSRF = csrf.TemplateField(r)
	session, err := ss.Get(r, "flash")
	if err != nil {
		log.Error().Err(err).Msg("could not get flash session")
		return
	}
	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, flash := range flashes {
			if msg, ok := flash.(string); ok {
				parts := strings.SplitN(msg, "$$", 2)
				if len(parts) == 2 {
					level := parts[0]
					message := parts[1]
					p.FlashMessages = append(p.FlashMessages, FlashMessage{Level: level, Message: message})
				}
			}
		}
	}
}

func (p *BasePage) Flash(r *http.Request, level string, message string) {
	session, err := ss.Get(r, "flash")
	if err != nil {
		log.Error().Err(err).Msg("could not get flash session")
	}
	session.AddFlash(fmt.Sprintf("%s$$%s", level, message))
}

func (p *BasePage) Handle(w http.ResponseWriter, r *http.Request) {
	// This is a no-op in the base page, but can be overridden by derived pages
}

func (p *BasePage) PostHandle(w http.ResponseWriter, r *http.Request) {
	session, err := ss.Get(r, "flash")
	if err != nil {
		log.Error().Err(err).Msg("could not get flash session")
		return
	}
	if err := session.Save(r, w); err != nil {
		log.Error().Err(err).Msg("could not save flash session")
	}
}

func PageHandler(factory func() AppPager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := factory()
		if page.TemplateName() == "" {
			http.NotFound(w, r)
			return
		}
		page.PreHandle(r)
		page.Handle(w, r)
		page.PostHandle(w, r)
		switch {
		case page.NotFound():
			http.NotFound(w, r)
			return
		case page.Redirect() != "":
			http.Redirect(w, r, page.Redirect(), http.StatusSeeOther)
			return
		default:
			templates.RenderHTML(w, page.TemplateName(), page)
		}
	})
}

// Helper function to send a template email
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
