package encoding

func TextToBytes(text string) []byte {
	return []byte(text)
}

func BytesToText(values []byte) string {
	return string(values)
}
