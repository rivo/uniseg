//go:build generate

// This program generates the grapheme_properties.go file containing Grapheme
// Break Properties, from the Unicode Character Database auxiliary data files.
//
//go:generate go run gen_graphemeproperties.go
package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	gbpURL   = `https://www.unicode.org/Public/14.0.0/ucd/auxiliary/GraphemeBreakProperty.txt`
	emojiURL = `https://unicode.org/Public/14.0.0/ucd/emoji/emoji-data.txt`
	target   = `graphemeproperties.go`
)

// The regular expression for a line containing a code point range property.
var propertyPattern = regexp.MustCompile(`^([0-9A-F]{4,6})(\.\.([0-9A-F]{4,6}))?\s+;\s+([A-Za-z0-9_]+)\s*#\s(.+)$`)

func main() {
	log.SetPrefix("gen_graphemeproperties: ")
	log.SetFlags(0)

	// Parse the text file and generate Go source code from it.
	src, err := parse(gbpURL, emojiURL)
	if err != nil {
		log.Fatal(err)
	}

	// Format the Go code.
	formatted, err := format.Source(src)
	if err != nil {
		log.Fatal("gofmt:", err)
	}

	// Save it to the (local) target file.
	log.Print("Writing to ", target)
	tmp := target + ".tmp.go"
	if err := os.WriteFile(tmp, formatted, 0644); err != nil {
		os.Remove(tmp)
		log.Fatal(err)
	}
	if err := os.Rename(tmp, target); err != nil {
		os.Remove(tmp)
		log.Fatal(err)
	}
}

// parse parses the Grapheme Break Properties text file located at the given
// URLs and returns its equivalent Go source code to be used in the uniseg
// package.
func parse(gbpURL, emojiURL string) ([]byte, error) {
	// Temporary buffer to hold properties.
	var properties [][4]string

	// Open the first URL.
	log.Printf("Parsing %s", gbpURL)
	res, err := http.Get(gbpURL)
	if err != nil {
		return nil, err
	}
	in1 := res.Body
	defer in1.Close()

	// Parse it.
	scanner := bufio.NewScanner(in1)
	num := 0
	for scanner.Scan() {
		num++
		line := scanner.Text()

		// Skip comments and empty lines.
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Everything else must be a code point range, a property and a comment.
		from, to, property, comment, err := parseProperty(line)
		if err != nil {
			return nil, fmt.Errorf("graphemes line %d: %v", num, err)
		}
		properties = append(properties, [4]string{from, to, property, comment})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Open the second URL.
	log.Printf("Parsing %s", emojiURL)
	res, err = http.Get(emojiURL)
	if err != nil {
		return nil, err
	}
	in2 := res.Body
	defer in2.Close()

	// Parse it.
	scanner = bufio.NewScanner(in2)
	num = 0
	for scanner.Scan() {
		num++
		line := scanner.Text()

		// Skip comments, empty lines, and everything not containing
		// "Extended_Pictographic".
		if strings.HasPrefix(line, "#") || line == "" || !strings.Contains(line, "Extended_Pictographic") {
			continue
		}

		// Everything else must be a code point range, a property and a comment.
		from, to, property, comment, err := parseProperty(line)
		if err != nil {
			return nil, fmt.Errorf("emojis line %d: %v", num, err)
		}
		properties = append(properties, [4]string{from, to, property, comment})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Sort properties.
	sort.Slice(properties, func(i, j int) bool {
		left, _ := strconv.ParseUint(properties[i][0], 16, 64)
		right, _ := strconv.ParseUint(properties[j][0], 16, 64)
		return left < right
	})

	// Header.
	var buf bytes.Buffer
	buf.WriteString(`// Code generated via go generate from gen_graphemeproperties.go. DO NOT EDIT.

package uniseg

type codePoint struct {
	lo, hi   rune
	property uint8
}

// graphemeCodePoints are taken from
// ` + gbpURL + `,
// and
// ` + emojiURL + `,
// ("Extended_Pictographic" only) on March 11, 2019. See
// https://www.unicode.org/license.html for the Unicode license agreement.
var graphemeCodePoints = []codePoint{
`)

	// Properties.
	for _, prop := range properties {
		fmt.Fprintf(&buf, "{0x%s,0x%s,%s}, // %s\n", prop[0], prop[1], translateProperty("pr", prop[2]), prop[3])
	}

	// Tail.
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

// parseProperty parses a line of the Grapheme Break Properties text file
// containing a property for a code point range and returns it along with its
// comment.
func parseProperty(line string) (from, to, property, comment string, err error) {
	fields := propertyPattern.FindStringSubmatch(line)
	if fields == nil {
		err = errors.New("no property found")
		return
	}
	from = fields[1]
	to = fields[3]
	if to == "" {
		to = from
	}
	property = fields[4]
	comment = fields[5]
	return
}

// translateProperty translates a property name as used in the Unicode data file
// to a variable used in the Go code.
func translateProperty(prefix, property string) string {
	return prefix + strings.ReplaceAll(property, "_", "")
}
