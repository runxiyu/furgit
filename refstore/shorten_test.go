package refstore_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/refstore"
)

func TestShortenName(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		got := refstore.ShortenName("refs/heads/main", []string{"refs/heads/main"})
		if got != "main" {
			t.Fatalf("ShortenName simple = %q, want %q", got, "main")
		}
	})

	t.Run("ambiguous with tags", func(t *testing.T) {
		t.Parallel()

		got := refstore.ShortenName(
			"refs/heads/main",
			[]string{
				"refs/heads/main",
				"refs/tags/main",
			},
		)
		if got != "heads/main" {
			t.Fatalf("ShortenName tags ambiguity = %q, want %q", got, "heads/main")
		}
	})

	t.Run("strict remote head ambiguity", func(t *testing.T) {
		t.Parallel()
		// In strict mode, refs/remotes/%s/HEAD blocks shortening to "%s".
		got := refstore.ShortenName(
			"refs/heads/main",
			[]string{
				"refs/heads/main",
				"refs/remotes/main/HEAD",
			},
		)
		if got != "heads/main" {
			t.Fatalf("ShortenName strict ambiguity = %q, want %q", got, "heads/main")
		}
	})

	t.Run("deep fallback still shortens", func(t *testing.T) {
		t.Parallel()
		// refs/remotes/origin/main conflicts with refs/heads/origin/main for
		// "origin/main", so it should fall back to "remotes/origin/main".
		got := refstore.ShortenName(
			"refs/remotes/origin/main",
			[]string{
				"refs/remotes/origin/main",
				"refs/heads/origin/main",
			},
		)
		if got != "remotes/origin/main" {
			t.Fatalf("ShortenName deep fallback = %q, want %q", got, "remotes/origin/main")
		}
	})

	t.Run("refs-prefix fallback", func(t *testing.T) {
		t.Parallel()

		name := "refs/notes/review/topic"

		got := refstore.ShortenName(name, []string{name})
		if got != "notes/review/topic" {
			t.Fatalf("ShortenName refs-prefix fallback = %q, want %q", got, "notes/review/topic")
		}
	})
}
