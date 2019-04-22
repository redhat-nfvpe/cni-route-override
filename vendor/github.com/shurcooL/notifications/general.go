package notifications

// TODO: Consider best (centralized?) place for RepoSpec?
//       Also consider replacing it with `RepoURI string`.
//       Also consider interface { RepoSpec() string }.

// RepoSpec is a specification for a repository.
type RepoSpec struct {
	URI string // URI is clean '/'-separated URI. E.g., "example.com/user/repo".
}

// String implements fmt.Stringer.
func (rs RepoSpec) String() string {
	return rs.URI
}
