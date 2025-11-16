package furgit

import (
	"bytes"
	"testing"
)

func TestIdentSerialize(t *testing.T) {
	tests := []struct {
		name  string
		ident Ident
	}{
		{
			name: "positive offset",
			ident: Ident{
				Name:          []byte("John Doe"),
				Email:         []byte("john@example.org"),
				WhenUnix:      1234567890,
				OffsetMinutes: 120,
			},
		},
		{
			name: "negative offset",
			ident: Ident{
				Name:          []byte("Jane Smith"),
				Email:         []byte("jane@example.org"),
				WhenUnix:      9876543210,
				OffsetMinutes: -300,
			},
		},
		{
			name: "zero offset",
			ident: Ident{
				Name:          []byte("UTC User"),
				Email:         []byte("utc@example.org"),
				WhenUnix:      1000000000,
				OffsetMinutes: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serialized, err := tt.ident.Serialize()
			if err != nil {
				t.Fatalf("Serialize failed: %v", err)
			}

			parsed, err := parseIdent(serialized)
			if err != nil {
				t.Fatalf("parseIdent failed: %v", err)
			}

			if !bytes.HasPrefix(parsed.Name, tt.ident.Name) {
				t.Errorf("name: got %q, want prefix %q", parsed.Name, tt.ident.Name)
			}
			if !bytes.Equal(parsed.Email, tt.ident.Email) {
				t.Errorf("email: got %q, want %q", parsed.Email, tt.ident.Email)
			}
			if parsed.WhenUnix != tt.ident.WhenUnix {
				t.Errorf("whenUnix: got %d, want %d", parsed.WhenUnix, tt.ident.WhenUnix)
			}
			if parsed.OffsetMinutes != tt.ident.OffsetMinutes {
				t.Errorf("offsetMinutes: got %d, want %d", parsed.OffsetMinutes, tt.ident.OffsetMinutes)
			}

			when := tt.ident.When()
			if when.Unix() != tt.ident.WhenUnix {
				t.Errorf("When().Unix(): got %d, want %d", when.Unix(), tt.ident.WhenUnix)
			}
		})
	}
}
