package uniseg

import (
	"testing"
	"unicode/utf8"
)

const benchmarkStr = "This is 🏳️‍🌈, a test string ツ for grapheme cluster testing. 🏋🏽‍♀️🙂🙂"

// Variables to avoid compiler optimizations.
var resultRunes []rune

type testCase = struct {
	original string
	expected [][]rune
}

// The test cases for the simple test function.
var testCases = []testCase{
	{original: "", expected: [][]rune{}},
	{original: "x", expected: [][]rune{{0x78}}},
	{original: "basic", expected: [][]rune{{0x62}, {0x61}, {0x73}, {0x69}, {0x63}}},
	{original: "möp", expected: [][]rune{{0x6d}, {0x6f, 0x308}, {0x70}}},
	{original: "\r\n", expected: [][]rune{{0xd, 0xa}}},
	{original: "\n\n", expected: [][]rune{{0xa}, {0xa}}},
	{original: "\t*", expected: [][]rune{{0x9}, {0x2a}}},
	{original: "뢴", expected: [][]rune{{0x1105, 0x116c, 0x11ab}}},
	{original: "ܐ܏ܒܓܕ", expected: [][]rune{{0x710}, {0x70f, 0x712}, {0x713}, {0x715}}},
	{original: "ำ", expected: [][]rune{{0xe33}}},
	{original: "ำำ", expected: [][]rune{{0xe33, 0xe33}}},
	{original: "สระอำ", expected: [][]rune{{0xe2a}, {0xe23}, {0xe30}, {0xe2d, 0xe33}}},
	{original: "*뢴*", expected: [][]rune{{0x2a}, {0x1105, 0x116c, 0x11ab}, {0x2a}}},
	{original: "*👩‍❤️‍💋‍👩*", expected: [][]rune{{0x2a}, {0x1f469, 0x200d, 0x2764, 0xfe0f, 0x200d, 0x1f48b, 0x200d, 0x1f469}, {0x2a}}},
	{original: "👩‍❤️‍💋‍👩", expected: [][]rune{{0x1f469, 0x200d, 0x2764, 0xfe0f, 0x200d, 0x1f48b, 0x200d, 0x1f469}}},
	{original: "🏋🏽‍♀️", expected: [][]rune{{0x1f3cb, 0x1f3fd, 0x200d, 0x2640, 0xfe0f}}},
	{original: "🙂", expected: [][]rune{{0x1f642}}},
	{original: "🙂🙂", expected: [][]rune{{0x1f642}, {0x1f642}}},
	{original: "🇩🇪", expected: [][]rune{{0x1f1e9, 0x1f1ea}}},
	{original: "🏳️‍🌈", expected: [][]rune{{0x1f3f3, 0xfe0f, 0x200d, 0x1f308}}},
	{original: "\t🏳️‍🌈", expected: [][]rune{{0x9}, {0x1f3f3, 0xfe0f, 0x200d, 0x1f308}}},
	{original: "\t🏳️‍🌈\t", expected: [][]rune{{0x9}, {0x1f3f3, 0xfe0f, 0x200d, 0x1f308}, {0x9}}},
}

