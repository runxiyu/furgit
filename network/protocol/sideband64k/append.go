package sideband64k

import (
	"fmt"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
)

// AppendBand appends one side-band-64k data frame to dst.
func AppendBand(dst []byte, band Band, payload []byte) ([]byte, error) {
	if !validBand(band) {
		return dst, fmt.Errorf("%w: %d", ErrInvalidBand, band)
	}

	maxData := effectiveMaxData(DataMax)
	if len(payload) > maxData {
		return dst, fmt.Errorf("%w: %d > %d", ErrTooLarge, len(payload), maxData)
	}

	framed := make([]byte, len(payload)+1)
	framed[0] = byte(band)
	copy(framed[1:], payload)

	return pktline.AppendData(dst, framed)
}

// AppendData appends one band-1 data frame to dst.
func AppendData(dst, payload []byte) ([]byte, error) {
	return AppendBand(dst, BandData, payload)
}

// AppendProgress appends one band-2 progress frame to dst.
func AppendProgress(dst, payload []byte) ([]byte, error) {
	return AppendBand(dst, BandProgress, payload)
}

// AppendError appends one band-3 error frame to dst.
func AppendError(dst, payload []byte) ([]byte, error) {
	return AppendBand(dst, BandError, payload)
}
