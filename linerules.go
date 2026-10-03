package uniseg

import "unicode/utf8"

// The states of the line break parser.
const (
	lbAny = iota
	lbBK
	lbCR
	lbLF
	lbNL
	lbSP
	lbZW
	lbWJ
	lbGL
	lbBA
	lbHY
	lbCL
	lbCP
	lbEX
	lbIS
	lbSY
	lbOP
	lbQU
	lbQUSP // lbQUPi followed by at least one SP (LB15a).
	lbNS
	lbCLCPSP
	lbB2
	lbB2SP
	lbCB
	lbBB
	lbLB21a
	lbHL
	lbAL
	lbNU
	lbPR
	lbEB
	lbIDEM
	lbNUNU
	lbNUSY
	lbNUIS
	lbNUCL
	lbNUCP
	lbPO
	lbJL
	lbJV
	lbJT
	lbH2
	lbH3
	lbOddRI
	lbEvenRI
	lbExtPicCn     // An unassigned Extended_Pictographic with class ID (LB30b).
	lbExtPicCnAL   // An unassigned Extended_Pictographic with class AL (LB30b).
	lbQUPi         // An initial punctuation QU in the context of LB15a.
	lbQUPf         // A final punctuation QU preceded by an East Asian character (LB19a).
	lbOPSP         // OP followed by at least one SP (LB14).
	lbHH           // Unambiguous hyphen.
	lbHYStart      // A word-initial HY (LB20a).
	lbHHStart      // A word-initial HH (LB20a).
	lbAK           // AK or AS (LB28a).
	lbDottedCircle // U+25CC DOTTED CIRCLE, an AL which is also used in LB28a.
	lbAP           // Aksara prebase (LB28a).
	lbAKVI         // (AK | ◌ | AS) VI (LB28a).

	// Bits which may be combined with the states above.
	lbZWJBit       = 64  // The last character was a zero width joiner (LB8a).
	lbEastAsianBit = 128 // The last base character's East_Asian_Width is F, W, or H.
)

// These constants define whether a given text may be broken into the next line.
// If the break is optional (LineCanBreak), you may choose to break or not based
// on your own criteria, for example, if the text has reached the available
// width.
const (
	LineDontBreak = iota // You may not break the line here.
	LineCanBreak         // You may or may not break the line here.
	LineMustBreak        // You must break the line here.
)