// decomposed returns a grapheme cluster decomposition.
func decomposed(s string) (runes [][]rune) {
	gr := NewGraphemes(s)
	for gr.Next() {
		runes = append(runes, gr.Runes())
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

// Run all lists of test cases using the Graphemes class.
func TestGraphemesClass(t *testing.T) {
	allCases := append(testCases, unicodeTestCases...)
	for testNum, testCase := range allCases {
		/*t.Logf(`Test case %d "%s": Expecting %x, getting %x, code points %x"`,
		testNum,
		strings.TrimSpace(testCase.original),
		testCase.expected,
		decomposed(testCase.original),
		[]rune(testCase.original))*/
		gr := NewGraphemes(testCase.original)
		var index int
	GraphemeLoop:
		for index = 0; gr.Next(); index++ {
			if index >= len(testCase.expected) {
				t.Errorf(`Test case %d "%s" failed: More grapheme clusters returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}
			cluster := gr.Runes()
			if len(cluster) != len(testCase.expected[index]) {
				t.Errorf(`Test case %d "%s" failed: Grapheme cluster at index %d has %d codepoints %x, %d expected %x`,
					testNum,
					testCase.original,
					index,
					len(cluster),
					cluster,
					len(testCase.expected[index]),
					testCase.expected[index])
				break
			}
			for i, r := range cluster {
				if r != testCase.expected[index][i] {
					t.Errorf(`Test case %d "%s" failed: Grapheme cluster at index %d is %x, expected %x`,
						testNum,
						testCase.original,
						index,
						cluster,
						testCase.expected[index])
					break GraphemeLoop
				}
			}
		}
		if index < len(testCase.expected) {
			t.Errorf(`Test case %d "%s" failed: Fewer grapheme clusters returned (%d) than expected (%d)`,
				testNum,
				testCase.original,
				index,
				len(testCase.expected))
		}
	}
}

// Test the Str() function.
func TestGraphemesStr(t *testing.T) {
	gr := NewGraphemes("möp")
	gr.Next()
	gr.Next()
	gr.Next()
	if str := gr.Str(); str != "p" {
		t.Errorf(`Expected "p", got "%s"`, str)
	}
}

// Test the Bytes() function.
func TestGraphemesBytes(t *testing.T) {
	gr := NewGraphemes("A👩‍❤️‍💋‍👩B")
	gr.Next()
	gr.Next()
	gr.Next()
	b := gr.Bytes()
	if len(b) != 1 {
		t.Fatalf(`Expected len("B") == 1, got %d`, len(b))
	}
	if b[0] != 'B' {
		t.Errorf(`Expected "B", got "%s"`, string(b[0]))
	}
}

// Test the Positions() function.
func TestGraphemesPositions(t *testing.T) {
	gr := NewGraphemes("A👩‍❤️‍💋‍👩B")
	gr.Next()
	gr.Next()
	from, to := gr.Positions()
	if from != 1 || to != 28 {
		t.Errorf(`Expected from=%d to=%d, got from=%d to=%d`, 1, 28, from, to)
	}
}

// Test the Reset() function.
func TestGraphemesReset(t *testing.T) {
	gr := NewGraphemes("möp")
	gr.Next()
	gr.Next()
	gr.Next()
	gr.Reset()
	gr.Next()
	if str := gr.Str(); str != "m" {
		t.Errorf(`Expected "m", got "%s"`, str)
	}
}

// Test retrieving clusters before calling Next().
func TestGraphemesEarly(t *testing.T) {
	gr := NewGraphemes("test")
	r := gr.Runes()
	if r != nil {
		t.Errorf(`Expected nil rune slice, got %x`, r)
	}
	str := gr.Str()
	if str != "" {
		t.Errorf(`Expected empty string, got "%s"`, str)
	}
	b := gr.Bytes()
	if b != nil {
		t.Errorf(`Expected byte rune slice, got %x`, b)
	}
	from, to := gr.Positions()
	if from != 0 || to != 0 {
		t.Errorf(`Expected from=%d to=%d, got from=%d to=%d`, 0, 0, from, to)
	}
}

// Test retrieving more clusters after retrieving the last cluster.
func TestGraphemesLate(t *testing.T) {
	gr := NewGraphemes("x")
	gr.Next()
	gr.Next()
	r := gr.Runes()
	if r != nil {
		t.Errorf(`Expected nil rune slice, got %x`, r)
	}
	str := gr.Str()
	if str != "" {
		t.Errorf(`Expected empty string, got "%s"`, str)
	}
	b := gr.Bytes()
	if b != nil {
		t.Errorf(`Expected byte rune slice, got %x`, b)
	}
	from, to := gr.Positions()
	if from != 1 || to != 1 {
		t.Errorf(`Expected from=%d to=%d, got from=%d to=%d`, 1, 1, from, to)
	}
}

// Test the GraphemeClusterCount function.
func TestGraphemesCount(t *testing.T) {
	if n := GraphemeClusterCount("🇩🇪🏳️‍🌈"); n != 2 {
		t.Errorf(`Expected 2 grapheme clusters, got %d`, n)
	}
}

// Run all lists of test cases using the Graphemes function for byte slices.
func TestGraphemesFunctionBytes(t *testing.T) {
	allCases := append(testCases, unicodeTestCases...)
	for testNum, testCase := range allCases {
		/*t.Logf(`Test case %d "%s": Expecting %x, getting %x, code points %x"`,
		testNum,
		strings.TrimSpace(testCase.original),
		testCase.expected,
		decomposed(testCase.original),
		[]rune(testCase.original))*/
		b := []byte(testCase.original)
		state := -1
		var (
			index int
			c     []byte
		)
	GraphemeLoop:
		for len(b) > 0 {
			c, b, state = firstGraphemeCluster(b, state)

			if index >= len(testCase.expected) {
				t.Errorf(`Test case %d "%s" failed: More grapheme clusters returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}

			cluster := []rune(string(c))
			if len(cluster) != len(testCase.expected[index]) {
				t.Errorf(`Test case %d "%s" failed: Grapheme cluster at index %d has %d codepoints %x, %d expected %x`,
					testNum,
					testCase.original,
					index,
					len(cluster),
					cluster,
					len(testCase.expected[index]),
					testCase.expected[index])
				break
			}
			for i, r := range cluster {
				if r != testCase.expected[index][i] {
					t.Errorf(`Test case %d "%s" failed: Grapheme cluster at index %d is %x, expected %x`,
						testNum,
						testCase.original,
						index,
						cluster,
						testCase.expected[index])
					break GraphemeLoop
				}
			}

			index++
		}
		if index < len(testCase.expected) {
			t.Errorf(`Test case %d "%s" failed: Fewer grapheme clusters returned (%d) than expected (%d)`,
				testNum,
				testCase.original,
				index,
				len(testCase.expected))
		}
	}
}

// Run all lists of test cases using the Graphemes function for strings.
func TestGraphemesFunctionString(t *testing.T) {
	allCases := append(testCases, unicodeTestCases...)
	for testNum, testCase := range allCases {
		/*t.Logf(`Test case %d "%s": Expecting %x, getting %x, code points %x"`,
		testNum,
		strings.TrimSpace(testCase.original),
		testCase.expected,
		decomposed(testCase.original),
		[]rune(testCase.original))*/
		str := testCase.original
		state := -1
		var (
			index int
			c     string
		)
	GraphemeLoop:
		for len(str) > 0 {
			c, str, state = firstGraphemeClusterInString(str, state)

			if index >= len(testCase.expected) {
				t.Errorf(`Test case %d "%s" failed: More grapheme clusters returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}

			cluster := []rune(c)
			if len(cluster) != len(testCase.expected[index]) {
				t.Errorf(`Test case %d "%s" failed: Grapheme cluster at index %d has %d codepoints %x, %d expected %x`,
					testNum,
					testCase.original,
					index,
					len(cluster),
					cluster,
					len(testCase.expected[index]),
					testCase.expected[index])
				break
			}
			for i, r := range cluster {
				if r != testCase.expected[index][i] {
					t.Errorf(`Test case %d "%s" failed: Grapheme cluster at index %d is %x, expected %x`,
						testNum,
						testCase.original,
						index,
						cluster,
						testCase.expected[index])
					break GraphemeLoop
				}
			}

			index++
		}
		if index < len(testCase.expected) {
			t.Errorf(`Test case %d "%s" failed: Fewer grapheme clusters returned (%d) than expected (%d)`,
				testNum,
				testCase.original,
				index,
				len(testCase.expected))
		}
	}
}

// Benchmark the use of the Graphemes class.
func BenchmarkGraphemesClass(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g := NewGraphemes(benchmarkStr)
		for g.Next() {
			resultRunes = g.Runes()
		}
	}
}

// Benchmark the use of the Graphemes function for byte slices.
func BenchmarkGraphemesFunctionBytes(b *testing.B) {
	str := []byte(benchmarkStr)
	for i := 0; i < b.N; i++ {
		var c []byte
		state := -1
		for len(str) > 0 {
			c, str, state = firstGraphemeCluster(str, state)
			resultRunes = []rune(string(c))
		}
	}
}

// Benchmark the use of the Graphemes function for strings.
func BenchmarkGraphemesFunctionString(b *testing.B) {
	str := benchmarkStr
	for i := 0; i < b.N; i++ {
		var c string
		state := -1
		for len(str) > 0 {
			c, str, state = firstGraphemeClusterInString(str, state)
			resultRunes = []rune(string(c))
		}
	}
}
