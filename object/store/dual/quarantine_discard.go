package dual

// Discard abandons both quarantine halves and invalidates the receiver.
func (quarantine *quarantine) Discard() error {
	err := quarantine.packQ.Discard()
	if err != nil {
		return err
	}

	return quarantine.objectQ.Discard()
}
