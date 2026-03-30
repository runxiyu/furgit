package objectstore

// WriterQuarantine represents one quarantined write that accepts both object-
// wise and pack-wise writes.
type WriterQuarantine interface {
	Quarantine
	ObjectWriter
	PackWriter
}

// QuarantineOptions controls the options for one coordinated quarantine creation.
type QuarantineOptions struct {
	Object ObjectQuarantineOptions
	Pack   PackQuarantineOptions
}

// WriterQuarantiner creates coordinated quarantines that support both object-
// wise and pack-wise writes.
type WriterQuarantiner interface {
	BeginQuarantine(opts QuarantineOptions) (WriterQuarantine, error)
}
