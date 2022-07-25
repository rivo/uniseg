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

func ExampleFirstGraphemeCluster() {
	b := []byte("🇩🇪🏳️‍🌈")
	state := -1
	var c []byte
	for len(b) > 0 {
		c, b, state = uniseg.FirstGraphemeCluster(b, state)
		fmt.Printf("%s\n", string(c))
	}
	// Output: 🇩🇪
	//🏳️‍🌈
}

func ExampleFirstGraphemeClusterInString() {
	str := "🇩🇪🏳️‍🌈"
	state := -1
	var c string
	for len(str) > 0 {
		c, str, state = uniseg.FirstGraphemeClusterInString(str, state)
		fmt.Printf("%s\n", c)
	}
	// Output: 🇩🇪
	//🏳️‍🌈
}

func ExampleFirstWord() {
	b := []byte("Hello, world!")
	state := -1
	var c []byte
	for len(b) > 0 {
		c, b, state = uniseg.FirstWord(b, state)
		fmt.Printf("(%s)\n", string(c))
	}
	// Output: (Hello)
	//(,)
	//( )
	//(world)
	//(!)
}

func ExampleFirstWordInString() {
	str := "Hello, world!"
	state := -1
	var c string
	for len(str) > 0 {
		c, str, state = uniseg.FirstWordInString(str, state)
		fmt.Printf("(%s)\n", c)
	}
	// Output: (Hello)
	//(,)
	//( )
	//(world)
	//(!)
}

func ExampleFirstSentence() {
	b := []byte("This is sentence 1.0. And this is sentence two.")
	state := -1
	var c []byte
	for len(b) > 0 {
		c, b, state = uniseg.FirstSentence(b, state)
		fmt.Printf("(%s)\n", string(c))
	}
	// Output: (This is sentence 1.0. )
	//(And this is sentence two.)
}

func ExampleFirstSentenceInString() {
	str := "This is sentence 1.0. And this is sentence two."
	state := -1
	var c string
	for len(str) > 0 {
		c, str, state = uniseg.FirstSentenceInString(str, state)
		fmt.Printf("(%s)\n", c)
	}
	// Output: (This is sentence 1.0. )
	//(And this is sentence two.)
}

func ExampleFirstLineSegment() {
	b := []byte("First line.\nSecond line.")
	state := -1
	var (
		c         []byte
		mustBreak bool
	)
	for len(b) > 0 {
		c, b, mustBreak, state = uniseg.FirstLineSegment(b, state)
		fmt.Printf("(%s)", string(c))
		if mustBreak {
			fmt.Println(" < must break")
		} else {
			fmt.Println(" < may break")
		}
	}
	// Output: (First ) < may break
	//(line.
	//) < must break
	//(Second ) < may break
	//(line.) < must break
}

func ExampleFirstLineSegmentInString() {
	str := "First line.\nSecond line."
	state := -1
	var (
		c         string
		mustBreak bool
	)
	for len(str) > 0 {
		c, str, mustBreak, state = uniseg.FirstLineSegmentInString(str, state)
		fmt.Printf("(%s)", c)
		if mustBreak {
			fmt.Println(" < must break")
		} else {
			fmt.Println(" < may break")
		}
	}
	// Output: (First ) < may break
	//(line.
	//) < must break
	//(Second ) < may break
	//(line.) < must break
}
