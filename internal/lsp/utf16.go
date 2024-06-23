package lsp

import "unicode/utf8"

// UTF16Len returns the number of UTF-16 code units required to encode the
// given UTF-8 byte slice.
func UTF16Len(s []byte) int {
	n := 0

	for len(s) > 0 {
		n++

		if s[0] < 0x80 {
			// ASCII optimization
			s = s[1:]
			continue
		}

		r, size := utf8.DecodeRune(s)

		if r >= 0x10000 {
			// UTF-16 surrogate pair
			n++
		}

		s = s[size:]
	}

	return n
}
