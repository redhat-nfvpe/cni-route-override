// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	f, err := os.Open(filepath.Join("_gendata", "src.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	node, err := html.Parse(f)
	if err != nil {
		return err
	}

	walk(node)

	return nil
}

func walk(n *html.Node) {
	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" && hasAttr(n, "class", "emoji-inner") {
			name := getAttr(n.Parent.Parent, "data-name")
			backgroundPosition := strings.TrimSuffix(strings.TrimPrefix(getAttr(n, "style"), "background: url(https://slack.global.ssl.fastly.net/d4bf/img/emoji_2015_2/sheet_apple_64_indexed_256colors.png);background-position:"), ";background-size:4100%")
			fmt.Printf("%q: %q,\n", name, backgroundPosition)
			x, y := parsePosition(backgroundPosition)
			fmt.Println(name, x, y)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
}

func hasAttr(n *html.Node, key, val string) bool {
	for _, attr := range n.Attr {
		if attr.Namespace == "" && attr.Key == key && attr.Val == val {
			return true
		}
	}
	return false
}

// getAttr returns an attribute of a node, or panics if not found.
func getAttr(n *html.Node, key string) (val string) {
	for _, attr := range n.Attr {
		if attr.Namespace == "" && attr.Key == key {
			return attr.Val
		}
	}
	panic("not found")
}

// parsePosition parses "97.5% 0%" to 39 0, etc.
func parsePosition(s string) (x int, y int) {
	xy := strings.Fields(s)
	return parsePercentage(xy[0]), parsePercentage(xy[1])
}

// parsePercentage parses "0%" to 0, "2.5%" to 1, "5%" to 2, "97.5%" to 39, etc.
func parsePercentage(in string) int {
	in = in[:len(in)-1] // Trim "%" suffix.
	f, err := strconv.ParseFloat(in, 64)
	if err != nil {
		panic(err)
	}
	f /= 2.5
	return near(f)
}

// near returns the nearest int to f.
func near(f float64) int {
	if f >= 0 {
		return int(f + 0.5)
	} else {
		return int(f - 0.5)
	}
}
