package resolve

type treeFSOp uint8

const (
	treeFSOpOpen treeFSOp = iota
	treeFSOpReadFile
	treeFSOpReadDir
	treeFSOpStat
	treeFSOpSub
)

func (op treeFSOp) pathErrorOp() string {
	switch op {
	case treeFSOpOpen:
		return "open"
	case treeFSOpReadFile:
		return "readfile"
	case treeFSOpReadDir:
		return "readdir"
	case treeFSOpStat:
		return "stat"
	case treeFSOpSub:
		return "sub"
	default:
		return "treefs"
	}
}