// lbTransitions implements the line break parser's state transitions. It's
// anologous to [grTransitions], see comments there for details.
//
// Unicode version 18.0.0.
func lbTransitions(state, prop int) (newState, lineBreak, rule int) {
	switch uint64(state) | uint64(prop)<<32 {
	// LB4.
	case lbBK | prAny<<32:
		return lbAny, LineMustBreak, 40

	// LB5.
	case lbCR | prLF<<32:
		return lbLF, LineDontBreak, 50
	case lbCR | prAny<<32:
		return lbAny, LineMustBreak, 50
	case lbLF | prAny<<32:
		return lbAny, LineMustBreak, 50
	case lbNL | prAny<<32:
		return lbAny, LineMustBreak, 50

	// LB6.
	case lbAny | prBK<<32:
		return lbBK, LineDontBreak, 60
	case lbAny | prCR<<32:
		return lbCR, LineDontBreak, 60
	case lbAny | prLF<<32:
		return lbLF, LineDontBreak, 60
	case lbAny | prNL<<32:
		return lbNL, LineDontBreak, 60

	// LB7.
	case lbAny | prSP<<32:
		return lbSP, LineDontBreak, 70
	case lbAny | prZW<<32:
		return lbZW, LineDontBreak, 70

	// LB8.
	case lbZW | prSP<<32:
		return lbZW, LineDontBreak, 70
	case lbZW | prAny<<32:
		return lbAny, LineCanBreak, 80

	// LB11.
	case lbAny | prWJ<<32:
		return lbWJ, LineDontBreak, 110
	case lbWJ | prAny<<32:
		return lbAny, LineDontBreak, 110

	// LB12.
	case lbAny | prGL<<32:
		return lbGL, LineCanBreak, 310
	case lbGL | prAny<<32:
		return lbAny, LineDontBreak, 120

	// LB13.
	case lbAny | prCL<<32:
		return lbCL, LineDontBreak, 130
	case lbAny | prCP<<32:
		return lbCP, LineDontBreak, 130
	case lbAny | prEX<<32:
		return lbEX, LineDontBreak, 130
	case lbAny | prSY<<32:
		return lbSY, LineDontBreak, 130

	// LB14.
	case lbAny | prOP<<32:
		return lbOP, LineCanBreak, 310
	case lbOP | prSP<<32:
		return lbOPSP, LineDontBreak, 70
	case lbOPSP | prSP<<32:
		return lbOPSP, LineDontBreak, 70
	case lbOP | prAny<<32:
		return lbAny, LineDontBreak, 140
	case lbOPSP | prAny<<32:
		return lbAny, LineDontBreak, 140

	// LB15a (the QU states are determined in transitionLineBreakState).
	case lbQUPi | prSP<<32:
		return lbQUSP, LineDontBreak, 70
	case lbQUSP | prSP<<32:
		return lbQUSP, LineDontBreak, 70
	case lbQUPi | prAny<<32:
		return lbAny, LineDontBreak, 151
	case lbQUSP | prAny<<32:
		return lbAny, LineDontBreak, 151

	// LB15d.
	case lbAny | prIS<<32:
		return lbIS, LineDontBreak, 154

	// LB16.
	case lbCL | prSP<<32:
		return lbCLCPSP, LineDontBreak, 70
	case lbNUCL | prSP<<32:
		return lbCLCPSP, LineDontBreak, 70
	case lbCP | prSP<<32:
		return lbCLCPSP, LineDontBreak, 70
	case lbNUCP | prSP<<32:
		return lbCLCPSP, LineDontBreak, 70
	case lbCL | prNS<<32:
		return lbNS, LineDontBreak, 160
	case lbNUCL | prNS<<32:
		return lbNS, LineDontBreak, 160
	case lbCP | prNS<<32:
		return lbNS, LineDontBreak, 160
	case lbNUCP | prNS<<32:
		return lbNS, LineDontBreak, 160
	case lbCLCPSP | prNS<<32:
		return lbNS, LineDontBreak, 160

	// LB17.
	case lbAny | prB2<<32:
		return lbB2, LineCanBreak, 310
	case lbB2 | prSP<<32:
		return lbB2SP, LineDontBreak, 70
	case lbB2 | prB2<<32:
		return lbB2, LineDontBreak, 170
	case lbB2SP | prB2<<32:
		return lbB2, LineDontBreak, 170

	// LB18.
	case lbSP | prAny<<32:
		return lbAny, LineCanBreak, 180
	case lbCLCPSP | prAny<<32:
		return lbAny, LineCanBreak, 180
	case lbB2SP | prAny<<32:
		return lbAny, LineCanBreak, 180

	// LB19 and LB19a (the remaining cases are handled in
	// transitionLineBreakState).
	case lbAny | prQU<<32:
		return lbQU, LineCanBreak, 310
	case lbQU | prAny<<32:
		return lbAny, LineDontBreak, 190

	// LB20.
	case lbAny | prCB<<32:
		return lbCB, LineCanBreak, 200
	case lbCB | prAny<<32:
		return lbAny, LineCanBreak, 200

	// LB20a (the HY and HH start states are determined in
	// transitionLineBreakState).
	case lbHYStart | prAL<<32:
		return lbAL, LineDontBreak, 201
	case lbHYStart | prHL<<32:
		return lbHL, LineDontBreak, 201
	case lbHHStart | prAL<<32:
		return lbAL, LineDontBreak, 201
	case lbHHStart | prHL<<32:
		return lbHL, LineDontBreak, 201

	// LB21.
	case lbAny | prBA<<32:
		return lbBA, LineDontBreak, 210
	case lbAny | prHY<<32:
		return lbHY, LineDontBreak, 210
	case lbAny | prHH<<32:
		return lbHH, LineDontBreak, 210
	case lbAny | prNS<<32:
		return lbNS, LineDontBreak, 210
	case lbAny | prBB<<32:
		return lbBB, LineCanBreak, 310
	case lbBB | prAny<<32:
		return lbAny, LineDontBreak, 210

	// LB21a.
	case lbAny | prHL<<32:
		return lbHL, LineCanBreak, 310
	case lbHL | prHY<<32:
		return lbLB21a, LineDontBreak, 210
	case lbHL | prHH<<32:
		return lbLB21a, LineDontBreak, 210
	case lbLB21a | prAny<<32:
		return lbAny, LineDontBreak, 211
	case lbLB21a | prHL<<32:
		return lbHL, LineCanBreak, 310

	// LB21b.
	case lbSY | prHL<<32:
		return lbHL, LineDontBreak, 212
	case lbNUSY | prHL<<32:
		return lbHL, LineDontBreak, 212

	// LB22.
	case lbAny | prIN<<32:
		return lbAny, LineDontBreak, 220

	// LB23.
	case lbAny | prAL<<32:
		return lbAL, LineCanBreak, 310
	case lbAny | prNU<<32:
		return lbNU, LineCanBreak, 310
	case lbAL | prNU<<32:
		return lbNU, LineDontBreak, 230
	case lbHL | prNU<<32:
		return lbNU, LineDontBreak, 230
	case lbNU | prAL<<32:
		return lbAL, LineDontBreak, 230
	case lbNU | prHL<<32:
		return lbHL, LineDontBreak, 230
	case lbNUNU | prAL<<32:
		return lbAL, LineDontBreak, 230
	case lbNUNU | prHL<<32:
		return lbHL, LineDontBreak, 230

	// LB23a.
	case lbAny | prPR<<32:
		return lbPR, LineCanBreak, 310
	case lbAny | prID<<32:
		return lbIDEM, LineCanBreak, 310
	case lbAny | prEB<<32:
		return lbEB, LineCanBreak, 310
	case lbAny | prEM<<32:
		return lbIDEM, LineCanBreak, 310
	case lbPR | prID<<32:
		return lbIDEM, LineDontBreak, 231
	case lbPR | prEB<<32:
		return lbEB, LineDontBreak, 231
	case lbPR | prEM<<32:
		return lbIDEM, LineDontBreak, 231
	case lbIDEM | prPO<<32:
		return lbPO, LineDontBreak, 231
	case lbEB | prPO<<32:
		return lbPO, LineDontBreak, 231

	// LB24.
	case lbAny | prPO<<32:
		return lbPO, LineCanBreak, 310
	case lbPR | prAL<<32:
		return lbAL, LineDontBreak, 240
	case lbPR | prHL<<32:
		return lbHL, LineDontBreak, 240
	case lbPO | prAL<<32:
		return lbAL, LineDontBreak, 240
	case lbPO | prHL<<32:
		return lbHL, LineDontBreak, 240
	case lbAL | prPR<<32:
		return lbPR, LineDontBreak, 240
	case lbAL | prPO<<32:
		return lbPO, LineDontBreak, 240
	case lbHL | prPR<<32:
		return lbPR, LineDontBreak, 240
	case lbHL | prPO<<32:
		return lbPO, LineDontBreak, 240

	// LB25 (simple transitions).
	case lbPR | prNU<<32:
		return lbNU, LineDontBreak, 250
	case lbPO | prNU<<32:
		return lbNU, LineDontBreak, 250
	case lbOP | prNU<<32:
		return lbNU, LineDontBreak, 250
	case lbHY | prNU<<32:
		return lbNU, LineDontBreak, 250
	case lbHYStart | prNU<<32:
		return lbNU, LineDontBreak, 250
	case lbIS | prNU<<32:
		return lbNU, LineDontBreak, 250
	case lbNU | prNU<<32:
		return lbNUNU, LineDontBreak, 250
	case lbNU | prSY<<32:
		return lbNUSY, LineDontBreak, 250
	case lbNU | prIS<<32:
		return lbNUIS, LineDontBreak, 250
	case lbNUNU | prNU<<32:
		return lbNUNU, LineDontBreak, 250
	case lbNUNU | prSY<<32:
		return lbNUSY, LineDontBreak, 250
	case lbNUNU | prIS<<32:
		return lbNUIS, LineDontBreak, 250
	case lbNUSY | prNU<<32:
		return lbNUNU, LineDontBreak, 250
	case lbNUSY | prSY<<32:
		return lbNUSY, LineDontBreak, 250
	case lbNUSY | prIS<<32:
		return lbNUIS, LineDontBreak, 250
	case lbNUIS | prNU<<32:
		return lbNUNU, LineDontBreak, 250
	case lbNUIS | prSY<<32:
		return lbNUSY, LineDontBreak, 250
	case lbNUIS | prIS<<32:
		return lbNUIS, LineDontBreak, 250
	case lbNU | prCL<<32:
		return lbNUCL, LineDontBreak, 250
	case lbNU | prCP<<32:
		return lbNUCP, LineDontBreak, 250
	case lbNUNU | prCL<<32:
		return lbNUCL, LineDontBreak, 250
	case lbNUNU | prCP<<32:
		return lbNUCP, LineDontBreak, 250
	case lbNUSY | prCL<<32:
		return lbNUCL, LineDontBreak, 250
	case lbNUSY | prCP<<32:
		return lbNUCP, LineDontBreak, 250
	case lbNUIS | prCL<<32:
		return lbNUCL, LineDontBreak, 250
	case lbNUIS | prCP<<32:
		return lbNUCP, LineDontBreak, 250
	case lbNU | prPO<<32:
		return lbPO, LineDontBreak, 250
	case lbNUNU | prPO<<32:
		return lbPO, LineDontBreak, 250
	case lbNUSY | prPO<<32:
		return lbPO, LineDontBreak, 250
	case lbNUIS | prPO<<32:
		return lbPO, LineDontBreak, 250
	case lbNUCL | prPO<<32:
		return lbPO, LineDontBreak, 250
	case lbNUCP | prPO<<32:
		return lbPO, LineDontBreak, 250
	case lbNU | prPR<<32:
		return lbPR, LineDontBreak, 250
	case lbNUNU | prPR<<32:
		return lbPR, LineDontBreak, 250
	case lbNUSY | prPR<<32:
		return lbPR, LineDontBreak, 250
	case lbNUIS | prPR<<32:
		return lbPR, LineDontBreak, 250
	case lbNUCL | prPR<<32:
		return lbPR, LineDontBreak, 250
	case lbNUCP | prPR<<32:
		return lbPR, LineDontBreak, 250

	// LB26.
	case lbAny | prJL<<32:
		return lbJL, LineCanBreak, 310
	case lbAny | prJV<<32:
		return lbJV, LineCanBreak, 310
	case lbAny | prJT<<32:
		return lbJT, LineCanBreak, 310
	case lbAny | prH2<<32:
		return lbH2, LineCanBreak, 310
	case lbAny | prH3<<32:
		return lbH3, LineCanBreak, 310
	case lbJL | prJL<<32:
		return lbJL, LineDontBreak, 260
	case lbJL | prJV<<32:
		return lbJV, LineDontBreak, 260
	case lbJL | prH2<<32:
		return lbH2, LineDontBreak, 260
	case lbJL | prH3<<32:
		return lbH3, LineDontBreak, 260
	case lbJV | prJV<<32:
		return lbJV, LineDontBreak, 260
	case lbJV | prJT<<32:
		return lbJT, LineDontBreak, 260
	case lbH2 | prJV<<32:
		return lbJV, LineDontBreak, 260
	case lbH2 | prJT<<32:
		return lbJT, LineDontBreak, 260
	case lbJT | prJT<<32:
		return lbJT, LineDontBreak, 260
	case lbH3 | prJT<<32:
		return lbJT, LineDontBreak, 260

	// LB27.
	case lbJL | prPO<<32:
		return lbPO, LineDontBreak, 270
	case lbJV | prPO<<32:
		return lbPO, LineDontBreak, 270
	case lbJT | prPO<<32:
		return lbPO, LineDontBreak, 270
	case lbH2 | prPO<<32:
		return lbPO, LineDontBreak, 270
	case lbH3 | prPO<<32:
		return lbPO, LineDontBreak, 270
	case lbPR | prJL<<32:
		return lbJL, LineDontBreak, 270
	case lbPR | prJV<<32:
		return lbJV, LineDontBreak, 270
	case lbPR | prJT<<32:
		return lbJT, LineDontBreak, 270
	case lbPR | prH2<<32:
		return lbH2, LineDontBreak, 270
	case lbPR | prH3<<32:
		return lbH3, LineDontBreak, 270

	// LB28.
	case lbAL | prAL<<32:
		return lbAL, LineDontBreak, 280
	case lbAL | prHL<<32:
		return lbHL, LineDontBreak, 280
	case lbHL | prAL<<32:
		return lbAL, LineDontBreak, 280
	case lbHL | prHL<<32:
		return lbHL, LineDontBreak, 280

	// LB28a (the remaining cases are handled in transitionLineBreakState).
	case lbAny | prAK<<32:
		return lbAK, LineCanBreak, 310
	case lbAny | prAS<<32:
		return lbAK, LineCanBreak, 310
	case lbAny | prAP<<32:
		return lbAP, LineCanBreak, 310

	// LB29.
	case lbIS | prAL<<32:
		return lbAL, LineDontBreak, 290
	case lbIS | prHL<<32:
		return lbHL, LineDontBreak, 290
	case lbNUIS | prAL<<32:
		return lbAL, LineDontBreak, 290
	case lbNUIS | prHL<<32:
		return lbHL, LineDontBreak, 290

	default:
		return -1, -1, -1
	}
}

