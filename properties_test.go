package uniseg

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

// Reference implementation using brute-force-search
func propertyReference(r rune) (property int, cp *codePoint) {
	for _, p := range graphemeCodePoints {
		if p.lo <= r && r <= p.hi {
			return int(p.property), &p
		}
	}
	return prAny, nil
}

var _propertyStrings = [...]string{
	"prAny",
	"prPrepend",
	"prCR",
	"prLF",
	"prControl",
	"prExtend",
	"prRegionalIndicator",
	"prSpacingMark",
	"prL",
	"prV",
	"prT",
	"prLV",
	"prLVT",
	"prZWJ",
	"prExtendedPictographic",
}

func propertyToString(p uint8) string {
	if p < uint8(len(_propertyStrings)) {
		return _propertyStrings[p]
	}
	return fmt.Sprintf("pr(%d)", p)
}

func formatCodePoint(p *codePoint) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{%X, %X, %s}", p.lo, p.hi, propertyToString(p.property))
}

func testPropertyRange(t *testing.T, wg *sync.WaitGroup, failures *int32, start, end rune) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	for r := start; r < end; r++ {
		want, cp := propertyReference(r)
		got := property(r)
		if got != want {
			t.Errorf("property(%q) = %s; want: %s\n\t%s", r,
				propertyToString(uint8(got)), propertyToString(uint8(want)), formatCodePoint(cp))
			if failures != nil && atomic.AddInt32(failures, 1) > 10 {
				return
			}
		}
	}
}

func TestProperty_ASCII(t *testing.T) {
	// Test ASCII and extended ASCII
	testPropertyRange(t, nil, nil, 0, 256)
}

func clampInt64(i, max int64) int64 {
	if i <= max {
		return i
	}
	return max
}

// Exhaustive test of all code points between 0 and the highest grapheme + 8.
func TestProperty(t *testing.T) {
	const MaxRune = 1<<32 - 1
	if testing.Short() {
		t.Skip("skipping: short test")
	}

	// Use int64 to prevent overflow on 32-bit systems
	end := int64(graphemeCodePoints[len(graphemeCodePoints)-1].hi) + 8
	if MaxRune-end > 8 {
		end += 8 // +8 to make sure it works outside of the range
	}
	delta := end/int64(runtime.NumCPU()) + 1

	var wg sync.WaitGroup
	var failures int32
	for i := int64(0); i <= end; i += delta {
		wg.Add(1)
		go testPropertyRange(t, &wg, &failures, rune(i), rune(clampInt64(i+delta, end)))
	}
	wg.Wait()
}

func BenchmarkProperty_ASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		property('a')
	}
}

func BenchmarkProperty_ExtendedASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		property('¿')
	}
}

func BenchmarkProperty_Unicode(b *testing.B) {
	var runes = [...]rune{
		0x0981,  // Mn    BENGALI SIGN CANDRABINDU
		0x2666,  // E0.6  [2]   (♥️..♦️)  heart suit..diamond suit
		0x1F986, // E3.0  [13]  (🦅..🦑)  eagle..squid
		0x1F602, // E0.6  [6]   (😁..😆)  beaming face with smiling eyes..grinning squinting face
	}
	for i := 0; i < b.N; i++ {
		property(runes[i%len(runes)])
	}
}
