package helpers

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// GetNodeOfType will retrieve a descendant "body" node/element from
// the passed-in docNode
func GetNodeOfType(docNode *html.Node, nodeType string) (node *html.Node) {
	var f func(*html.Node) *html.Node
	f = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode {
			switch n.Data {
			case nodeType:
				return n
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			result := f(c)
			if result != nil {
				return result
			}
		}
		return nil
	}

	return f(docNode)
}

// Converts a string to a lowercase, hyphen-separated string of max length 36
// Unused currently
func GetTitleURLFromString(title string, maxLength int, lowerCase bool) (output string) {
	// first, strip out any special characters
	re := regexp.MustCompile(`(?m)[^\d^A-Z^a-z^\-^\s]`)
	substitution := ""
	output = re.ReplaceAllString(title, substitution)

	// set to lowercase
	if lowerCase {
		output = strings.ToLower(output)
	}

	// next, replace all whitespace characters with dashes
	re = regexp.MustCompile(`(?m)[\s]`)
	substitution = "-"
	output = re.ReplaceAllString(output, substitution)

	// replace "clumps" of 2 or more hyphens with 1 hyphen
	re = regexp.MustCompile(`(?m)-{2,}`)
	substitution = "-"
	output = re.ReplaceAllString(output, substitution)

	// result is only up to x characters (or the whole thing if less than x)
	output = output[:int(math.Min(float64(len(output)), float64(maxLength)))]

	// remove trailing hyphens from the final output
	re = regexp.MustCompile(`(?m)-*$`)
	substitution = ""
	output = re.ReplaceAllString(output, substitution)

	return output
}

// RenderNode is a quick helper function to render an HTML node as a string
func RenderNode(n *html.Node) (output string, err error) {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	err = html.Render(w, n)
	if err != nil {
		return output, fmt.Errorf("failed to render node: %v", err.Error())
	}
	return buf.String(), nil
}
