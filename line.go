package uniseg

import "unicode/utf8"

//TODO: Adapt from firstWord() when making it public.
//TODO: Add description of "mustBreak".
func firstLineSegment(b []byte, state int) (segment, rest []byte, mustBreak bool, newState int) {
	// An empty byte slice returns nothing.
	if len(b) == 0 {
		return
	}

	// Extract the first rune.
	r, length := utf8.DecodeRune(b)
	if len(b) <= length { // If we're already past the end, there is nothing else to parse.
		return b, nil, true, lbAny // LB3.
	}

	// If we don't know the state, determine it now.
	if state < 0 {
		state, _ = transitionLineBreakState(state, r, b[length:], "")
	}

	// Transition until we find a boundary.
	var boundary int
	for {
		r, l := utf8.DecodeRune(b[length:])
		state, boundary = transitionLineBreakState(state, r, b[length+l:], "")

		if boundary != LineDontBreak {
			return b[:length], b[length:], boundary == LineMustBreak, state
		}

		length += l
		if len(b) <= length {
			return b, nil, true, lbAny // LB3
		}
	}
}

// firstLineSegmentInString is like firstLineSegment() but its input and outputs
// are a string.
func firstLineSegmentInString(str string, state int) (sentence, rest string, mustBreak bool, newState int) {
	// An empty byte slice returns nothing.
	if len(str) == 0 {
		return
	}

	// Extract the first rune.
	r, length := utf8.DecodeRuneInString(str)
	if len(str) <= length { // If we're already past the end, there is nothing else to parse.
		return str, "", true, lbAny // LB3.
	}

	// If we don't know the state, determine it now.
	if state < 0 {
		state, _ = transitionLineBreakState(state, r, nil, str[length:])
	}

	// Transition until we find a boundary.
	var boundary int
	for {
		r, l := utf8.DecodeRuneInString(str[length:])
		state, boundary = transitionLineBreakState(state, r, nil, str[length+l:])

		if boundary != LineDontBreak {
			return str[:length], str[length:], boundary == LineMustBreak, state
		}

		length += l
		if len(str) <= length {
			return str, "", true, lbAny // LB3.
		}
	}
}
