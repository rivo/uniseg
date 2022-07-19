package uniseg

// The states of the word break parser.
const (
	wbAny = iota
	wbCR
	wbLF
	wbNewline
	wbZWJ
	wbWSegSpace
)

// The word break parser's breaking instructions.
const (
	wbDontBreak = iota
	wbBreak
)

// The word break parser's state transitions. It's anologous to wbTransitions,
// see comments there for details.
var wbTransitions = map[[2]int][3]int{
	// WB3b.
	{wbAny, prNewline}: {wbNewline, wbBreak, 32},
	{wbAny, prCR}:      {wbCR, wbBreak, 32},
	{wbAny, prLF}:      {wbLF, wbBreak, 32},

	// WB3a.
	{wbNewline, prAny}: {wbAny, wbBreak, 31},
	{wbCR, prAny}:      {wbAny, wbBreak, 31},
	{wbLF, prAny}:      {wbAny, wbBreak, 31},

	// WB3.
	{wbCR, prLF}: {wbLF, wbDontBreak, 30},

	// WB3c.
	{wbAny, prZWJ}:                  {wbZWJ, wbBreak, 9990},
	{wbZWJ, prExtendedPictographic}: {wbAny, wbDontBreak, 33},

	// WB3d.
	{wbAny, prWSegSpace}:       {wbWSegSpace, wbBreak, 9990},
	{wbWSegSpace, prWSegSpace}: {wbWSegSpace, wbDontBreak, 34},
}

// transitionWordBreakState determines the new state of the word break parser
// given the current state and the next code point. It also returns whether a
// word break was detected.
func transitionWordBreakState(state int, r rune) (newState int, wordBreak bool) {
	// Determine the property of the next character.
	nextProperty := property(workBreakCodePoints, r)

	// Find the applicable transition.
	transition, ok := wbTransitions[[2]int{state, nextProperty}]
	if ok {
		// We have a specific transition. We'll use it.
		return transition[0], transition[1] == wbBreak
	}

	// No specific transition found. Try the less specific ones.
	transAnyProp, okAnyProp := wbTransitions[[2]int{state, prAny}]
	transAnyState, okAnyState := wbTransitions[[2]int{wbAny, nextProperty}]
	if okAnyProp && okAnyState {
		// Both apply. We'll use a mix (see comments for wbTransitions).
		newState = transAnyState[0]
		wordBreak = transAnyState[1] == wbBreak
		if transAnyProp[2] < transAnyState[2] {
			wordBreak = transAnyProp[1] == wbBreak
		}
		return
	}

	if okAnyProp {
		// We only have a specific state.
		return transAnyProp[0], transAnyProp[1] == wbBreak
		// This branch will probably never be reached because okAnyState will
		// always be true given the current transition map. But we keep it here
		// for future modifications to the transition map where this may not be
		// true anymore.
	}

	if okAnyState {
		// We only have a specific property.
		return transAnyState[0], transAnyState[1] == wbBreak
	}

	// No known transition. WB999: Any ÷ Any.
	return wbAny, true
}
