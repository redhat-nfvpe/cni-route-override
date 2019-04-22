// Package htmlg contains helper funcs for generating HTML nodes and rendering them.
// Context-aware escaping is done just like in html/template, making it safe against code injection.
//
// Note: This package is quite experimental in nature, so its API is susceptible to more frequent
// changes than the average package. This is necessary in order to keep this package useful.
package htmlg

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

// Render renders HTML nodes, returning result as a string.
// Context-aware escaping is done just like in html/template when rendering nodes.
func Render(nodes ...*html.Node) string {
	var buf bytes.Buffer
	for _, node := range nodes {
		err := html.Render(&buf, node)
		if err != nil {
			// html.Render should only return a non-nil error if there's a problem writing to the supplied io.Writer.
			// We don't expect that to ever be the case (unless there's not enough memory), so panic.
			// If this ever happens in other situations, it's a bug in this library that should be reported and fixed.
			panic(fmt.Errorf("internal error: html.Render returned non-nil error, this is not expected to happen: %v", err))
		}
	}
	return buf.String()
}

// Component is anything that can render itself into HTML nodes.
type Component interface {
	Render() []*html.Node
}

// RenderComponents renders components into HTML, writing result to w.
// Context-aware escaping is done just like in html/template when rendering nodes.
func RenderComponents(w io.Writer, components ...Component) error {
	for _, c := range components {
		for _, node := range c.Render() {
			err := html.Render(w, node)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RenderComponentsString renders components into HTML, returning result as a string.
// Context-aware escaping is done just like in html/template when rendering nodes.
func RenderComponentsString(components ...Component) string {
	var buf bytes.Buffer
	for _, c := range components {
		for _, node := range c.Render() {
			err := html.Render(&buf, node)
			if err != nil {
				// html.Render should only return a non-nil error if there's a problem writing to the supplied io.Writer.
				// We don't expect that to ever be the case (unless there's not enough memory), so panic.
				// If this ever happens in other situations, it's a bug in this library that should be reported and fixed.
				panic(fmt.Errorf("internal error: html.Render returned non-nil error, this is not expected to happen: %v", err))
			}
		}
	}
	return buf.String()
}

// Nodes implements the Component interface from a slice of HTML nodes.
//
// The Render method always returns the same references to existing nodes,
// and as a result, it is unsuitable to be rendered and attached to other
// HTML nodes more than once. It is suitable to rendered directly into HTML
// multiple times, or attached to an existing node once.
//
// Nodes is experimental and may be changed or removed.
type Nodes []*html.Node

func (ns Nodes) Render() []*html.Node {
	return []*html.Node(ns)
}

// NodeComponent is a wrapper that makes a Component from a single html.Node.
type NodeComponent html.Node

func (n NodeComponent) Render() []*html.Node {
	node := html.Node(n)
	return []*html.Node{&node}
}

// AppendChildren adds nodes cs as children of n.
//
// It will panic if any of cs already has a parent or siblings.
func AppendChildren(n *html.Node, cs ...*html.Node) {
	for _, c := range cs {
		n.AppendChild(c)
	}
}
