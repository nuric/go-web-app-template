package components

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var tpl *template.Template

func init() {
	_, sourcePath, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal().Msg("Could not determine source path for templates")
	}
	// sourcePath  is something like .../go-web-app-template/routes/routes.go
	// We want .../go-web-app-template/templates/*/*.html
	tplPath := filepath.Join(filepath.Dir(filepath.Dir(sourcePath)), "templates", "*", "*.html")
	log.Debug().Str("tplPath", tplPath).Msg("Loading templates")
	tpl = template.Must(template.ParseGlob(tplPath))
	fmt.Println(tpl.DefinedTemplates())
}

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Render the template
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

// This is the global database connection exposed to the components. It should
// be thought of as a dependency of the components. Because the nested structure
// can be complex, we are using a global variable. During unit testing it may be
// difficult to inject these dependencies to run tests in parallel, but the
// quality of life improvements are worth it i.e. no need to pass the database
// connection to every component.
var db *gorm.DB
var ss sessions.Store

// SetDB sets the global database connection
func Set(database *gorm.DB, store sessions.Store) {
	db = database
	ss = store
	log.Debug().Str("database", db.Name()).Str("store", fmt.Sprintf("%T", store)).Msg("Database and session store set")
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
