package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates
var fs embed.FS

type templateSystem struct {
	fs        embed.FS
	templates map[string]*template.Template
}

func NewTemplates() (*templateSystem, error) {
	ts := templateSystem{
		fs:        fs,
		templates: make(map[string]*template.Template),
	}

	if t, err := loadTemplate("templates/index.html"); err == nil {
		ts.templates["index"] = t
	} else {
		return nil, err
	}
	if t, err := loadTemplate("templates/auth.html"); err == nil {
		ts.templates["auth"] = t
	} else {
		return nil, err
	}
	if t, err := loadTemplate("templates/error.html"); err == nil {
		ts.templates["error"] = t
	} else {
		return nil, err
	}

	return &ts, nil
}

func (ts *templateSystem) Render(w io.Writer, name string, data interface{}) error {
	t, ok := ts.templates[name]
	if !ok {
		return &templateNotFoundError{name}
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return err
	}

	return nil
}

func loadTemplate(path string) (*template.Template, error) {
	t, err := template.ParseFS(fs, "templates/base.html", path)
	if err != nil {
		return nil, err
	}
	return t, nil
}
