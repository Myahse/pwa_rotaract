package membercard

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"
)

//go:embed templates/*.html
var cardTemplateFS embed.FS

var cardTemplate = template.Must(
	template.New("card.html").Funcs(template.FuncMap{
		"upper": func(value string) string {
			return strings.ToUpper(strings.TrimSpace(value))
		},
		"formatCardNumber": formatCardNumber,
		"formatDate": func(value time.Time) string {
			return value.Format("02/01/2006")
		},
	}).ParseFS(cardTemplateFS, "templates/*.html"),
)

func RenderHTML(data CardData) (string, error) {
	if strings.TrimSpace(data.FullName) == "" {
		data.FullName = "Membre"
	}
	if strings.TrimSpace(data.ClubName) == "" {
		data.ClubName = "Club Rotaract"
	}

	var buf bytes.Buffer
	if err := cardTemplate.ExecuteTemplate(&buf, "card.html", data); err != nil {
		return "", fmt.Errorf("render member card html: %w", err)
	}
	return buf.String(), nil
}

func RenderHTMLSafe(data CardData) (template.HTML, error) {
	html, err := RenderHTML(data)
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}
