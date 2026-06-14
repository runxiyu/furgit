package main

import (
	"bytes"
	"encoding/hex"
	"io"

	"lindenii.org/go/furgit/internal/utils"
)

func indentBlock(out io.Writer, indent string, block []byte) {
	lines := bytes.Split(block, []byte("\n"))
	if n := len(lines); n > 0 && len(lines[n-1]) == 0 {
		lines = lines[:n-1]
	}

	for _, line := range lines {
		utils.BestEffortFprintf(out, "%s%s\n", indent, line)
	}
}

func hexBlock(out io.Writer, indent string, data []byte) {
	var buf bytes.Buffer

	dumper := hex.Dumper(&buf)
	_, _ = dumper.Write(data)
	_ = dumper.Close()

	indentBlock(out, indent, buf.Bytes())
}
