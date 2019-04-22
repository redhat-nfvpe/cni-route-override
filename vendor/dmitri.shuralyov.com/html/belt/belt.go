// Package belt is an opinionated collection of HTML components
// for shared use by multiple web apps.
package belt

import (
	"fmt"
	"strings"

	"dmitri.shuralyov.com/state"
	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/octicon"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Issue is a component that displays an issue, with a state icon and title.
type Issue struct {
	State   state.Issue
	Title   string
	HTMLURL string
	Short   bool
}

func (i Issue) Render() []*html.Node {
	n := iconLink{
		Text:    i.Title,
		Tooltip: i.Title,
		URL:     i.HTMLURL,
	}
	if i.Short {
		n.Text = shortTitle(i.Title)
	}
	switch i.State {
	case state.IssueOpen:
		n.Icon = octicon.IssueOpened
		n.IconColor = &rgb{R: 0x6c, G: 0xc6, B: 0x44} // Green.
	case state.IssueClosed:
		n.Icon = octicon.IssueClosed
		n.IconColor = &rgb{R: 0xbd, G: 0x2c, B: 0x00} // Red.
	}
	return n.Render()
}

// Change is a component that displays a change, with a state icon and title.
type Change struct {
	State   state.Change
	Title   string
	HTMLURL string
	Short   bool
}

func (c Change) Render() []*html.Node {
	n := iconLink{
		Text:    c.Title,
		Tooltip: c.Title,
		URL:     c.HTMLURL,
	}
	if c.Short {
		n.Text = shortTitle(c.Title)
	}
	switch c.State {
	case state.ChangeOpen:
		n.Icon = octicon.GitPullRequest
		n.IconColor = &rgb{R: 0x6c, G: 0xc6, B: 0x44} // Green.
	case state.ChangeClosed:
		n.Icon = octicon.GitPullRequest
		n.IconColor = &rgb{R: 0xbd, G: 0x2c, B: 0x00} // Red.
	case state.ChangeMerged:
		n.Icon = octicon.GitMerge
		n.IconColor = &rgb{R: 0x6e, G: 0x54, B: 0x94} // Purple.
	}
	return n.Render()
}

func shortTitle(s string) string {
	if len(s) <= 36 {
		return s
	}
	return s[:35] + "…"
}

// iconLink consists of an icon and a text link.
// Icon must be not nil.
type iconLink struct {
	Text      string
	Tooltip   string
	URL       string
	Black     bool              // Black link.
	Icon      func() *html.Node // Not nil.
	IconColor *rgb              // Optional icon color override.
}

func (d iconLink) Render() []*html.Node {
	a := &html.Node{
		Type: html.ElementNode, Data: atom.A.String(),
		Attr: []html.Attribute{{Key: atom.Href.String(), Val: d.URL}},
	}
	if d.Tooltip != "" {
		a.Attr = append(a.Attr, html.Attribute{Key: atom.Title.String(), Val: d.Tooltip})
	}
	if d.Black {
		a.Attr = append(a.Attr, html.Attribute{Key: atom.Class.String(), Val: "black"})
	}
	iconSpanStyle := "margin-right: 4px;"
	if d.IconColor != nil {
		iconSpanStyle += fmt.Sprintf(" color: %s;", d.IconColor.HexString())
	}
	a.AppendChild(&html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr:       []html.Attribute{{Key: atom.Style.String(), Val: iconSpanStyle}},
		FirstChild: d.Icon(),
	})
	a.AppendChild(htmlg.Text(d.Text))
	return []*html.Node{a}
}

// rgb represents a 24-bit color without alpha channel.
type rgb struct {
	R, G, B uint8
}

// HexString returns a hexadecimal color string. For example, "#ff0000" for red.
func (c rgb) HexString() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// Commit is a component that displays a commit, with an author and title.
type Commit struct {
	SHA             string
	Message         string
	AuthorAvatarURL string
	HTMLURL         string // Optional.
	Short           bool
}

func (c Commit) Render() []*html.Node {
	avatar := &html.Node{
		Type: html.ElementNode, Data: atom.Img.String(),
		Attr: []html.Attribute{
			{Key: atom.Src.String(), Val: c.AuthorAvatarURL},
			{Key: atom.Style.String(), Val: "width: 16px; height: 16px; vertical-align: top; margin-right: 4px;"},
		},
	}
	commitID := CommitID{SHA: c.SHA, HTMLURL: c.HTMLURL}
	message := &html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "margin-left: 4px;"},
			{Key: atom.Title.String(), Val: c.Message},
		},
	}
	switch c.Short {
	case false:
		message.AppendChild(htmlg.Text(firstParagraph(c.Message)))
	case true:
		message.AppendChild(htmlg.Text(shortCommit(firstParagraph(c.Message))))
	}

	var ns []*html.Node
	ns = append(ns, avatar)
	ns = append(ns, commitID.Render()...)
	ns = append(ns, message)
	return ns
}

func shortCommit(s string) string {
	if len(s) <= 24 {
		return s
	}
	return s[:23] + "…"
}

// firstParagraph returns the first paragraph of text s.
func firstParagraph(s string) string {
	i := strings.Index(s, "\n\n")
	if i == -1 {
		return s
	}
	return s[:i]
}

// CommitID is a component that displays a commit ID. E.g., "c0de1234".
type CommitID struct {
	SHA     string
	HTMLURL string // Optional.
}

func (c CommitID) Render() []*html.Node {
	sha := &html.Node{
		Type: html.ElementNode, Data: atom.Code.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "width: 8ch; overflow: hidden; display: inline-grid; white-space: nowrap;"},
			{Key: atom.Title.String(), Val: c.SHA},
		},
		FirstChild: htmlg.Text(c.SHA),
	}
	if c.HTMLURL != "" {
		sha = &html.Node{
			Type: html.ElementNode, Data: atom.A.String(),
			Attr: []html.Attribute{
				{Key: atom.Href.String(), Val: c.HTMLURL},
			},
			FirstChild: sha,
		}
	}
	return []*html.Node{sha}
}

// Reference is a component that displays a reference (branch or tag). E.g., "master".
type Reference struct {
	Name          string
	Strikethrough bool
}

func (r Reference) Render() []*html.Node {
	codeStyle := `padding: 2px 6px;
background-color: rgb(232, 241, 246);
border-radius: 3px;`
	if r.Strikethrough {
		codeStyle += `text-decoration: line-through; color: gray;`
	}
	code := &html.Node{
		Type: html.ElementNode, Data: atom.Code.String(),
		Attr:       []html.Attribute{{Key: atom.Style.String(), Val: codeStyle}},
		FirstChild: htmlg.Text(r.Name),
	}
	return []*html.Node{code}
}
