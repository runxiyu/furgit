package server

import (
	"fmt"
	"strings"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// AdvertiseRefs writes one server ref advertisement.
func (session *Session) AdvertiseRefs(ad Advertisement, capabilityTokens []string) error {
	if session.opts.Version == Version1 {
		err := session.enc.WriteData([]byte("version 1\n"))
		if err != nil {
			return err
		}
	}

	capList := strings.Join(capabilityTokens, " ")

	refs := sortAdvertisedRefs(ad.Refs)
	if len(refs) == 0 {
		line := fmt.Sprintf("%s capabilities^{}\x00%s\n", objectid.Zero(session.opts.Algorithm), capList)

		err := session.enc.WriteData([]byte(line))
		if err != nil {
			return err
		}

		return session.WriteFlush()
	}

	for i, entry := range refs {
		line := fmt.Sprintf("%s %s", entry.ID, entry.Name)
		if i == 0 {
			line += "\x00" + capList
		}

		err := session.enc.WriteData([]byte(line + "\n"))
		if err != nil {
			return err
		}

		if entry.Peeled != nil {
			peeled := fmt.Sprintf("%s %s^{}\n", *entry.Peeled, entry.Name)

			err = session.enc.WriteData([]byte(peeled))
			if err != nil {
				return err
			}
		}
	}

	return session.WriteFlush()
}
