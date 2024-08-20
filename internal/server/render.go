package server

import (
	"html/template"
	"io"
	"path/filepath"
)

type TemplatesRepo interface {
	Render(w io.Writer, name string, data interface{}) error
}

type Templates struct {
	templates *template.Template
}

func viewsPath() (string, error) {
	absPath, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	return filepath.Join(absPath, "views", "*.html"), nil
}

func newTemplates(views string) *Templates {
	templates := template.Must(template.ParseGlob(views))
	return &Templates{templates: templates}
}

func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
