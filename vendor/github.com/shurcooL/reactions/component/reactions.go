// Package component contains individual components that can render themselves as HTML.
package component

import (
	"fmt"

	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/octicon"
	"github.com/shurcooL/reactions"
	"github.com/shurcooL/users"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ReactionsBar is a component next to anything that can be reacted to, with reactable ID.
// It displays all reactions for that reactable ID, and a NewReaction component for adding new reactions.
type ReactionsBar struct {
	Reactions   []reactions.Reaction
	CurrentUser users.User
	ID          string // ID is the reactable ID.
}

func (r ReactionsBar) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		<div id="reactable-container-{{.ID}}" class="reactable-container" data-reactableID="{{.ID}}">
			ReactionsBarInner{Reactions: reactions, CurrentUser: r.CurrentUser, ReactableID: r.ID}
		</div>
	*/
	div := &html.Node{
		Type: html.ElementNode, Data: atom.Div.String(),
		Attr: []html.Attribute{
			{Key: atom.Id.String(), Val: "reactable-container-" + r.ID},
			{Key: atom.Class.String(), Val: "reactable-container"},
			{Key: "data-reactableID", Val: r.ID},
		},
	}
	htmlg.AppendChildren(div, ReactionsBarInner{Reactions: r.Reactions, CurrentUser: r.CurrentUser, ReactableID: r.ID}.Render()...)
	return []*html.Node{div}
}

// ReactionsBarInner is a static component that displays all reactions, and
// a NewReaction component with ReactableID for adding new reactions.
type ReactionsBarInner struct {
	Reactions   []reactions.Reaction
	CurrentUser users.User
	ReactableID string
}

func (r ReactionsBarInner) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		{{template "reactions" prioritizeThumbsUpDown(.Reactions)}}
		NewReaction{ReactableID: r.ReactableID}
	*/
	var nodes []*html.Node
	spacingAfter := prioritizeThumbsUpDown(r.Reactions)
	for i, reaction := range r.Reactions {
		nodes = append(nodes, Reaction{Reaction: reaction, CurrentUser: r.CurrentUser}.Render()...)
		if i == spacingAfter {
			nodes = append(nodes, &html.Node{
				Type: html.ElementNode, Data: atom.Span.String(),
				Attr: []html.Attribute{{Key: atom.Style.String(), Val: "margin-right: 4px;"}},
			})
		}
	}
	nodes = append(nodes, NewReaction{ReactableID: r.ReactableID}.Render()...)
	return nodes
}

// prioritizeThumbsUpDown bubbles thumbs up/down reactions to the front. It returns
// an index after which spacing should be inserted to visually separate
// thumbs up/down reactions from the rest, or -1 if no need.
func prioritizeThumbsUpDown(reactions []reactions.Reaction) (spacingAfter int) {
	spacingAfter = -1
	for i, reaction := range reactions { // Move thumbs down reaction to the front first.
		if reaction.Reaction == "-1" {
			thumbsDown := reaction
			for ; i > 0; i-- {
				reactions[i] = reactions[i-1]
			}
			reactions[0] = thumbsDown
			spacingAfter++
			break
		}
	}
	for i, reaction := range reactions { // Move thumbs up reaction to the front last.
		if reaction.Reaction == "+1" {
			thumbsUp := reaction
			for ; i > 0; i-- {
				reactions[i] = reactions[i-1]
			}
			reactions[0] = thumbsUp
			spacingAfter++
			break
		}
	}
	if spacingAfter == len(reactions)-1 {
		spacingAfter = -1 // No need for spacing if there are no other reactions after +1 and -1.
	}
	return spacingAfter
}

// Reaction is a component for displaying a single Reaction, as seen by CurrentUser.
type Reaction struct {
	Reaction    reactions.Reaction
	CurrentUser users.User
}

