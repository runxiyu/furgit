package pktline

import "fmt"

// AppendData appends one data frame to dst.
//
// Empty payload is encoded as 0004.
func AppendData(dst, payload []byte) ([]byte, error) {
	if len(payload) > LargePacketDataMax {
		return dst, fmt.Errorf("%w: %d > %d", ErrTooLarge, len(payload), LargePacketDataMax)
	}

	var hdr [4]byte

	err := EncodeLengthHeader(&hdr, len(payload)+4)
	if err != nil {
		return dst, err
	}

	dst = append(dst, hdr[:]...)
	dst = append(dst, payload...)

	return dst, nil
}

// AppendFlushPkt appends control frame 0000 (flush-pkt).
func AppendFlushPkt(dst []byte) []byte {
	return append(dst, '0', '0', '0', '0')
}

// AppendDelimPkt appends control frame 0001 (delim-pkt).
func AppendDelimPkt(dst []byte) []byte {
	return append(dst, '0', '0', '0', '1')
}

// AppendResponseEndPkt appends control frame 0002 (response-end-pkt).
func AppendResponseEndPkt(dst []byte) []byte {
	return append(dst, '0', '0', '0', '2')
}
