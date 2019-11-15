package utils

import (
	"html"
	"regexp"
)

// FormatMessage - Форматирование текста в около html
func FormatMessage(message string) (string, error) {
	message = EscapeString(message)
	message = Nl2br(message)
	message = ReduceNewLines(message)
	return message, nil
}

// EscapeString - Why not?
func EscapeString(str string) string {
	return html.EscapeString(str)
}

// Nl2br - Change new line to br
func Nl2br(str string) string {
	re := regexp.MustCompile(`\r?\n`)
	return re.ReplaceAllString(str, "<br>")
	// Hz chto lutshe
	// return strings.Replace(str, "\n", "<br />", -1)
}

// ReduceNewLines - Reduce line breaks
func ReduceNewLines(str string) string {
	re := regexp.MustCompile(`(<br(?: \/)?>\s*){3,}`)
	return re.ReplaceAllString(str, "<br><br>")
}
