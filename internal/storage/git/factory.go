package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage/fs"
)

const (
	// Only allow loading Git repos from current directory. This will remove
	// lots of corner cases related with the repo loading and make the app more
	// reliable.
	gitRepoPath         = "."
	gitClonesRemoteName = "origin"
)

// RepositoriesConfig is the configuration for NewRepositories
type RepositoriesConfig struct {
	ExcludeRegex      []string
	IncludeRegex      []string
	OldRelPath        string
	NewRelPath        string
	KubernetesDecoder fs.K8sObjectDecoder
	AppConfig         *model.AppConfig
	Logger            log.Logger

	// GitBeforeCommitSHA Used to set the Git old repo state.
	// If empty it will use merge-base to get the common ancestor
	// of HEAD.
	// If we are on the default branch this will be neccesary.
	GitBeforeCommitSHA string
	// GitDefaultBranch is the base branch that will be used when no GitBeforeCommitSHA setting is passed.
	// This branch will be the one used against HEAD to get the common parent commit  by using merge-base.
	// Only local branches are support, other refs are not supported (remote branch, tag, hash...)
	GitDefaultBranch string
	// GitDiffIncludeFilter will filter everything except the files modified (edit, create, delete)
	// between the two Git repository states, this is, before-commit and HEAD.
	// This could be translated as: `git diff --name-only before-commit HEAD`.
	GitDiffIncludeFilter bool

	// Don't use these, used for testing purposes. Normally `nil` because are created correctly by the factory.
	// Go-git loaded repositories, if `nil` it will clone them into memory from the current disk path (`.`).
	GoGitOldRepo GoGitRepoClient
	GoGitNewRepo GoGitRepoClient
}

func (c *RepositoriesConfig) defaults() error {
	if c.GitDefaultBranch == "" {
		c.GitDefaultBranch = "origin/master"
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "git.Repository"})

	if filepath.IsAbs(c.OldRelPath) {
		return fmt.Errorf("old path %q is absolute, must be relative to the git repository", c.OldRelPath)
	}

	if filepath.IsAbs(c.NewRelPath) {
		return fmt.Errorf("new path %q is absolute, must be relative to the git repository", c.NewRelPath)
	}

	return nil
}

// NewRepositories is a factory that knows how to return two fs repositories based on Git.
//
// 1. Loads/clones a git repository (with all its files/worktree) from the fs into memory.
// 	  twice (one for old repo and one for new repo)
// 2. If no before commit passed, it will search the common ancestor of the new repository HEAD
//    against the default branch.
// 3. Set the old repo worktree (and its FS) in the `before commit` state.
// 4. Create 2 `fs.Repository` using the internal worktree FS of the git repositories.
//
// We end with 2 fs.Repositories that are based on memory and each one has different data based
// on the commit history that we have been preparing before creating them.
//
// Note: Cloned repos (memory) will have original repo (fs) local branches as remotes because of
//       the clone, so to use local branch refs, we need to use remote notation, and remote
//       branches are not supported.
func NewRepositories(config RepositoriesConfig) (old, new *fs.Repository, err error) {
	err = config.defaults()
	if err != nil {
		return nil, nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Load git repo in memory if required.
	oldGitRepo, newGitRepo, err := loadGitRepositories(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not load git repositores: %w", err)
	}

	// Search git before commit if required.
	var gitBeforeHash plumbing.Hash
	if config.GitBeforeCommitSHA == "" {
		config.Logger.Debugf("searching git before commit using common parent of HEAD and %s", config.GitDefaultBranch)
		hash, err := getBeforeCommit(newGitRepo, config.GitDefaultBranch)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get git before commit: %w", err)
		}
		gitBeforeHash = *hash
	} else {
		gitBeforeHash = plumbing.NewHash(config.GitBeforeCommitSHA)
	}

	// Get both file systems.
	oldRepoFs, err := oldGitRepo.FileSystem()
	if err != nil {
		return nil, nil, err
	}
	newRepoFs, err := newGitRepo.FileSystem()
	if err != nil {
		return nil, nil, err
	}

	// Set old git repo in the `before commit` state.
	err = oldGitRepo.Checkout(&git.CheckoutOptions{Hash: gitBeforeHash})
	if err != nil {
		return nil, nil, err
	}

	// Get both repository HEAD refs.
	oldRef, err := oldGitRepo.Head()
	if err != nil {
		return nil, nil, err
	}
	newRef, err := newGitRepo.Head()
	if err != nil {
		return nil, nil, err
	}

	if config.GitDiffIncludeFilter {
		config.Logger.Infof("using git diff include filter")
		includes, err := includeFiltersFromGitDiff(newGitRepo, oldRef.Hash(), newRef.Hash())
		if err != nil {
			return nil, nil, fmt.Errorf("could not get git diff: %w", err)
		}
		config.IncludeRegex = append(config.IncludeRegex, includes...)
	}

	// Validations to help the user in case of misusage.
	// Check old and new repos are not in the same state.
	if oldRef.Hash() == newRef.Hash() {
		return nil, nil, fmt.Errorf("old and new repo HEAD ref can't be the same (%s) use 'before commit'", oldRef.Hash())
	}

	// Check our paths on each repo exist.
	_, err = oldRepoFs.Stat(config.OldRelPath)
	if err != nil {
		return nil, nil, fmt.Errorf("old git repo path %q: %w", config.OldRelPath, err)
	}
	_, err = newRepoFs.Stat(config.NewRelPath)
	if err != nil {
		return nil, nil, fmt.Errorf("new git repo path %q: %w", config.NewRelPath, err)
	}

	config.Logger.Debugf("old repository worktree in %q commit", oldRef.Hash())
	config.Logger.Debugf("new repository worktree in %q commit", newRef.Hash())

	// At this point our memory git repositories (and its internal file system) are
	// in the correct states (new == current, and old == before commit).
	// We create regular fs.Repositories except the data comes from memory instead of disk.
	// We use the git memory worktree file system with a custom FileSystemManager that
	// understands go-git internal File system (go-billy) implementation.
	oldRepo, err := fs.NewRepository(fs.RepositoryConfig{
		ExcludeRegex:      config.ExcludeRegex,
		IncludeRegex:      config.IncludeRegex,
		Path:              config.OldRelPath,
		KubernetesDecoder: config.KubernetesDecoder,
		AppConfig:         config.AppConfig,
		Logger: config.Logger.WithValues(log.Kv{
			"repo-state": "old",
			"git-rev":    oldRef.Hash().String(),
		}),
		FSManager: newBillyFsManager(oldRepoFs),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create old Git fs %q repository storage: %w", config.OldRelPath, err)
	}

	newRepo, err := fs.NewRepository(fs.RepositoryConfig{
		ExcludeRegex:      config.ExcludeRegex,
		IncludeRegex:      config.IncludeRegex,
		Path:              config.NewRelPath,
		KubernetesDecoder: config.KubernetesDecoder,
		AppConfig:         config.AppConfig,
		Logger: config.Logger.WithValues(log.Kv{
			"repo-state": "new",
			"git-rev":    newRef.Hash().String(),
		}),
		FSManager: newBillyFsManager(newRepoFs),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create new Git fs %q repository storage: %w", config.OldRelPath, err)
	}

	return oldRepo, newRepo, nil
}

// loadGitRepositories will clone the repositories into memory.
// NOTE: When cloning our local repository from disk into memory, original (fs)
//       repository remote refs will be lost. The remotes on our new memory clones
//       will be the local refs of our disk repository.
func loadGitRepositories(config RepositoriesConfig) (old, new GoGitRepoClient, err error) {
	oldRepo := config.GoGitOldRepo
	newRepo := config.GoGitNewRepo

	if oldRepo == nil {
		config.Logger.Debugf("loading old git repository into memory")
		goGitRepo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{URL: gitRepoPath, RemoteName: gitClonesRemoteName})
		if err != nil {
			return nil, nil, fmt.Errorf("could not clone Git repository into memory: %w", err)
		}
		oldRepo = goGitRepoClient{repo: *goGitRepo}
	}

	if newRepo == nil {
		config.Logger.Debugf("loading new git repository into memory")
		goGitRepo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{URL: gitRepoPath, RemoteName: gitClonesRemoteName})
		if err != nil {
			return nil, nil, fmt.Errorf("could not clone Git repository into memory: %w", err)
		}
		newRepo = goGitRepoClient{repo: *goGitRepo}
	}

	return oldRepo, newRepo, nil
}

