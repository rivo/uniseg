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
	prNewline
	prWSegSpace
	prDoubleQuote
	prSingleQuote
	prMidNumLet
	prNumeric
	prMidLetter
	prMidNum
	prExtendNumLet
	prALetter
	prFormat
	prHebrewLetter
	prKatakana
)

// property returns the Unicode property value (see constants above) of the
// given code point.
func property(dictionary [][3]int, r rune) int {
	// Run a binary search.
	from := 0
	to := len(dictionary)
	for to > from {
		middle := (from + to) / 2
		cpRange := dictionary[middle]
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
