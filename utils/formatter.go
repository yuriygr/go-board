package utils

import (
	"html"
	"regexp"
	"strings"
)

const (
	reNewLines       = `\r?\n`
	reReduceNewLines = `(<br(?: \/)?>\s*){3,}`
	reURL            = `(http|ftp|https):\/\/([\w\p{L}\-_]+(?:(?:\.[\w\p{L}\-_]+)+))([\w\p{L}\-\.,@?^=%&amp;:/~\+#]*[\w\p{L}\-\@?^=%&amp;/~\+#])?`
	reTags           = `#[A-Za-z0-9\_]*` // TODO: Make it better
)

// FormatMessage - Форматирование текста в около html
func FormatMessage(str string) (string, error) {
	str = EscapeString(str)
	str = Nl2br(str)
	str = ReduceNewLines(str)
	str = MarkupURLs(str)
	str = MarkupHashtags(str)
	str = Markup(str)
	return str, nil
}

// EscapeString - Why not?
func EscapeString(str string) string {
	return html.EscapeString(str)
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

// MarkupURLs - Долгожданная функция
func MarkupURLs(str string) string {
	re := regexp.MustCompile(reURL)
	return re.ReplaceAllString(str, `<a href="$0">$0</a>`)
}

// ExtractHashtags - Извлекает список всех хештегов
// и возвращает _уникальный_ массив тегов.
// В разметке поста не используется.
func ExtractHashtags(str string) []string {
	re := regexp.MustCompile(reTags)
	tags := re.FindAllString(str, -1)

	findedTags := map[string]bool{}
	for _, tag := range tags {
		findedTags[strings.TrimLeft(tag, "#")] = true
	}

	uniqueTags := []string{}
	for tag := range findedTags {
		uniqueTags = append(uniqueTags, tag)
	}

	return uniqueTags
}

// MarkupHashtags - Форматирование хештегов
func MarkupHashtags(str string) string {
	return str
}

// Markup - Форматирование разметки
// a.k.a. корень всего зла
func Markup(str string) string {
	return str
}