// getBeforeCommit searches common parent using merge-base on HEAD and a
// default branch (e.g master).
//
//              o---o---o (HEAD)
//             /
// ---o---o---1 (default branch)
//
// We would get 1 as the before commit.
func getBeforeCommit(newRepo GoGitRepoClient, branch string) (*plumbing.Hash, error) {
	// Get HEAD commit.
	ref, err := newRepo.Head()
	if err != nil {
		return nil, err
	}

	if ref.Name().IsBranch() && ref.Name().Short() == branch {
		return nil, fmt.Errorf("can't get common parent from same branch (%q), use before commit", branch)
	}

	currentCommit, err := newRepo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	// Get branch commit.
	// We use the branch name as a remote beacause our cloned repositories into memory
	// have the original repo (fs) local refs as remotes.
	branchRef := plumbing.NewRemoteReferenceName(gitClonesRemoteName, branch)
	revHash, err := newRepo.ResolveRevision(plumbing.Revision(branchRef))
	if err != nil {
		return nil, fmt.Errorf("could not translate branch ref %q to commit hash: %w", branch, err)
	}
	branchCommit, err := newRepo.CommitObject(*revHash)
	if err != nil {
		return nil, err
	}

	// Get the common parent commit using `merge-base`.
	commit, err := newRepo.MergeBase(currentCommit, branchCommit)
	if err != nil {
		return nil, err
	}

	if len(commit) < 1 {
		return nil, fmt.Errorf("could not get common parent")
	}
	cHash := commit[0].Hash

	return &cHash, nil
}

func includeFiltersFromGitDiff(newRepo GoGitRepoClient, old, new plumbing.Hash) ([]string, error) {
	// Get commits from refs.
	oldCommit, err := newRepo.CommitObject(old)
	if err != nil {
		return nil, err
	}
	newCommit, err := newRepo.CommitObject(new)
	if err != nil {
		return nil, err
	}
	// Get patch.
	patch, err := newRepo.Patch(oldCommit, newCommit)
	if err != nil {
		return nil, err
	}

	// Get all file paths from the changes.
	filePaths := map[string]struct{}{}
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		if from != nil {
			filePaths[from.Path()] = struct{}{}
		}

		if to != nil {
			filePaths[to.Path()] = struct{}{}
		}
	}

	// Although is likely that raw paths will filter correctly, we will
	// convert them to correct regexes to increase filtering reliability.
	res := make([]string, 0, len(filePaths))
	for path := range filePaths {
		if path == "" {
			continue
		}
		path = strings.ReplaceAll(path, "/", `\/`)
		res = append(res, fmt.Sprintf(`.*\/%s$`, path))
	}

	return res, nil
}