// transitionLineBreakState determines the new state of the line break parser
// given the current state and the next code point. It also returns the type of
// line break: LineDontBreak, LineCanBreak, or LineMustBreak. If more than one
// code point is needed to determine the new state, the byte slice or the string
// starting after rune "r" can be used (whichever is not nil or empty) for
// further lookups.
func transitionLineBreakState(state int, r rune, b []byte, str string) (newState int, lineBreak int) {
	// Determine the property of the next character.
	nextProperty, generalCategory := propertyLineBreak(r)

	// Prepare.
	var forceNoBreak, isEastAsian bool
	if state >= 0 {
		forceNoBreak = state&lbZWJBit != 0 // LB8a.
		isEastAsian = state&lbEastAsianBit != 0
		state &^= lbZWJBit | lbEastAsianBit
	}

	// LB1.
	nextProperty = resolveLineBreakClass(nextProperty, generalCategory)

	// Combining marks.
	var zwjBit int
	if nextProperty == prZWJ || nextProperty == prCM {
		if nextProperty == prZWJ {
			zwjBit = lbZWJBit
		}
		if state >= 0 && !lbIsSpaceOrMandatoryBreak(state) {
			// LB9.
			if isEastAsian {
				return state | zwjBit | lbEastAsianBit, LineDontBreak
			}
			return state | zwjBit, LineDontBreak
		}

		// LB10. Treat it as if it were U+0041 LATIN CAPITAL LETTER A.
		r, nextProperty, generalCategory = 'A', prAL, gcLu
	}

	// The following rules need to know whether the next character is East
	// Asian (LB19a, LB30).
	ea := propertyEastAsianWidth(r)
	nextIsEastAsian := ea == prF || ea == prW || ea == prH

	// Some states behave like other states except for specific rules.
	prevState := state
	switch state {
	case lbDottedCircle, lbExtPicCnAL:
		state = lbAL
	case lbExtPicCn:
		state = lbIDEM
	}
	afterDottedCircle := prevState == lbDottedCircle

	// Find the applicable transition in the table.
	var rule int
	newState, lineBreak, rule = lbTransitions(state, nextProperty)
	if newState < 0 {
		// No specific transition found. Try the less specific ones.
		anyPropProp, anyPropLineBreak, anyPropRule := lbTransitions(state, prAny)
		anyStateProp, anyStateLineBreak, anyStateRule := lbTransitions(lbAny, nextProperty)
		if anyPropProp >= 0 && anyStateProp >= 0 {
			// Both apply. We'll use a mix (see comments for grTransitions).
			newState, lineBreak, rule = anyStateProp, anyStateLineBreak, anyStateRule
			if anyPropRule < anyStateRule {
				lineBreak, rule = anyPropLineBreak, anyPropRule
			}
		} else if anyPropProp >= 0 {
			// We only have a specific state.
			newState, lineBreak, rule = anyPropProp, anyPropLineBreak, anyPropRule
			// This branch will probably never be reached because okAnyState will
			// always be true given the current transition map. But we keep it here
			// for future modifications to the transition map where this may not be
			// true anymore.
		} else if anyStateProp >= 0 {
			// We only have a specific property.
			newState, lineBreak, rule = anyStateProp, anyStateLineBreak, anyStateRule
		} else {
			// No known transition. LB31: ALL ÷ ALL.
			newState, lineBreak, rule = lbAny, LineCanBreak, 310
		}
	}

	// LB12a.
	if rule > 121 &&
		nextProperty == prGL &&
		!lbIsSpace(state) && state != lbHY && state != lbHYStart && state != lbHH && state != lbHHStart && state != lbLB21a {
		newState, lineBreak, rule = lbGL, LineDontBreak, 121
	}

	// LB15b.
	if rule > 152 && nextProperty == prQU && generalCategory == gcPf {
		switch lbLookAhead(b, str) {
		case prSP, prGL, prWJ, prCL, prQU, prCP, prEX, prIS, prSY, prBK, prCR, prLF, prNL, prZW, -1:
			lineBreak, rule = LineDontBreak, 152
		}
	}

	// LB15c.
	if rule > 153 &&
		nextProperty == prIS &&
		(state == lbSP || state == lbCLCPSP || state == lbB2SP) &&
		lbLookAhead(b, str) == prNU {
		newState, lineBreak, rule = lbIS, LineCanBreak, 153
	}

	// LB19 and LB19a (before QU).
	if rule > 191 && nextProperty == prQU {
		if generalCategory != gcPi || !isEastAsian || !lbLookAheadIsEastAsian(b, str) {
			lineBreak, rule = LineDontBreak, 190
		}
	}

	// LB19a (after QU).
	if rule > 191 && state == lbQUPf && !nextIsEastAsian {
		lineBreak, rule = LineDontBreak, 191
	}

	// LB25 (look ahead).
	if rule > 250 && (state == lbPR || state == lbPO) && nextProperty == prOP {
		pr, _, rb, rs := lbLookAheadRest(b, str)
		if pr == prIS {
			pr, _, _, _ = lbLookAheadRest(rb, rs)
		}
		if pr == prNU {
			newState, lineBreak, rule = lbOP, LineDontBreak, 250
		}
	}

	// LB28a.
	if rule > 285 {
		nextIsAK := nextProperty == prAK || r == 0x25cc || nextProperty == prAS
		afterAK := state == lbAK || afterDottedCircle
		if state == lbAP && nextIsAK ||
			afterAK && (nextProperty == prVF || nextProperty == prVI) ||
			state == lbAKVI && (nextProperty == prAK || r == 0x25cc) {
			lineBreak, rule = LineDontBreak, 285
		} else if afterAK && nextIsAK {
			if lbLookAhead(b, str) == prVF {
				lineBreak, rule = LineDontBreak, 285
			}
		}
	}

	// LB30.
	if rule > 300 {
		if (state == lbAL || state == lbHL || state == lbNU || state == lbNUNU) && nextProperty == prOP && !nextIsEastAsian {
			newState, lineBreak, rule = lbOP, LineDontBreak, 300
		} else if (state == lbCP || state == lbNUCP) && !isEastAsian &&
			(nextProperty == prAL || nextProperty == prHL || nextProperty == prNU) {
			lineBreak, rule = LineDontBreak, 300
		}
	}

	// LB30a.
	if nextProperty == prRI {
		if state == lbOddRI {
			// Don't break pairs of Regional Indicators.
			newState, lineBreak = lbEvenRI, LineDontBreak
		} else {
			newState = lbOddRI
		}
	}

	// LB30b.
	if rule > 302 && nextProperty == prEM &&
		(state == lbEB || prevState == lbExtPicCn || prevState == lbExtPicCnAL) {
		lineBreak, rule = LineDontBreak, 302
	}

	// Determine the states which depend on more context.
	switch {
	case nextProperty == prQU:
		if generalCategory == gcPi && lbIsLB15aContext(state) {
			newState = lbQUPi
		} else if generalCategory == gcPf && isEastAsian {
			newState = lbQUPf
		}
	case newState == lbHY && lbIsLB20aContext(state):
		newState = lbHYStart
	case newState == lbHH && lbIsLB20aContext(state):
		newState = lbHHStart
	case nextProperty == prVI && (state == lbAK || afterDottedCircle):
		newState = lbAKVI
	case r == 0x25cc && newState == lbAL:
		newState = lbDottedCircle
	case (generalCategory == gcCn || generalCategory == gcNone) && propertyGraphemes(r) == prExtendedPictographic:
		// Remember this for LB30b. (Code points without a general category
		// are not listed in the line break table and are unassigned.)
		switch newState {
		case lbIDEM:
			newState = lbExtPicCn
		case lbAL:
			newState = lbExtPicCnAL
		}
	}

	// Add bits.
	newState |= zwjBit
	if nextIsEastAsian {
		newState |= lbEastAsianBit
	}

	// LB8a.
	if forceNoBreak {
		lineBreak = LineDontBreak
	}

	return
}

