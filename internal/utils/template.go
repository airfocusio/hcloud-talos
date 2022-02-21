package utils

import (
	"bytes"
	"fmt"
	"text/template"
)

func RenderTemplate(input string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(input)
	if err != nil {
		return "", fmt.Errorf("unable to parse template: %w", err)
	}
	buf := new(bytes.Buffer)
	if buf == nil {
		return "", fmt.Errorf("unable to initialize render buffer")
	}
	err = tmpl.Execute(buf, data)
	if err != nil {
		return "", fmt.Errorf("unable to render template: %w", err)
	}
	output := buf.String()
	return output, nil
}
