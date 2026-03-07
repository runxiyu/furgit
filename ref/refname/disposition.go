package refname

func refnameDisposition(ch byte) byte {
	switch {
	case ch == '/':
		return 1
	case ch == '.':
		return 2
	case ch == '{':
		return 3
	case ch == '*':
		return 5
	case ch < 0x20 || ch == 0x7f:
		return 4
	case ch == ':' || ch == '?' || ch == '[' || ch == '\\' || ch == '^' || ch == '~' || ch == ' ' || ch == '\t':
		return 4
	default:
		return 0
	}
}
