package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// Generator creates short random identifiers with a stable prefix.
type Generator struct{}

// New returns a prefix-scoped identifier suitable for database records.
func (Generator) New(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "_fallback"
	}

	var builder strings.Builder
	builder.Grow(len(prefix) + 1 + hex.EncodedLen(len(buf)))
	builder.WriteString(prefix)
	builder.WriteByte('_')
	builder.WriteString(hex.EncodeToString(buf))
	return builder.String()
}
