package encoding

import "testing"

func TestBaseFiveExample(t *testing.T) {
	got, err := DecimalToBase(72, 5)
	if err != nil {
		t.Fatal(err)
	}
	if got != "242" {
		t.Fatalf("expected 242, got %s", got)
	}
	back, err := BaseToDecimal(got, 5)
	if err != nil || back != 72 {
		t.Fatalf("round trip failed: %d %v", back, err)
	}
}
func TestPayloadRoundTrip(t *testing.T) {
	tokens, binary, err := EncodePayload("Hello world!", 14)
	if err != nil {
		t.Fatal(err)
	}
	decodedTokens, err := DeserializeTokens(binary)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != len(decodedTokens) {
		t.Fatal("token count changed")
	}
	payload, err := DecodePayload(decodedTokens, 14)
	if err != nil {
		t.Fatal(err)
	}
	if payload != "Hello world!" {
		t.Fatalf("got %q", payload)
	}
}
