package encoding

import (
	"fmt"
	"strings"
)

func SerializeTokens(tokens []string) string {
	framed := strings.Join(tokens, ",")
	parts := make([]string, len([]byte(framed)))
	for i, b := range []byte(framed) {
		parts[i] = fmt.Sprintf("%08b", b)
	}
	return strings.Join(parts, " ")
}
