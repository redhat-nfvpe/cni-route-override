// Package maintner implements a read-only issues.Service using
// a x/build/maintner corpus.
package maintner

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shurcooL/issues"
	"github.com/shurcooL/users"
	"golang.org/x/build/maintner"
)

// NewService creates an issues.Service backed with the given corpus.
func NewService(corpus *maintner.Corpus) issues.Service {
	return service{
		c: corpus,
	}
}

type service struct {
	c *maintner.Corpus
}

func (s service) List(_ context.Context, rs issues.RepoSpec, opt issues.IssueListOptions) ([]issues.Issue, error) {
	// TODO: Pagination.

	repoID, err := ghRepoID(rs)
	if err != nil {
		return nil, err
	}
	s.c.RLock()
	defer s.c.RUnlock()
	repo := s.c.GitHub().Repo(repoID.Owner, repoID.Repo)
	if repo == nil {
		return nil, fmt.Errorf("repo %v not found", rs)
	}

	var is []issues.Issue
	err = repo.ForeachIssue(func(i *maintner.GitHubIssue) error {
		if i.NotExist || i.PullRequest {
			return nil
		}

		state := ghState(i)
		switch {
		case opt.State == issues.StateFilter(issues.OpenState) && state != issues.OpenState:
			return nil
		case opt.State == issues.StateFilter(issues.ClosedState) && state != issues.ClosedState:
			return nil
		}

		var labels []issues.Label
		for _, l := range i.Labels {
			labels = append(labels, issues.Label{
				Name: l.Name,
				// TODO: Can we use label ID to figure out its color?
				Color: issues.RGB{R: 0xed, G: 0xed, B: 0xed}, // maintner.Corpus doesn't support GitHub issue label colors, so fall back to a default light gray.
			})
		}
		replies := 0
		err := i.ForeachComment(func(*maintner.GitHubComment) error {
			replies++
			return nil
		})
		if err != nil {
			return err
		}
		is = append(is, issues.Issue{
			ID:     uint64(i.Number),
			State:  state,
			Title:  i.Title,
			Labels: labels,
			Comment: issues.Comment{
				User:      ghUser(i.User),
				CreatedAt: i.Created,
			},
			Replies: replies,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(is, func(i, j int) bool { return is[i].ID > is[j].ID })
	return is, nil
}

func (s service) Count(_ context.Context, rs issues.RepoSpec, opt issues.IssueListOptions) (uint64, error) {
	repoID, err := ghRepoID(rs)
	if err != nil {
		return 0, err
	}
	s.c.RLock()
	defer s.c.RUnlock()
	repo := s.c.GitHub().Repo(repoID.Owner, repoID.Repo)
	if repo == nil {
		return 0, fmt.Errorf("repo %v not found", rs)
	}

	var count uint64
	err = repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
		if issue.NotExist || issue.PullRequest {
			return nil
		}

		state := ghState(issue)
		switch {
		case opt.State == issues.StateFilter(issues.OpenState) && state != issues.OpenState:
			return nil
		case opt.State == issues.StateFilter(issues.ClosedState) && state != issues.ClosedState:
			return nil
		}

		count++

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s service) Get(_ context.Context, rs issues.RepoSpec, id uint64) (issues.Issue, error) {
	repoID, err := ghRepoID(rs)
	if err != nil {
		return issues.Issue{}, err
	}
	s.c.RLock()
	defer s.c.RUnlock()
	repo := s.c.GitHub().Repo(repoID.Owner, repoID.Repo)
	if repo == nil {
		return issues.Issue{}, fmt.Errorf("repo %v not found", rs)
	}
	i := repo.Issue(int32(id))
	if i == nil || i.NotExist || i.PullRequest {
		return issues.Issue{}, os.ErrNotExist
	}

	return issues.Issue{
		ID:    uint64(i.Number),
		State: ghState(i),
		Title: i.Title,
		Comment: issues.Comment{
			User:      ghUser(i.User),
			CreatedAt: i.Created,
		},
	}, nil
}

func (s service) ListComments(_ context.Context, rs issues.RepoSpec, id uint64, opt *issues.ListOptions) ([]issues.Comment, error) {
	repoID, err := ghRepoID(rs)
	if err != nil {
		return nil, err
	}
	s.c.RLock()
	defer s.c.RUnlock()
	repo := s.c.GitHub().Repo(repoID.Owner, repoID.Repo)
	if repo == nil {
		return nil, fmt.Errorf("repo %v not found", rs)
	}
	i := repo.Issue(int32(id))
	if i == nil || i.NotExist || i.PullRequest {
		return nil, os.ErrNotExist
	}

	var cs []issues.Comment
	cs = append(cs, issues.Comment{
		ID:        0, // We use 0 as a special ID for the comment that is the issue description.
		User:      ghUser(i.User),
		CreatedAt: i.Created,
		// Can't use i.Updated for issue body because of false positives, since it includes the entire issue (e.g., if it was closed, that changes its Updated time).
		Body:      i.Body,
		Reactions: nil, // maintner.Corpus doesn't support GitHub issue reactions.
	})
	err = i.ForeachComment(func(c *maintner.GitHubComment) error {
		var edited *issues.Edited
		if !c.Updated.Equal(c.Created) {
			edited = &issues.Edited{
				By: users.User{Login: "Someone"}, // maintner.Corpus doesn't expose GitHub issue comment editor user.
				At: c.Updated,
			}
		}
		cs = append(cs, issues.Comment{
			ID:        uint64(c.ID),
			User:      ghUser(c.User),
			CreatedAt: c.Created,
			Edited:    edited,
			Body:      c.Body,
			Reactions: nil, // maintner.Corpus doesn't support GitHub issue reactions.
		})
		return nil
	})
	if opt != nil {
		// Pagination.
		start := opt.Start
		if start > len(cs) {
			start = len(cs)
		}
		end := opt.Start + opt.Length
		if end > len(cs) {
			end = len(cs)
		}
		cs = cs[start:end]
	}
	return cs, err
}

func (s service) ListEvents(_ context.Context, rs issues.RepoSpec, id uint64, opt *issues.ListOptions) ([]issues.Event, error) {
	repoID, err := ghRepoID(rs)
	if err != nil {
		return nil, err
	}
	s.c.RLock()
	defer s.c.RUnlock()
	repo := s.c.GitHub().Repo(repoID.Owner, repoID.Repo)
	if repo == nil {
		return nil, fmt.Errorf("repo %v not found", rs)
	}
	i := repo.Issue(int32(id))
	if i == nil || i.NotExist || i.PullRequest {
		return nil, os.ErrNotExist
	}

	var es []issues.Event
	err = i.ForeachEvent(func(e *maintner.GitHubIssueEvent) error {
		et := issues.EventType(e.Type)
		if !et.Valid() {
			return nil
		}
		ev := issues.Event{
			ID:        uint64(e.ID),
			Actor:     ghUser(e.Actor),
			CreatedAt: e.Created,
			Type:      et,
		}
		switch et {
		case issues.Renamed:
			ev.Rename = &issues.Rename{
				From: e.From,
				To:   e.To,
			}
		case issues.Labeled, issues.Unlabeled:
			ev.Label = &issues.Label{
				Name:  e.Label,
				Color: issues.RGB{R: 0xED, G: 0xED, B: 0xED}, // maintner.Corpus doesn't support GitHub issue label colors, so fall back to a default light gray.
			}
		}
		es = append(es, ev)
		return nil
	})
	if opt != nil {
		// Pagination.
		start := opt.Start
		if start > len(es) {
			start = len(es)
		}
		end := opt.Start + opt.Length
		if end > len(es) {
			end = len(es)
		}
		es = es[start:end]
	}
	return es, err
}

func (service) CreateComment(_ context.Context, rs issues.RepoSpec, id uint64, c issues.Comment) (issues.Comment, error) {
	return issues.Comment{}, fmt.Errorf("CreateComment: not implemented")
}

func (service) Create(_ context.Context, rs issues.RepoSpec, i issues.Issue) (issues.Issue, error) {
	return issues.Issue{}, fmt.Errorf("Create: not implemented")
}

func (service) Edit(_ context.Context, rs issues.RepoSpec, id uint64, ir issues.IssueRequest) (issues.Issue, []issues.Event, error) {
	return issues.Issue{}, nil, fmt.Errorf("Edit: not implemented")
}

func (service) EditComment(_ context.Context, rs issues.RepoSpec, id uint64, cr issues.CommentRequest) (issues.Comment, error) {
	return issues.Comment{}, fmt.Errorf("EditComment: not implemented")
}

// ghRepoID converts a RepoSpec into a maintner.GitHubRepoID.
func ghRepoID(repo issues.RepoSpec) (maintner.GitHubRepoID, error) {
	elems := strings.Split(repo.URI, "/")
	if len(elems) != 2 || elems[0] == "" || elems[1] == "" {
		return maintner.GitHubRepoID{}, fmt.Errorf(`RepoSpec is not of form "owner/repo": %q`, repo.URI)
	}
	return maintner.GitHubRepoID{
		Owner: elems[0],
		Repo:  elems[1],
	}, nil
}

// ghState converts a GitHub issue state into a issues.State.
func ghState(issue *maintner.GitHubIssue) issues.State {
	switch issue.Closed {
	case false:
		return issues.OpenState
	case true:
		return issues.ClosedState
	default:
		panic("unreachable")
	}
}

// ghUser converts a GitHub user into a users.User.
func ghUser(user *maintner.GitHubUser) users.User {
	return users.User{
		UserSpec: users.UserSpec{
			ID:     uint64(user.ID),
			Domain: "github.com",
		},
		Login:     user.Login,
		AvatarURL: fmt.Sprintf("https://avatars.githubusercontent.com/u/%d?v=4&s=96", user.ID),
		HTMLURL:   fmt.Sprintf("https://github.com/%v", user.Login),
	}
}
