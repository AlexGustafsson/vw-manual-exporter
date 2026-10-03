package htmlutil

import (
	"errors"

	"golang.org/x/net/html"
)

// SkipNode is used as a return value from a walk func to indicate that the node
// is to be skipped. It is not returned as an error by any function.
var SkipNode = errors.New("skip this directory")

// Walk walks a node by performing a depth-first search.
func Walk(node *html.Node, walkFunc func(node *html.Node) error) error {
	if err := walkFunc(node); err != nil {
		return err
	}

	// Check subtree
	if node.FirstChild != nil {
		err := Walk(node.FirstChild, walkFunc)
		if err != nil {
			return err
		}
	}

	// Go to next child
	if node.NextSibling != nil {
		err := Walk(node.NextSibling, walkFunc)
		if err != nil {
			return err
		}
	}

	return nil
}
