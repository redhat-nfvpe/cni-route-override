// Package event defines event types.
package event

import (
	"encoding/json"
	"fmt"
	"time"

	"dmitri.shuralyov.com/state"
	"github.com/shurcooL/users"
)

// Event represents an event.
type Event struct {
	Time      time.Time
	Actor     users.User // UserSpec and Login fields populated.
	Container string     // URL of container without schema. E.g., "github.com/user/repo".

	// Payload specifies the event type. It's one of:
	// Issue, Change, IssueComment, ChangeComment, CommitComment,
	// Push, Star, Create, Fork, Delete, Wiki.
	Payload interface{}
}

// MarshalJSON implements the json.Marshaler interface.
func (e Event) MarshalJSON() ([]byte, error) {
	v := struct {
		Time      time.Time
		Actor     users.User
		Container string
		Type      string
		Payload   interface{}
	}{
		Time:      e.Time,
		Actor:     e.Actor,
		Container: e.Container,
		Payload:   e.Payload,
	}
	switch e.Payload.(type) {
	case Issue:
		v.Type = "Issue"
	case Change:
		v.Type = "Change"
	case IssueComment:
		v.Type = "IssueComment"
	case ChangeComment:
		v.Type = "ChangeComment"
	case CommitComment:
		v.Type = "CommitComment"
	case Push:
		v.Type = "Push"
	case Star:
		v.Type = "Star"
	case Create:
		v.Type = "Create"
	case Fork:
		v.Type = "Fork"
	case Delete:
		v.Type = "Delete"
	case Wiki:
		v.Type = "Wiki"
	default:
		return nil, fmt.Errorf("Event.MarshalJSON: invalid payload type %T", e.Payload)
	}
	return json.Marshal(v)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Event) UnmarshalJSON(b []byte) error {
	// Ignore null, like in the main JSON package.
	if string(b) == "null" {
		return nil
	}
	var v struct {
		Time      time.Time
		Actor     users.User
		Container string
		Type      string
		Payload   json.RawMessage
	}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	*e = Event{
		Time:      v.Time,
		Actor:     v.Actor,
		Container: v.Container,
	}
	switch v.Type {
	case "Issue":
		var p Issue
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "Change":
		var p Change
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "IssueComment":
		var p IssueComment
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "ChangeComment":
		var p ChangeComment
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "CommitComment":
		var p CommitComment
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "Push":
		var p Push
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "Star":
		var p Star
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "Create":
		var p Create
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "Fork":
		var p Fork
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "Delete":
		var p Delete
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	case "Wiki":
		var p Wiki
		err := json.Unmarshal(v.Payload, &p)
		if err != nil {
			return err
		}
		e.Payload = p
	default:
		return fmt.Errorf("Event.UnmarshalJSON: invalid payload type %q", v.Type)
	}
	return nil
}

// Issue is an issue event.
type Issue struct {
	Action       string // "opened", "closed", "reopened".
	IssueTitle   string
	IssueHTMLURL string
}

// Change is a change event.
type Change struct {
	Action        string // "opened", "closed", "merged", "reopened".
	ChangeTitle   string
	ChangeHTMLURL string
}

// IssueComment is an issue comment event.
type IssueComment struct {
	IssueTitle     string
	IssueState     state.Issue
	CommentBody    string
	CommentHTMLURL string
}

// ChangeComment is a change comment event.
// A change comment is a review iff CommentReview is non-zero.
type ChangeComment struct {
	ChangeTitle    string
	ChangeState    state.Change
	CommentBody    string
	CommentReview  state.Review
	CommentHTMLURL string
}

// CommitComment is a commit comment event.
type CommitComment struct {
	Commit      Commit
	CommentBody string
}

// Push is a push event.
type Push struct {
	Branch  string   // Name of branch pushed to. E.g., "master".
	Head    string   // SHA of the most recent commit after the push.
	Before  string   // SHA of the most recent commit before the push.
	Commits []Commit // Ordered from earliest to most recent (head).

	HeadHTMLURL   string // Optional.
	BeforeHTMLURL string // Optional.
}

// Star is a star event.
type Star struct{}

// Create is a create event.
type Create struct {
	Type        string // "repository", "package", "branch", "tag".
	Name        string // Only for "branch", "tag" types.
	Description string // Only for "repository", "package" types.
}

// Fork is a fork event.
type Fork struct {
	Container string // URL of forkee container without schema. E.g., "github.com/user/repo".
}

// Delete is a delete event.
type Delete struct {
	Type string // "branch", "tag".
	Name string
}

// Wiki is a wiki event. It happens when an actor updates a wiki.
type Wiki struct {
	Pages []Page // Wiki pages that are affected.
}
