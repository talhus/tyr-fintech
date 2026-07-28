package utils

import "strings"

func MaskString(s string) string {
	runes := []rune(s)

	if len(runes) <= 2 {
		return string(runes)
	}

	visible := string(runes[:2])
	masked := strings.Repeat("*", len(runes)-2)
	return visible + masked
}
