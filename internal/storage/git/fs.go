package git

import (
	"fmt"
	"io/ioutil"
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
func newBillyFsManager(repoFs billy.Filesystem) billyFsManager {
	return billyFsManager{fs: repoFs}
}

func (b billyFsManager) Walk(root string, walkFn filepath.WalkFunc) error {
	files, err := b.fs.ReadDir(root)
	if err != nil {
		return err
	}

	for _, file := range files {
		path := b.fs.Join(root, file.Name())

		// If is a file, use our walkf func.
		if !file.IsDir() {
			err := walkFn(path, file, nil)
			if err != nil {
				return err
			}
			continue
		}

		// Continue walking.
		err := b.Walk(path, walkFn)
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