// resolveLineBreakClass resolves the line breaking classes AI, CJ, SA, SG, and
// XX according to LB1.
func resolveLineBreakClass(property, generalCategory int) int {
	switch property {
	case prAI, prSG, prXX:
		return prAL
	case prSA:
		if generalCategory == gcMn || generalCategory == gcMc {
			return prCM
		}
		return prAL
	case prCJ:
		return prNS
	}
	return property
}

// lbIsSpace returns true if the given line break state (without bits) follows
// a space character.
func lbIsSpace(state int) bool {
	switch state {
	case lbSP, lbQUSP, lbCLCPSP, lbB2SP, lbOPSP:
		return true
	}
	return false
}

// lbIsSpaceOrMandatoryBreak returns true if the given line break state (without
// bits) follows one of the classes BK, CR, LF, NL, SP, or ZW, i.e. the states
// to which LB9 does not apply.
func lbIsSpaceOrMandatoryBreak(state int) bool {
	switch state {
	case lbBK, lbCR, lbLF, lbNL, lbZW:
		return true
	}
	return lbIsSpace(state)
}

// lbIsLB15aContext returns true if the given line break state (without bits)
// matches the left side of LB15a: (sot | BK | CR | LF | NL | OP | QU | GL | SP |
// ZW).
func lbIsLB15aContext(state int) bool {
	switch state {
	case lbOP, lbQU, lbQUPi, lbQUPf, lbGL:
		return true
	}
	return state < 0 || lbIsSpaceOrMandatoryBreak(state)
}

