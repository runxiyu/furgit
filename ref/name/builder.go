package name

import "strings"

func overwriteBuilderAt(builder *strings.Builder, index int) {
	value := builder.String()
	truncateBuilder(builder, index)
	builder.WriteByte('-')
	builder.WriteString(value[index+1:])
}

func truncateBuilder(builder *strings.Builder, n int) {
	value := builder.String()
	builder.Reset()
	builder.WriteString(value[:n])
}
