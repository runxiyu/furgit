package receivepack

import (
	"fmt"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
)

// WriteReportStatus writes one classic report-status response.
func (session *Session) WriteReportStatus(result ReportStatusResult) error {
	unpackResult := "ok"
	if result.UnpackError != "" {
		unpackResult = result.UnpackError
	}

	if !session.negotiated.SideBand64K {
		err := session.base.WriteData(fmt.Appendf(nil, "unpack %s\n", unpackResult))
		if err != nil {
			return err
		}

		for _, command := range result.Commands {
			line := fmt.Sprintf("ok %s\n", command.Name)
			if command.Error != "" {
				line = fmt.Sprintf("ng %s %s\n", command.Name, command.Error)
			}

			err = session.base.WriteData([]byte(line))
			if err != nil {
				return err
			}
		}

		return session.base.WriteFlush()
	}

	buf, err := pktline.AppendData(nil, fmt.Appendf(nil, "unpack %s\n", unpackResult))
	if err != nil {
		return err
	}

	for _, command := range result.Commands {
		line := fmt.Sprintf("ok %s\n", command.Name)
		if command.Error != "" {
			line = fmt.Sprintf("ng %s %s\n", command.Name, command.Error)
		}

		buf, err = pktline.AppendData(buf, []byte(line))
		if err != nil {
			return err
		}
	}

	buf = pktline.AppendFlushPkt(buf)

	w := session.base.PrimaryDataWriter()

	_, err = w.Write(buf)
	if err != nil {
		return err
	}

	return session.base.WriteFlush()
}

// WriteReportStatusV2 writes one report-status-v2 response.
func (session *Session) WriteReportStatusV2(result ReportStatusResult) error {
	unpackResult := "ok"
	if result.UnpackError != "" {
		unpackResult = result.UnpackError
	}

	if !session.negotiated.SideBand64K { //nolint:nestif
		err := session.base.WriteData(fmt.Appendf(nil, "unpack %s\n", unpackResult))
		if err != nil {
			return err
		}

		for _, command := range result.Commands {
			if command.Error != "" {
				err = session.base.WriteData(fmt.Appendf(nil, "ng %s %s\n", command.Name, command.Error))
				if err != nil {
					return err
				}

				continue
			}

			err = session.base.WriteData(fmt.Appendf(nil, "ok %s\n", command.Name))
			if err != nil {
				return err
			}

			if command.RefName != "" {
				err = session.base.WriteData(fmt.Appendf(nil, "option refname %s\n", command.RefName))
				if err != nil {
					return err
				}
			}

			if command.OldID != nil {
				err = session.base.WriteData(fmt.Appendf(nil, "option old-oid %s\n", *command.OldID))
				if err != nil {
					return err
				}
			}

			if command.NewID != nil {
				err = session.base.WriteData(fmt.Appendf(nil, "option new-oid %s\n", *command.NewID))
				if err != nil {
					return err
				}
			}

			if command.ForcedUpdate {
				err = session.base.WriteData([]byte("option forced-update\n"))
				if err != nil {
					return err
				}
			}
		}

		return session.base.WriteFlush()
	}

	buf, err := pktline.AppendData(nil, fmt.Appendf(nil, "unpack %s\n", unpackResult))
	if err != nil {
		return err
	}

	for _, command := range result.Commands {
		if command.Error != "" {
			buf, err = pktline.AppendData(buf, fmt.Appendf(nil, "ng %s %s\n", command.Name, command.Error))
			if err != nil {
				return err
			}

			continue
		}

		buf, err = pktline.AppendData(buf, fmt.Appendf(nil, "ok %s\n", command.Name))
		if err != nil {
			return err
		}

		if command.RefName != "" {
			buf, err = pktline.AppendData(buf, fmt.Appendf(nil, "option refname %s\n", command.RefName))
			if err != nil {
				return err
			}
		}

		if command.OldID != nil {
			buf, err = pktline.AppendData(buf, fmt.Appendf(nil, "option old-oid %s\n", *command.OldID))
			if err != nil {
				return err
			}
		}

		if command.NewID != nil {
			buf, err = pktline.AppendData(buf, fmt.Appendf(nil, "option new-oid %s\n", *command.NewID))
			if err != nil {
				return err
			}
		}

		if command.ForcedUpdate {
			buf, err = pktline.AppendData(buf, []byte("option forced-update\n"))
			if err != nil {
				return err
			}
		}
	}

	buf = pktline.AppendFlushPkt(buf)

	w := session.base.PrimaryDataWriter()

	_, err = w.Write(buf)
	if err != nil {
		return err
	}

	return session.base.WriteFlush()
}
