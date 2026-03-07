package refname

import (
	"strings"
)

func overwriteLastByte(builder *strings.Builder, ch byte) {
	overwriteBuilderAt(builder, builder.Len()-1, ch)
}

func overwriteBuilderAt(builder *strings.Builder, index int, ch byte) {
	value := builder.String()
	truncateBuilder(builder, index)
	builder.WriteByte(ch)
	builder.WriteString(value[index+1:])
}

func truncateBuilder(builder *strings.Builder, n int) {
	value := builder.String()
	builder.Reset()
	builder.WriteString(value[:n])
}
