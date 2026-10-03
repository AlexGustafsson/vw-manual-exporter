package htmlutil

import (
	"golang.org/x/net/html"
)

// DeleteChildren recursively deletes matching child nodes in a subtree.
// The root node is never deleted.
func DeleteChildren(node *html.Node, matchFunc func(node *html.Node) (bool, error)) error {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling

		match, err := matchFunc(child)
		if err != nil {
			return err
		}

		if match {
			// Unlink child from its siblings.
			if child.PrevSibling != nil {
				child.PrevSibling.NextSibling = child.NextSibling
			} else {
				node.FirstChild = child.NextSibling
			}

			if child.NextSibling != nil {
				child.NextSibling.PrevSibling = child.PrevSibling
			} else {
				node.LastChild = child.PrevSibling
			}

			// Fully detach the deleted subtree.
			child.Parent = nil
			child.PrevSibling = nil
			child.NextSibling = nil
		} else {
			if err := DeleteChildren(child, matchFunc); err != nil {
				return err
			}
		}

		child = next
	}

	return nil
}