func (r Reaction) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		<a class="reaction" href="javascript:" title="{{reactionTooltip .}}" onclick="ToggleReaction(this, event, {{.Reaction | json}});">
			<div class="reaction {{if (not (containsCurrentUser .Users))}}others{{end}}">
				<span class="emoji-outer emoji-sizer">
					<span class="emoji-inner" style="background-position: {{reactionPosition .Reaction}};">
					</span>
				</span>
				<strong>{{len .Users}}</strong>
			</div>
		</a>
	*/
	innerSpan := &html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{
			{Key: atom.Class.String(), Val: "emoji-inner"},
			{Key: atom.Style.String(), Val: fmt.Sprintf("background-position: %s;", reactions.Position(":"+string(r.Reaction.Reaction)+":"))},
		},
	}
	outerSpan := &html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: "emoji-outer emoji-sizer"}},
	}
	outerSpan.AppendChild(innerSpan)
	strong := htmlg.Strong(fmt.Sprint(len(r.Reaction.Users)))
	divClass := "reaction"
	if !r.containsCurrentUser(r.Reaction.Users) {
		divClass += " others"
	}
	div := &html.Node{
		Type: html.ElementNode, Data: atom.Div.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: divClass}},
	}
	div.AppendChild(outerSpan)
	div.AppendChild(strong)
	a := &html.Node{
		Type: html.ElementNode, Data: atom.A.String(),
		Attr: []html.Attribute{
			{Key: atom.Class.String(), Val: "reaction"},
			{Key: atom.Href.String(), Val: "javascript:"},
			{Key: atom.Title.String(), Val: r.reactionTooltip(r.Reaction)},
			{Key: atom.Onclick.String(), Val: fmt.Sprintf("ToggleReaction(this, event, '%q');", r.Reaction.Reaction)},
		},
	}
	a.AppendChild(div)
	return []*html.Node{a}
}

func (r Reaction) containsCurrentUser(users []users.User) bool {
	if r.CurrentUser.ID == 0 {
		return false
	}
	for _, u := range users {
		if u.UserSpec == r.CurrentUser.UserSpec {
			return true
		}
	}
	return false
}

func (r Reaction) reactionTooltip(reaction reactions.Reaction) string {
	var users string
	for i, u := range reaction.Users {
		if i != 0 {
			if i < len(reaction.Users)-1 {
				users += ", "
			} else {
				users += " and "
			}
		}
		if r.CurrentUser.ID != 0 && u.UserSpec == r.CurrentUser.UserSpec {
			if i == 0 {
				users += "You"
			} else {
				users += "you"
			}
		} else {
			users += u.Login
		}
	}
	// TODO: Handle when there are too many users and their details are left out by backend.
	//       Count them and add "and N others" here.
	return fmt.Sprintf("%v reacted with :%v:.", users, reaction.Reaction)
}

// NewReaction is a component for adding new reactions to a Reactable with ReactableID id.
type NewReaction struct {
	ReactableID string
}

func (nr NewReaction) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		<a href="javascript:" title="React" onclick="ShowReactionMenu(this, event, {{.}});">
			<div class="new-reaction">
				<octicon.Smiley() class="smiley" />
				<octicon.PlusSmall() class="plus" />
			</div>
		</a>
	*/
	smiley := octicon.Smiley()
	smiley.Attr = append(smiley.Attr, html.Attribute{
		Key: atom.Class.String(), Val: "smiley",
	})
	plus := octicon.PlusSmall()
	plus.Attr = append(plus.Attr, html.Attribute{
		Key: atom.Class.String(), Val: "plus",
	})
	div := &html.Node{
		Type: html.ElementNode, Data: atom.Div.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: "new-reaction"}},
	}
	div.AppendChild(smiley)
	div.AppendChild(plus)
	a := &html.Node{
		Type: html.ElementNode, Data: atom.A.String(),
		Attr: []html.Attribute{
			{Key: atom.Class.String(), Val: "new-reaction"},
			{Key: atom.Href.String(), Val: "javascript:"},
			{Key: atom.Title.String(), Val: "React"},
			{Key: atom.Onclick.String(), Val: fmt.Sprintf("ShowReactionMenu(this, event, '%q');", nr.ReactableID)},
		},
	}
	a.AppendChild(div)
	return []*html.Node{a}
}
