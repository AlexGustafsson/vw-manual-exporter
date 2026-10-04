package vwapps

import (
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/AlexGustafsson/vw-manual-exporter/internal/htmlutil"
	"golang.org/x/net/html"
)

type HTMLBuilder struct {
	Stylesheet string

	once sync.Once

	documentNode *html.Node
	bodyNode     *html.Node
}

func (b *HTMLBuilder) init() {
	documentNode := &html.Node{
		Type: html.DocumentNode,
	}

	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "html",
	}
	documentNode.AppendChild(htmlNode)

	headNode := &html.Node{
		Type: html.ElementNode,
		Data: "head",
	}
	htmlNode.AppendChild(headNode)

	// Meta language
	// <meta http-equiv="content-type" content="text/html;charset=utf-8" />
	{
		metaNode := &html.Node{
			Type: html.ElementNode,
			Data: "meta",
		}
		headNode.AppendChild(metaNode)

		metaNode.Attr = []html.Attribute{
			{
				Key: "http-equiv",
				Val: "content-type",
			},
			{
				Key: "content",
				Val: "text/html;charset=utf-8",
			},
		}
	}

	// First party style
	{
		styleNode := &html.Node{
			Type: html.ElementNode,
			Data: "style",
		}
		headNode.AppendChild(styleNode)

		textNode := &html.Node{
			Type: html.TextNode,
			Data: `h3 { font-size: 28px; font-weight: 700; } td { white-space: normal!important; }`,
		}
		styleNode.AppendChild(textNode)
	}

	// Vendored style
	{
		styleNode := &html.Node{
			Type: html.ElementNode,
			Data: "style",
		}
		headNode.AppendChild(styleNode)

		textNode := &html.Node{
			Type: html.TextNode,
			Data: b.Stylesheet,
		}
		styleNode.AppendChild(textNode)
	}

	bodyNode := &html.Node{
		Type: html.ElementNode,
		Data: "body",
	}
	htmlNode.AppendChild(bodyNode)

	b.documentNode = documentNode
	b.bodyNode = bodyNode
}

func (b *HTMLBuilder) Render(w io.Writer) error {
	return html.Render(w, b.documentNode)
}

func (b *HTMLBuilder) AppendDocument(r io.Reader) error {
	b.once.Do(b.init)

	documentNode, err := html.Parse(r)
	if err != nil {
		return err
	}

	htmlNode := documentNode.FirstChild
	if htmlNode == nil {
		return fmt.Errorf("invalid html document: missing html element")
	}

	headNode := htmlNode.FirstChild
	if headNode == nil {
		return fmt.Errorf("invalid html document: missing head element")
	}

	title := ""
	if titleNode, _ := htmlutil.Match(headNode, htmlutil.MatchTag("title")); titleNode != nil {
		title = titleNode.FirstChild.Data
	}
	_ = title

	bodyNode := headNode.NextSibling
	if bodyNode == nil {
		return fmt.Errorf("invalid html document: missing body element")
	}

	// Strip HTML elements
	htmlutil.DeleteChildren(bodyNode, func(node *html.Node) (bool, error) {
		// Allow list tags
		switch node.Data {
		case "head", "meta":
			return true, nil
		default:
			return false, nil
		}
	})

	// Strip attributes
	htmlutil.Walk(bodyNode, func(node *html.Node) error {
		// Allow list attributes
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

	// Append a title to the document
	{
		h3Node := &html.Node{
			Type: html.ElementNode,
			Data: "h3",
		}
		b.bodyNode.AppendChild(h3Node)

		textNode := &html.Node{
			Type: html.TextNode,
			Data: title,
		}
		h3Node.AppendChild(textNode)
	}

	// Append body to the document
	children := slices.Collect(bodyNode.ChildNodes())
	for _, child := range children {
		child.Parent = nil
		child.PrevSibling = nil
		child.NextSibling = nil
		b.bodyNode.AppendChild(child)
	}

	return nil
}
