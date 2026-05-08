package validator

import "strings"

func HasTextOrImage(content string, imageURL string) bool {
	return strings.TrimSpace(content) != "" || strings.TrimSpace(imageURL) != ""
}
