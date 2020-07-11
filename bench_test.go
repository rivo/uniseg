package uniseg

import (
	"testing"
)

func BenchmarkCountOriginal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, bcase := range testCases {
			g := originalNewGraphemes(bcase.original)
			var n int
			for g.Next() {
				n++
			}
		}
	}
}

func originalNewGraphemes(s string) *Graphemes {
	g := &Graphemes{}
	for index, codePoint := range s {
		g.codePoints = append(g.codePoints, codePoint)
		g.indices = append(g.indices, index)
	}
	g.indices = append(g.indices, len(s))
	g.Next()
	return g
}

func BenchmarkCountCodeHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, bcase := range testCases {
			g := codeHexNewGraphemes(bcase.original)
			var n int
			for g.Next() {
				n++
			}
		}
	}
}

// https://github.com/rivo/uniseg/pull/5
func codeHexNewGraphemes(s string) *Graphemes {
	ln := len(s)
	g := &Graphemes{
		codePoints: make([]rune, 0, ln),
		indices:    make([]int, 0, ln+1),
	}
	for index, codePoint := range s {
		g.codePoints = append(g.codePoints, codePoint)
		g.indices = append(g.indices, index)
	}
	g.indices = append(g.indices, len(s))
	g.Next()
	return g
}

func BenchmarkCountDolmen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, bcase := range testCases {
			g := NewGraphemes(bcase.original)
			var n int
			for g.Next() {
				n++
			}
		}
	}
}
