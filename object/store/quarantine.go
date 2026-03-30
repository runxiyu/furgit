package objectstore

// WriterQuarantine represents one quarantined write that accepts both object-
// wise and pack-wise writes.
type Quarantine interface {
	BaseQuarantine
	Writer
}

// QuarantineOptions controls the options for one coordinated quarantine creation.
type QuarantineOptions struct {
	Object ObjectQuarantineOptions
	Pack   PackQuarantineOptions
}

// WriterQuarantiner creates coordinated quarantines that support both object-
// wise and pack-wise writes.
type Quarantiner interface {
	BeginQuarantine(opts QuarantineOptions) (Quarantine, error)
}
