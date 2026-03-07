package receivepack

import (
	"io/fs"

	"codeberg.org/lindenii/furgit/receivepack/internal/service"
)

// PromotedObjectPermissions configures the destination permissions applied to
// objects and directories promoted out of quarantine.
type PromotedObjectPermissions struct {
	DirMode  fs.FileMode
	FileMode fs.FileMode
}

func translatePromotedObjectPermissions(
	perms *PromotedObjectPermissions,
) *service.PromotedObjectPermissions {
	if perms == nil {
		return nil
	}

	return &service.PromotedObjectPermissions{
		DirMode:  perms.DirMode,
		FileMode: perms.FileMode,
	}
}
