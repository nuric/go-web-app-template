package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/auth"
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

func render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if data == nil {
		data = make(map[string]any)
	}
	// Add common context
	currentUser := auth.GetCurrentUser(r)
	data["User"] = currentUser
	data[csrf.TemplateTag] = csrf.TemplateField(r)
	// Render the template
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
	mux.Handle("/", auth.AuthenticatedOnly(authBlock, db, ss))
	return mux
}

func GetLoginPage(w http.ResponseWriter, r *http.Request) {
	render(w, r, "login.html", nil)
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
			"error": err, // You can set an error message if needed
		}
		render(w, r, "login.html", data)
		return
	}
	var user models.User
	if err := dbs.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.Debug().Err(err).Msg("could not find user")
		data := map[string]any{
			"error": "invalid email or password",
		}
		render(w, r, "login.html", data)
		return
	}
	if !utils.VerifyPassword(user.Password, req.Password) {
		log.Debug().Msg("password verification failed")
		data := map[string]any{
			"error": "invalid email or password",
		}
		render(w, r, "login.html", data)
		return
	}
	if err := auth.LogUserIn(w, r, user.ID, dbs.ss); err != nil {
		log.Error().Err(err).Msg("could not log user in")
		http.Error(w, "could not log user in", http.StatusInternalServerError)
		return
	}
	log.Debug().Str("email", req.Email).Msg("User logged in successfully")
	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (dbs *dbRoutes) LogoutPage(w http.ResponseWriter, r *http.Request) {
	if err := auth.LogUserOut(w, r, dbs.ss); err != nil {
		log.Error().Err(err).Msg("could not log user out")
		http.Error(w, "could not log user out", http.StatusInternalServerError)
		return
	}
	log.Debug().Msg("User logged out successfully")
	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func GetSignUpPage(w http.ResponseWriter, r *http.Request) {
	render(w, r, "signup.html", nil)
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
			"error": err, // You can set an error message if needed
		}
		render(w, r, "signup.html", data)
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
			"error": "could not create user: " + err.Error(),
		}
		render(w, r, "signup.html", data)
		return
	}

	if err := auth.LogUserIn(w, r, newUser.ID, dbs.ss); err != nil {
		log.Error().Err(err).Msg("could not log user in after signup")
		http.Error(w, "could not log user in after signup", http.StatusInternalServerError)
		return
	}
	// Here you would typically save the user to the database.
	log.Debug().Str("email", req.Email).Msg("User signed up successfully")
	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (dbs *dbRoutes) GetDashboardPage(w http.ResponseWriter, r *http.Request) {
	render(w, r, "dashboard.html", nil)
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
