package encoding

import (
	"fmt"
	"strconv"
	"strings"
)

func DecimalToBase(value, base int) (string, error) {
	if base < 2 || base > 36 {
		return "", fmt.Errorf("unsupported base %d; expected 2..36", base)
	}
	if value < 0 {
		return "", fmt.Errorf("negative values are not supported")
	}
	return strings.ToUpper(strconv.FormatInt(int64(value), base)), nil
}

func BaseToDecimal(value string, base int) (int, error) {
	if base < 2 || base > 36 {
		return 0, fmt.Errorf("unsupported base %d; expected 2..36", base)
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), base, 16)
	if err != nil {
		return 0, fmt.Errorf("parse %q as base %d: %w", value, base, err)
	}
	if parsed < 0 || parsed > 255 {
		return 0, fmt.Errorf("decoded value %d is outside a byte", parsed)
	}
	return int(parsed), nil
}
