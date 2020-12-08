package email

import (
	"fmt"
	"html/template"
	"path"
)

// TemplateFileSystem is load and serve templates from local file system
type TemplateFileSystem struct {
	location  string
	templates map[string]*template.Template
}

// NewTemplateFileSystem ...
func NewTemplateFileSystem(location string) *TemplateFileSystem {
	return &TemplateFileSystem{
		templates: make(map[string]*template.Template),
		location:  location,
	}
}

// Load parses and prepare a template for further use of it
// It will reloads the template if called many times with same input
func (t *TemplateFileSystem) Load(name string) error {
	// add tpl file extension
	withExtension := fmt.Sprintf("%s.tpl", name)
	filename := path.Join(t.location, withExtension)
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		return err
	}
	t.templates[name] = tmpl
	return nil
}

// Get return previously loaded templates
func (t *TemplateFileSystem) Get(name string) (*template.Template, error) {
	tmpl, found := t.templates[name]
	if !found {
		return nil, fmt.Errorf("unknown template: %s", name)
	}
	return tmpl, nil
}
