package htmlutil

import "golang.org/x/net/html"

func MatchTag(tag string) func(node *html.Node) (bool, error) {
	return func(node *html.Node) (bool, error) {
		return node.Data == tag, nil
	}
}

// Match returns the first matching node by performing a depth-first search.
func Match(node *html.Node, matchFunc func(node *html.Node) (bool, error)) (*html.Node, error) {
	ok, err := matchFunc(node)
	if err != nil {
		return nil, err
	}

	if ok {
		return node, nil
	}

	// Check subtree
	if node.FirstChild != nil {
		match, err := Match(node.FirstChild, matchFunc)
		if err != nil {
			return nil, err
		}
		return match, nil
	}

	// Go to next child
	if node.NextSibling != nil {
		match, err := Match(node.NextSibling, matchFunc)
		if err != nil {
			return nil, err
		}
		return match, nil
	}

	return nil, nil
}
