package vwapps

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"github.com/AlexGustafsson/vw-manual-exporter/internal/htmlutil"
	"golang.org/x/net/html"
)

func RenderHTML(r io.Reader, stylesheet string) (string, error) {
	// <title> i head har titel för sida
	//
	// TODO: We need to extract image URLs

	document, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	htmlElement := document.FirstChild
	if htmlElement == nil {
		return "", fmt.Errorf("invalid html document: missing html element")
	}

	headElement := htmlElement.FirstChild
	if headElement == nil {
		return "", fmt.Errorf("invalid html document: missing head element")
	}

	styleTextElement := &html.Node{
		Type: html.TextNode,
		Data: stylesheet,
	}

	styleElement := &html.Node{
		Type: html.ElementNode,
		Data: "style",
	}
	styleElement.AppendChild(styleTextElement)

	headElement.AppendChild(styleElement)

	bodyElement := headElement.NextSibling
	if bodyElement == nil {
		return "", fmt.Errorf("invalid html document: missing body element")
	}

	title := ""
	if titleNode, _ := htmlutil.Match(headElement, htmlutil.MatchTag("title")); titleNode != nil {
		title = titleNode.FirstChild.Data
	}
	_ = title

	// Strip HTML elements
	htmlutil.DeleteChildren(bodyElement, func(node *html.Node) (bool, error) {
		switch node.Data {
		case "head", "meta":
			return true, nil
		default:
			return false, nil
		}
	})

	// Strip attributes
	htmlutil.Walk(bodyElement, func(node *html.Node) error {
		node.Attr = slices.DeleteFunc(node.Attr, func(a html.Attribute) bool {
			switch a.Key {
			case "href", "src", "alt", "class", "data-dl-wrapper", "data-type", "data-role", "data-class", "dy-link":
				return false
			default:
				return true
			}
		})

		return nil
	})

	var buffer bytes.Buffer
	if err := html.Render(&buffer, htmlElement); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
