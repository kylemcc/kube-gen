package kubegen

import (
	"bytes"
	"path/filepath"
	"text/template"
)

func newTemplate(name string) *template.Template {
	return template.New(name).Funcs(Funcs)
}

// Executes a template located at path with the specified data
func execTemplateFile(path string, data interface{}) ([]byte, error) {
	tmpl, err := newTemplate(filepath.Base(path)).ParseFiles(path)
	if err != nil {
		return nil, err
	}
	return execTemplate(tmpl, data)
}

// Executes a template string with the specified data
func execTemplateString(text string, data interface{}) ([]byte, error) {
	tmpl, err := newTemplate("stdin").Parse(text)
	if err != nil {
		return nil, err
	}
	return execTemplate(tmpl, data)
}

// Helper for execTemplateFile and execTemplateString - actually executes the template
func execTemplate(tmpl *template.Template, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
