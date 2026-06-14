package packed

import (
	"fmt"
	"io/fs"
	"strings"
)

// Refresh rescans the pack directory
// and replaces the store's view of available packs.
//
// Every index found must parse
// and have its pack data present and consistent;
// otherwise Refresh fails without changing the view.
func (packed *Packed) Refresh() error {
	packed.refreshMu.Lock()
	defer packed.refreshMu.Unlock()

	dirEntries, err := fs.ReadDir(packed.root.FS(), ".")
	if err != nil {
		return fmt.Errorf("object/store/packed: %w", err)
	}

	next := make(map[string]*pack, len(packed.byName))

	opened := make([]*pack, 0, len(dirEntries))

	for _, dirEntry := range dirEntries {
		name, ok := strings.CutSuffix(dirEntry.Name(), ".idx")
		if !ok || dirEntry.IsDir() {
			continue
		}

		if existing, ok := packed.byName[name]; ok {
			next[name] = existing

			continue
		}

		p, err := openPack(packed.root, name, packed.objectFormat)
		if err != nil {
			for _, p := range opened {
				_ = p.close()
			}

			return err
		}

		opened = append(opened, p)
		next[name] = p
	}

	for name, p := range packed.byName {
		if _, ok := next[name]; !ok {
			packed.retired = append(packed.retired, p)
		}
	}

	packed.byName = next

	present := make(map[*pack]struct{}, len(next))
	for _, p := range next {
		present[p] = struct{}{}
	}

	packed.order.Sync(present)

	return nil
}
