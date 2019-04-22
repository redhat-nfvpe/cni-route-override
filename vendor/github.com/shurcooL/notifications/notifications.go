// Package notifications provides a notifications service definition.
package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/users"
)

// Service for notifications.
type Service interface {
	// List notifications for authenticated user.
	// Returns a permission error if no authenticated user.
	List(ctx context.Context, opt ListOptions) (Notifications, error)

	// Count notifications for authenticated user.
	// Returns a permission error if no authenticated user.
	Count(ctx context.Context, opt interface{}) (uint64, error)

	// MarkAllRead marks all notifications in the specified repository as read.
	// Returns a permission error if no authenticated user.
	MarkAllRead(ctx context.Context, repo RepoSpec) error

	ExternalService
}

// ExternalService for notifications.
type ExternalService interface {
	// Subscribe subscribes subscribers to the specified thread.
	// If threadType and threadID are zero, subscribers are subscribed
	// to watch the entire repo.
	// Returns a permission error if no authenticated user.
	//
	// THINK: Why is MarkRead and MarkAllRead 2 separate methods instead of 1,
	//        but this is combined into one method? Maybe there should be:
	//        SubscribeAll(ctx context.Context, repo RepoSpec, subscribers []users.UserSpec) error
	//        Or maybe MarkAllRead should be merged into MarkRead?
	Subscribe(ctx context.Context, repo RepoSpec, threadType string, threadID uint64, subscribers []users.UserSpec) error

	// MarkRead marks the specified thread as read.
	// Returns a permission error if no authenticated user.
	MarkRead(ctx context.Context, repo RepoSpec, threadType string, threadID uint64) error

	// Notify notifies subscribers of the specified thread of a notification.
	// Returns a permission error if no authenticated user.
	Notify(ctx context.Context, repo RepoSpec, threadType string, threadID uint64, nr NotificationRequest) error
}

// CopierFrom is an optional interface that allows copying notifications between services.
type CopierFrom interface {
	// CopyFrom copies all accessible notifications from src to dst user.
	// ctx should provide permission to access all notifications in src.
	CopyFrom(ctx context.Context, src Service, dst users.UserSpec) error
}

// ListOptions are options for List operation.
type ListOptions struct {
	// Repo is an optional filter. If not nil, only notifications from Repo will be listed.
	Repo *RepoSpec

	// All specifies whether to include read notifications in addition to unread ones.
	All bool
}

// Notification represents a notification.
type Notification struct {
	RepoSpec   RepoSpec
	ThreadType string
	ThreadID   uint64
	Title      string
	Icon       OcticonID // TODO: Some notifications can exist for a long time. OcticonID may change when frontend updates to newer versions of octicons. Think of a better long term solution?
	Color      RGB
	Actor      users.User
	UpdatedAt  time.Time
	Read       bool
	HTMLURL    string // Address of notification target.

	Participating bool // Whether user is participating in the thread, or just watching.
	Mentioned     bool // Whether user was specifically @mentioned in the content.
}

// NotificationRequest represents a request to create a notification.
type NotificationRequest struct {
	Title     string
	Icon      OcticonID
	Color     RGB
	Actor     users.UserSpec // Actor that triggered the notification. TODO: Maybe not needed? Why not use current user?
	UpdatedAt time.Time      // TODO: Maybe not needed? Why not use time.Now()? Could do it, but time.Now() will be slightly later than original request time.
	HTMLURL   string         // Address of notification target.
}

// Octicon ID. E.g., "issue-opened".
type OcticonID string

// RGB represents a 24-bit color without alpha channel.
type RGB struct {
	R, G, B uint8
}

// HexString returns a hexadecimal color string. For example, "#ff0000" for red.
func (c RGB) HexString() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// Notifications implements sort.Interface.
type Notifications []Notification

func (s Notifications) Len() int           { return len(s) }
func (s Notifications) Less(i, j int) bool { return !s[i].UpdatedAt.Before(s[j].UpdatedAt) }
func (s Notifications) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
