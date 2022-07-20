package uniseg

import "unicode/utf8"

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

	// The current state of the Grapheme code point parser.
	graphemeState int
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
		g.graphemeState, boundary = transitionGraphemeState(g.graphemeState, g.codePoints[g.pos])

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
	g.start, g.end, g.pos, g.graphemeState = 0, 0, 0, grAny
	g.Next() // Parse ahead again.
}

// GraphemeClusterCount returns the number of user-perceived characters
// (grapheme clusters) for the given string. To calculate this number, it
// iterates through the string using the Graphemes iterator.
func GraphemeClusterCount(s string) (n int) {
	state := -1
	for len(s) > 0 {
		_, s, state = firstGraphemeClusterInString(s, state)
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
// for the first time, you must pass -1. For consecutive calls, pass the state
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
