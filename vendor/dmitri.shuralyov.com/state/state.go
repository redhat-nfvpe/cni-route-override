// Package state defines states for domain types.
package state

// Issue represents the possible states of an issue.
type Issue string

// The possible states of an issue.
const (
	IssueOpen   Issue = "open"   // An issue that is still open.
	IssueClosed Issue = "closed" // An issue that has been closed.
)

// Change represents the possible states of a change.
type Change string

// The possible states of a change.
const (
	ChangeOpen   Change = "open"   // A change that is still open.
	ChangeClosed Change = "closed" // A change that has been closed without being merged.
	ChangeMerged Change = "merged" // A change that has been closed by being merged.
)

// Review represents the possible states of a change review.
type Review int8

const (
	ReviewPlus2   Review = +2 // Looks good to me, approved.
	ReviewPlus1   Review = +1 // Looks good to me, but someone else must approve.
	ReviewNoScore Review = 0  // No score, just a comment.
	ReviewMinus1  Review = -1 // I would prefer this is not merged as is, and here's why.
	ReviewMinus2  Review = -2 // This shall not be merged, and here's why.
)
