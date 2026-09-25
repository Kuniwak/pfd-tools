package pfddrawio

import (
	"slices"
	"strings"
)

func StyleTokens(style string) []string {
	var tokens []string
	for _, token := range strings.Split(style, ";") {
		if token == "" {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func StyleTokenName(token string) string {
	name, _, _ := strings.Cut(token, "=")
	return name
}

func WithStyleToken(tokens []string, name, value string) []string {
	token := name + "=" + value

	replaced := false
	result := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		if StyleTokenName(t) != name {
			result = append(result, t)
			continue
		}
		if replaced {
			continue
		}
		result = append(result, token)
		replaced = true
	}
	if !replaced {
		result = append(result, token)
	}
	return result
}

func WithoutStyleTokens(tokens []string, names ...string) []string {
	var result []string
	for _, t := range tokens {
		if slices.Contains(names, StyleTokenName(t)) {
			continue
		}
		result = append(result, t)
	}
	return result
}

func FormatStyle(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, ";") + ";"
}

func StraightEdgeStyle(style string) string {
	tokens := StyleTokens(style)
	tokens = WithStyleToken(tokens, "edgeStyle", "none")
	tokens = WithStyleToken(tokens, "jumpStyle", "gap")
	tokens = WithoutStyleTokens(tokens,
		"entryX", "entryY", "entryDx", "entryDy", "entryPerimeter",
		"exitX", "exitY", "exitDx", "exitDy", "exitPerimeter",
	)
	return FormatStyle(tokens)
}
