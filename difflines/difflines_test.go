package difflines

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestDiffLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		oldInput string
		newInput string
		expected []LinesDiffChunk
	}{
		{
			name:     "empty inputs produce no chunks",
			oldInput: "",
			newInput: "",
			expected: []LinesDiffChunk{},
		},
		{
			name:     "only additions",
			oldInput: "",
			newInput: "alpha\nbeta\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindAdded, Data: []byte("alpha\nbeta\n")},
			},
		},
		{
			name:     "only deletions",
			oldInput: "alpha\nbeta\n",
			newInput: "",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("alpha\nbeta\n")},
			},
		},
		{
			name:     "unchanged content is grouped",
			oldInput: "same\nlines\n",
			newInput: "same\nlines\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("same\nlines\n")},
			},
		},
		{
			name:     "insertion in the middle",
			oldInput: "a\nb\nc\n",
			newInput: "a\nb\nX\nc\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("a\nb\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("X\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("c\n")},
			},
		},
		{
			name:     "replacement without trailing newline",
			oldInput: "first\nsecond",
			newInput: "first\nsecond\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("first\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("second")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("second\n")},
			},
		},
		{
			name:     "line replacement",
			oldInput: "a\nb\nc\n",
			newInput: "a\nB\nc\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("a\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("b\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("B\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("c\n")},
			},
		},
		{
			name:     "swap adjacent lines",
			oldInput: "A\nB\n",
			newInput: "B\nA\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("A\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("B\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("A\n")},
			},
		},
		{
			name:     "indentation change is a full line replacement",
			oldInput: "func main() {\n\treturn\n}\n",
			newInput: "func main() {\n    return\n}\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("func main() {\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("\treturn\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("    return\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("}\n")},
			},
		},
		{
			name:     "commenting out lines",
			oldInput: "code\n",
			newInput: "// code\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("code\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("// code\n")},
			},
		},
		{
			name:     "reducing repeating lines",
			oldInput: "log\nlog\nlog\n",
			newInput: "log\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("log\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("log\nlog\n")},
			},
		},
		{
			name:     "expanding repeating lines",
			oldInput: "tick\n",
			newInput: "tick\ntick\ntick\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("tick\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("tick\ntick\n")},
			},
		},
		{
			name:     "interleaved modifications",
			oldInput: "keep\nchange\nkeep\nchange\n",
			newInput: "keep\nfixed\nkeep\nfixed\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("keep\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("change\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("fixed\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("keep\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("change\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("fixed\n")},
			},
		},
		{
			name:     "large common header and footer",
			oldInput: "header\nheader\nheader\nOLD\nfooter\nfooter\n",
			newInput: "header\nheader\nheader\nNEW\nfooter\nfooter\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("header\nheader\nheader\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("OLD\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("NEW\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("footer\nfooter\n")},
			},
		},
		{
			name:     "completely different content",
			oldInput: "apple\nbanana\n",
			newInput: "cherry\ndate\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("apple\nbanana\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("cherry\ndate\n")},
			},
		},
		{
			name:     "unicode and emoji changes",
			oldInput: "Hello 🌍\nYay\n",
			newInput: "Hello 🌎\nYay\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("Hello 🌍\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("Hello 🌎\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("Yay\n")},
			},
		},
		{
			name:     "binary data with embedded newlines",
			oldInput: "\x00\x01\n\x02\x03\n",
			newInput: "\x00\x01\n\x02\xFF\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("\x00\x01\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("\x02\x03\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("\x02\xFF\n")},
			},
		},
		{
			name:     "adding trailing newline to last line",
			oldInput: "Line 1\nLine 2",
			newInput: "Line 1\nLine 2\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("Line 1\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("Line 2")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("Line 2\n")},
			},
		},
		{
			name:     "removing trailing newline",
			oldInput: "A\nB\n",
			newInput: "A\nB",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("B\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("B")},
			},
		},
		{
			name:     "inserting blank lines",
			oldInput: "A\nB\n",
			newInput: "A\n\n\nB\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("\n\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("B\n")},
			},
		},
		{
			name:     "collapsing blank lines",
			oldInput: "A\n\n\n\nB\n",
			newInput: "A\nB\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("\n\n\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("B\n")},
			},
		},
		{
			name:     "case sensitivity check",
			oldInput: "FOO\nbar\n",
			newInput: "foo\nbar\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("FOO\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("foo\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("bar\n")},
			},
		},
		{
			name:     "partial line match is full mismatch",
			oldInput: "The quick brown fox\n",
			newInput: "The quick brown fox jumps\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("The quick brown fox\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("The quick brown fox jumps\n")},
			},
		},
		{
			name:     "inserting middle content",
			oldInput: "Top\nBottom\n",
			newInput: "Top\nMiddle\nBottom\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("Top\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("Middle\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("Bottom\n")},
			},
		},
		{
			name:     "block move simulated",
			oldInput: "BlockA\nBlockB\nBlockC\n",
			newInput: "BlockA\nBlockC\nBlockB\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("BlockA\n")},
				{Kind: LinesDiffChunkKindDeleted, Data: []byte("BlockB\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("BlockC\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("BlockB\n")},
			},
		},
		{
			name:     "alternating additions",
			oldInput: "A\nB\nC\n",
			newInput: "A\n1\nB\n2\nC\n",
			expected: []LinesDiffChunk{
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("A\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("1\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("B\n")},
				{Kind: LinesDiffChunkKindAdded, Data: []byte("2\n")},
				{Kind: LinesDiffChunkKindUnchanged, Data: []byte("C\n")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chunks, err := DiffLines([]byte(tt.oldInput), []byte(tt.newInput))
			if err != nil {
				t.Fatalf("DiffLines returned error: %v", err)
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

func formatChunks(chunks []LinesDiffChunk) string {
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

func chunkKindName(kind LinesDiffChunkKind) string {
	switch kind {
	case LinesDiffChunkKindUnchanged:
		return "U"
	case LinesDiffChunkKindDeleted:
		return "D"
	case LinesDiffChunkKindAdded:
		return "A"
	default:
		return "?"
	}
}
