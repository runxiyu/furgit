package diffbytes

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestDiffBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		oldInput string
		newInput string
		expected []BytesDiffChunk
	}{
		{
			name:     "empty inputs produce no chunks",
			oldInput: "",
			newInput: "",
			expected: []BytesDiffChunk{},
		},
		{
			name:     "only additions",
			oldInput: "",
			newInput: "alpha\nbeta\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindAdded, Data: []byte("alpha\nbeta\n")},
			},
		},
		{
			name:     "only deletions",
			oldInput: "alpha\nbeta\n",
			newInput: "",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("alpha\nbeta\n")},
			},
		},
		{
			name:     "unchanged content is grouped",
			oldInput: "same\nlines\n",
			newInput: "same\nlines\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("same\nlines\n")},
			},
		},
		{
			name:     "insertion in the middle",
			oldInput: "a\nb\nc\n",
			newInput: "a\nb\nX\nc\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("a\nb\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("X\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("c\n")},
			},
		},
		{
			name:     "replacement without trailing newline",
			oldInput: "first\nsecond",
			newInput: "first\nsecond\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("first\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("second")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("second\n")},
			},
		},
		{
			name:     "line replacement",
			oldInput: "a\nb\nc\n",
			newInput: "a\nB\nc\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("a\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("b\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("B\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("c\n")},
			},
		},
		{
			name:     "swap adjacent lines",
			oldInput: "A\nB\n",
			newInput: "B\nA\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("A\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("B\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("A\n")},
			},
		},
		{
			name:     "indentation change is a full line replacement",
			oldInput: "func main() {\n\treturn\n}\n",
			newInput: "func main() {\n    return\n}\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("func main() {\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("\treturn\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("    return\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("}\n")},
			},
		},
		{
			name:     "commenting out lines",
			oldInput: "code\n",
			newInput: "// code\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("code\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("// code\n")},
			},
		},
		{
			name:     "reducing repeating lines",
			oldInput: "log\nlog\nlog\n",
			newInput: "log\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("log\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("log\nlog\n")},
			},
		},
		{
			name:     "expanding repeating lines",
			oldInput: "tick\n",
			newInput: "tick\ntick\ntick\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("tick\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("tick\ntick\n")},
			},
		},
		{
			name:     "interleaved modifications",
			oldInput: "keep\nchange\nkeep\nchange\n",
			newInput: "keep\nfixed\nkeep\nfixed\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("keep\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("change\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("fixed\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("keep\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("change\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("fixed\n")},
			},
		},
		{
			name:     "large common header and footer",
			oldInput: "header\nheader\nheader\nOLD\nfooter\nfooter\n",
			newInput: "header\nheader\nheader\nNEW\nfooter\nfooter\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("header\nheader\nheader\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("OLD\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("NEW\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("footer\nfooter\n")},
			},
		},
		{
			name:     "completely different content",
			oldInput: "apple\nbanana\n",
			newInput: "cherry\ndate\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("apple\nbanana\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("cherry\ndate\n")},
			},
		},
		{
			name:     "unicode and emoji changes",
			oldInput: "Hello 🌍\nYay\n",
			newInput: "Hello 🌎\nYay\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("Hello 🌍\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("Hello 🌎\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("Yay\n")},
			},
		},
		{
			name:     "binary data with embedded newlines",
			oldInput: "\x00\x01\n\x02\x03\n",
			newInput: "\x00\x01\n\x02\xFF\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("\x00\x01\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("\x02\x03\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("\x02\xFF\n")},
			},
		},
		{
			name:     "adding trailing newline to last line",
			oldInput: "Line 1\nLine 2",
			newInput: "Line 1\nLine 2\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("Line 1\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("Line 2")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("Line 2\n")},
			},
		},
		{
			name:     "removing trailing newline",
			oldInput: "A\nB\n",
			newInput: "A\nB",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("B\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("B")},
			},
		},
		{
			name:     "inserting blank lines",
			oldInput: "A\nB\n",
			newInput: "A\n\n\nB\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("\n\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("B\n")},
			},
		},
		{
			name:     "collapsing blank lines",
			oldInput: "A\n\n\n\nB\n",
			newInput: "A\nB\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("\n\n\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("B\n")},
			},
		},
		{
			name:     "case sensitivity check",
			oldInput: "FOO\nbar\n",
			newInput: "foo\nbar\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("FOO\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("foo\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("bar\n")},
			},
		},
		{
			name:     "partial line match is full mismatch",
			oldInput: "The quick brown fox\n",
			newInput: "The quick brown fox jumps\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("The quick brown fox\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("The quick brown fox jumps\n")},
			},
		},
		{
			name:     "inserting middle content",
			oldInput: "Top\nBottom\n",
			newInput: "Top\nMiddle\nBottom\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("Top\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("Middle\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("Bottom\n")},
			},
		},
		{
			name:     "block move simulated",
			oldInput: "BlockA\nBlockB\nBlockC\n",
			newInput: "BlockA\nBlockC\nBlockB\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("BlockA\n")},
				{Kind: BytesDiffChunkKindDeleted, Data: []byte("BlockB\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("BlockC\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("BlockB\n")},
			},
		},
		{
			name:     "alternating additions",
			oldInput: "A\nB\nC\n",
			newInput: "A\n1\nB\n2\nC\n",
			expected: []BytesDiffChunk{
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("1\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("B\n")},
				{Kind: BytesDiffChunkKindAdded, Data: []byte("2\n")},
				{Kind: BytesDiffChunkKindUnchanged, Data: []byte("C\n")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chunks, err := DiffBytes([]byte(tt.oldInput), []byte(tt.newInput))
			if err != nil {
				t.Fatalf("DiffBytes returned error: %v", err)
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

func formatChunks(chunks []BytesDiffChunk) string {
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

func chunkKindName(kind BytesDiffChunkKind) string {
	switch kind {
	case BytesDiffChunkKindUnchanged:
		return "U"
	case BytesDiffChunkKindDeleted:
		return "D"
	case BytesDiffChunkKindAdded:
		return "A"
	default:
		return "?"
	}
}
