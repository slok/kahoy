package git

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GoGitRepoClient helper interface to use go-git and make testable by mocking
// common operations on a Git repo.
type GoGitRepoClient interface {
	Head() (*plumbing.Reference, error)
	Checkout(opts *git.CheckoutOptions) error
	FileSystem() (billy.Filesystem, error)
	MergeBase(current, other *object.Commit) ([]*object.Commit, error)
	CommitObject(h plumbing.Hash) (*object.Commit, error)
	ResolveRevision(rev plumbing.Revision) (*plumbing.Hash, error)
}

//go:generate mockery --case underscore --output gitmock --outpkg gitmock --name GoGitRepoClient

type goGitRepoClient struct {
	repo git.Repository
}

// Interface assertion.
var _ GoGitRepoClient = goGitRepoClient{}

func (g goGitRepoClient) Head() (*plumbing.Reference, error) {
	return g.repo.Head()
}

func (g goGitRepoClient) Checkout(opts *git.CheckoutOptions) error {
	wt, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	return wt.Checkout(opts)
}

func (g goGitRepoClient) FileSystem() (billy.Filesystem, error) {
	wt, err := g.repo.Worktree()
	if err != nil {
		return nil, err
	}
	return wt.Filesystem, nil
}

func (g goGitRepoClient) MergeBase(current, other *object.Commit) ([]*object.Commit, error) {
	return current.MergeBase(other)
}

func (g goGitRepoClient) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	return g.repo.CommitObject(h)
}

func (g goGitRepoClient) ResolveRevision(rev plumbing.Revision) (*plumbing.Hash, error) {
	return g.repo.ResolveRevision(rev)
}
