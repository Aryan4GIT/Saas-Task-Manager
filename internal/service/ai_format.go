package service

import "strings"

func sanitizeAIReport(s string) string {
	if s == "" {
		return s
	}

	// Remove common "noisy" punctuation while keeping periods/commas/dashes for readability.
	replacer := strings.NewReplacer(
		"!", "",
		"?", "",
		";", "",
		"…", "",
		"\u2026", "",
		"*", "",
	)

	out := replacer.Replace(s)
	out = strings.ReplaceAll(out, "\r\n", "\n")
	out = strings.TrimSpace(out)
	return out
}
