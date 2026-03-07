package service

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"slices"
)

// createQuarantineRoot creates one per-push quarantine directory beneath the
// permanent objects root.
func (service *Service) createQuarantineRoot() (string, *os.Root, error) {
	name := "tmp_objdir-incoming-" + rand.Text()

	err := service.opts.ObjectsRoot.Mkdir(name, 0o700)
	if err != nil {
		return "", nil, err
	}

	root, err := service.opts.ObjectsRoot.OpenRoot(name)
	if err != nil {
		_ = service.opts.ObjectsRoot.RemoveAll(name)

		return "", nil, err
	}

	return name, root, nil
}

func (service *Service) openQuarantinePackRoot(quarantineRoot *os.Root) (*os.Root, error) {
	err := quarantineRoot.Mkdir("pack", 0o755)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	return quarantineRoot.OpenRoot("pack")
}

func (service *Service) promoteQuarantine(quarantineName string, quarantineRoot *os.Root) error {
	if quarantineName == "" || quarantineRoot == nil {
		return nil
	}

	return service.promoteQuarantineDir(quarantineName, quarantineRoot, ".")
}

func (service *Service) promoteQuarantineDir(quarantineName string, quarantineRoot *os.Root, rel string) error {
	entries, err := fs.ReadDir(quarantineRoot.FS(), rel)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	slices.SortFunc(entries, func(left, right fs.DirEntry) int {
		return packCopyPriority(left.Name()) - packCopyPriority(right.Name())
	})

	for _, entry := range entries {
		childRel := entry.Name()
		if rel != "." {
			childRel = path.Join(rel, entry.Name())
		}

		if entry.IsDir() {
			err = service.opts.ObjectsRoot.Mkdir(childRel, 0o755)
			if err != nil && !os.IsExist(err) {
				return err
			}

			err = service.applyPromotedDirectoryPermissions(childRel)
			if err != nil {
				return err
			}

			err = service.promoteQuarantineDir(quarantineName, quarantineRoot, childRel)
			if err != nil {
				return err
			}

			continue
		}

		err = finalizeQuarantineFile(
			service.opts.ObjectsRoot,
			path.Join(quarantineName, childRel),
			childRel,
			isLooseObjectShardPath(rel),
			service.opts.PromotedObjectPermissions,
		)
		if err == nil {
			continue
		}

		return err
	}

	return nil
}

func packCopyPriority(name string) int {
	if !pathHasPackPrefix(name) {
		return 0
	}

	switch {
	case path.Ext(name) == ".keep":
		return 1
	case path.Ext(name) == ".pack":
		return 2
	case path.Ext(name) == ".rev":
		return 3
	case path.Ext(name) == ".idx":
		return 4
	default:
		return 5
	}
}

func pathHasPackPrefix(name string) bool {
	return len(name) >= 4 && name[:4] == "pack"
}

func isLooseObjectShardPath(rel string) bool {
	return len(rel) == 2 && isHex(rel[0]) && isHex(rel[1])
}

func isHex(ch byte) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func (service *Service) applyPromotedDirectoryPermissions(name string) error {
	if service.opts.PromotedObjectPermissions == nil {
		return nil
	}

	return service.opts.ObjectsRoot.Chmod(name, service.opts.PromotedObjectPermissions.DirMode)
}

func applyPromotedFilePermissions(
	root *os.Root,
	name string,
	perms *PromotedObjectPermissions,
) error {
	if perms == nil {
		return nil
	}

	return root.Chmod(name, perms.FileMode)
}

func finalizeQuarantineFile(
	root *os.Root,
	src, dst string,
	skipCollisionCheck bool,
	perms *PromotedObjectPermissions,
) error {
	const maxVanishedRetries = 5

	for retries := 0; ; retries++ {
		err := root.Link(src, dst)
		switch {
		case err == nil:
			_ = root.Remove(src)

			return applyPromotedFilePermissions(root, dst, perms)
		case !errors.Is(err, fs.ErrExist):
			_, statErr := root.Stat(dst)
			if statErr == nil {
				err = fs.ErrExist
			} else if errors.Is(statErr, fs.ErrNotExist) {
				renameErr := root.Rename(src, dst)
				if renameErr == nil {
					return applyPromotedFilePermissions(root, dst, perms)
				}

				err = renameErr
			} else {
				_ = root.Remove(src)

				return statErr
			}
		}

		if !errors.Is(err, fs.ErrExist) {
			_ = root.Remove(src)

			return fmt.Errorf("promote quarantine %q -> %q: %w", src, dst, err)
		}

		if skipCollisionCheck {
			_ = root.Remove(src)

			return applyPromotedFilePermissions(root, dst, perms)
		}

		equal, vanished, cmpErr := compareRootFiles(root, src, dst)
		if vanished {
			if retries >= maxVanishedRetries {
				return fmt.Errorf("promote quarantine %q -> %q: destination repeatedly vanished", src, dst)
			}

			continue
		}

		if cmpErr != nil {
			return cmpErr
		}

		if !equal {
			return fmt.Errorf("promote quarantine %q -> %q: files differ in contents", src, dst)
		}

		_ = root.Remove(src)

		return applyPromotedFilePermissions(root, dst, perms)
	}
}

func compareRootFiles(root *os.Root, left, right string) (equal bool, vanished bool, err error) {
	leftFile, err := root.Open(left)
	if err != nil {
		return false, false, err
	}

	defer func() {
		_ = leftFile.Close()
	}()

	rightFile, err := root.Open(right)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, true, nil
		}

		return false, false, err
	}

	defer func() {
		_ = rightFile.Close()
	}()

	var leftBuf, rightBuf [4096]byte

	for {
		leftN, leftErr := leftFile.Read(leftBuf[:])
		rightN, rightErr := rightFile.Read(rightBuf[:])

		if leftErr != nil && !errors.Is(leftErr, io.EOF) {
			return false, false, leftErr
		}

		if rightErr != nil && !errors.Is(rightErr, io.EOF) {
			return false, false, rightErr
		}

		if leftN != rightN || !bytes.Equal(leftBuf[:leftN], rightBuf[:rightN]) {
			return false, false, nil
		}

		if leftErr != nil || rightErr != nil {
			return true, false, nil
		}
	}
}
