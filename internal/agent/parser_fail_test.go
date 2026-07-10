package agent

import "testing"

func newTestParser() *DeterministicParser {
	return NewParser([]string{
		"Aegis",
		"Boreas",
		"Dawn",
		"Elysium",
		"Fenix",
		"Caelum",
	})
}

func TestDeterministicParser_Parse_CoreCases(t *testing.T) {
	parser := newTestParser()

	tests := []struct {
		name        string
		raw         string
		origin      string
		destination string
		payload     string
		wantErr     bool
	}{
		{
			name:        "standard quoted payload",
			raw:         `Send "Hello" from Aegis to Caelum`,
			origin:      "Aegis",
			destination: "Caelum",
			payload:     "Hello",
			wantErr:     false,
		},
		{
			name:        "payload contains destination keyword and planet name",
			raw:         `Send "go to Dawn" from Aegis to Caelum`,
			origin:      "Aegis",
			destination: "Caelum",
			payload:     "go to Dawn",
			wantErr:     false,
		},
		{
			name:        "payload contains origin keyword and planet name",
			raw:         `Send "from Dawn" from Aegis to Caelum`,
			origin:      "Aegis",
			destination: "Caelum",
			payload:     "from Dawn",
			wantErr:     false,
		},
		{
			name:        "input contains word starting with to",
			raw:         `Today send "Hello" from Aegis to Caelum`,
			origin:      "Aegis",
			destination: "Caelum",
			payload:     "Hello",
			wantErr:     false,
		},
		{
			name:        "lowercase planets should parse",
			raw:         `send "Hello" from aegis to caelum`,
			origin:      "Aegis",
			destination: "Caelum",
			payload:     "Hello",
			wantErr:     false,
		},
		{
			name:        "punctuation after planets should parse",
			raw:         `Send "Hello" from Aegis, to Caelum.`,
			origin:      "Aegis",
			destination: "Caelum",
			payload:     "Hello",
			wantErr:     false,
		},
		{
			name:    "missing origin should fail",
			raw:     `Send "Hello" to Caelum`,
			wantErr: true,
		},
		{
			name:    "ambiguous origin should fail",
			raw:     `Send "Hello" from Aegis and Dawn to Caelum`,
			wantErr: true,
		},
		{
			name:    "missing payload should fail",
			raw:     `Send from Aegis to Caelum`,
			wantErr: true,
		},
		{
			name:    "empty request should fail",
			raw:     ``,
			wantErr: true,
		},
		{
			name:    "whitespace request should fail",
			raw:     `        `,
			wantErr: true,
		},
		{
			name:    "unknown origin should fail",
			raw:     `Send "Hello" from Earth to Caelum`,
			wantErr: true,
		},
		{
			name:    "unknown destination should fail",
			raw:     `Send "Hello" from Aegis to Earth`,
			wantErr: true,
		},
		{
			name:    "same origin and destination should fail",
			raw:     `Send "Hello" from Aegis to Aegis`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.raw)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil result: %+v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.OriginID != tt.origin {
				t.Fatalf("origin mismatch: got %q, want %q", got.OriginID, tt.origin)
			}

			if got.DestinationID != tt.destination {
				t.Fatalf("destination mismatch: got %q, want %q", got.DestinationID, tt.destination)
			}

			if got.Payload != tt.payload {
				t.Fatalf("payload mismatch: got %q, want %q", got.Payload, tt.payload)
			}
		})
	}
}

func TestDeterministicParser_Parse_AllPlanetPairs(t *testing.T) {
	planets := []string{
		"Aegis",
		"Boreas",
		"Dawn",
		"Elysium",
		"Fenix",
		"Caelum",
	}

	parser := NewParser(planets)

	for _, origin := range planets {
		for _, destination := range planets {
			if origin == destination {
				continue
			}

			raw := `Send "Hello" from ` + origin + ` to ` + destination

			t.Run(origin+"_to_"+destination, func(t *testing.T) {
				got, err := parser.Parse(raw)
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", raw, err)
				}

				if got.OriginID != origin {
					t.Fatalf("origin mismatch: got %q, want %q", got.OriginID, origin)
				}

				if got.DestinationID != destination {
					t.Fatalf("destination mismatch: got %q, want %q", got.DestinationID, destination)
				}

				if got.Payload != "Hello" {
					t.Fatalf("payload mismatch: got %q, want %q", got.Payload, "Hello")
				}
			})
		}
	}
}

func TestDeterministicParser_Parse_RejectSameOriginDestinationForAllPlanets(t *testing.T) {
	planets := []string{
		"Aegis",
		"Boreas",
		"Dawn",
		"Elysium",
		"Fenix",
		"Caelum",
	}

	parser := NewParser(planets)

	for _, planet := range planets {
		raw := `Send "Hello" from ` + planet + ` to ` + planet

		t.Run(planet+"_to_itself", func(t *testing.T) {
			_, err := parser.Parse(raw)
			if err == nil {
				t.Fatalf("expected error for same origin/destination: %q", raw)
			}
		})
	}
}

func TestDeterministicParser_Parse_RejectAmbiguousDestinations(t *testing.T) {
	parser := newTestParser()

	tests := []string{
		`Send "Hello" from Aegis to Caelum and Dawn`,
		`Send "Hello" from Aegis to Caelum or Fenix`,
		`Send "Hello" from Aegis to Caelum, not Dawn`,
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			_, err := parser.Parse(raw)
			if err == nil {
				t.Fatalf("expected ambiguous destination error for %q", raw)
			}
		})
	}
}

func TestDeterministicParser_Parse_RejectPlanetsOnlyInsidePayload(t *testing.T) {
	parser := newTestParser()

	tests := []string{
		`Send "Report about Dawn" to Caelum`,
		`Send "Message from Aegis to Caelum"`,
		`Send "Caelum Dawn Aegis"`,
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			_, err := parser.Parse(raw)
			if err == nil {
				t.Fatalf("expected error because route planets are missing or only inside payload: %q", raw)
			}
		})
	}
}
