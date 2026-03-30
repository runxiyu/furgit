package dual

// Promote publishes both quarantine halves and invalidates the receiver.
//
// Promotion is coordinated and ordered, but not atomic.
func (quarantine *quarantine) Promote() error {
	err := quarantine.packQ.Promote()
	if err != nil {
		return err
	}

	return quarantine.objectQ.Promote()
}
