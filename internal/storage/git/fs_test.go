package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/stretchr/testify/assert"
)

func TestBillyFsManagerWalk(t *testing.T) {
	tests := map[string]struct {
		root     string
		fs       func() billy.Filesystem
		walkFunc func(gotPaths map[string]string) filepath.WalkFunc
		expPaths map[string]string
		expErr   bool
	}{
		"Missing root path walk should fail.": {
			root: "/",
			fs: func() billy.Filesystem {
				return memfs.New()
			},
			walkFunc: func(gotPaths map[string]string) filepath.WalkFunc {
				return func(path string, info os.FileInfo, err error) error {
					gotPaths[path] = info.Name()
					return nil
				}
			},
			expErr: true,
		},

		"Root files should be handled correctly.": {
			root: "/",
			fs: func() billy.Filesystem {
				fs := memfs.New()
				_, _ = fs.Create("/a")
				_, _ = fs.Create("/b")
				_, _ = fs.Create("/c")
				return fs
			},
			walkFunc: func(gotPaths map[string]string) filepath.WalkFunc {
				return func(path string, info os.FileInfo, err error) error {
					gotPaths[path] = info.Name()
					return nil
				}
			},
			expPaths: map[string]string{
				"/a": "a",
				"/b": "b",
				"/c": "c",
			},
		},

		"Files at different levels should be handled correctly.": {
			root: "/",
			fs: func() billy.Filesystem {
				fs := memfs.New()
				_, _ = fs.Create("/a/a")
				_, _ = fs.Create("/a")
				_, _ = fs.Create("/b/b")
				_, _ = fs.Create("/c/c/c")
				_, _ = fs.Create("/c/c/d")
				_, _ = fs.Create("/c/c/d/e")
				_, _ = fs.Create("/c/c/e/e")
				return fs
			},
			walkFunc: func(gotPaths map[string]string) filepath.WalkFunc {
				return func(path string, info os.FileInfo, err error) error {
					gotPaths[path] = info.Name()
					return nil
				}
			},
			expPaths: map[string]string{
				"/a/a":     "a",
				"/b/b":     "b",
				"/c/c/c":   "c",
				"/c/c/d":   "d",
				"/c/c/e/e": "e",
			},
		},

		"Walk should stop on the first error.": {
			root: "/",
			fs: func() billy.Filesystem {
				fs := memfs.New()
				_, _ = fs.Create("/a/a")
				_, _ = fs.Create("/a")
				_, _ = fs.Create("/b/b")
				_, _ = fs.Create("/c/c/c")
				_, _ = fs.Create("/c/c/d")
				_, _ = fs.Create("/c/c/d/e")
				_, _ = fs.Create("/c/c/e/e")
				return fs
			},
			walkFunc: func(gotPaths map[string]string) filepath.WalkFunc {
				return func(path string, info os.FileInfo, err error) error {
					if path == "/c/c/c" {
						return fmt.Errorf("whatever")
					}

					gotPaths[path] = info.Name()
					return nil
				}
			},
			expPaths: map[string]string{},
			expErr:   true,
		},

		"If the root file is a file it should handle that file.": {
			root: "/a/a",
			fs: func() billy.Filesystem {
				fs := memfs.New()
				_, _ = fs.Create("/a/a")
				return fs
			},
			walkFunc: func(gotPaths map[string]string) filepath.WalkFunc {
				return func(path string, info os.FileInfo, err error) error {
					gotPaths[path] = info.Name()
					return nil
				}
			},
			expPaths: map[string]string{
				"/a/a": "a",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fs := test.fs()
			gotPaths := map[string]string{}
			f := test.walkFunc(gotPaths)

			fm := newBillyFsManager(fs)
			err := fm.Walk(test.root, f)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(test.expPaths, gotPaths)
			}
		})
	}
}

func TestBillyFsManagerReadFile(t *testing.T) {
	tests := map[string]struct {
		path    string
		fs      func() billy.Filesystem
		expData string
		expErr  bool
	}{
		"Missing file should fail.": {
			path: "/missing/path",
			fs: func() billy.Filesystem {
				return memfs.New()
			},
			expErr: true,
		},

		"A file should be read correctly.": {
			path: "/readable/file",
			fs: func() billy.Filesystem {
				fs := memfs.New()
				f, _ := fs.Create("/readable/file")
				_, _ = f.Write([]byte("whatever"))
				f.Close()
				return fs
			},
			expData: "whatever",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fs := test.fs()
			fm := newBillyFsManager(fs)
			gotData, err := fm.ReadFile(test.path)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(test.expData, string(gotData))
			}
		})
	}
}
