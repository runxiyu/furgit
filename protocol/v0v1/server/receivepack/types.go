package receivepack

import "codeberg.org/lindenii/furgit/objectid"

// Command is one requested reference update.
type Command struct {
	OldID objectid.ObjectID
	NewID objectid.ObjectID
	Name  string
}

// PushCertificate is one parsed push certificate block.
type PushCertificate struct {
	HeaderLines    []string
	EmbeddedOption []string
	Commands       []Command
	SignatureLines []string
}

// Request is one parsed receive-pack request.
type Request struct {
	Capabilities Capabilities
	Shallow      []objectid.ObjectID
	Commands     []Command
	PushCert     *PushCertificate
	PushOptions  []string
	PackExpected bool
	DeleteOnly   bool
}

// CommandResult is one per-command report-status result.
type CommandResult struct {
	Name         string
	Error        string
	RefName      string
	OldID        *objectid.ObjectID
	NewID        *objectid.ObjectID
	ForcedUpdate bool
}

// ReportStatusResult is one report-status payload.
type ReportStatusResult struct {
	UnpackError string
	Commands    []CommandResult
}
