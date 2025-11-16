package furgit

import (
	"strings"
	"testing"
)

func TestParseIdentRoundTrip(t *testing.T) {
	line := []byte("Alice Example <alice@example.com> 1700000000 -0700")
	id, err := parseIdent(line)
	if err != nil {
		t.Fatalf("parseIdent error: %v", err)
	}
	if got := string(id.Email); got != "alice@example.com" {
		t.Fatalf("email mismatch: %q", got)
	}
	ids, err := id.Serialize()
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	serialized := string(ids)
	if !strings.Contains(serialized, "alice@example.com") {
		t.Fatalf("Serialize missing email: %q", serialized)
	}
	when := id.When()
	if when.Unix() != 1700000000 {
		t.Fatalf("When unix mismatch: %d", when.Unix())
	}
	if _, offset := when.Zone(); offset != -7*3600 {
		t.Fatalf("When offset mismatch: %d", offset)
	}
}

func TestParseIdentInvalidInputs(t *testing.T) {
	cases := []string{
		"MissingEmail 1700000000 +0000",
		"Name <email> notanumber +0000",
		"Name <email> 1700000000 123",
	}
	for _, tc := range cases {
		if _, err := parseIdent([]byte(tc)); err == nil {
			t.Fatalf("expected error for %q", tc)
		}
	}
}

func TestIdentSerializeUsesCanonicalSpacing(t *testing.T) {
	id := Ident{
		Name:          []byte("Bob"),
		Email:         []byte("bob@example.com"),
		WhenUnix:      1000,
		OffsetMinutes: 90,
	}
	ids, err := id.Serialize()
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	got := string(ids)
	if !strings.Contains(got, "Bob <bob@example.com>") {
		t.Fatalf("unexpected serialize output: %q", got)
	}
	if !strings.HasSuffix(got, "+0130") {
		t.Fatalf("expected timezone in +0130 form: %q", got)
	}
	loc := id.When()
	if loc.Unix() != 1000 {
		t.Fatalf("When unix mismatch: %d", loc.Unix())
	}
	if _, offset := loc.Zone(); offset != 90*60 {
		t.Fatalf("When offset mismatch: %d", offset)
	}
}
