package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/middleware"
	"github.com/nuric/go-api-template/models"
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

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

type dbRoutes struct {
	db *gorm.DB
	ss sessions.Store
}

func SetupRoutes(db *gorm.DB, ss sessions.Store) http.Handler {
	mux := http.NewServeMux()
	dbs := &dbRoutes{db: db, ss: ss}
	log.Debug().Any("dbs", dbs).Msg("Setting up routes")
	mux.HandleFunc("POST /greetings/{firstName}", GreetingHandler)
	mux.HandleFunc("GET /login", GetLoginPage)
	mux.HandleFunc("POST /login", dbs.PostLoginPage)
	mux.HandleFunc("GET /logout", dbs.LogoutPage)
	mux.HandleFunc("GET /signup", GetSignUpPage)
	mux.HandleFunc("POST /signup", dbs.PostSignUpPage)
	authBlock := http.NewServeMux()
	authBlock.HandleFunc("GET /dashboard", dbs.GetDashboardPage)
	mux.Handle("/", middleware.AuthenticatedOnly(authBlock, ss))
	return mux
}

func GetLoginPage(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the login page handler.
	// You can render a login page template here.
	data := map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
		"error":          "memes", // You can set an error message if needed
	}
	if err := tpl.ExecuteTemplate(w, "login.html", data); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

type LoginRequest struct {
	Email    string
	Password string
}

func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (dbs *dbRoutes) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeValidForm[LoginRequest](r)
	if err != nil {
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"error":          err, // You can set an error message if needed
		}
		render(w, "login.html", data)
		return
	}
	var user models.User
	if err := dbs.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.Error().Err(err).Msg("could not find user")
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"error":          "invalid email or password",
		}
		render(w, "login.html", data)
		return
	}
	if !utils.VerifyPassword(user.Password, req.Password) {
		log.Error().Msg("password verification failed")
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"error":          "invalid email or password",
		}
		render(w, "login.html", data)
		return
	}
	newSession, err := dbs.ss.New(r, "app-session")
	if err != nil {
		log.Error().Err(err).Msg("could not create session")
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	newSession.Values["userId"] = user.ID
	if err := newSession.Save(r, w); err != nil {
		log.Error().Err(err).Msg("could not save session")
		http.Error(w, "could not save session", http.StatusInternalServerError)
		return
	}
	log.Debug().Str("email", req.Email).Msg("User logged in successfully")
	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (dbs *dbRoutes) LogoutPage(w http.ResponseWriter, r *http.Request) {
	session, err := dbs.ss.Get(r, "app-session")
	if err != nil {
		log.Error().Err(err).Msg("could not get session")
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}
	// Clear the session values
	session.Values = make(map[any]any)
	if err := session.Save(r, w); err != nil {
		log.Error().Err(err).Msg("could not save session")
		http.Error(w, "could not save session", http.StatusInternalServerError)
		return
	}
	log.Debug().Msg("User logged out successfully")
	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func GetSignUpPage(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the signup page handler.
	// You can render a signup page template here.
	data := map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
		"error":          "memes", // You can set an error message if needed
	}
	if err := tpl.ExecuteTemplate(w, "signup.html", data); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

type SignUpRequest struct {
	Email           string
	Password        string
	ConfirmPassword string
}

func (r SignUpRequest) Validate() error {
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	if r.Password != r.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}
	return nil
}

func (dbs *dbRoutes) PostSignUpPage(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeValidForm[SignUpRequest](r)
	if err != nil {
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"error":          err, // You can set an error message if needed
		}
		render(w, "signup.html", data)
		return
	}

	newUser := models.User{
		Email:    req.Email,
		Password: utils.HashPassword(req.Password),
		Role:     "basic", // Default role
	}

	if err := dbs.db.Create(&newUser).Error; err != nil {
		log.Error().Err(err).Msg("could not create user")
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"error":          "could not create user: " + err.Error(),
		}
		render(w, "signup.html", data)
		return
	}

	newSession, err := dbs.ss.New(r, "app-session")
	if err != nil {
		log.Error().Err(err).Msg("could not create session")
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	newSession.Values["userId"] = newUser.ID
	if err := newSession.Save(r, w); err != nil {
		log.Error().Err(err).Msg("could not save session")
		http.Error(w, "could not save session", http.StatusInternalServerError)
		return
	}

	// Here you would typically save the user to the database.
	log.Debug().Str("email", req.Email).Msg("User signed up successfully")

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (dbs *dbRoutes) GetDashboardPage(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.UserIDKey).(uint)
	var user models.User
	if err := dbs.db.First(&user, userId).Error; err != nil {
		log.Error().Err(err).Msg("could not find user")
		http.Error(w, "could not find user", http.StatusInternalServerError)
		return
	}
	// This is a placeholder for the dashboard page handler.
	// You can render a dashboard page template here.
	data := map[string]any{
		"User": user,
	}
	if err := tpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
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
	req, err := utils.DecodeValidJSON[GreetingRequest](r)
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
