package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// GenerateID creates a unique ID for a news item
func GenerateID(source, content string) string {
	h := sha256.New()
	h.Write([]byte(source))
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// CleanContent removes extra whitespace and normalizes text
func CleanContent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	return s
}
