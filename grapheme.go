package uniseg

import "unicode/utf8"

// The states of the grapheme cluster parser.
const (
	grAny = iota
	grCR
	grControlLF
	grL
	grLVV
	grLVTT
	grPrepend
	grExtendedPictographic
	grExtendedPictographicZWJ
	grRIOdd
	grRIEven

	_grLast
)

type transition struct {
	newState uint8
	boundry  bool // breaking instruction
	rule     uint16
}

func (s *transition) isValid() bool { return s != nil }

// The grapheme cluster parser's state transitions. Maps (state, property) to
// (new state, breaking instruction, rule number). The breaking instruction
// always refers to the boundary between the last and next code point.
//
// This map is queried as follows:
//
//   1. Find specific state + specific property. Stop if found.
//   2. Find specific state + any property.
//   3. Find any state + specific property.
//   4. If only (2) or (3) (but not both) was found, stop.
//   5. If both (2) and (3) were found, use state from (3) and breaking instruction
//      from the transition with the lower rule number, prefer (3) if rule numbers
//      are equal. Stop.
//   6. Assume grAny and grBoundary.
var grStateTransitions = [_grLast][_prLast]*transition{
	grAny: {
		prPrepend:              {grPrepend, true, 9990},
		prCR:                   {grCR, true, 50},
		prLF:                   {grControlLF, true, 50},
		prControl:              {grControlLF, true, 50},
		prExtend:               {grAny, false, 90},
		prRegionalIndicator:    {grRIOdd, true, 9990},
		prSpacingMark:          {grAny, false, 91},
		prL:                    {grL, true, 9990},
		prV:                    {grLVV, true, 9990},
		prT:                    {grLVTT, true, 9990},
		prLV:                   {grLVV, true, 9990},
		prLVT:                  {grLVTT, true, 9990},
		prZWJ:                  {grAny, false, 90},
		prExtendedPictographic: {grExtendedPictographic, true, 9990},
	},
	grCR: {
		prAny: {grAny, true, 40},
		prLF:  {grAny, false, 30},
	},
	grControlLF: {
		prAny: {grAny, true, 40},
	},
	grL: {
		prL:   {grL, false, 60},
		prV:   {grLVV, false, 60},
		prLV:  {grLVV, false, 60},
		prLVT: {grLVTT, false, 60},
	},
	grLVV: {
		prV: {grLVV, false, 70},
		prT: {grLVTT, false, 70},
	},
	grLVTT: {
		prT: {grLVTT, false, 80},
	},
	grPrepend: {
		prAny: {grAny, false, 92},
	},
	grExtendedPictographic: {
		prExtend: {grExtendedPictographic, false, 110},
		prZWJ:    {grExtendedPictographicZWJ, false, 110},
	},
	grExtendedPictographicZWJ: {
		prExtendedPictographic: {grExtendedPictographic, false, 110},
	},
	grRIOdd: {
		prRegionalIndicator: {grRIEven, false, 120},
	},
	grRIEven: {
		prRegionalIndicator: {grRIOdd, true, 120},
	},
}

func lookupStateTransition(state, property int) *transition {
	return grStateTransitions[state][property]
}

// Graphemes implements an iterator over Unicode extended grapheme clusters,
// specified in the Unicode Standard Annex #29. Grapheme clusters correspond to
// "user-perceived characters". These characters often consist of multiple
// code points (e.g. the "woman kissing woman" emoji consists of 8 code points:
// woman + ZWJ + heavy black heart (2 code points) + ZWJ + kiss mark + ZWJ +
// woman) and the rules described in Annex #29 must be applied to group those
// code points into clusters perceived by the user as one character.
type Graphemes struct {
	// The code points over which this class iterates.
	codePoints []rune

	// The (byte-based) indices of the code points into the original string plus
	// len(original string). Thus, len(indices) = len(codePoints) + 1.
	indices []int

	// The current grapheme cluster to be returned. These are indices into
	// codePoints/indices. If start == end, we either haven't started iterating
	// yet (0) or the iteration has already completed (1).
	start, end int

	// The index of the next code point to be parsed.
	pos int

	// The current state of the code point parser.
	state int
}

