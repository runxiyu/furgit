package packed

// import (
// 	"os"
//
// 	"lindenii.org/go/furgit/object/store"
// )
//
// var (
// 	_ store.PackQuarantiner = (*Packed)(nil)
// 	_ store.PackQuarantine  = (*packQuarantine)(nil)
// )
//
// // packQuarantine is one quarantined packed store
// // rooted privately beneath a destination pack root.
// type packQuarantine struct {
// 	*Packed
//
// 	parent   *Packed
// 	tempName string
// 	tempRoot *os.Root
// }
//
// // BeginPackQuarantine creates one quarantined packed store
// // rooted privately beneath the destination pack root.
// //
// // Labels: Deps-Borrowed, Life-Parent, Close-No.
// func (packed *Packed) BeginPackQuarantine(opts store.PackQuarantineOptions) (store.PackQuarantine, error)
//
// // Discard removes the quarantine
// // and invalidates the receiver.
// func (quarantine *packQuarantine) Discard() error
//
// // Promote publishes all finalized pack artifacts in the quarantine
// // into the parent packed store,
// // and invalidates the receiver.
// func (quarantine *packQuarantine) Promote() error
//
// // promoteAll links every pack artifact in the quarantine
// // into the parent packed store,
// // in pack/rev/idx dependency order.
// func (quarantine *packQuarantine) promoteAll() error
//
// // promoteFile links one quarantined pack artifact
// // into the parent packed store,
// // treating an already-present destination as success.
// func (quarantine *packQuarantine) promoteFile(name string) error
//
// // createPackQuarantineRoot creates a private quarantine directory
// // beneath parent,
// // and returns its name and an os.Root over it.
// func createPackQuarantineRoot(parent *os.Root) (string, *os.Root, error)
//
// // packPromotionPriority orders pack artifacts
// // so that data files are linked
// // before the index that publishes them.
// func packPromotionPriority(name string) int
