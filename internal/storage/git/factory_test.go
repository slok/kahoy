package git_test

import (
	"fmt"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/slok/kahoy/internal/storage/fs/fsmock"
	"github.com/slok/kahoy/internal/storage/git"
	"github.com/slok/kahoy/internal/storage/git/gitmock"
)

func TestNewRepositories(t *testing.T) {
	tests := map[string]struct {
		config git.RepositoriesConfig
		mock   func(mOld, mNew *gitmock.GoGitRepoClient)
		expErr bool
	}{
		"If old repo path is absolute it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "/manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock:   func(mOld, mNew *gitmock.GoGitRepoClient) {},
			expErr: true,
		},

		"If new repo path is absolute it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "/manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock:   func(mOld, mNew *gitmock.GoGitRepoClient) {},
			expErr: true,
		},

		"If getting the old repo FS fails, it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				mOld.On("FileSystem").Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"If getting the new repo FS fails, it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				mOld.On("FileSystem").Once().Return(memfs.New(), nil)
				mNew.On("FileSystem").Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"If checkout on the old repo fails, it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = oldFs.Create("/manifests")
				_, _ = newFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				mOld.On("Checkout", mock.Anything).Once().Return(fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"If getting the old repo HEAD fails, it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = oldFs.Create("/manifests")
				_, _ = newFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				mOld.On("Checkout", mock.Anything).Once().Return(nil)

				mOld.On("Head").Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"If getting the new repo HEAD fails, it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = oldFs.Create("/manifests")
				_, _ = newFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				mOld.On("Checkout", mock.Anything).Once().Return(nil)

				mOld.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "111"), nil)
				mNew.On("Head").Once().Return(nil, fmt.Errorf("whatever"))
			},
			expErr: true,
		},

		"If old and new repos have the same HEAD it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = oldFs.Create("/manifests")
				_, _ = newFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				mOld.On("Checkout", mock.Anything).Once().Return(nil)

				mOld.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "448aa55561b85897120f611e2da67ac2d0e7a8bf"), nil)
				mNew.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "448aa55561b85897120f611e2da67ac2d0e7a8bf"), nil)
			},
			expErr: true,
		},

		"If the old relative path does not exist, should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = newFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				mOld.On("Checkout", mock.Anything).Once().Return(nil)

				mOld.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "448aa55561b85897120f611e2da67ac2d0e7a8bf"), nil)
				mNew.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "f69577b7a14bb6d112188cd75029f4aa6605b944"), nil)
			},
			expErr: true,
		},

		"If the new relative path does not exist, should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = oldFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				mOld.On("Checkout", mock.Anything).Once().Return(nil)

				mOld.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "448aa55561b85897120f611e2da67ac2d0e7a8bf"), nil)
				mNew.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "f69577b7a14bb6d112188cd75029f4aa6605b944"), nil)
			},
			expErr: true,
		},

		"If before commit is passed, it should not search for it and create correctly the repositories (Happy path without search before commit).": {
			config: git.RepositoriesConfig{
				OldRelPath:         "./manifests",
				NewRelPath:         "./manifests",
				GitBeforeCommitSHA: "cfd490e528d1c49b093cee3818cb63e2e0e8f3ef",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = oldFs.Create("/manifests")
				_, _ = newFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				expCheckout := &gogit.CheckoutOptions{Hash: plumbing.NewHash("cfd490e528d1c49b093cee3818cb63e2e0e8f3ef")}
				mOld.On("Checkout", expCheckout).Once().Return(nil)

				mOld.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "448aa55561b85897120f611e2da67ac2d0e7a8bf"), nil)
				mNew.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "f69577b7a14bb6d112188cd75029f4aa6605b944"), nil)
			},
		},

		"If not before commit is passed, it should search for it and create correctly the repositories (Happy path with search before commit).": {
			config: git.RepositoriesConfig{
				OldRelPath:       "./manifests",
				NewRelPath:       "./manifests",
				GitDefaultBranch: "main",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				oldFs, newFs := memfs.New(), memfs.New()
				_, _ = oldFs.Create("/manifests")
				_, _ = newFs.Create("/manifests")
				mOld.On("FileSystem").Once().Return(oldFs, nil)
				mNew.On("FileSystem").Once().Return(newFs, nil)

				// Check get HEAD commit.
				mNew.On("Head").Twice().Return(plumbing.NewReferenceFromStrings("", "f69577b7a14bb6d112188cd75029f4aa6605b944"), nil)
				expHash := plumbing.NewHash("f69577b7a14bb6d112188cd75029f4aa6605b944")
				commit := &object.Commit{Hash: expHash}
				mNew.On("CommitObject", expHash).Once().Return(commit, nil)

				// Check get other branch commit.
				expOtherBranch := plumbing.Revision("main")
				otherHEADHash := plumbing.NewHash("1dd111de9c9b61d7955e08078ef58a92460f7cca")
				mNew.On("ResolveRevision", expOtherBranch).Once().Return(&otherHEADHash, nil)
				otherCommit := &object.Commit{Hash: otherHEADHash}
				mNew.On("CommitObject", otherHEADHash).Once().Return(otherCommit, nil)

				// Check merge-base getting before commit from common parent.
				beforeCommitHash := plumbing.NewHash("cfd490e528d1c49b093cee3818cb63e2e0e8f3ef")
				beforeCommit := &object.Commit{Hash: beforeCommitHash}
				mNew.On("MergeBase", commit, otherCommit).Once().Return([]*object.Commit{beforeCommit}, nil)

				// Check what we found is what we checkout on old repo.
				expCheckout := &gogit.CheckoutOptions{Hash: beforeCommitHash}
				mOld.On("Checkout", expCheckout).Once().Return(nil)

				mOld.On("Head").Once().Return(plumbing.NewReferenceFromStrings("", "448aa55561b85897120f611e2da67ac2d0e7a8bf"), nil)
			},
		},

		"If search before commit in done in the default branch it should fail.": {
			config: git.RepositoriesConfig{
				OldRelPath:       "./manifests",
				NewRelPath:       "./manifests",
				GitDefaultBranch: "main",
			},
			mock: func(mOld, mNew *gitmock.GoGitRepoClient) {
				// Check get HEAD commit.
				mNew.On("Head").Once().Return(plumbing.NewReferenceFromStrings("refs/heads/main", "f69577b7a14bb6d112188cd75029f4aa6605b944"), nil)
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mOld := &gitmock.GoGitRepoClient{}
			mNew := &gitmock.GoGitRepoClient{}
			test.mock(mOld, mNew)

			// Execute.
			test.config.KubernetesDecoder = &fsmock.K8sObjectDecoder{}
			test.config.GoGitOldRepo = mOld
			test.config.GoGitNewRepo = mNew
			_, _, err := git.NewRepositories(test.config)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			mOld.AssertExpectations(t)
			mNew.AssertExpectations(t)
		})
	}
}
