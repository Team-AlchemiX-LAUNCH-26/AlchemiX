package packet

import "fmt"

func VerifyPayload(original, decoded string) error {
	if original != decoded {
		return fmt.Errorf("payload integrity check failed: decoded payload does not match original")
	}
	return nil
}
