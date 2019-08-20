package utils

import (
	"regexp"
)

func ParseMessage(message string) (string, error) {
	//$message = preg_replace('#(<br(?: \/)?>\s*){3,}#i', '<br /><br />', $message);

	//s := strings.Replace(message, "\n", "\r\n", -1)

	re := regexp.MustCompile(`\r?\n`)
	input := re.ReplaceAllString(message, "<br>")

	return input, nil
}
