package reading

// RefreshPolicy configures when candidate pack/index discovery refreshes.
type RefreshPolicy uint8

const (
	// RefreshPolicyOnMissing refreshes candidates once after a lookup miss.
	RefreshPolicyOnMissing RefreshPolicy = iota
	// RefreshPolicyNever disables automatic refresh after lookup misses.
	RefreshPolicyNever
)

// Options configures a packed object store.
type Options struct {
	RefreshPolicy RefreshPolicy
}
