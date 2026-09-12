package clubname

import "strings"

// NormalizeKey returns a compact lookup key so variants like
// "ROTARACT IUGB", "Rotaract IUGB Club", and "UUGB" can match.
func NormalizeKey(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, "uugb", "iugb")

	for _, word := range []string{"rotaract", "club", "de", "du", "la", "le", "les", "rotary"} {
		s = strings.ReplaceAll(s, word, "")
	}

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