// lbIsLB20aContext returns true if the given line break state (without bits)
// matches the left side of LB20a: (sot | BK | CR | LF | NL | SP | ZW | CB |
// GL).
func lbIsLB20aContext(state int) bool {
	switch state {
	case lbCB, lbGL:
		return true
	}
	return state < 0 || lbIsSpaceOrMandatoryBreak(state)
}

// lbLookAhead returns the resolved line breaking class (see LB1) of the first
// character in the byte slice "b" (if not nil) or the string "str", skipping
// over leading combining marks and zero width joiners which belong to the
// preceding character (LB9). It returns -1 if there is no such character.
func lbLookAhead(b []byte, str string) int {
	property, _, _, _ := lbLookAheadRest(b, str)
	return property
}

// lbLookAheadIsEastAsian is like [lbLookAhead] but returns whether the found
// character's East_Asian_Width is F, W, or H.
func lbLookAheadIsEastAsian(b []byte, str string) bool {
	property, r, _, _ := lbLookAheadRest(b, str)
	if property < 0 {
		return false
	}
	ea := propertyEastAsianWidth(r)
	return ea == prF || ea == prW || ea == prH
}

// lbLookAheadRest is like [lbLookAhead] but also returns the found character
// and the remaining byte slice or string after it.
func lbLookAheadRest(b []byte, str string) (property int, r rune, restB []byte, restStr string) {
	for {
		var length int
		if b != nil { // Byte slice version.
			r, length = utf8.DecodeRune(b)
			b = b[length:]
		} else { // String version.
			r, length = utf8.DecodeRuneInString(str)
			str = str[length:]
		}
		if length == 0 {
			return -1, 0, b, str
		}
		pr, gc := propertyLineBreak(r)
		property = resolveLineBreakClass(pr, gc)
		if property != prCM && property != prZWJ {
			return property, r, b, str
		}
	}
}
