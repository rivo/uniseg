package uniseg

import (
	"testing"
)

const benchmarkStr = "This is 🏳️‍🌈, a test string ツ for grapheme cluster testing. 🏋🏽‍♀️🙂🙂 It's only relevant for benchmark tests."

var benchmarkBytes = []byte(benchmarkStr)

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
	{original: "\r\n\uFE0E", expected: [][]rune{{13, 10}, {0xfe0e}}},
	{original: "\u0915\u094d\u200d\u0937", expected: [][]rune{{0x0915, 0x094d, 0x200d, 0x0937}}},
	{original: "\u0915\u094d\u0937", expected: [][]rune{{0x0915, 0x094d, 0x0937}}},
	{original: "\u0915\u094d\u200d\u0937\u0915\u094d\u0937", expected: [][]rune{{0x0915, 0x094d, 0x200d, 0x0937}, {0x0915, 0x094d, 0x0937}}},
	{original: "\u0915\u0937", expected: [][]rune{{0x0915}, {0x0937}}},
	{original: "\u0995\u09cd\u09b7", expected: [][]rune{{0x0995, 0x09cd, 0x09b7}}},
}

// decomposed returns a grapheme cluster decomposition.
func decomposed(s string) (runes [][]rune) {
	gr := NewGraphemes(s)
	for gr.Next() {
		runes = append(runes, gr.Runes())
	}
	return
}

