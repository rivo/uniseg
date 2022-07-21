package uniseg

// The unicode properties. Only the ones needed in the context of this package
// are included.
const (
	prAny = iota
	prPrepend
	prCR
	prLF
	prControl
	prExtend
	prRegionalIndicator
	prSpacingMark
	prL
	prV
	prT
	prLV
	prLVT
	prZWJ
	prExtendedPictographic

	_prLast
)

// ASCII and Extended ASCII properties
var asciiGraphemeProperties = [256]uint8{
	'\x00': prControl, '\x01': prControl, '\x02': prControl, '\x03': prControl,
	'\x04': prControl, '\x05': prControl, '\x06': prControl, '\x07': prControl,
	'\x08': prControl, '\x09': prControl, '\x0a': prLF, '\x0b': prControl,
	'\x0c': prControl, '\x0d': prCR, '\x0e': prControl, '\x0f': prControl,
	'\x10': prControl, '\x11': prControl, '\x12': prControl, '\x13': prControl,
	'\x14': prControl, '\x15': prControl, '\x16': prControl, '\x17': prControl,
	'\x18': prControl, '\x19': prControl, '\x1a': prControl, '\x1b': prControl,
	'\x1c': prControl, '\x1d': prControl, '\x1e': prControl, '\x1f': prControl,
	'\x7f': prControl, '\x80': prControl, '\x81': prControl, '\x82': prControl,
	'\x83': prControl, '\x84': prControl, '\x85': prControl, '\x86': prControl,
	'\x87': prControl, '\x88': prControl, '\x89': prControl, '\x8a': prControl,
	'\x8b': prControl, '\x8c': prControl, '\x8d': prControl, '\x8e': prControl,
	'\x8f': prControl, '\x90': prControl, '\x91': prControl, '\x92': prControl,
	'\x93': prControl, '\x94': prControl, '\x95': prControl, '\x96': prControl,
	'\x97': prControl, '\x98': prControl, '\x99': prControl, '\x9a': prControl,
	'\x9b': prControl, '\x9c': prControl, '\x9d': prControl, '\x9e': prControl,
	'\x9f': prControl, '\xa9': prExtendedPictographic, '\xad': prControl,
	'\xae': prExtendedPictographic,
}

// property returns the Unicode property value (see constants above) of the
// given code point.
func property(r rune) int {
	// ASCII and Extended ASCII fast path
	if r < rune(len(asciiGraphemeProperties)) {
		return int(asciiGraphemeProperties[r])
	}

	// Run a binary search.
	from := 0
	to := len(graphemeCodePoints)
	for to > from {
		middle := int(uint(from+to) >> 1)
		p := graphemeCodePoints[middle]
		if r < p.lo {
			to = middle
		} else if r > p.hi {
			from = middle + 1
		} else {
			return int(p.property)
		}
	}
	return prAny
}
