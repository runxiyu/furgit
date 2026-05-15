package id_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/object/id"
)

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
