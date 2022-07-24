package uniseg

import "unicode/utf8"

//TODO: Adapt from firstWord() when making it public.
func firstSentence(b []byte, state int) (sentence, rest []byte, newState int) {
	// An empty byte slice returns nothing.
	if len(b) == 0 {
		return
	}

	// Extract the first rune.
	r, length := utf8.DecodeRune(b)
	if len(b) <= length { // If we're already past the end, there is nothing else to parse.
		return b, nil, sbAny
	}

	// If we don't know the state, determine it now.
	if state < 0 {
		state, _ = transitionSentenceBreakState(state, r, b[length:], "")
	}

	// Transition until we find a boundary.
	var boundary bool
	for {
		r, l := utf8.DecodeRune(b[length:])
		state, boundary = transitionSentenceBreakState(state, r, b[length+l:], "")

		if boundary {
			return b[:length], b[length:], state
		}

		length += l
		if len(b) <= length {
			return b, nil, sbAny
		}
	}
}

// firstSentenceInString is like firstSentence() but its input and outputs are a
// string.
func firstSentenceInString(str string, state int) (sentence, rest string, newState int) {
	// An empty byte slice returns nothing.
	if len(str) == 0 {
		return
	}

	// Extract the first rune.
	r, length := utf8.DecodeRuneInString(str)
	if len(str) <= length { // If we're already past the end, there is nothing else to parse.
		return str, "", sbAny
	}

	// If we don't know the state, determine it now.
	if state < 0 {
		state, _ = transitionSentenceBreakState(state, r, nil, str[length:])
	}

	// Transition until we find a boundary.
	var boundary bool
	for {
		r, l := utf8.DecodeRuneInString(str[length:])
		state, boundary = transitionSentenceBreakState(state, r, nil, str[length+l:])

		if boundary {
			return str[:length], str[length:], state
		}

		length += l
		if len(str) <= length {
			return str, "", sbAny
		}
	}
}