// NewGraphemes returns a new grapheme cluster iterator.
func NewGraphemes(s string) *Graphemes {
	l := utf8.RuneCountInString(s)
	codePoints := make([]rune, l)
	indices := make([]int, l+1)
	i := 0
	for pos, r := range s {
		codePoints[i] = r
		indices[i] = pos
		i++
	}
	indices[l] = len(s)
	g := &Graphemes{
		codePoints: codePoints,
		indices:    indices,
	}
	g.Next() // Parse ahead.
	return g
}

// Next advances the iterator by one grapheme cluster and returns false if no
// clusters are left. This function must be called before the first cluster is
// accessed.
func (g *Graphemes) Next() bool {
	g.start = g.end

	// The state transition gives us a boundary instruction BEFORE the next code
	// point so we always need to stay ahead by one code point.

	// Parse the next code point.
	for g.pos <= len(g.codePoints) {
		// GB2.
		if g.pos == len(g.codePoints) {
			g.end = g.pos
			g.pos++
			break
		}

		// Calculate the next state.
		var boundary bool
		g.state, boundary = transitionGraphemeState(g.state, g.codePoints[g.pos])

		// If we found a cluster boundary, let's stop here. The current cluster will
		// be the one that just ended.
		if g.pos == 0 /* GB1 */ || boundary {
			g.end = g.pos
			g.pos++
			break
		}

		g.pos++
	}

	return g.start != g.end
}

// transitionGraphemeState determines the new state of the grapheme cluster
// parser given the current state and the next code point. It also returns
// whether a cluster boundary was detected.
func transitionGraphemeState(state int, r rune) (newState int, boundary bool) {
	// Determine the property of the next character.
	nextProperty := property(r)

	// Find the applicable transition.
	transition := lookupStateTransition(state, nextProperty)
	if transition.isValid() {
		// We have a specific transition. We'll use it.
		return int(transition.newState), transition.boundry
	}

	// No specific transition found. Try the less specific ones.
	transAnyProp := lookupStateTransition(state, prAny)
	transAnyState := lookupStateTransition(grAny, nextProperty)
	if transAnyProp.isValid() && transAnyState.isValid() {
		// Both apply. We'll use a mix (see comments for grTransitions).
		boundary = transAnyState.boundry
		if transAnyProp.rule < transAnyState.rule {
			boundary = transAnyProp.boundry
		}
		return int(transAnyState.newState), boundary
	}

	if transAnyState.isValid() {
		// We only have a specific property.
		return int(transAnyState.newState), transAnyState.boundry
	}

	if transAnyProp.isValid() {
		// We only have a specific state.
		return int(transAnyProp.newState), transAnyProp.boundry
		// This branch will probably never be reached because okAnyState will
		// always be true given the current transition map. But we keep it here
		// for future modifications to the transition map where this may not be
		// true anymore.
	}

	// No known transition. GB999: Any x Any.
	return grAny, true
}

// Runes returns a slice of runes (code points) which corresponds to the current
// grapheme cluster. If the iterator is already past the end or Next() has not
// yet been called, nil is returned.
func (g *Graphemes) Runes() []rune {
	if g.start == g.end {
		return nil
	}
	return g.codePoints[g.start:g.end]
}

// Str returns a substring of the original string which corresponds to the
// current grapheme cluster. If the iterator is already past the end or Next()
// has not yet been called, an empty string is returned.
func (g *Graphemes) Str() string {
	if g.start == g.end {
		return ""
	}
	return string(g.codePoints[g.start:g.end])
}

// Bytes returns a byte slice which corresponds to the current grapheme cluster.
// If the iterator is already past the end or Next() has not yet been called,
// nil is returned.
func (g *Graphemes) Bytes() []byte {
	if g.start == g.end {
		return nil
	}
	return []byte(string(g.codePoints[g.start:g.end]))
}

