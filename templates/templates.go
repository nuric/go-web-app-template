package templates

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

/* When we embed, our binary effectively contains the templates. This allows us
 * to serve them without needing a separate file system. */

//go:embed */*.html */*.txt
var templatesFS embed.FS

var tpl *template.Template

func init() {
	// Parse all templates from the embedded filesystem
	var err error
	if tpl == nil {
		tpl, err = template.ParseFS(templatesFS, "*/*.html", "*/*.txt")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse templates")
		}
	}
	log.Debug().Str("templates", tpl.DefinedTemplates()).Msg("Templates loaded successfully")
}

func RenderHTML(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Render the template
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		log.Error().Err(err).Msg("could not write template error response")
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

func RenderEmail(templateName string, data any) (string, error) {
	// Render the template to a string
	var body strings.Builder
	if err := tpl.ExecuteTemplate(&body, templateName, data); err != nil {
		log.Error().Err(err).Msg("could not render email template")
		return "", err
	}
	return body.String(), nil
}
