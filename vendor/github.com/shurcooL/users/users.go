// Package users provides a users service definition.
package users

import (
	"context"
)

// Service for users.
type Service interface {
	// Get fetches the specified user.
	Get(ctx context.Context, user UserSpec) (User, error)

	// GetAuthenticatedSpec fetches the currently authenticated
	// user specification, or UserSpec{ID: 0} if there is no
	// authenticated user. A non-nil error is returned if the
	// the authentication process was not able to successfully
	// determine if a user or no user is currently authenticated.
	GetAuthenticatedSpec(ctx context.Context) (UserSpec, error)

	// GetAuthenticated fetches the currently authenticated user,
	// or User{UserSpec: UserSpec{ID: 0}} if there is no authenticated user.
	GetAuthenticated(ctx context.Context) (User, error)

	// Edit the authenticated user.
	Edit(ctx context.Context, er EditRequest) (User, error)
}

// Store for users.
type Store interface {
	// Create creates the specified user.
	// It returns os.ErrExist if the user already exists.
	Create(ctx context.Context, user User) error

	// Get fetches the specified user.
	Get(ctx context.Context, user UserSpec) (User, error)

	// CONSIDER: Delete user.
	//Delete(ctx context.Context, user UserSpec) error
}

// UserSpec specifies a user.
// ID value 0 represents no user. Valid users may not use 0 as their ID.
type UserSpec struct {
	ID     uint64
	Domain string
}

// User represents a user.
type User struct {
	UserSpec
	Elsewhere []UserSpec // THINK: Consider merging Elsewhere with root-most UserSpec. Maybe even use a set, order of linked accounts shouldn't matter, should it?

	Login     string
	Name      string
	Email     string // Public email.
	AvatarURL string
	HTMLURL   string

	SiteAdmin bool
}

// EditRequest represents a request to edit a user.
type EditRequest struct {
	// Currently nothing, but editable fields will be added here in the future.
}
