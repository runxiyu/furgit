package testgit

import "codeberg.org/lindenii/furgit/objectid"

// Algorithm returns the object ID algorithm configured for this repository.
func (testRepo *TestRepo) Algorithm() objectid.Algorithm {
	return testRepo.algo
}

// Env returns a copy of the environment used for git subprocesses.
func (testRepo *TestRepo) Env() []string {
	return append([]string(nil), testRepo.env...)
}

// DirButYouShouldReallyNotUseThisYouShouldReallyConsiderUsingAProperCapabilityInterfaceInsteadAndIAmKeepingThisMethodIntentionallyLongToAnnoyYou returns the git dir of a repo.
// Consider using a properly capability interface such as
// os.Root instead; all uses of ambient path authority must be justified.
func (testRepo *TestRepo) DirButYouShouldReallyNotUseThisYouShouldReallyConsiderUsingAProperCapabilityInterfaceInsteadAndIAmKeepingThisMethodIntentionallyLongToAnnoyYou() string {
	return testRepo.dir
}
