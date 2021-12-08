package backup

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

type backupManager struct {
	dir            string
	maxBackupFiles int
	dbType         db.Type
	timeNow        func() time.Time
}

func newBackupManager(dir string, dbType db.Type) *backupManager {
	const maxBackupFiles = 10

	return &backupManager{
		dir:            filepath.Clean(dir),
		maxBackupFiles: maxBackupFiles,
		dbType:         dbType,
		timeNow:        time.Now,
	}
}

func (m backupManager) NewBackupFile() (io.WriteCloser, error) {
	type file struct {
		path    string
		modTime time.Time
	}
	var files []file

	err := filepath.Walk(m.dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		path = filepath.Clean(path)
		if info.IsDir() {
			if m.dir != path {
				// Ignore files in nested dirs
				return fs.SkipDir
			}
			return nil
		}

		files = append(files, file{
			path:    path,
			modTime: info.ModTime(),
		})

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get previous backup files")
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	for len(files) >= m.maxBackupFiles {
		if err := os.Remove(files[0].path); err != nil {
			return nil, errors.Wrapf(err, "couldn't remove old backup file %q", files[0].path)
		}
		files = files[1:]
	}

	newBackupFilepath := filepath.Join(m.dir, m.generateBackupFilename())
	f, err := os.Create(newBackupFilepath)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't create new backup file %q", newBackupFilepath)
	}
	return f, nil
}

func (m backupManager) generateBackupFilename() string {
	return m.timeNow().Format("2006-01-02_15-04-05") + "." + m.dbType.String() + ".sql"
}
