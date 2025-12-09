package server

import (
	"encoding/json"
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

var templateFuncs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"inc": func(i int) int { return i + 1 },
	"json": func(v interface{}) template.JS {
		a, _ := json.Marshal(v)
		return template.JS(a)
	},
}

func newTemplates(views string) *Templates {
	templates := template.Must(
		template.New("").
			Funcs(templateFuncs).
			ParseGlob(views),
	)
	return &Templates{templates: templates}
}

func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
