// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zlib_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"codeberg.org/lindenii/furgit/internal/compress/zlib"
)

var filenames = []string{
	"../testdata/gettysburg.txt",
	"../testdata/e.txt",
	"../testdata/pi.txt",
}

var data = []string{
	"test a reasonable sized string that can be compressed",
}

// Tests that compressing and then decompressing the given file at the given compression level and dictionary
// yields equivalent bytes to the original file.
func testFileLevelDict(t *testing.T, fn string, level int, d string) {
	t.Helper()

	// Read the file, as golden output.
	golden, err := os.Open(fn)
	if err != nil {
		t.Errorf("%s (level=%d, dict=%q): %v", fn, level, d, err)

		return
	}

	defer func() {
		err := golden.Close()
		if err != nil {
			t.Fatalf("%s (level=%d, dict=%q): close golden: %v", fn, level, d, err)
		}
	}()

	b0, err0 := io.ReadAll(golden)
	if err0 != nil {
		t.Errorf("%s (level=%d, dict=%q): %v", fn, level, d, err0)

		return
	}

	testLevelDict(t, fn, b0, level, d)
}

func testLevelDict(t *testing.T, fn string, b0 []byte, level int, d string) {
	t.Helper()

	// Make dictionary, if given.
	var dict []byte
	if d != "" {
		dict = []byte(d)
	}

	// Push data through a pipe that compresses at the write end, and decompresses at the read end.
	piper, pipew := io.Pipe()

	defer func() {
		err := piper.Close()
		if err != nil {
			t.Fatalf("%s (level=%d, dict=%q): close piper: %v", fn, level, d, err)
		}
	}()

	go func() {
		defer func() {
			err := pipew.Close()
			if err != nil {
				t.Errorf("%s (level=%d, dict=%q): close pipew: %v", fn, level, d, err)
			}
		}()

		zlibw, err := zlib.NewWriterLevelDict(pipew, level, dict)
		if err != nil {
			t.Errorf("%s (level=%d, dict=%q): %v", fn, level, d, err)

			return
		}

		defer func() {
			err := zlibw.Close()
			if err != nil {
				t.Errorf("%s (level=%d, dict=%q): close zlibw: %v", fn, level, d, err)
			}
		}()

		_, err = zlibw.Write(b0)
		if err != nil {
			t.Errorf("%s (level=%d, dict=%q): %v", fn, level, d, err)

			return
		}
	}()

	zlibr, err := zlib.NewReaderDict(piper, dict)
	if err != nil {
		t.Errorf("%s (level=%d, dict=%q): %v", fn, level, d, err)

		return
	}

	defer func() {
		err := zlibr.Close()
		if err != nil {
			t.Fatalf("%s (level=%d, dict=%q): close zlibr: %v", fn, level, d, err)
		}
	}()

	// Compare the decompressed data.
	b1, err1 := io.ReadAll(zlibr)
	if err1 != nil {
		t.Errorf("%s (level=%d, dict=%q): %v", fn, level, d, err1)

		return
	}

	if len(b0) != len(b1) {
		t.Errorf("%s (level=%d, dict=%q): length mismatch %d versus %d", fn, level, d, len(b0), len(b1))

		return
	}

	for i := range b0 {
		if b0[i] != b1[i] {
			t.Errorf("%s (level=%d, dict=%q): mismatch at %d, 0x%02x versus 0x%02x\n", fn, level, d, i, b0[i], b1[i])

			return
		}
	}
}

func TestWriter(t *testing.T) {
	t.Parallel()

	for i, s := range data {
		b := []byte(s)
		tag := fmt.Sprintf("#%d", i)
		testLevelDict(t, tag, b, zlib.DefaultCompression, "")
		testLevelDict(t, tag, b, zlib.NoCompression, "")
		testLevelDict(t, tag, b, zlib.HuffmanOnly, "")

		for level := zlib.BestSpeed; level <= zlib.BestCompression; level++ {
			testLevelDict(t, tag, b, level, "")
		}
	}
}

func TestWriterBig(t *testing.T) {
	t.Parallel()

	for i, fn := range filenames {
		testFileLevelDict(t, fn, zlib.DefaultCompression, "")
		testFileLevelDict(t, fn, zlib.NoCompression, "")
		testFileLevelDict(t, fn, zlib.HuffmanOnly, "")

		for level := zlib.BestSpeed; level <= zlib.BestCompression; level++ {
			testFileLevelDict(t, fn, level, "")

			if level >= 1 && testing.Short() {
				break
			}
		}

		if i == 0 && testing.Short() {
			break
		}
	}
}

func TestWriterDict(t *testing.T) {
	t.Parallel()

	const dictionary = "0123456789."
	for i, fn := range filenames {
		testFileLevelDict(t, fn, zlib.DefaultCompression, dictionary)
		testFileLevelDict(t, fn, zlib.NoCompression, dictionary)
		testFileLevelDict(t, fn, zlib.HuffmanOnly, dictionary)

		for level := zlib.BestSpeed; level <= zlib.BestCompression; level++ {
			testFileLevelDict(t, fn, level, dictionary)

			if level >= 1 && testing.Short() {
				break
			}
		}

		if i == 0 && testing.Short() {
			break
		}
	}
}

func TestWriterDictIsUsed(t *testing.T) {
	t.Parallel()

	var (
		input = []byte("Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
		buf   bytes.Buffer
	)

	compressor, err := zlib.NewWriterLevelDict(&buf, zlib.BestCompression, input)
	if err != nil {
		t.Errorf("error in NewWriterLevelDict: %s", err)

		return
	}

	_, err = compressor.Write(input)
	if err != nil {
		t.Errorf("error in compressor.Write: %s", err)

		return
	}

	err = compressor.Close()
	if err != nil {
		t.Errorf("error in compressor.Close: %s", err)

		return
	}

	const expectedMaxSize = 25

	output := buf.Bytes()
	if len(output) > expectedMaxSize {
		t.Errorf("result too large (got %d, want <= %d bytes). Is the dictionary being used?", len(output), expectedMaxSize)
	}
}
