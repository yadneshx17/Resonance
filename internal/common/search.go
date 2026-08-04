package common

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func FuzzyMatch(query, s string) bool {
	query = strings.ToLower(query)
	s = strings.ToLower(s)
	queryIdx := 0
	for _, ch := range s {
		if queryIdx < len(query) && byte(ch) == query[queryIdx] {
			queryIdx++
		}
		for queryIdx < len(query) && query[queryIdx] == ' ' {
			queryIdx++
		}
	}
	// for queryIdx < len(query) && query[queryIdx] == ' ' {
	// 	queryIdx++
	// }
	return queryIdx == len(query)
}

func SearchBar(query string, width int) string {
	bar := Search + " " + query + "█"
	if w := lipgloss.Width(bar); w < width {
		bar += strings.Repeat(" ", width-w)
	}
	return bar
}
