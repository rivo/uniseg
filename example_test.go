package uniseg_test

import (
	"fmt"

	"github.com/rivo/uniseg"
)

func ExampleGraphemes() {
	gr := uniseg.NewGraphemes("👍🏼!")
	for gr.Next() {
		fmt.Printf("%x ", gr.Runes())
	}
	// Output: [1f44d 1f3fc] [21]
}

func ExampleGraphemeClusterCount() {
	n := uniseg.GraphemeClusterCount("🇩🇪🏳️‍🌈")
	fmt.Println(n)
	// Output: 2
}
