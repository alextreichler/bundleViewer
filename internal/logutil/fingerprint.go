package logutil

import (
	"strings"
)

// FingerprinterOverride can be set by other packages (like internal/parser) to use
// an accelerated implementation if available.
var FingerprinterOverride func(string) string

// GenerateFingerprint creates a generic signature for a log message.
// It uses a token-aware approach inspired by the Drain algorithm, splitting by common
// log delimiters and masking tokens that contain variable data (digits, hex, etc.)
func GenerateFingerprint(message string) string {
	if FingerprinterOverride != nil {
		return FingerprinterOverride(message)
	}
	return defaultGenerateFingerprint(message)
}

func defaultGenerateFingerprint(message string) string {
	if len(message) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(message))

	n := len(message)
	for i := 0; i < n; i++ {
		c := message[i]

		// 1. Quoted Strings: "..." or '...'
		if c == '"' || c == '\'' {
			quote := c
			b.WriteString("<STR>")
			i++
			// Skip until closing quote or end
			for i < n && message[i] != quote {
				i++
			}
			continue
		}

		// 2. Delimiters (Space and common Punctuation)
		if isDelimiter(c) {
			b.WriteByte(c)
			continue
		}

		// 3. Token/Word
		start := i
		for i < n && !isDelimiter(message[i]) {
			i++
		}
		token := message[start:i]

		// Identify if this token is a variable part
		if marker, isVar := getVariableMarker(token); isVar {
			b.WriteString(marker)
		} else {
			b.WriteString(token)
		}
		
		// Back up one since the outer loop increments
		i--
	}
	return b.String()
}

func isDelimiter(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '[', ']', '{', '}', '(', ')', ',', ':', '=', ';', '<', '>', '/', '\\', '|', '-', '#':
		return true
	default:
		return false
	}
}

func getVariableMarker(s string) (string, bool) {
	if len(s) == 0 {
		return "", false
	}

	// Hex check
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return "<HEX>", true
	}

	// Pure numeric check (including decimals/negatives if they aren't delimiters)
	isNumeric := true
	hasDigit := false
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			hasDigit = true
		} else if s[i] == '.' {
			// fine
		} else {
			isNumeric = false
		}
	}

	if isNumeric && hasDigit {
		return "<NUM>", true
	}

	if hasDigit {
		return "<*>", true
	}

	return "", false
}
