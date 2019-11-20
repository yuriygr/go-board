package utils

import (
	"html"
	"regexp"
)

const (
	reNewLines       = `\r?\n`
	reReduceNewLines = `(<br(?: \/)?>\s*){3,}`
	reURL            = `(http|ftp|https):\/\/([\w\p{L}\-_]+(?:(?:\.[\w\p{L}\-_]+)+))([\w\p{L}\-\.,@?^=%&amp;:/~\+#]*[\w\p{L}\-\@?^=%&amp;/~\+#])?`
)

// FormatMessage - Форматирование текста в около html
func FormatMessage(str string) (string, error) {
	str = EscapeString(str)
	str = MakeURL(str)
	str = Nl2br(str)
	str = ReduceNewLines(str)
	return str, nil
}

// EscapeString - Why not?
func EscapeString(str string) string {
	return html.EscapeString(str)
}

// MakeURL - Долгожданная функция
func MakeURL(str string) string {
	re := regexp.MustCompile(reURL)
	return re.ReplaceAllString(str, `<a href="$0">$0</a>`)
}

// Nl2br - Change new line to br
func Nl2br(str string) string {
	re := regexp.MustCompile(reNewLines)
	return re.ReplaceAllString(str, `<br>`)
	// Hz chto lutshe
	// return strings.Replace(str, "\n", "<br />", -1)
}

// ReduceNewLines - Reduce line breaks
func ReduceNewLines(str string) string {
	re := regexp.MustCompile(reReduceNewLines)
	return re.ReplaceAllString(str, `<br><br>`)
}

// Markup - Форматирование разметки
// a.k.a. корень всего зла
func Markup(str string) string {
	return str
}
