package templates

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
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
			panic("could not parse templates: " + err.Error())
		}
	}
	slog.Debug("Templates loaded", "template_count", len(tpl.Templates()))
}

func RenderHTML(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Render the template
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("could not write template response", "error", err)
		http.Error(w, "could not generate page", http.StatusInternalServerError)
	}
}

func RenderEmail(templateName string, data any) (string, error) {
	// Render the template to a string
	var body strings.Builder
	if err := tpl.ExecuteTemplate(&body, templateName, data); err != nil {
		slog.Error("could not render email template", "error", err)
		return "", err
	}
	return body.String(), nil
}
