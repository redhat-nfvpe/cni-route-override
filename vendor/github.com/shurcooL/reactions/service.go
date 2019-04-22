package reactions

import (
	"context"

	"github.com/shurcooL/users"
)

// Service defines methods of a reactions service.
type Service interface {
	// List all reactions at uri. Map key is reactable ID.
	// uri is clean '/'-separated URI. E.g., "example.com/page".
	// If uri isn't a valid reactable URI, a not exist error is returned.
	List(ctx context.Context, uri string) (map[string][]Reaction, error)

	// Get reactions for id at uri.
	// uri is clean '/'-separated URI. E.g., "example.com/page".
	// If uri/id doesn't point at a valid reactable target,
	// a not exist error is returned.
	Get(ctx context.Context, uri string, id string) ([]Reaction, error)

	// Toggle a reaction for id at uri.
	// If uri/id doesn't point at a valid reactable target,
	// a not exist error is returned.
	Toggle(ctx context.Context, uri string, id string, tr ToggleRequest) ([]Reaction, error)
}

// Reaction represents a single reaction, backed by 1 or more users.
type Reaction struct {
	Reaction EmojiID
	Users    []users.User // Length is 1 or more. First entry is first person who reacted.
}

// EmojiID is the id of a reaction. For example, "+1".
// TODO, THINK: Maybe keep the colons, i.e., ":+1:".
type EmojiID string

// ToggleRequest is a request to toggle a reaction.
type ToggleRequest struct {
	Reaction EmojiID
}

// Validate returns non-nil error if the request is invalid.
func (ToggleRequest) Validate() error {
	// TODO: Maybe validate that the emojiID is one of supported ones.
	//       Or maybe not (unsupported ones can be handled by frontend component).
	//       That way custom emoji can be added/removed, etc. Figure out what the best thing to do is and do it.
	return nil
}
