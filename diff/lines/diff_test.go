package lines_test

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"lindenii.org/go/furgit/diff/lines"
)

func TestDiff(t *testing.T) { //nolint:maintidx
	t.Parallel()

	tests := []struct {
		name     string
		oldInput string
		newInput string
		expected []lines.Chunk
	}{
		{
			name:     "empty inputs produce no chunks",
			oldInput: "",
			newInput: "",
			expected: []lines.Chunk{},
		},
		{
			name:     "only additions",
			oldInput: "",
			newInput: "alpha\nbeta\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindAdded, Data: []byte("alpha\nbeta\n")},
			},
		},
		{
			name:     "only deletions",
			oldInput: "alpha\nbeta\n",
			newInput: "",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindDeleted, Data: []byte("alpha\nbeta\n")},
			},
		},
		{
			name:     "unchanged content is grouped",
			oldInput: "same\nlines\n",
			newInput: "same\nlines\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("same\nlines\n")},
			},
		},
		{
			name:     "insertion in the middle",
			oldInput: "a\nb\nc\n",
			newInput: "a\nb\nX\nc\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("a\nb\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("X\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("c\n")},
			},
		},
		{
			name:     "replacement without trailing newline",
			oldInput: "first\nsecond",
			newInput: "first\nsecond\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("first\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("second")},
				{Kind: lines.ChunkKindAdded, Data: []byte("second\n")},
			},
		},
		{
			name:     "line replacement",
			oldInput: "a\nb\nc\n",
			newInput: "a\nB\nc\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("a\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("b\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("B\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("c\n")},
			},
		},
		{
			name:     "swap adjacent lines",
			oldInput: "A\nB\n",
			newInput: "B\nA\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindDeleted, Data: []byte("A\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("B\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("A\n")},
			},
		},
		{
			name:     "indentation change is a full line replacement",
			oldInput: "func main() {\n\treturn\n}\n",
			newInput: "func main() {\n    return\n}\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("func main() {\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("\treturn\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("    return\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("}\n")},
			},
		},
		{
			name:     "commenting out lines",
			oldInput: "code\n",
			newInput: "// code\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindDeleted, Data: []byte("code\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("// code\n")},
			},
		},
		{
			name:     "reducing repeating lines",
			oldInput: "log\nlog\nlog\n",
			newInput: "log\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("log\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("log\nlog\n")},
			},
		},
		{
			name:     "expanding repeating lines",
			oldInput: "tick\n",
			newInput: "tick\ntick\ntick\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("tick\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("tick\ntick\n")},
			},
		},
		{
			name:     "interleaved modifications",
			oldInput: "keep\nchange\nkeep\nchange\n",
			newInput: "keep\nfixed\nkeep\nfixed\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("keep\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("change\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("fixed\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("keep\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("change\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("fixed\n")},
			},
		},
		{
			name:     "large common header and footer",
			oldInput: "header\nheader\nheader\nOLD\nfooter\nfooter\n",
			newInput: "header\nheader\nheader\nNEW\nfooter\nfooter\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("header\nheader\nheader\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("OLD\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("NEW\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("footer\nfooter\n")},
			},
		},
		{
			name:     "completely different content",
			oldInput: "apple\nbanana\n",
			newInput: "cherry\ndate\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindDeleted, Data: []byte("apple\nbanana\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("cherry\ndate\n")},
			},
		},
		{
			name:     "unicode and emoji changes",
			oldInput: "Hello 🌍\nYay\n",
			newInput: "Hello 🌎\nYay\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindDeleted, Data: []byte("Hello 🌍\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("Hello 🌎\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("Yay\n")},
			},
		},
		{
			name:     "binary data with embedded newlines",
			oldInput: "\x00\x01\n\x02\x03\n",
			newInput: "\x00\x01\n\x02\xFF\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("\x00\x01\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("\x02\x03\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("\x02\xFF\n")},
			},
		},
		{
			name:     "adding trailing newline to last line",
			oldInput: "Line 1\nLine 2",
			newInput: "Line 1\nLine 2\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("Line 1\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("Line 2")},
				{Kind: lines.ChunkKindAdded, Data: []byte("Line 2\n")},
			},
		},
		{
			name:     "removing trailing newline",
			oldInput: "A\nB\n",
			newInput: "A\nB",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("B\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("B")},
			},
		},
		{
			name:     "inserting blank lines",
			oldInput: "A\nB\n",
			newInput: "A\n\n\nB\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("\n\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("B\n")},
			},
		},
		{
			name:     "collapsing blank lines",
			oldInput: "A\n\n\n\nB\n",
			newInput: "A\nB\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("\n\n\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("B\n")},
			},
		},
		{
			name:     "case sensitivity check",
			oldInput: "FOO\nbar\n",
			newInput: "foo\nbar\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindDeleted, Data: []byte("FOO\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("foo\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("bar\n")},
			},
		},
		{
			name:     "partial line match is full mismatch",
			oldInput: "The quick brown fox\n",
			newInput: "The quick brown fox jumps\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindDeleted, Data: []byte("The quick brown fox\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("The quick brown fox jumps\n")},
			},
		},
		{
			name:     "inserting middle content",
			oldInput: "Top\nBottom\n",
			newInput: "Top\nMiddle\nBottom\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("Top\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("Middle\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("Bottom\n")},
			},
		},
		{
			name:     "block move simulated",
			oldInput: "BlockA\nBlockB\nBlockC\n",
			newInput: "BlockA\nBlockC\nBlockB\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("BlockA\n")},
				{Kind: lines.ChunkKindDeleted, Data: []byte("BlockB\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("BlockC\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("BlockB\n")},
			},
		},
		{
			name:     "alternating additions",
			oldInput: "A\nB\nC\n",
			newInput: "A\n1\nB\n2\nC\n",
			expected: []lines.Chunk{
				{Kind: lines.ChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("1\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("B\n")},
				{Kind: lines.ChunkKindAdded, Data: []byte("2\n")},
				{Kind: lines.ChunkKindUnchanged, Data: []byte("C\n")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chunks, err := lines.Diff([]byte(tt.oldInput), []byte(tt.newInput))
			if err != nil {
				t.Fatalf("Diff returned error: %v", err)
			}

			if len(chunks) != len(tt.expected) {
				t.Fatalf("expected %d chunks, got %d: %s", len(tt.expected), len(chunks), formatChunks(chunks))
			}

			for i := range tt.expected {
				if chunks[i].Kind != tt.expected[i].Kind {
					t.Fatalf("chunk %d kind mismatch: got %v, want %v; chunks: %s", i, chunks[i].Kind, tt.expected[i].Kind, formatChunks(chunks))
				}

				if !bytes.Equal(chunks[i].Data, tt.expected[i].Data) {
					t.Fatalf("chunk %d data mismatch: got %q, want %q; chunks: %s", i, string(chunks[i].Data), string(tt.expected[i].Data), formatChunks(chunks))
				}
			}
		})
	}
}

func formatChunks(chunks []lines.Chunk) string {
	var b strings.Builder
	b.WriteByte('[')

	for i, chunk := range chunks {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(chunkKindName(chunk.Kind))
		b.WriteByte(':')
		b.WriteString(strconv.Quote(string(chunk.Data)))
	}

	b.WriteByte(']')

	return b.String()
}

func chunkKindName(kind lines.ChunkKind) string {
	switch kind {
	case lines.ChunkKindUnchanged:
		return "U"
	case lines.ChunkKindDeleted:
		return "D"
	case lines.ChunkKindAdded:
		return "A"
	default:
		return "?"
	}
}
