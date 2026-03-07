package files

import (
	"fmt"
	"strings"
)

func (tx *Transaction) verifyCurrent(item preparedTxOp) error {
	switch item.op.kind {
	case txCreate:
		if item.target.ref.kind != directMissing {
			return fmt.Errorf("refstore/files: reference %q already exists", item.target.name)
		}

		return nil
	case txUpdate, txDelete, txVerify:
		if item.target.ref.kind == directMissing {
			return fmt.Errorf("refstore/files: reference %q is missing", item.target.name)
		}

		if item.target.ref.kind != directDetached {
			return fmt.Errorf("refstore/files: reference %q is not detached", item.target.name)
		}

		if item.target.ref.id != item.op.oldID {
			return fmt.Errorf("refstore/files: reference %q is at %s but expected %s", item.target.name, item.target.ref.id, item.op.oldID)
		}

		return nil
	case txCreateSymbolic:
		if item.target.ref.kind != directMissing {
			return fmt.Errorf("refstore/files: reference %q already exists", item.target.name)
		}

		return nil
	case txUpdateSymbolic, txDeleteSymbolic, txVerifySymbolic:
		if item.target.ref.kind == directMissing {
			return fmt.Errorf("refstore/files: symbolic reference %q is missing", item.target.name)
		}

		if item.target.ref.kind != directSymbolic {
			return fmt.Errorf("refstore/files: reference %q is not symbolic", item.target.name)
		}

		if strings.TrimSpace(item.target.ref.target) != strings.TrimSpace(item.op.oldTarget) {
			return fmt.Errorf("refstore/files: reference %q points at %q, expected %q", item.target.name, item.target.ref.target, item.op.oldTarget)
		}

		return nil
	default:
		return fmt.Errorf("refstore/files: unsupported transaction operation %d", item.op.kind)
	}
}
