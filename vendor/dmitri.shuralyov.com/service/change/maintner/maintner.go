// Package maintner implements a read-only change.Service using
// a x/build/maintner corpus.
package maintner

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"dmitri.shuralyov.com/service/change"
	"dmitri.shuralyov.com/state"
	"github.com/shurcooL/issues"
	"github.com/shurcooL/users"
	"github.com/sourcegraph/go-diff/diff"
	"golang.org/x/build/maintner"
)

// NewService creates a change.Service backed with the given corpus.
func NewService(corpus *maintner.Corpus) change.Service {
	return service{
		c: corpus,
	}
}

type service struct {
	c *maintner.Corpus
}

func (s service) List(_ context.Context, repo string, opt change.ListOptions) ([]change.Change, error) {
	s.c.RLock()
	defer s.c.RUnlock()
	project := s.c.Gerrit().Project(serverProject(repo))
	if project == nil {
		return nil, os.ErrNotExist
	}
	var is []change.Change
	err := project.ForeachCLUnsorted(func(cl *maintner.GerritCL) error {
		if cl.Private {
			return nil
		}
		changeState := changeState(cl.Status)
		switch {
		case opt.Filter == change.FilterOpen && changeState != change.OpenState:
			return nil
		case opt.Filter == change.FilterClosedMerged && !(changeState == change.ClosedState || changeState == change.MergedState):
			return nil
		}
		var labels []issues.Label
		cl.Meta.Hashtags().Foreach(func(hashtag string) {
			labels = append(labels, issues.Label{
				Name:  hashtag,
				Color: issues.RGB{R: 0xed, G: 0xed, B: 0xed}, // A default light gray.
			})
		})
		is = append(is, change.Change{
			ID:        uint64(cl.Number),
			State:     changeState,
			Title:     firstParagraph(cl.Commit.Msg),
			Labels:    labels,
			Author:    gerritUser(cl.Commit.Author),
			CreatedAt: cl.Created,
			Replies:   len(cl.Messages),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.After(is[j].CreatedAt)
	})
	return is, nil
}

func (s service) Count(_ context.Context, repo string, opt change.ListOptions) (uint64, error) {
	s.c.RLock()
	defer s.c.RUnlock()
	project := s.c.Gerrit().Project(serverProject(repo))
	if project == nil {
		return 0, os.ErrNotExist
	}
	var count uint64
	err := project.ForeachCLUnsorted(func(cl *maintner.GerritCL) error {
		if cl.Private {
			return nil
		}
		changeState := changeState(cl.Status)
		switch {
		case opt.Filter == change.FilterOpen && changeState != change.OpenState:
			return nil
		case opt.Filter == change.FilterClosedMerged && !(changeState == change.ClosedState || changeState == change.MergedState):
			return nil
		}
		count++
		return nil
	})
	return count, err
}

func (s service) Get(_ context.Context, repo string, id uint64) (change.Change, error) {
	s.c.RLock()
	defer s.c.RUnlock()
	project := s.c.Gerrit().Project(serverProject(repo))
	if project == nil {
		return change.Change{}, os.ErrNotExist
	}
	cl := project.CL(int32(id))
	if cl == nil || cl.Private {
		return change.Change{}, os.ErrNotExist
	}
	var labels []issues.Label
	cl.Meta.Hashtags().Foreach(func(hashtag string) {
		labels = append(labels, issues.Label{
			Name:  hashtag,
			Color: issues.RGB{R: 0xed, G: 0xed, B: 0xed}, // A default light gray.
		})
	})
	return change.Change{
		ID:           uint64(cl.Number),
		State:        changeState(cl.Status),
		Title:        firstParagraph(cl.Commit.Msg),
		Labels:       labels,
		Author:       gerritUser(cl.Commit.Author),
		CreatedAt:    cl.Created,
		Replies:      len(cl.Messages),
		Commits:      int(cl.Version),
		ChangedFiles: 0, // TODO.
	}, nil
}

func (s service) ListTimeline(_ context.Context, repo string, id uint64, opt *change.ListTimelineOptions) ([]interface{}, error) {
	s.c.RLock()
	defer s.c.RUnlock()
	project := s.c.Gerrit().Project(serverProject(repo))
	if project == nil {
		return nil, os.ErrNotExist
	}
	cl := project.CL(int32(id))
	if cl == nil || cl.Private {
		return nil, os.ErrNotExist
	}
	var timeline []interface{}
	timeline = append(timeline, change.Comment{ // CL description.
		ID:        "0",
		User:      gerritUser(cl.Commit.Author),
		CreatedAt: cl.Created,
		Body:      "", // THINK: Include commit message or no?
	})
	for _, m := range cl.Messages {
		labels, body, ok := parseMessage(m.Message)
		if !ok {
			timeline = append(timeline, change.Comment{
				User:      gerritUser(m.Author),
				CreatedAt: m.Date,
				Body:      m.Message,
			})
			continue
		}
		timeline = append(timeline, change.Review{
			User:      gerritUser(m.Author),
			CreatedAt: m.Date,
			State:     reviewState(labels),
			Body:      body,
		})
	}
	return timeline, nil
}

func parseMessage(m string) (labels string, body string, ok bool) {
	// "Patch Set ".
	if !strings.HasPrefix(m, "Patch Set ") {
		return "", "", false
	}
	m = m[len("Patch Set "):]

	// "123".
	i := strings.IndexFunc(m, func(c rune) bool { return !unicode.IsNumber(c) })
	if i == -1 {
		return "", "", false
	}
	m = m[i:]

	// ":".
	if len(m) < 1 || m[0] != ':' {
		return "", "", false
	}
	m = m[1:]

	switch i = strings.IndexByte(m, '\n'); i {
	case -1:
		labels = m
	default:
		labels = m[:i]
		body = m[i+1:]
	}

	if labels != "" {
		// " ".
		if len(labels) < 1 || labels[0] != ' ' {
			return "", "", false
		}
		labels = labels[1:]
	}

	if body != "" {
		// "\n".
		if len(body) < 1 || body[0] != '\n' {
			return "", "", false
		}
		body = body[1:]
	}

	return labels, body, true
}

func reviewState(labels string) state.Review {
	for _, label := range strings.Split(labels, " ") {
		switch label {
		case "Code-Review+2":
			return state.ReviewPlus2
		case "Code-Review+1":
			return state.ReviewPlus1
		case "Code-Review-1":
			return state.ReviewMinus1
		case "Code-Review-2":
			return state.ReviewMinus2
		}
	}
	return state.ReviewNoScore
}

func (s service) ListCommits(_ context.Context, repo string, id uint64) ([]change.Commit, error) {
	s.c.RLock()
	defer s.c.RUnlock()
	project := s.c.Gerrit().Project(serverProject(repo))
	if project == nil {
		return nil, os.ErrNotExist
	}
	cl := project.CL(int32(id))
	if cl == nil || cl.Private {
		return nil, os.ErrNotExist
	}
	commits := make([]change.Commit, int(cl.Version))
	for n := int32(1); n <= cl.Version; n++ {
		c := cl.CommitAtVersion(n)
		commits[n-1] = change.Commit{
			SHA:        c.Hash.String(),
			Message:    fmt.Sprintf("Patch Set %d", n),
			Author:     gerritUser(c.Author),
			AuthorTime: c.AuthorTime,
		}
	}
	return commits, nil
}

func (s service) GetDiff(_ context.Context, repo string, id uint64, opt *change.GetDiffOptions) ([]byte, error) {
	s.c.RLock()
	defer s.c.RUnlock()
	project := s.c.Gerrit().Project(serverProject(repo))
	if project == nil {
		return nil, os.ErrNotExist
	}
	cl := project.CL(int32(id))
	if cl == nil || cl.Private {
		return nil, os.ErrNotExist
	}
	var c *maintner.GitCommit
	switch opt {
	case nil:
		c = cl.Commit
	default:
		c = project.GitCommit(opt.Commit)
	}
	var fds []*diff.FileDiff
	for _, f := range c.Files {
		fds = append(fds, &diff.FileDiff{
			OrigName: f.File,
			NewName:  f.File,
			Hunks:    []*diff.Hunk{}, // Hunk data isn't present in maintner.Corpus.
		})
	}
	return diff.PrintMultiFileDiff(fds)
}

func (service) EditComment(_ context.Context, repo string, id uint64, cr change.CommentRequest) (change.Comment, error) {
	return change.Comment{}, fmt.Errorf("EditComment: not implemented")
}

func changeState(status string) change.State {
	switch status {
	case "new":
		return change.OpenState
	case "abandoned":
		return change.ClosedState
	case "merged":
		return change.MergedState
	case "draft":
		panic("not sure how to deal with draft status")
	default:
		panic(fmt.Errorf("unrecognized status %q", status))
	}
}

func gerritUser(user *maintner.GitPerson) users.User {
	return users.User{
		UserSpec: users.UserSpec{
			ID:     0,  // TODO.
			Domain: "", // TODO.
		},
		Login: user.Name(), //user.Username, // TODO.
		Name:  user.Name(),
		Email: user.Email(),
		//AvatarURL: fmt.Sprintf("https://%s/accounts/%d/avatar?s=96", s.domain, user.AccountID),
	}
}

func serverProject(repo string) (server, project string) {
	i := strings.IndexByte(repo, '/')
	if i == -1 {
		return "", ""
	}
	return repo[:i], repo[i+1:]
}

// firstParagraph returns the first paragraph of text s.
func firstParagraph(s string) string {
	i := strings.Index(s, "\n\n")
	if i == -1 {
		return s
	}
	return s[:i]
}
