package encoding

import "fmt"

func EncodePayload(payload string, targetCodex int) ([]string, string, error) {
	bytes := TextToBytes(payload)
	tokens := make([]string, len(bytes))
	for i, b := range bytes {
		token, err := DecimalToBase(int(b), targetCodex)
		if err != nil {
			return nil, "", err
		}
		tokens[i] = token
	}
	return tokens, SerializeTokens(tokens), nil
}

func DecodePayload(tokens []string, sourceCodex int) (string, error) {
	bytes := make([]byte, len(tokens))
	for i, token := range tokens {
		value, err := BaseToDecimal(token, sourceCodex)
		if err != nil {
			return "", fmt.Errorf("token %d: %w", i, err)
		}
		bytes[i] = byte(value)
	}
	return BytesToText(bytes), nil
}

func DecodeBinary(binary string, sourceCodex int) (string, []string, error) {
	tokens, err := DeserializeTokens(binary)
	if err != nil {
		return "", nil, err
	}
	payload, err := DecodePayload(tokens, sourceCodex)
	if err != nil {
		return "", nil, err
	}
	return payload, tokens, nil
}
