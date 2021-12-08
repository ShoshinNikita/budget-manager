package backup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

func TestBackupManager(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dir := t.TempDir()

	manager := newBackupManager(dir, db.Postgres)
	manager.maxBackupFiles = 3
	manager.timeNow = func() time.Time {
		return time.Date(2021, 12, 8, 21, 40, 0, 0, time.UTC)
	}

	generateFiles(t, dir, 5)

	nestedDir := filepath.Join(dir, "nested")
	err := os.Mkdir(nestedDir, 0o700)
	require.NoError(err)

	generateFiles(t, nestedDir, 5)

	compareFiles(t, dir,
		"1.test", "2.test", "3.test", "4.test", "5.test",
		"nested/1.test", "nested/2.test", "nested/3.test", "nested/4.test", "nested/5.test",
	)

	f, err := manager.NewBackupFile()
	require.NoError(err)
	require.NoError(f.Close())

	compareFiles(t, dir,
		// "1.test", "2.test" and "3.test" must be removed
		"4.test", "5.test",
		"2021-12-08_21-40-00.postgres.sql",
		"nested/1.test", "nested/2.test", "nested/3.test", "nested/4.test", "nested/5.test",
	)
}

func generateFiles(t *testing.T, dir string, n int) {
	t.Helper()

	for i := 1; i <= n; i++ {
		filename := filepath.Join(dir, fmt.Sprintf("%d.test", i))
		f, err := os.Create(filename)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		if i+1 != n {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func compareFiles(t *testing.T, dir string, expectedFilenames ...string) {
	for i := range expectedFilenames {
		expectedFilenames[i] = filepath.Join(dir, expectedFilenames[i])
	}

	var gotFiles []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			gotFiles = append(gotFiles, path)
		}
		return nil
	})
	require.NoError(t, err)

	require.ElementsMatch(t, expectedFilenames, gotFiles)
}
