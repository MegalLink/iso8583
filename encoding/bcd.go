package encoding

import (
	"github.com/franizus/go-util/bcd"
)

var BCD Encoder = &bcdEncoder{}

type bcdEncoder struct{}

func (e *bcdEncoder) Encode(src []byte) ([]byte, error) {
	if len(src)%2 != 0 {
		src = append([]byte("0"), src...)
	}

	enc := bcd.NewEncoder(bcd.Visa)
	dst := make([]byte, bcd.EncodedLen(len(src)))
	n, err := enc.Encode(dst, src)
	if err != nil {
		return nil, err
	}

	return dst[:n], nil
}

func (e *bcdEncoder) Decode(src []byte, length int) ([]byte, int, error) {
	// for BCD encoding the length should be even
	decodedLen := length
	if length%2 != 0 {
		decodedLen += 1
	}

	// how many bytes we will read
	read := bcd.EncodedLen(decodedLen)

	dec := bcd.NewDecoder(bcd.Visa)
	dst := make([]byte, decodedLen)
	_, err := dec.Decode(dst, src)
	if err != nil {
		return nil, 0, err
	}

	// becase BCD is right aligned, we skip first bytes and
	// read only what we need
	// e.g. 0643 => 643
	return dst[decodedLen-length:], read, nil
}
