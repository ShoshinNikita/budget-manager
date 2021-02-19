package embed

import (
	"io/fs"
	"os"
	"path/filepath"
)

// DirFS is like http.Dir but implements fs.ReadDirFS
type DirFS string

var _ fs.ReadDirFS = (*DirFS)(nil)

func (root DirFS) Open(name string) (fs.File, error) {
	return os.Open(join(root, name))
}

func (root DirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(join(root, name))
}

func join(root DirFS, name string) string {
	if root == "" {
		root = "."
	}

	return filepath.Join(string(root), name)
}
