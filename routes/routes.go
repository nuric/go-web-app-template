package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/nuric/go-api-template/utils"
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

type dbRoutes struct {
	db *gorm.DB
}

func SetupRoutes(db *gorm.DB) http.Handler {
	mux := http.NewServeMux()
	dbs := &dbRoutes{db: db}
	log.Debug().Any("dbs", dbs).Msg("Setting up routes")
	mux.HandleFunc("POST /greetings/{firstName}", GreetingHandler)
	mux.HandleFunc("GET /login", GetLoginPage)
	mux.HandleFunc("GET /signup", GetSignUpPage)
	mux.HandleFunc("GET /dashboard", GetDashboardPage)
	return mux
}

func GetLoginPage(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the login page handler.
	// You can render a login page template here.
	if err := tpl.ExecuteTemplate(w, "login.html", nil); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

func GetSignUpPage(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the signup page handler.
	// You can render a signup page template here.
	if err := tpl.ExecuteTemplate(w, "signup.html", nil); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

func GetDashboardPage(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the dashboard page handler.
	// You can render a dashboard page template here.
	if err := tpl.ExecuteTemplate(w, "dashboard.html", nil); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

/* Key things to note:
- Request and response types are nearby the handler for easy debugging, you know
what is coming and going.
- We validate the request explicitly so it knows what is expected.
*/

type GreetingRequest struct {
	LastName string `json:"lastName"`
}

type GreetingResponse struct {
	Greeting string `json:"greeting"`
}

func (r GreetingRequest) Validate() error {
	if r.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	return nil
}

func GreetingHandler(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeValid[GreetingRequest](r)
	/* You'll find this error checking pattern repeats in all handlers. It's not
	 * that much if it really annoys you then you can do it inside the
	 * DecodeValid function. Either way is fine. I prefer explicit error checking
	 * in case we can do something else. */
	if err != nil {
		utils.Encode(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	// Do path validation here.
	firstName := r.PathValue("firstName")
	if len(firstName) < 2 {
		utils.Encode(w, http.StatusBadRequest, map[string]string{"error": "first name must be at least 2 characters"})
		return
	}
	// Construct the response and encode as JSON.
	response := GreetingResponse{
		Greeting: fmt.Sprintf("Hello, %s %s!", firstName, req.LastName),
	}
	utils.Encode(w, http.StatusOK, response)
}
