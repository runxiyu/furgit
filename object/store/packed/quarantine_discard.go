package packed

// Discard removes the quarantine and invalidates the receiver.
func (quarantine *packQuarantine) Discard() error {
	closeErr := quarantine.Close()
	tempRootErr := quarantine.tempRoot.Close()
	removeErr := quarantine.parent.root.RemoveAll(quarantine.tempName)

	if closeErr != nil {
		return closeErr
	}

	if tempRootErr != nil {
		return tempRootErr
	}

	return removeErr
}
