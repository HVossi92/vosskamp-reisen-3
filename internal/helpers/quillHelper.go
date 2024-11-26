package helpers

import (
	"html/template"
	"strings"

	quill "github.com/dchenk/go-render-quill"
)

func ConvertQuillToHtml(quillText string) (template.HTML, error) {
	start := strings.Index(quillText, `"ops":[`) + len(`"ops":[`)
	end := strings.Index(quillText[start:], `]`) + start

	// Extract the ops array
	opsArray := "[" + quillText[start:end] + "]"

	byteAry, err := quill.Render([]byte(opsArray))
	if err != nil {
		return "", err
	}
	return template.HTML(byteAry), nil
}

func ConvertQuillToHtmlIgnoreError(quillText string) template.HTML {
	html, err := ConvertQuillToHtml(quillText)
	if err != nil {
		return template.HTML("")
	}
	return html
}