// Run all lists of test cases using the Graphemes class.
func TestGraphemesClass(t *testing.T) {
	allCases := append(testCases, graphemeBreakTestCases...)
	for testNum, testCase := range allCases {
		/*t.Logf(`Test case %d %q: Expecting %x, getting %x, code points %x"`,
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
				t.Errorf(`Test case %d %q failed: More grapheme clusters returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}
			cluster := gr.Runes()
			if len(cluster) != len(testCase.expected[index]) {
				t.Errorf(`Test case %d %q failed: Grapheme cluster at index %d has %d codepoints %x, %d expected %x`,
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
					t.Errorf(`Test case %d %q failed: Grapheme cluster at index %d is %x, expected %x`,
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
			t.Errorf(`Test case %d %q failed: Fewer grapheme clusters returned (%d) than expected (%d)`,
				testNum,
				testCase.original,
				index,
				len(testCase.expected))
		}
	}
}

// Run the standard Unicode test cases for word boundaries using the Graphemes
// class.
func TestGraphemesClassWord(t *testing.T) {
	for testNum, testCase := range wordBreakTestCases {
		if testNum == 1700 {
			// This test case reveals an inconsistency in the Unicode rule set,
			// namely the handling of ZWJ within two RI graphemes. (Grapheme
			// rules will restart the RI count, word rules will ignore the ZWJ.)
			// An error has been reported.
			continue
		}
		/*t.Logf(`Test case %d %q: Expecting %x, getting %x, code points %x"`,
		testNum,
		strings.TrimSpace(testCase.original),
		testCase.expected,
		decomposed(testCase.original),
		[]rune(testCase.original))*/
		gr := NewGraphemes(testCase.original)
		var (
			index   int
			cluster []rune
		)
		if !gr.IsWordBoundary() {
			t.Error("Expected initial IsWordBoundary to be true, got false")
		}
	GraphemeLoop:
		for gr.Next() {
			if index >= len(testCase.expected) {
				t.Errorf(`Test case %d %q failed: More words returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}
			cluster = append(cluster, gr.Runes()...)
			if gr.IsWordBoundary() {
				if len(cluster) != len(testCase.expected[index]) {
					t.Errorf(`Test case %d %q failed: Word at index %d has %d codepoints %x, %d expected %x`,
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
						t.Errorf(`Test case %d %q failed: Word at index %d is %x, expected %x`,
							testNum,
							testCase.original,
							index,
							cluster,
							testCase.expected[index])
						break GraphemeLoop
					}
				}
				cluster = nil
				index++
			}
		}
		if index < len(testCase.expected) {
			t.Errorf(`Test case %d %q failed: Fewer words returned (%d) than expected (%d)`,
				testNum,
				testCase.original,
				index,
				len(testCase.expected))
		}
	}
}

// Run the standard Unicode test cases for sentence boundaries using the
// Graphemes class.
func TestGraphemesClassSentence(t *testing.T) {
	for testNum, testCase := range sentenceBreakTestCases {
		/*t.Logf(`Test case %d %q: Expecting %x, getting %x, code points %x"`,
		testNum,
		strings.TrimSpace(testCase.original),
		testCase.expected,
		decomposed(testCase.original),
		[]rune(testCase.original))*/
		gr := NewGraphemes(testCase.original)
		var (
			index   int
			cluster []rune
		)
		if !gr.IsSentenceBoundary() {
			t.Error("Expected initial IsSentenceBoundary to be true, got false")
		}
	GraphemeLoop:
		for gr.Next() {
			if index >= len(testCase.expected) {
				t.Errorf(`Test case %d %q failed: More sentences returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}
			cluster = append(cluster, gr.Runes()...)
			if gr.IsSentenceBoundary() {
				if len(cluster) != len(testCase.expected[index]) {
					t.Errorf(`Test case %d %q failed: Sentence at index %d has %d codepoints %x, %d expected %x`,
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
						t.Errorf(`Test case %d %q failed: Sentence at index %d is %x, expected %x`,
							testNum,
							testCase.original,
							index,
							cluster,
							testCase.expected[index])
						break GraphemeLoop
					}
				}
				cluster = nil
				index++
			}
		}
		if index < len(testCase.expected) {
			t.Errorf(`Test case %d %q failed: Fewer sentences returned (%d) than expected (%d)`,
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
		t.Errorf(`Expected "p", got %q`, str)
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
		t.Errorf(`Expected "B", got %q`, string(b[0]))
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
		t.Errorf(`Expected "m", got %q`, str)
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
		t.Errorf(`Expected empty string, got %q`, str)
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
		t.Errorf(`Expected empty string, got %q`, str)
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

// Test the ReverseString function.
func TestReverseString(t *testing.T) {
	for _, testCase := range testCases {
		var r []rune
		for index := len(testCase.expected) - 1; index >= 0; index-- {
			r = append(r, testCase.expected[index]...)
		}
		if string(r) != ReverseString(testCase.original) {
			t.Errorf(`Exepected reverse of %q to be %q, got %q`, testCase.original, string(r), ReverseString(testCase.original))
		}
	}

	// Three additional ones, for good measure.
	if ReverseString("🇩🇪🏳️‍🌈") != "🏳️‍🌈🇩🇪" {
		t.Error("Flags weren't reversed correctly")
	}
	if ReverseString("🏳️‍🌈") != "🏳️‍🌈" {
		t.Error("Flag wasn't reversed correctly")
	}
	if ReverseString("") != "" {
		t.Error("Empty string wasn't reversed correctly")
	}
}

// Run all lists of test cases using the Graphemes function for byte slices.
func TestGraphemesFunctionBytes(t *testing.T) {
	allCases := append(testCases, graphemeBreakTestCases...)
	for testNum, testCase := range allCases {
		/*t.Logf(`Test case %d %q: Expecting %x, getting %x, code points %x"`,
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
			c, b, _, state = FirstGraphemeCluster(b, state)

			if index >= len(testCase.expected) {
				t.Errorf(`Test case %d %q failed: More grapheme clusters returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}

			cluster := []rune(string(c))
			if len(cluster) != len(testCase.expected[index]) {
				t.Errorf(`Test case %d %q failed: Grapheme cluster at index %d has %d codepoints %x, %d expected %x`,
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
					t.Errorf(`Test case %d %q failed: Grapheme cluster at index %d is %x, expected %x`,
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
			t.Errorf(`Test case %d %q failed: Fewer grapheme clusters returned (%d) than expected (%d)`,
				testNum,
				testCase.original,
				index,
				len(testCase.expected))
		}
	}
	cluster, rest, width, newState := FirstGraphemeCluster([]byte{}, 0)
	if len(cluster) > 0 {
		t.Errorf(`Expected cluster to be empty byte slice, got %q`, cluster)
	}
	if len(rest) > 0 {
		t.Errorf(`Expected rest to be empty byte slice, got %q`, rest)
	}
	if width != 0 {
		t.Errorf(`Expected width to be 0, got %d`, width)
	}
	if newState != 0 {
		t.Errorf(`Expected newState to be 0, got %d`, newState)
	}
}

// Run all lists of test cases using the Graphemes function for strings.
func TestGraphemesFunctionString(t *testing.T) {
	allCases := append(testCases, graphemeBreakTestCases...)
	for testNum, testCase := range allCases {
		/*t.Logf(`Test case %d %q: Expecting %x, getting %x, code points %x"`,
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
			c, str, _, state = FirstGraphemeClusterInString(str, state)

			if index >= len(testCase.expected) {
				t.Errorf(`Test case %d %q failed: More grapheme clusters returned than expected %d`,
					testNum,
					testCase.original,
					len(testCase.expected))
				break
			}

			cluster := []rune(c)
			if len(cluster) != len(testCase.expected[index]) {
				t.Errorf(`Test case %d %q failed: Grapheme cluster at index %d has %d codepoints %x, %d expected %x`,
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
					t.Errorf(`Test case %d %q failed: Grapheme cluster at index %d is %x, expected %x`,
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
			t.Errorf(`Test case %d %q failed: Fewer grapheme clusters returned (%d) than expected (%d)`,
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
	for i := 0; i < b.N; i++ {
		var c []byte
		state := -1
		str := benchmarkBytes
		for len(str) > 0 {
			c, str, _, state = FirstGraphemeCluster(str, state)
			resultRunes = []rune(string(c))
		}
	}
}

// Benchmark the use of the Graphemes function for strings.
func BenchmarkGraphemesFunctionString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var c string
		state := -1
		str := benchmarkStr
		for len(str) > 0 {
			c, str, _, state = FirstGraphemeClusterInString(str, state)
			resultRunes = []rune(c)
		}
	}
}

// TestIssue60 tests the grapheme cluster segmentation of Indic conjunct sequences
// (Unicode rule GB9c) as reported in issue #60.
func TestIssue60(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantCount    int
		wantClusters []string
	}{
		{
			name:         "Devanagari conjunct with ZWJ (issue #60 reproduction)",
			input:        "\u0915\u094d\u200d\u0937",
			wantCount:    1,
			wantClusters: []string{"\u0915\u094d\u200d\u0937"},
		},
		{
			name:         "Devanagari conjunct without ZWJ",
			input:        "\u0915\u094d\u0937",
			wantCount:    1,
			wantClusters: []string{"\u0915\u094d\u0937"},
		},
		{
			name:         "Two Devanagari conjuncts in sequence",
			input:        "\u0915\u094d\u200d\u0937\u0915\u094d\u0937",
			wantCount:    2,
			wantClusters: []string{"\u0915\u094d\u200d\u0937", "\u0915\u094d\u0937"},
		},
		{
			name:         "Devanagari consonants without virama linker",
			input:        "\u0915\u0937",
			wantCount:    2,
			wantClusters: []string{"\u0915", "\u0937"},
		},
		{
			name:         "Bengali conjunct",
			input:        "\u0995\u09cd\u09b7",
			wantCount:    1,
			wantClusters: []string{"\u0995\u09cd\u09b7"},
		},
		{
			name:         "Devanagari conjunct with trailing text",
			input:        "\u0915\u094d\u200d\u0937 text",
			wantCount:    6,
			wantClusters: []string{"\u0915\u094d\u200d\u0937", " ", "t", "e", "x", "t"},
		},
		{
			name:         "Devanagari triple consonant conjunct",
			input:        "\u0938\u094d\u0924\u094d\u0930",
			wantCount:    1,
			wantClusters: []string{"\u0938\u094d\u0924\u094d\u0930"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount := GraphemeClusterCount(tt.input)
			if gotCount != tt.wantCount {
				t.Errorf("GraphemeClusterCount(%q) = %d, want %d", tt.input, gotCount, tt.wantCount)
			}

			// Verify via Graphemes iterator
			var clusters []string
			g := NewGraphemes(tt.input)
			for g.Next() {
				clusters = append(clusters, g.Str())
			}
			if len(clusters) != len(tt.wantClusters) {
				t.Fatalf("Graphemes iterator returned %d clusters, want %d: %q", len(clusters), len(tt.wantClusters), clusters)
			}
			for i, cluster := range clusters {
				if cluster != tt.wantClusters[i] {
					t.Errorf("Cluster[%d] = %q, want %q", i, cluster, tt.wantClusters[i])
				}
			}

			// Verify via FirstGraphemeClusterInString
			var firstClusters []string
			str := tt.input
			state := -1
			for len(str) > 0 {
				var cluster string
				cluster, str, _, state = FirstGraphemeClusterInString(str, state)
				firstClusters = append(firstClusters, cluster)
			}
			if len(firstClusters) != len(tt.wantClusters) {
				t.Fatalf("FirstGraphemeClusterInString returned %d clusters, want %d: %q", len(firstClusters), len(tt.wantClusters), firstClusters)
			}
			for i, cluster := range firstClusters {
				if cluster != tt.wantClusters[i] {
					t.Errorf("FirstGraphemeClusterInString cluster[%d] = %q, want %q", i, cluster, tt.wantClusters[i])
				}
			}

			// Verify via StepString
			var stepClusters []string
			str = tt.input
			state = -1
			for len(str) > 0 {
				var cluster string
				cluster, str, _, state = StepString(str, state)
				stepClusters = append(stepClusters, cluster)
			}
			if len(stepClusters) != len(tt.wantClusters) {
				t.Fatalf("StepString returned %d clusters, want %d: %q", len(stepClusters), len(tt.wantClusters), stepClusters)
			}
			for i, cluster := range stepClusters {
				if cluster != tt.wantClusters[i] {
					t.Errorf("StepString cluster[%d] = %q, want %q", i, cluster, tt.wantClusters[i])
				}
			}
		})
	}
}
