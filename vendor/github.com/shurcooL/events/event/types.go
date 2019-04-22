package event

// Commit describes a commit in a CommitComment or Push event.
type Commit struct {
	SHA             string
	Message         string
	AuthorAvatarURL string
	HTMLURL         string // Optional.
}

// Page describes a page action in a Wiki event.
type Page struct {
	Action         string // "created", "edited".
	SHA            string
	Title          string
	HTMLURL        string
	CompareHTMLURL string
}
