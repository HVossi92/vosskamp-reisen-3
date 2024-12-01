package helpers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	quill "github.com/dchenk/go-render-quill"
)

func ConvertQuillToHtml(quillText string) (template.HTML, error) {
	// First verify we have a valid JSON string
	if !strings.Contains(quillText, `"ops":[`) {
		return "", fmt.Errorf("invalid Quill delta format: missing ops array")
	}

	// Find the start of the ops array
	start := strings.Index(quillText, `"ops":[`)
	if start == -1 {
		return "", fmt.Errorf("couldn't find ops array start")
	}
	start += len(`"ops":[`)

	// Find the end of the ops array
	end := strings.Index(quillText[start:], `]`)
	if end == -1 {
		return "", fmt.Errorf("couldn't find ops array end")
	}
	end += start // Adjust end position relative to full string

	// Verify the slice bounds
	if start > end {
		return "", fmt.Errorf("invalid ops array bounds: start(%d) > end(%d)", start, end)
	}

	// Extract the ops array
	opsArray := "[" + quillText[start:end] + "]"

	// Attempt to render the Quill content
	byteAry, err := quill.Render([]byte(opsArray))
	if err != nil {
		return "", fmt.Errorf("failed to render Quill content: %w", err)
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

func ConvertQuillToPlainText(quillText string) (string, error) {
	// First verify we have a valid JSON string
	if !strings.Contains(quillText, `"ops":[`) {
		return "", fmt.Errorf("invalid Quill delta format: missing ops array")
	}

	// Find the start of the ops array
	start := strings.Index(quillText, `"ops":[`)
	if start == -1 {
		return "", fmt.Errorf("couldn't find ops array start")
	}
	start += len(`"ops":[`)

	// Find the end of the ops array
	end := strings.Index(quillText[start:], `]`)
	if end == -1 {
		return "", fmt.Errorf("couldn't find ops array end")
	}
	end += start // Adjust end position relative to full string

	// Verify the slice bounds
	if start > end {
		return "", fmt.Errorf("invalid ops array bounds: start(%d) > end(%d)", start, end)
	}

	// Extract the ops array
	opsArray := "[" + quillText[start:end] + "]"

	// Parse the JSON array into a slice of operations
	var ops []map[string]interface{}
	if err := json.Unmarshal([]byte(opsArray), &ops); err != nil {
		return "", fmt.Errorf("failed to parse Quill ops: %w", err)
	}

	// Build plain text by concatenating insert operations
	var result strings.Builder
	for _, op := range ops {
		if insert, ok := op["insert"].(string); ok {
			result.WriteString(insert)
		}
	}

	if len(result.String()) > 256 {
		return result.String()[:256] + "...", nil
	}
	return result.String(), nil
}
