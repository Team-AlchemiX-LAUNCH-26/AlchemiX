package encoding

import (
	"fmt"
	"strconv"
	"strings"
)

func DeserializeTokens(binary string) ([]string, error) {
	binary = strings.TrimSpace(binary)
	if binary == "" {
		return nil, nil
	}
	bitGroups := strings.Fields(binary)
	decoded := make([]byte, len(bitGroups))
	for i, group := range bitGroups {
		if len(group) != 8 {
			return nil, fmt.Errorf("binary group %d has length %d, expected 8", i, len(group))
		}
		value, err := strconv.ParseUint(group, 2, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid binary group %q: %w", group, err)
		}
		decoded[i] = byte(value)
	}
	framed := string(decoded)
	if framed == "" {
		return nil, nil
	}
	return strings.Split(framed, ","), nil
}
