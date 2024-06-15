package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"path/filepath"
)


type Templates struct {
	templates *template.Template
}

func newTemplates() *Templates {
	absPath, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}
	fmt.Println("Current Directory:", absPath)
	fmt.Println("Looking for templates in:", filepath.Join("views", "*.html"))

	templates := template.Must(template.ParseGlob(filepath.Join("views", "*.html")))
	return &Templates{templates: templates}
}

func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
