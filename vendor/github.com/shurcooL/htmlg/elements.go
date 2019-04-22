package htmlg

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Text returns a plain text node.
func Text(s string) *html.Node {
	return &html.Node{
		Type: html.TextNode, Data: s,
	}
}

// Strong returns a strong text node.
func Strong(s string) *html.Node {
	strong := &html.Node{
		Type: html.ElementNode, Data: atom.Strong.String(),
	}
	strong.AppendChild(Text(s))
	return strong
}

// A returns an anchor element <a href="{{.href}}">{{.s}}</a>.
func A(s, href string) *html.Node {
	a := &html.Node{
		Type: html.ElementNode, Data: atom.A.String(),
		Attr: []html.Attribute{{Key: atom.Href.String(), Val: href}},
	}
	a.AppendChild(Text(s))
	return a
}

// H1 returns a h1 element <h1>{{range .nodes}}{{.}}{{end}}</h1>.
func H1(nodes ...*html.Node) *html.Node {
	h1 := &html.Node{
		Type: html.ElementNode, Data: atom.H1.String(),
	}
	for _, n := range nodes {
		h1.AppendChild(n)
	}
	return h1
}

// H2 returns a h2 element <h2>{{range .nodes}}{{.}}{{end}}</h2>.
func H2(nodes ...*html.Node) *html.Node {
	h2 := &html.Node{
		Type: html.ElementNode, Data: atom.H2.String(),
	}
	for _, n := range nodes {
		h2.AppendChild(n)
	}
	return h2
}

// H3 returns a h3 element <h3>{{range .nodes}}{{.}}{{end}}</h3>.
func H3(nodes ...*html.Node) *html.Node {
	h3 := &html.Node{
		Type: html.ElementNode, Data: atom.H3.String(),
	}
	for _, n := range nodes {
		h3.AppendChild(n)
	}
	return h3
}

// H4 returns a h4 element <h4>{{range .nodes}}{{.}}{{end}}</h4>.
func H4(nodes ...*html.Node) *html.Node {
	h4 := &html.Node{
		Type: html.ElementNode, Data: atom.H4.String(),
	}
	for _, n := range nodes {
		h4.AppendChild(n)
	}
	return h4
}

// P returns a p element <p>{{range .nodes}}{{.}}{{end}}</p>.
func P(nodes ...*html.Node) *html.Node {
	p := &html.Node{
		Type: html.ElementNode, Data: atom.P.String(),
	}
	for _, n := range nodes {
		p.AppendChild(n)
	}
	return p
}

// DL returns a dl element <dl>{{range .nodes}}{{.}}{{end}}</dl>.
func DL(nodes ...*html.Node) *html.Node {
	dl := &html.Node{
		Type: html.ElementNode, Data: atom.Dl.String(),
	}
	for _, n := range nodes {
		dl.AppendChild(n)
	}
	return dl
}

// DT returns a dt element <dt>{{range .nodes}}{{.}}{{end}}</dt>.
func DT(nodes ...*html.Node) *html.Node {
	dt := &html.Node{
		Type: html.ElementNode, Data: atom.Dt.String(),
	}
	for _, n := range nodes {
		dt.AppendChild(n)
	}
	return dt
}

// DD returns a dd element <dd>{{range .nodes}}{{.}}{{end}}</dd>.
func DD(nodes ...*html.Node) *html.Node {
	dd := &html.Node{
		Type: html.ElementNode, Data: atom.Dd.String(),
	}
	for _, n := range nodes {
		dd.AppendChild(n)
	}
	return dd
}

// UL returns a ul element <ul>{{range .nodes}}{{.}}{{end}}</ul>.
func UL(nodes ...*html.Node) *html.Node {
	ul := &html.Node{
		Type: html.ElementNode, Data: atom.Ul.String(),
	}
	for _, n := range nodes {
		ul.AppendChild(n)
	}
	return ul
}

// LI returns a li element <li>{{range .nodes}}{{.}}{{end}}</li>.
func LI(nodes ...*html.Node) *html.Node {
	li := &html.Node{
		Type: html.ElementNode, Data: atom.Li.String(),
	}
	for _, n := range nodes {
		li.AppendChild(n)
	}
	return li
}

// TR returns a tr element <tr>{{range .nodes}}{{.}}{{end}}</tr>.
func TR(nodes ...*html.Node) *html.Node {
	tr := &html.Node{
		Type: html.ElementNode, Data: atom.Tr.String(),
	}
	for _, n := range nodes {
		tr.AppendChild(n)
	}
	return tr
}

// TD returns a td element <td>{{range .nodes}}{{.}}{{end}}</td>.
func TD(nodes ...*html.Node) *html.Node {
	td := &html.Node{
		Type: html.ElementNode, Data: atom.Td.String(),
	}
	for _, n := range nodes {
		td.AppendChild(n)
	}
	return td
}

// Div returns a div element <div>{{range .nodes}}{{.}}{{end}}</div>.
//
// Div is experimental and may be changed or removed.
func Div(nodes ...*html.Node) *html.Node {
	div := &html.Node{
		Type: html.ElementNode, Data: atom.Div.String(),
	}
	for _, n := range nodes {
		div.AppendChild(n)
	}
	return div
}

// DivClass returns a div element <div class="{{.class}}">{{range .nodes}}{{.}}{{end}}</div>.
//
// DivClass is experimental and may be changed or removed.
func DivClass(class string, nodes ...*html.Node) *html.Node {
	div := &html.Node{
		Type: html.ElementNode, Data: atom.Div.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: class}},
	}
	for _, n := range nodes {
		div.AppendChild(n)
	}
	return div
}

// Span returns a span element <span>{{range .nodes}}{{.}}{{end}}</span>.
//
// Span is experimental and may be changed or removed.
func Span(nodes ...*html.Node) *html.Node {
	span := &html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
	}
	for _, n := range nodes {
		span.AppendChild(n)
	}
	return span
}

// SpanClass returns a span element <span class="{{.class}}">{{range .nodes}}{{.}}{{end}}</span>.
//
// SpanClass is experimental and may be changed or removed.
func SpanClass(class string, nodes ...*html.Node) *html.Node {
	span := &html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: class}},
	}
	for _, n := range nodes {
		span.AppendChild(n)
	}
	return span
}

// ULClass returns a div element <ul class="{{.class}}">{{range .nodes}}{{.}}{{end}}</ul>.
//
// ULClass is experimental and may be changed or removed.
func ULClass(class string, nodes ...*html.Node) *html.Node {
	ul := &html.Node{
		Type: html.ElementNode, Data: atom.Ul.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: class}},
	}
	for _, n := range nodes {
		ul.AppendChild(n)
	}
	return ul
}

// LIClass returns a div element <li class="{{.class}}">{{range .nodes}}{{.}}{{end}}</li>.
//
// LIClass is experimental and may be changed or removed.
func LIClass(class string, nodes ...*html.Node) *html.Node {
	li := &html.Node{
		Type: html.ElementNode, Data: atom.Li.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: class}},
	}
	for _, n := range nodes {
		li.AppendChild(n)
	}
	return li
}