// Positions returns the interval of the current grapheme cluster as byte
// positions into the original string. The first returned value "from" indexes
// the first byte and the second returned value "to" indexes the first byte that
// is not included anymore, i.e. str[from:to] is the current grapheme cluster of
// the original string "str". If Next() has not yet been called, both values are
// 0. If the iterator is already past the end, both values are 1.
func (g *Graphemes) Positions() (int, int) {
	return g.indices[g.start], g.indices[g.end]
}

// Reset puts the iterator into its initial state such that the next call to
// Next() sets it to the first grapheme cluster again.
func (g *Graphemes) Reset() {
	g.start, g.end, g.pos, g.state = 0, 0, 0, grAny
	g.Next() // Parse ahead again.
}

// GraphemeClusterCount returns the number of user-perceived characters
// (grapheme clusters) for the given string. To calculate this number, it
// iterates through the string using the Graphemes iterator.
func GraphemeClusterCount(s string) (n int) {
	g := NewGraphemes(s)
	for g.Next() {
		n++
	}
	return
}

// firstGraphemeCluster returns the first grapheme cluster (as a slice of bytes)
// found in the given byte slice. This function can be called continuously to
// extract all grapheme clusters from a byte slice, as follows:
//
//   state := -1
//   for len(b) > 0 {
//       c, b, state = firstGraphemeCluster(b, state)
//       // Do something with c.
//   }
//
// If you don't know the current state, for example when calling the function
// for the first time, you must pass -1. Always passing -1 will work but it will
// slow down the function. For consecutive calls, you should pass the state
// returned by the previous call.
//
// The "rest" slice is the subslice of the original byte slice "b" starting
// after the last byte of the identified grapheme cluster. If the length of the
// "rest" slice is 0, the entire byte slice "b" has been processed.
//
// For an empty byte slice "b", the function returns nil values.
//
// Using this function is the preferred method of extracting grapheme clusters
// when working exclusively with byte slices and/or with large byte slices, as
// no large allocations are made.
//
// For the time being, this function is private because its signature might
// still change.
func firstGraphemeCluster(b []byte, state int) (cluster, rest []byte, newState int) {
	// An empty byte slice returns nothing.
	if len(b) == 0 {
		return
	}

	// Extract the first rune.
	r, length := utf8.DecodeRune(b)
	if len(b) <= length { // If we're already past the end, there is nothing else to parse.
		return b, nil, grAny
	}

	// If we don't know the state, determine it now.
	if state < 0 {
		state, _ = transitionGraphemeState(grAny, r)
	}

	// Transition until we find a boundary.
	var boundary bool
	for {
		r, l := utf8.DecodeRune(b[length:])
		state, boundary = transitionGraphemeState(state, r)

		if boundary {
			return b[:length], b[length:], state
		}

		length += l
		if len(b) <= length {
			return b, nil, grAny
		}
	}
}

// firstGraphemeClusterInString is like firstGraphemeCluster() but its input and
// outputs are a string.
func firstGraphemeClusterInString(str string, state int) (cluster, rest string, newState int) {
	// An empty string returns nothing.
	if len(str) == 0 {
		return
	}

	// Extract the first rune.
	r, length := utf8.DecodeRuneInString(str)
	if len(str) <= length { // If we're already past the end, there is nothing else to parse.
		return str, "", grAny
	}

	// If we don't know the state, determine it now.
	if state < 0 {
		state, _ = transitionGraphemeState(grAny, r)
	}

	// Transition until we find a boundary.
	var boundary bool
	for {
		r, l := utf8.DecodeRuneInString(str[length:])
		state, boundary = transitionGraphemeState(state, r)

		if boundary {
			return str[:length], str[length:], state
		}

		length += l
		if len(str) <= length {
			return str, "", grAny
		}
	}
}
