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
)

// property returns the Unicode property value (see constants above) of the
// given code point.
func property(r rune) int {
	// Run a binary search.
	from := 0
	to := len(graphemeCodePoints)
	for to > from {
		middle := (from + to) / 2
		cpRange := graphemeCodePoints[middle]
		if int(r) < cpRange[0] {
			to = middle
			continue
		}
		if int(r) > cpRange[1] {
			from = middle + 1
			continue
		}
		return cpRange[2]
	}
	return prAny
}
