package id_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/object/id"
)

func TestAlgorithmEmptyTree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		algo id.Algorithm
		want string
	}{
		{
			name: "sha1",
			algo: id.AlgorithmSHA1,
			want: "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
		},
		{
			name: "sha256",
			algo: id.AlgorithmSHA256,
			want: "6ef19b41225c5369f1c104d45d8d85efa9b057b53b14b4b9b939dd74decc5321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.algo.EmptyTree()

			if got.String() != tt.want {
				t.Fatalf("EmptyTree() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestParseAlgorithm(t *testing.T) {
	t.Parallel()

	algo, ok := id.ParseAlgorithm("sha1")
	if !ok || algo != id.AlgorithmSHA1 {
		t.Fatalf("ParseAlgorithm(sha1) = (%v,%v)", algo, ok)
	}

	algo, ok = id.ParseAlgorithm("sha256")
	if !ok || algo != id.AlgorithmSHA256 {
		t.Fatalf("ParseAlgorithm(sha256) = (%v,%v)", algo, ok)
	}

	if _, ok := id.ParseAlgorithm("md5"); ok {
		t.Fatalf("ParseAlgorithm(md5) should fail")
	}
}

func TestAlgorithmSum(t *testing.T) {
	t.Parallel()

	id1 := id.AlgorithmSHA1.Sum([]byte("hello"))
	if id1.Algorithm() != id.AlgorithmSHA1 || id1.Algorithm().Size() != id.AlgorithmSHA1.Size() {
		t.Fatalf("sha1 sum produced invalid object id")
	}

	id2 := id.AlgorithmSHA256.Sum([]byte("hello"))
	if id2.Algorithm() != id.AlgorithmSHA256 || id2.Algorithm().Size() != id.AlgorithmSHA256.Size() {
		t.Fatalf("sha256 sum produced invalid object id")
	}

	if id1.String() == id2.String() {
		t.Fatalf("sha1 and sha256 should differ")
	}
}

func TestUnknownAlgorithmEmptyTree(t *testing.T) {
	t.Parallel()

	got := id.AlgorithmUnknown.EmptyTree()
	if got != (id.ObjectID{}) {
		t.Fatalf("EmptyTree() for unknown algorithm = %#v, want zero value", got)
	}
}
