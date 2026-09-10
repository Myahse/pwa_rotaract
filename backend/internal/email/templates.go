package email

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFiles embed.FS

var emailTemplates = template.Must(template.ParseFS(templateFiles, "templates/*.html"))
