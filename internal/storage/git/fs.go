package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"

	"github.com/slok/kahoy/internal/storage/fs"
)

type billyFsManager struct {
	fs billy.Filesystem
}

var _ fs.FileSystemManager = billyFsManager{}

// newbillyFsManager returns a new compatible fs.FileSystemManager based on a
// go-git repo fs system (go-billy).
// TODO(slok): Implement SkipDir error.
func newBillyFsManager(repoFs billy.Filesystem) billyFsManager {
	return billyFsManager{fs: repoFs}
}

func (b billyFsManager) Walk(root string, walkFn filepath.WalkFunc) error {
	info, err := b.fs.Lstat(root)
	if err != nil {
		return err
	}

	return b.walk(root, info, walkFn)
}

func (b billyFsManager) walk(path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	files, err := b.fs.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		path := b.fs.Join(path, file.Name())
		err := b.walk(path, file, walkFn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b billyFsManager) ReadFile(path string) ([]byte, error) {
	f, err := b.fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	return data, nil
}

func (b billyFsManager) Abs(path string) (string, error) {
	return filepath.Abs(path)
}
