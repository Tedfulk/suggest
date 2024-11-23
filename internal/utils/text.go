package utils

import (
	"regexp"
)

func TruncateText(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length-3] + "..."
}

func ExtractVariables(content string) []string {
	var vars []string
	varMap := make(map[string]bool)

	matches := regexp.MustCompile(`\[(.*?)\]`).FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 && !varMap[match[1]] {
			vars = append(vars, match[1])
			varMap[match[1]] = true
		}
	}
	return vars
} 